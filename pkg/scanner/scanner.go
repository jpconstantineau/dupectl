package scanner

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"time"

	"github.com/jpconstantineau/dupectl/internal/worker"
	"github.com/jpconstantineau/dupectl/pkg/checkpoint"
	"github.com/jpconstantineau/dupectl/pkg/datastore"
	"github.com/jpconstantineau/dupectl/pkg/hash"
	"github.com/jpconstantineau/dupectl/pkg/logger"
)

// Scanner orchestrates the scanning process
type Scanner struct {
	db            *sql.DB
	rootFolderID  int64
	rootPath      string
	scanMode      string // "all", "folders", "files"
	hasher        hash.Hasher
	workerCount   int
	progress      *ProgressIndicator
	checkpointMgr *checkpoint.Manager
	traverseLinks bool
}

// Config holds scanner configuration
type Config struct {
	RootFolderID     int64
	RootPath         string
	ScanMode         string
	HashAlgorithm    string
	WorkerCount      int
	ShowProgress     bool
	ProgressInterval time.Duration
	TraverseLinks    bool
}

// NewScanner creates a new scanner
func NewScanner(db *sql.DB, config *Config) (*Scanner, error) {
	// Create hasher
	hasher, err := hash.NewHasher(config.HashAlgorithm)
	if err != nil {
		return nil, err
	}

	// Default worker count to CPU cores
	workerCount := config.WorkerCount
	if workerCount == 0 {
		workerCount = runtime.NumCPU()
	}

	// Debug: Log configuration
	logger.Info("Scanner config: hash=%s, workers=%d, progress_interval=%v",
		config.HashAlgorithm, workerCount, config.ProgressInterval)

	return &Scanner{
		db:            db,
		rootFolderID:  config.RootFolderID,
		rootPath:      config.RootPath,
		scanMode:      config.ScanMode,
		hasher:        hasher,
		workerCount:   workerCount,
		progress:      NewProgressIndicator(config.ShowProgress, config.ProgressInterval),
		checkpointMgr: checkpoint.NewManager(db, config.RootFolderID, config.ScanMode),
		traverseLinks: config.TraverseLinks,
	}, nil
}

// Scan performs the scan operation
func (s *Scanner) Scan(ctx context.Context, restart bool) error {
	logger.Info("Starting scan: mode=%s, root=%s", s.scanMode, s.rootPath)

	// Handle restart
	var resuming bool
	if restart {
		if err := s.checkpointMgr.Clear(); err != nil {
			return fmt.Errorf("failed to clear checkpoint: %w", err)
		}
	} else {
		// Check for existing checkpoint
		state, err := s.checkpointMgr.Resume()
		if err != nil {
			return fmt.Errorf("failed to check checkpoint: %w", err)
		}
		if state != nil {
			logger.Info("Resuming from checkpoint")
			resuming = true
			// TODO: Implement resume logic to skip already processed files
		}
	}

	// Start new checkpoint only if not resuming
	if !resuming {
		if err := s.checkpointMgr.Start(); err != nil {
			return fmt.Errorf("failed to start checkpoint: %w", err)
		}
	}

	// Start progress indicator
	s.progress.Start()
	defer s.progress.Stop()

	// Execute scan based on mode
	var err error
	switch s.scanMode {
	case "all":
		err = s.scanAll(ctx)
	case "folders":
		err = s.scanFolders(ctx)
	case "files":
		err = s.scanFiles(ctx)
	default:
		return fmt.Errorf("invalid scan mode: %s", s.scanMode)
	}

	if err != nil {
		return err
	}

	// Phase 3: Update root folder statistics
	logger.Info("Phase 3: Updating statistics...")
	if err := s.updateRootStatistics(); err != nil {
		logger.Error("Failed to update statistics: %v", err)
		// Don't fail the scan if statistics update fails
	}

	// Complete checkpoint
	if err := s.checkpointMgr.Complete(); err != nil {
		return fmt.Errorf("failed to complete checkpoint: %w", err)
	}

	logger.Info("Scan completed successfully")
	return nil
}

// scanAll performs folder traversal + file hashing
func (s *Scanner) scanAll(ctx context.Context) error {
	// Phase 1: Traverse folders and register files
	logger.Info("Phase 1: Traversing folders...")
	traverser := NewTraverser(s.db, s.rootFolderID, s.rootPath, s.traverseLinks)

	// Collect files to hash
	var filesToHash []struct {
		ID   int64
		Path string
	}

	err := traverser.Traverse(ctx, func(folderInfo *FolderInfo) error {
		s.progress.IncrementFolders()

		// Register folder and files
		folderID, err := RegisterFolder(s.db, s.rootFolderID, folderInfo.Path,
			folderInfo.ParentPath, folderInfo.ErrorStatus)
		if err != nil {
			logger.Error("Failed to register folder %s: %v", folderInfo.Path, err)
			return err
		}

		// Register files
		for _, fileInfo := range folderInfo.Files {
			fileID, err := RegisterFile(s.db, folderID, s.rootFolderID, &fileInfo, nil)
			if err != nil {
				logger.Warn("Failed to register file %s: %v", fileInfo.Path, err)
				continue
			}

			filesToHash = append(filesToHash, struct {
				ID   int64
				Path string
			}{fileID, fileInfo.Path})

			s.progress.IncrementFiles()
		}

		// Save checkpoint after each folder
		s.checkpointMgr.Save(&folderInfo.Path, nil)
		return nil
	})

	if err != nil {
		return fmt.Errorf("folder traversal failed: %w", err)
	}

	// Phase 2: Hash files in parallel
	logger.Info("Phase 2: Hashing %d files...", len(filesToHash))
	hashPool, err := worker.NewWorkerPool(ctx, s.workerCount)
	if err != nil {
		return fmt.Errorf("failed to create worker pool: %w", err)
	}

	for _, file := range filesToHash {
		workItem := NewFileHashingWorkItem(s.db, file.ID, file.Path, s.hasher, s.progress)
		if err := hashPool.Submit(workItem); err != nil {
			logger.Error("Failed to submit file %s for hashing: %v", file.Path, err)
		}
	}

	// Wait for hashing to complete
	hashErrors := hashPool.Wait()
	if len(hashErrors) > 0 {
		logger.Warn("Hashing completed with %d errors", len(hashErrors))
		for _, err := range hashErrors {
			logger.Error("Hash error: %v", err)
		}
	}

	return nil
}

// scanFolders performs folder traversal only (no hashing)
func (s *Scanner) scanFolders(ctx context.Context) error {
	logger.Info("Scanning folders only...")
	traverser := NewTraverser(s.db, s.rootFolderID, s.rootPath, s.traverseLinks)

	return traverser.Traverse(ctx, func(folderInfo *FolderInfo) error {
		s.progress.IncrementFolders()

		// Register folder only
		_, err := RegisterFolder(s.db, s.rootFolderID, folderInfo.Path,
			folderInfo.ParentPath, folderInfo.ErrorStatus)
		if err != nil {
			logger.Error("Failed to register folder %s: %v", folderInfo.Path, err)
			return err
		}

		// Save checkpoint after each folder
		s.checkpointMgr.Save(&folderInfo.Path, nil)
		return nil
	})
}

// scanFiles performs file hashing only (assumes folders exist)
func (s *Scanner) scanFiles(ctx context.Context) error {
	logger.Info("Scanning files only...")

	// Get all folders for this root
	folders, err := datastore.GetFoldersByRootID(s.db, s.rootFolderID)
	if err != nil {
		return fmt.Errorf("failed to get folders: %w", err)
	}

	if len(folders) == 0 {
		return fmt.Errorf("no folders found - run 'scan folders' first")
	}

	logger.Info("Found %d folders to scan", len(folders))

	// Create worker pool for hashing
	hashPool, err := worker.NewWorkerPool(ctx, s.workerCount)
	if err != nil {
		return fmt.Errorf("failed to create worker pool: %w", err)
	}

	// For each folder, register files and hash them
	for _, folder := range folders {
		// Read directory
		traverser := NewTraverser(s.db, s.rootFolderID, folder.Path, s.traverseLinks)

		err := traverser.Traverse(ctx, func(folderInfo *FolderInfo) error {
			// Register and hash files
			for _, fileInfo := range folderInfo.Files {
				fileID, err := RegisterFile(s.db, folder.ID, s.rootFolderID, &fileInfo, nil)
				if err != nil {
					logger.Warn("Failed to register file %s: %v", fileInfo.Path, err)
					continue
				}

				workItem := NewFileHashingWorkItem(s.db, fileID, fileInfo.Path, s.hasher, s.progress)
				if err := hashPool.Submit(workItem); err != nil {
					logger.Error("Failed to submit file %s for hashing: %v", fileInfo.Path, err)
				}

				s.progress.IncrementFiles()
			}
			return nil
		})

		if err != nil {
			logger.Error("Failed to scan folder %s: %v", folder.Path, err)
		}
	}

	// Wait for hashing to complete
	hashErrors := hashPool.Wait()
	if len(hashErrors) > 0 {
		logger.Warn("File scanning completed with %d errors", len(hashErrors))
	}

	return nil
}

// GetSummary returns scan summary statistics
func (s *Scanner) GetSummary() (folders, files int64, duration string) {
	f, fi, d := s.progress.Summary()
	return f, fi, formatDuration(d)
}

// updateRootStatistics calculates and updates root folder statistics
func (s *Scanner) updateRootStatistics() error {
	// Calculate statistics from database
	var folderCount, fileCount, totalSize int64

	err := s.db.QueryRow("SELECT COUNT(*) FROM folders WHERE root_folder_id = ? AND removed = 0", s.rootFolderID).Scan(&folderCount)
	if err != nil {
		return fmt.Errorf("failed to count folders: %w", err)
	}

	err = s.db.QueryRow("SELECT COUNT(*) FROM files WHERE root_folder_id = ? AND removed = 0", s.rootFolderID).Scan(&fileCount)
	if err != nil {
		return fmt.Errorf("failed to count files: %w", err)
	}

	err = s.db.QueryRow("SELECT COALESCE(SUM(size), 0) FROM files WHERE root_folder_id = ? AND removed = 0", s.rootFolderID).Scan(&totalSize)
	if err != nil {
		return fmt.Errorf("failed to calculate total size: %w", err)
	}

	// Update root folder statistics
	_, err = s.db.Exec(`
		UPDATE root_folders 
		SET folder_count = ?, 
			file_count = ?, 
			total_size = ?, 
			last_scan_date = CURRENT_TIMESTAMP
		WHERE id = ?
	`, folderCount, fileCount, totalSize, s.rootFolderID)

	if err != nil {
		return fmt.Errorf("failed to update statistics: %w", err)
	}

	logger.Info("Statistics updated: %d folders, %d files, %d bytes", folderCount, fileCount, totalSize)
	return nil
}
