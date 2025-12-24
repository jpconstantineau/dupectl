package datastore

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jpconstantineau/dupectl/pkg/entities"
)

// InsertFolder inserts a new folder record into the database
func InsertFolder(folder entities.ScanFolder) (int, error) {
	db, err := GetDB()
	if err != nil {
		return 0, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	var scannedAt *int64
	if folder.ScannedAt != nil {
		ts := folder.ScannedAt.Unix()
		scannedAt = &ts
	}

	result, err := db.Exec(`
		INSERT INTO folders (root_folder_id, parent_folder_id, path, name, scanned_at)
		VALUES (?, ?, ?, ?, ?)
	`, folder.RootFolderID, folder.ParentFolderID, folder.Path, folder.Name, scannedAt)

	if err != nil {
		return 0, fmt.Errorf("failed to insert folder: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get inserted folder ID: %w", err)
	}

	return int(id), nil
}

// GetFolderByPath retrieves a folder by its path
func GetFolderByPath(path string) (*entities.ScanFolder, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	var folder entities.ScanFolder
	var scannedAt *int64

	err = db.QueryRow(`
		SELECT id, root_folder_id, parent_folder_id, path, name, scanned_at
		FROM folders
		WHERE path = ?
	`, path).Scan(&folder.ID, &folder.RootFolderID, &folder.ParentFolderID, &folder.Path, &folder.Name, &scannedAt)

	if err == sql.ErrNoRows {
		return nil, nil // Folder not found, not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query folder: %w", err)
	}

	// Convert Unix timestamp to time.Time
	if scannedAt != nil {
		t := time.Unix(*scannedAt, 0)
		folder.ScannedAt = &t
	}

	return &folder, nil
}

// GetChildFolders retrieves all child folders of a parent folder
func GetChildFolders(parentID int) ([]entities.ScanFolder, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, root_folder_id, parent_folder_id, path, name, scanned_at
		FROM folders
		WHERE parent_folder_id = ?
		ORDER BY name
	`, parentID)

	if err != nil {
		return nil, fmt.Errorf("failed to query child folders: %w", err)
	}
	defer rows.Close()

	var folders []entities.ScanFolder
	for rows.Next() {
		var folder entities.ScanFolder
		var scannedAt *int64

		err = rows.Scan(&folder.ID, &folder.RootFolderID, &folder.ParentFolderID, &folder.Path, &folder.Name, &scannedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan folder row: %w", err)
		}

		// Convert Unix timestamp to time.Time
		if scannedAt != nil {
			t := time.Unix(*scannedAt, 0)
			folder.ScannedAt = &t
		}

		folders = append(folders, folder)
	}

	return folders, nil
}

// UpdateFolderScannedAt updates the scanned_at timestamp for a folder
func UpdateFolderScannedAt(id int, scannedAt time.Time) error {
	db, err := GetDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	timestamp := scannedAt.Unix()

	_, err = db.Exec(`
		UPDATE folders
		SET scanned_at = ?
		WHERE id = ?
	`, timestamp, id)

	if err != nil {
		return fmt.Errorf("failed to update folder scanned_at: %w", err)
	}

	return nil
}

// GetFoldersInRootFolder retrieves all folders under a root folder
func GetFoldersInRootFolder(rootFolderID int) ([]entities.ScanFolder, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, root_folder_id, parent_folder_id, path, name, scanned_at
		FROM folders
		WHERE root_folder_id = ?
		ORDER BY path
	`, rootFolderID)

	if err != nil {
		return nil, fmt.Errorf("failed to query folders: %w", err)
	}
	defer rows.Close()

	var folders []entities.ScanFolder
	for rows.Next() {
		var folder entities.ScanFolder
		var scannedAt *int64

		err = rows.Scan(&folder.ID, &folder.RootFolderID, &folder.ParentFolderID, &folder.Path, &folder.Name, &scannedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan folder row: %w", err)
		}

		// Convert Unix timestamp to time.Time
		if scannedAt != nil {
			t := time.Unix(*scannedAt, 0)
			folder.ScannedAt = &t
		}

		folders = append(folders, folder)
	}

	return folders, nil
}
