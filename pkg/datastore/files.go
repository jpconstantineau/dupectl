package datastore

import (
	"database/sql"
	"path/filepath"
	"strings"
)

const CreateFilesTableSQL = `
CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL UNIQUE,
    size INTEGER NOT NULL,
    mtime INTEGER NOT NULL,
    hash_value TEXT,
    hash_algorithm TEXT,
    error_status TEXT,
    first_scanned_at INTEGER NOT NULL,
    last_scanned_at INTEGER NOT NULL,
    removed INTEGER NOT NULL DEFAULT 0,
    folder_id INTEGER NOT NULL,
    root_folder_id INTEGER NOT NULL,
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE CASCADE,
    FOREIGN KEY (root_folder_id) REFERENCES root_folders(id) ON DELETE CASCADE
);`

const CreateFilesIndexesSQL = `
CREATE INDEX IF NOT EXISTS idx_files_hash ON files(hash_value, size) 
    WHERE removed = 0 AND error_status IS NULL AND hash_value IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_files_folder ON files(folder_id);
CREATE INDEX IF NOT EXISTS idx_files_root ON files(root_folder_id);
CREATE INDEX IF NOT EXISTS idx_files_removed ON files(removed);
CREATE INDEX IF NOT EXISTS idx_files_path ON files(path);
`

// File represents a file record in the database
type File struct {
	ID             int64
	Path           string
	Size           int64
	Mtime          int64
	HashValue      *string
	HashAlgorithm  *string
	ErrorStatus    *string
	FirstScannedAt int64
	LastScannedAt  int64
	Removed        bool
	FolderID       int64
	RootFolderID   int64
	RootFolderPath string // For display purposes
}

// InsertFile inserts a new file record
func InsertFile(db *sql.DB, file *File) (int64, error) {
	query := `
	INSERT INTO files (path, size, mtime, hash_value, hash_algorithm, error_status, 
	                   first_scanned_at, last_scanned_at, removed, folder_id, root_folder_id)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(path) DO UPDATE SET
		size = excluded.size,
		mtime = excluded.mtime,
		hash_value = excluded.hash_value,
		hash_algorithm = excluded.hash_algorithm,
		error_status = excluded.error_status,
		last_scanned_at = excluded.last_scanned_at,
		removed = 0,
		folder_id = excluded.folder_id,
		root_folder_id = excluded.root_folder_id
	RETURNING id
	`

	removed := 0
	if file.Removed {
		removed = 1
	}

	var id int64
	err := db.QueryRow(query, file.Path, file.Size, file.Mtime, file.HashValue,
		file.HashAlgorithm, file.ErrorStatus, file.FirstScannedAt, file.LastScannedAt,
		removed, file.FolderID, file.RootFolderID).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// UpdateFileHash updates the hash value for a file
func UpdateFileHash(db *sql.DB, fileID int64, hashValue, hashAlgorithm string) error {
	query := `UPDATE files SET hash_value = ?, hash_algorithm = ?, last_scanned_at = strftime('%s', 'now') WHERE id = ?`
	_, err := db.Exec(query, hashValue, hashAlgorithm, fileID)
	return err
}

// SetFileRemoved marks a file as removed
func SetFileRemoved(db *sql.DB, path string, removed bool) error {
	removedInt := 0
	if removed {
		removedInt = 1
	}
	query := `UPDATE files SET removed = ?, last_scanned_at = strftime('%s', 'now') WHERE path = ?`
	_, err := db.Exec(query, removedInt, path)
	return err
}

// GetFilesByHash returns all files with matching hash and size
func GetFilesByHash(db *sql.DB, hashValue string, size int64) ([]*File, error) {
	query := `
	SELECT f.id, f.path, f.size, f.mtime, f.hash_value, f.hash_algorithm, f.error_status,
	       f.first_scanned_at, f.last_scanned_at, f.removed, f.folder_id, f.root_folder_id,
	       COALESCE(rf.path, '') as root_folder_path
	FROM files f
	LEFT JOIN root_folders rf ON f.root_folder_id = rf.id
	WHERE f.hash_value = ? AND f.size = ? AND f.removed = 0 AND f.error_status IS NULL
	ORDER BY f.path
	`

	rows, err := db.Query(query, hashValue, size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*File
	seenPaths := make(map[string]bool) // Track normalized paths to avoid duplicates

	for rows.Next() {
		file := &File{}
		var removed int
		err := rows.Scan(&file.ID, &file.Path, &file.Size, &file.Mtime, &file.HashValue,
			&file.HashAlgorithm, &file.ErrorStatus, &file.FirstScannedAt, &file.LastScannedAt,
			&removed, &file.FolderID, &file.RootFolderID, &file.RootFolderPath)
		if err != nil {
			return nil, err
		}
		file.Removed = removed != 0

		// Normalize path for comparison (case-insensitive on Windows)
		normalizedPath := normalizePathForComparison(file.Path)

		// Skip if we've already seen this path (case-insensitive duplicate)
		if seenPaths[normalizedPath] {
			continue
		}

		seenPaths[normalizedPath] = true
		files = append(files, file)
	}

	return files, rows.Err()
}

// normalizePathForComparison normalizes paths for case-insensitive comparison
func normalizePathForComparison(path string) string {
	// On Windows, convert to lowercase for comparison
	// On Unix, paths are case-sensitive so return as-is
	return strings.ToLower(filepath.Clean(path))
}
