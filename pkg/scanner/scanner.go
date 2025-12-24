package scanner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jpconstantineau/dupectl/pkg/datastore"
	"github.com/jpconstantineau/dupectl/pkg/entities"
	"github.com/jpconstantineau/dupectl/pkg/hash"
	"github.com/spf13/viper"
)

// Scanner is the interface for scanning files and folders
type Scanner interface {
	Scan(ctx context.Context, rootPath string, rootFolderID int, scanMode entities.ScanMode) error
}

// FileSystemScanner implements Scanner for filesystem scanning
type FileSystemScanner struct {
	progress  *ProgressTracker
	hasher    hash.Hasher
	verbose   bool
	batchSize int
}

// NewFileSystemScanner creates a new filesystem scanner
func NewFileSystemScanner(verbose bool) (*FileSystemScanner, error) {
	// Get hash algorithm from config
	algorithm := viper.GetString("scan.hash_algorithm")
	hasher, err := hash.NewHasher(algorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to create hasher: %w", err)
	}

	// Get progress interval from config
	intervalStr := viper.GetString("scan.progress_interval")
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		return nil, fmt.Errorf("invalid progress_interval: %w", err)
	}

	batchSize := viper.GetInt("scan.batch_size")
	if batchSize <= 0 {
		batchSize = 1000
	}

	return &FileSystemScanner{
		progress:  NewProgressTracker(interval),
		hasher:    hasher,
		verbose:   verbose,
		batchSize: batchSize,
	}, nil
}

// Scan performs the filesystem scan based on the scan mode
func (s *FileSystemScanner) Scan(ctx context.Context, rootPath string, rootFolderID int, scanMode entities.ScanMode) error {
	// Start progress reporting
	s.progress.Start()
	defer s.progress.Stop()

	// Check for existing scan state (resume capability)
	scanState, err := s.checkExistingScanState(rootFolderID, scanMode)
	if err != nil {
		return err
	}

	// Create or resume scan state
	if scanState == nil {
		scanState, err = s.createNewScanState(rootFolderID, scanMode)
		if err != nil {
			return err
		}
	} else {
		if s.verbose {
			fmt.Printf("Resuming scan from checkpoint: %s\n", *scanState.CurrentFolderPath)
		}
	}

	// Perform scan based on mode
	switch scanMode {
	case entities.ScanModeAll:
		err = s.scanAll(ctx, rootPath, rootFolderID, scanState)
	case entities.ScanModeFolders:
		err = s.scanFolders(ctx, rootPath, rootFolderID, scanState)
	case entities.ScanModeFiles:
		err = s.scanFiles(ctx, rootPath, rootFolderID, scanState)
	default:
		return fmt.Errorf("unsupported scan mode: %s", scanMode)
	}

	if err != nil {
		// Mark scan as interrupted
		if scanState != nil {
			datastore.MarkInterrupted(scanState.ID)
		}
		return err
	}

	// Mark scan as completed
	if scanState != nil {
		err = datastore.CompleteScanState(scanState.ID)
		if err != nil {
			return fmt.Errorf("failed to mark scan complete: %w", err)
		}
	}

	return nil
}

// checkExistingScanState checks for an existing active scan state
func (s *FileSystemScanner) checkExistingScanState(rootFolderID int, scanMode entities.ScanMode) (*entities.ScanState, error) {
	scanState, err := datastore.GetActiveScanState(rootFolderID)
	if err != nil {
		return nil, fmt.Errorf("failed to check scan state: %w", err)
	}

	// If found and mode matches, resume
	if scanState != nil && scanState.ScanMode == scanMode {
		return scanState, nil
	}

	return nil, nil
}

// createNewScanState creates a new scan state record
func (s *FileSystemScanner) createNewScanState(rootFolderID int, scanMode entities.ScanMode) (*entities.ScanState, error) {
	now := time.Now()
	scanState := entities.ScanState{
		RootFolderID:     rootFolderID,
		ScanMode:         scanMode,
		StartedAt:        now,
		UpdatedAt:        now,
		Status:           entities.ScanStatusRunning,
		FilesProcessed:   0,
		FoldersProcessed: 0,
	}

	id, err := datastore.CreateScanState(scanState)
	if err != nil {
		return nil, fmt.Errorf("failed to create scan state: %w", err)
	}

	scanState.ID = id
	return &scanState, nil
}

// scanAll performs a complete scan of folders and files
func (s *FileSystemScanner) scanAll(ctx context.Context, rootPath string, rootFolderID int, scanState *entities.ScanState) error {
	// Scan folders and hash files in one pass
	return s.walkDirectoryTree(ctx, rootPath, rootFolderID, scanState, true)
}

// scanFolders performs a folder-only scan
func (s *FileSystemScanner) scanFolders(ctx context.Context, rootPath string, rootFolderID int, scanState *entities.ScanState) error {
	// Scan folders only, no file hashing
	return s.walkDirectoryTree(ctx, rootPath, rootFolderID, scanState, false)
}

// scanFiles performs a file-only scan (hashing files in already scanned folders)
func (s *FileSystemScanner) scanFiles(ctx context.Context, rootPath string, rootFolderID int, scanState *entities.ScanState) error {
	// Get all folders for this root
	folders, err := datastore.GetFoldersInRootFolder(rootFolderID)
	if err != nil {
		return fmt.Errorf("failed to get folders: %w", err)
	}

	if len(folders) == 0 {
		return fmt.Errorf("no folders found - run 'dupectl scan folders' first")
	}

	// Process each folder
	for _, folder := range folders {
		select {
		case <-ctx.Done():
			return fmt.Errorf("scan interrupted")
		default:
		}

		err = s.hashFilesInFolder(folder.Path, rootFolderID, folder.ID)
		if err != nil {
			if s.verbose {
				fmt.Printf("Error hashing files in %s: %v\n", folder.Path, err)
			}
		}

		// Update checkpoint
		folderPath := folder.Path
		scanState.CurrentFolderPath = &folderPath
		err = datastore.UpdateCheckpoint(scanState.ID, folderPath,
			int(s.progress.filesProcessed.Load()),
			int(s.progress.foldersProcessed.Load()))
		if err != nil && s.verbose {
			fmt.Printf("Warning: failed to save checkpoint: %v\n", err)
		}
	}

	return nil
}

// walkDirectoryTree recursively walks the directory tree
func (s *FileSystemScanner) walkDirectoryTree(ctx context.Context, rootPath string, rootFolderID int, scanState *entities.ScanState, hashFiles bool) error {
	return filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("scan interrupted")
		default:
		}

		// Handle permission errors
		if err != nil {
			if s.verbose {
				fmt.Printf("Warning: %v\n", err)
			}
			if info != nil && info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Process directory
		if info.IsDir() {
			return s.processDirectory(path, rootPath, rootFolderID, scanState)
		}

		// Process file
		if hashFiles {
			return s.processFile(path, info, rootFolderID, scanState)
		}

		return nil
	})
}

// processDirectory processes a directory entry
func (s *FileSystemScanner) processDirectory(path string, rootPath string, rootFolderID int, scanState *entities.ScanState) error {
	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if folder already exists
	existingFolder, err := datastore.GetFolderByPath(absPath)
	if err != nil {
		return fmt.Errorf("failed to check existing folder: %w", err)
	}

	if existingFolder == nil {
		// Create folder record
		now := time.Now()
		folder := entities.ScanFolder{
			RootFolderID: rootFolderID,
			Path:         absPath,
			Name:         filepath.Base(absPath),
			ScannedAt:    &now,
		}

		// Set parent folder ID if not root
		if absPath != rootPath {
			parentPath := filepath.Dir(absPath)
			parentFolder, err := datastore.GetFolderByPath(parentPath)
			if err == nil && parentFolder != nil {
				folder.ParentFolderID = &parentFolder.ID
			}
		}

		_, err = datastore.InsertFolder(folder)
		if err != nil && s.verbose {
			fmt.Printf("Warning: failed to insert folder %s: %v\n", absPath, err)
		}
	}

	s.progress.IncrementFolders()

	// Save checkpoint every folder
	err = datastore.UpdateCheckpoint(scanState.ID, absPath,
		int(s.progress.filesProcessed.Load()),
		int(s.progress.foldersProcessed.Load()))
	if err != nil && s.verbose {
		fmt.Printf("Warning: failed to save checkpoint: %v\n", err)
	}

	return nil
}

// processFile processes a file entry
func (s *FileSystemScanner) processFile(path string, info os.FileInfo, rootFolderID int, scanState *entities.ScanState) error {
	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if file already exists and is hashed
	existingFile, err := datastore.GetFileByPath(absPath)
	if err != nil {
		return fmt.Errorf("failed to check existing file: %w", err)
	}

	if existingFile != nil && existingFile.IsHashed() {
		// File already hashed, skip
		s.progress.IncrementFiles()
		return nil
	}

	// Get folder ID
	folderPath := filepath.Dir(absPath)
	folder, err := datastore.GetFolderByPath(folderPath)
	var folderID *int
	if err == nil && folder != nil {
		folderID = &folder.ID
	}

	// Create or update file record
	file := entities.File{
		RootFolderID: rootFolderID,
		FolderID:     folderID,
		Path:         absPath,
		Name:         filepath.Base(absPath),
		Size:         info.Size(),
		Mtime:        info.ModTime(),
	}

	var fileID int
	if existingFile == nil {
		fileID, err = datastore.InsertFile(file)
		if err != nil {
			if s.verbose {
				fmt.Printf("Warning: failed to insert file %s: %v\n", absPath, err)
			}
			return nil
		}
	} else {
		fileID = existingFile.ID
	}

	// Hash the file
	err = s.hashFile(absPath, fileID)
	if err != nil {
		// Mark file with error
		datastore.MarkFileError(fileID, err.Error())
		if s.verbose {
			fmt.Printf("Warning: failed to hash file %s: %v\n", absPath, err)
		}
	}

	s.progress.IncrementFiles()
	return nil
}

// hashFile calculates the hash of a file
func (s *FileSystemScanner) hashFile(path string, fileID int) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hashValue, err := s.hasher.HashFile(file)
	if err != nil {
		return fmt.Errorf("failed to hash file: %w", err)
	}

	err = datastore.UpdateFileHash(fileID, hashValue, s.hasher.Algorithm())
	if err != nil {
		return fmt.Errorf("failed to update file hash: %w", err)
	}

	return nil
}

// hashFilesInFolder hashes all files in a specific folder
func (s *FileSystemScanner) hashFilesInFolder(folderPath string, rootFolderID int, folderID int) error {
	// Read directory
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(folderPath, entry.Name())
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			continue
		}

		// Get file info
		info, err := entry.Info()
		if err != nil {
			if s.verbose {
				fmt.Printf("Warning: failed to get file info for %s: %v\n", absPath, err)
			}
			continue
		}

		// Check if file exists in database
		existingFile, err := datastore.GetFileByPath(absPath)
		if err != nil {
			continue
		}

		var fileID int
		if existingFile == nil {
			// Create file record
			file := entities.File{
				RootFolderID: rootFolderID,
				FolderID:     &folderID,
				Path:         absPath,
				Name:         entry.Name(),
				Size:         info.Size(),
				Mtime:        info.ModTime(),
			}

			fileID, err = datastore.InsertFile(file)
			if err != nil {
				continue
			}
		} else {
			fileID = existingFile.ID
			if existingFile.IsHashed() {
				s.progress.IncrementFiles()
				continue
			}
		}

		// Hash the file
		err = s.hashFile(absPath, fileID)
		if err != nil {
			datastore.MarkFileError(fileID, err.Error())
			if s.verbose {
				fmt.Printf("Warning: failed to hash file %s: %v\n", absPath, err)
			}
		}

		s.progress.IncrementFiles()
	}

	return nil
}
