package datastore

import (
	"database/sql"
	"fmt"
	"time"
)

// EnsureRootFolder ensures a root folder exists and returns its ID
// If the path doesn't exist, creates a new entry
func EnsureRootFolder(path, name string) (int, error) {
	db, err := GetDB()
	if err != nil {
		return 0, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Check if root folder already exists
	var id int
	err = db.QueryRow("SELECT id FROM root_folders WHERE path = ?", path).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to query root folder: %w", err)
	}

	// Create new root folder
	createdAt := time.Now().Unix()
	result, err := db.Exec(`
		INSERT INTO root_folders (path, name, created_at)
		VALUES (?, ?, ?)
	`, path, name, createdAt)

	if err != nil {
		return 0, fmt.Errorf("failed to create root folder: %w", err)
	}

	insertID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get root folder ID: %w", err)
	}

	return int(insertID), nil
}

// GetRootFolderByPath retrieves a root folder by its path
func GetRootFolderByPath(path string) (int, error) {
	db, err := GetDB()
	if err != nil {
		return 0, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	var id int
	err = db.QueryRow("SELECT id FROM root_folders WHERE path = ?", path).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("root folder not found: %s", path)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to query root folder: %w", err)
	}

	return id, nil
}
