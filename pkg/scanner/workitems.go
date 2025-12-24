package scanner

import (
	"context"
	"database/sql"

	"github.com/jpconstantineau/dupectl/pkg/datastore"
	"github.com/jpconstantineau/dupectl/pkg/errors"
	"github.com/jpconstantineau/dupectl/pkg/hash"
	"github.com/jpconstantineau/dupectl/pkg/logger"
)

// FileHashingWorkItem processes file hashing
type FileHashingWorkItem struct {
	db       *sql.DB
	fileID   int64
	filePath string
	hasher   hash.Hasher
	progress *ProgressIndicator
}

// NewFileHashingWorkItem creates a file hashing work item
func NewFileHashingWorkItem(db *sql.DB, fileID int64, filePath string, hasher hash.Hasher, progress *ProgressIndicator) *FileHashingWorkItem {
	return &FileHashingWorkItem{
		db:       db,
		fileID:   fileID,
		filePath: filePath,
		hasher:   hasher,
		progress: progress,
	}
}

// Process calculates file hash and updates database
func (w *FileHashingWorkItem) Process(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Calculate hash
	hashValue, err := w.hasher.Hash(ctx, w.filePath)
	if err != nil {
		logger.Warn("Failed to hash file %s: %v", w.filePath, err)

		// Update file with error status
		errMsg := err.Error()
		_, updateErr := datastore.InsertFile(w.db, &datastore.File{
			ID:          w.fileID,
			Path:        w.filePath,
			ErrorStatus: &errMsg,
		})
		if updateErr != nil {
			return errors.NewDatabaseError("update file error status", updateErr)
		}
		return nil // Don't fail worker pool
	}

	// Update database with hash
	err = datastore.UpdateFileHash(w.db, w.fileID, hashValue, w.hasher.Algorithm())
	if err != nil {
		logger.Error("Failed to update hash for %s: %v", w.filePath, err)
		return errors.NewDatabaseError("update file hash", err)
	}

	// Update progress
	if w.progress != nil {
		w.progress.IncrementFilesHashed()
	}

	logger.Debug("Hashed file %s: %s", w.filePath, hashValue[:16]+"...")
	return nil
}

// ID returns work item identifier
func (w *FileHashingWorkItem) ID() string {
	return w.filePath
}
