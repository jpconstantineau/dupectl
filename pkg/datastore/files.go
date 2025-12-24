package datastore

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jpconstantineau/dupectl/pkg/entities"
)

// InsertFile inserts a new file record into the database
func InsertFile(file entities.File) (int, error) {
	db, err := GetDB()
	if err != nil {
		return 0, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Convert time.Time to Unix timestamp (int64)
	mtime := file.Mtime.Unix()
	var scannedAt *int64
	if file.ScannedAt != nil {
		ts := file.ScannedAt.Unix()
		scannedAt = &ts
	}

	result, err := db.Exec(`
		INSERT INTO files (root_folder_id, folder_id, path, name, size, mtime, hash_value, hash_algorithm, error_status, scanned_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, file.RootFolderID, file.FolderID, file.Path, file.Name, file.Size, mtime, file.HashValue, file.HashAlgorithm, file.ErrorStatus, scannedAt)

	if err != nil {
		return 0, fmt.Errorf("failed to insert file: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get inserted file ID: %w", err)
	}

	return int(id), nil
}

// GetFileByPath retrieves a file by its path
func GetFileByPath(path string) (*entities.File, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	var file entities.File
	var mtime, scannedAt *int64

	err = db.QueryRow(`
		SELECT id, root_folder_id, folder_id, path, name, size, mtime, hash_value, hash_algorithm, error_status, scanned_at
		FROM files
		WHERE path = ?
	`, path).Scan(&file.ID, &file.RootFolderID, &file.FolderID, &file.Path, &file.Name, &file.Size, &mtime, &file.HashValue, &file.HashAlgorithm, &file.ErrorStatus, &scannedAt)

	if err == sql.ErrNoRows {
		return nil, nil // File not found, not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query file: %w", err)
	}

	// Convert Unix timestamps to time.Time
	if mtime != nil {
		t := time.Unix(*mtime, 0)
		file.Mtime = t
	}
	if scannedAt != nil {
		t := time.Unix(*scannedAt, 0)
		file.ScannedAt = &t
	}

	return &file, nil
}

// UpdateFileHash updates the hash value and algorithm for a file
func UpdateFileHash(id int, hashValue, hashAlgorithm string) error {
	db, err := GetDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	scannedAt := time.Now().Unix()

	_, err = db.Exec(`
		UPDATE files
		SET hash_value = ?, hash_algorithm = ?, scanned_at = ?
		WHERE id = ?
	`, hashValue, hashAlgorithm, scannedAt, id)

	if err != nil {
		return fmt.Errorf("failed to update file hash: %w", err)
	}

	return nil
}

// MarkFileError marks a file with an error status
func MarkFileError(id int, errorStatus string) error {
	db, err := GetDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	scannedAt := time.Now().Unix()

	_, err = db.Exec(`
		UPDATE files
		SET error_status = ?, scanned_at = ?
		WHERE id = ?
	`, errorStatus, scannedAt, id)

	if err != nil {
		return fmt.Errorf("failed to mark file error: %w", err)
	}

	return nil
}

// GetFilesInFolder retrieves all files in a specific folder
func GetFilesInFolder(folderID int) ([]entities.File, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, root_folder_id, folder_id, path, name, size, mtime, hash_value, hash_algorithm, error_status, scanned_at
		FROM files
		WHERE folder_id = ?
		ORDER BY name
	`, folderID)

	if err != nil {
		return nil, fmt.Errorf("failed to query files: %w", err)
	}
	defer rows.Close()

	var files []entities.File
	for rows.Next() {
		var file entities.File
		var mtime, scannedAt *int64

		err = rows.Scan(&file.ID, &file.RootFolderID, &file.FolderID, &file.Path, &file.Name, &file.Size, &mtime, &file.HashValue, &file.HashAlgorithm, &file.ErrorStatus, &scannedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file row: %w", err)
		}

		// Convert Unix timestamps to time.Time
		if mtime != nil {
			t := time.Unix(*mtime, 0)
			file.Mtime = t
		}
		if scannedAt != nil {
			t := time.Unix(*scannedAt, 0)
			file.ScannedAt = &t
		}

		files = append(files, file)
	}

	return files, nil
}

// GetFilesInRootFolder retrieves all files under a root folder
func GetFilesInRootFolder(rootFolderID int) ([]entities.File, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, root_folder_id, folder_id, path, name, size, mtime, hash_value, hash_algorithm, error_status, scanned_at
		FROM files
		WHERE root_folder_id = ?
		ORDER BY path
	`, rootFolderID)

	if err != nil {
		return nil, fmt.Errorf("failed to query files: %w", err)
	}
	defer rows.Close()

	var files []entities.File
	for rows.Next() {
		var file entities.File
		var mtime, scannedAt *int64

		err = rows.Scan(&file.ID, &file.RootFolderID, &file.FolderID, &file.Path, &file.Name, &file.Size, &mtime, &file.HashValue, &file.HashAlgorithm, &file.ErrorStatus, &scannedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file row: %w", err)
		}

		// Convert Unix timestamps to time.Time
		if mtime != nil {
			t := time.Unix(*mtime, 0)
			file.Mtime = t
		}
		if scannedAt != nil {
			t := time.Unix(*scannedAt, 0)
			file.ScannedAt = &t
		}

		files = append(files, file)
	}

	return files, nil
}
