package scanner

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/jpconstantineau/dupectl/pkg/datastore"
	"github.com/jpconstantineau/dupectl/pkg/errors"
	"github.com/jpconstantineau/dupectl/pkg/logger"
	"github.com/jpconstantineau/dupectl/pkg/pathutil"
)

// Traverser handles folder tree traversal
type Traverser struct {
	db            *sql.DB
	rootFolderID  int64
	rootPath      string
	traverseLinks bool
}

// NewTraverser creates a folder traverser
func NewTraverser(db *sql.DB, rootFolderID int64, rootPath string, traverseLinks bool) *Traverser {
	return &Traverser{
		db:            db,
		rootFolderID:  rootFolderID,
		rootPath:      filepath.Clean(rootPath),
		traverseLinks: traverseLinks,
	}
}

// FolderInfo contains discovered folder information
type FolderInfo struct {
	Path        string
	ParentPath  *string
	Files       []FileInfo
	ErrorStatus *string
}

// FileInfo contains discovered file information
type FileInfo struct {
	Path  string
	Size  int64
	Mtime int64
}

// Traverse walks the directory tree and returns folders with their files
func (t *Traverser) Traverse(ctx context.Context, callback func(*FolderInfo) error) error {
	return t.traverseDir(ctx, t.rootPath, nil, callback)
}

func (t *Traverser) traverseDir(ctx context.Context, dirPath string, parentPath *string, callback func(*FolderInfo) error) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Check if we should skip this directory (symlink handling)
	if !t.traverseLinks {
		info, err := os.Lstat(dirPath)
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			logger.Debug("Skipping symlink: %s", dirPath)
			return nil
		}
	}

	folderInfo := &FolderInfo{
		Path:       pathutil.NormalizePathForStorage(dirPath),
		ParentPath: parentPath,
		Files:      []FileInfo{},
	}

	// Read directory entries
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		// Permission denied or other error
		errMsg := err.Error()
		folderInfo.ErrorStatus = &errMsg
		logger.Warn("Cannot read directory %s: %v", dirPath, err)

		// Still callback with error status
		if callbackErr := callback(folderInfo); callbackErr != nil {
			return callbackErr
		}
		return nil // Don't fail entire scan
	}

	// Collect files and subdirectories
	var subdirs []string
	for _, entry := range entries {
		entryPath := pathutil.NormalizePathForStorage(pathutil.Join(dirPath, entry.Name()))

		if entry.IsDir() {
			subdirs = append(subdirs, entryPath)
		} else {
			// Get file info
			info, err := entry.Info()
			if err != nil {
				logger.Warn("Cannot stat file %s: %v", entryPath, err)
				continue
			}

			folderInfo.Files = append(folderInfo.Files, FileInfo{
				Path:  entryPath,
				Size:  info.Size(),
				Mtime: info.ModTime().Unix(),
			})
		}
	}

	// Callback with this folder's info
	if err := callback(folderInfo); err != nil {
		return err
	}

	// Recursively traverse subdirectories
	for _, subdir := range subdirs {
		if err := t.traverseDir(ctx, subdir, &dirPath, callback); err != nil {
			return err
		}
	}

	return nil
}

// RegisterFolder registers a folder in the database
func RegisterFolder(db *sql.DB, rootFolderID int64, folderPath string, parentPath *string, errorStatus *string) (int64, error) {
	now := time.Now().Unix()

	// Get parent folder ID if parent exists
	var parentFolderID *int64
	if parentPath != nil {
		parent, err := datastore.GetFolderByPath(db, *parentPath)
		if err != nil && err != sql.ErrNoRows {
			return 0, errors.NewDatabaseError("get parent folder", err)
		}
		if parent != nil {
			parentFolderID = &parent.ID
		}
	}

	folder := &datastore.Folder{
		Path:           folderPath,
		ParentFolderID: parentFolderID,
		RootFolderID:   rootFolderID,
		ErrorStatus:    errorStatus,
		FirstScannedAt: now,
		LastScannedAt:  now,
		Removed:        false,
	}

	id, err := datastore.InsertFolder(db, folder)
	if err != nil {
		return 0, errors.NewDatabaseError("insert folder", err)
	}

	return id, nil
}

// RegisterFile registers a file in the database (without hash)
func RegisterFile(db *sql.DB, folderID, rootFolderID int64, fileInfo *FileInfo, errorStatus *string) (int64, error) {
	now := time.Now().Unix()

	file := &datastore.File{
		Path:           fileInfo.Path,
		Size:           fileInfo.Size,
		Mtime:          fileInfo.Mtime,
		FirstScannedAt: now,
		LastScannedAt:  now,
		Removed:        false,
		FolderID:       folderID,
		RootFolderID:   rootFolderID,
		ErrorStatus:    errorStatus,
	}

	id, err := datastore.InsertFile(db, file)
	if err != nil {
		return 0, errors.NewDatabaseError("insert file", err)
	}

	return id, nil
}
