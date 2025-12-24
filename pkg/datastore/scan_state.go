package datastore

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jpconstantineau/dupectl/pkg/entities"
)

// CreateScanState creates a new scan state record for tracking scan progress
func CreateScanState(state entities.ScanState) (int, error) {
	db, err := GetDB()
	if err != nil {
		return 0, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	startedAt := state.StartedAt.Unix()
	updatedAt := state.UpdatedAt.Unix()
	var completedAt *int64
	if state.CompletedAt != nil {
		ts := state.CompletedAt.Unix()
		completedAt = &ts
	}

	result, err := db.Exec(`
		INSERT INTO scan_state (root_folder_id, scan_mode, current_folder_path, started_at, updated_at, completed_at, status, files_processed, folders_processed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, state.RootFolderID, state.ScanMode, state.CurrentFolderPath, startedAt, updatedAt, completedAt, state.Status, state.FilesProcessed, state.FoldersProcessed)

	if err != nil {
		return 0, fmt.Errorf("failed to create scan state: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get inserted scan state ID: %w", err)
	}

	return int(id), nil
}

// GetActiveScanState retrieves the active or interrupted scan state for a root folder
func GetActiveScanState(rootFolderID int) (*entities.ScanState, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	var state entities.ScanState
	var startedAt, updatedAt, completedAt *int64
	var scanMode, status string

	err = db.QueryRow(`
		SELECT id, root_folder_id, scan_mode, current_folder_path, started_at, updated_at, completed_at, status, files_processed, folders_processed
		FROM scan_state
		WHERE root_folder_id = ? AND (status = ? OR status = ?)
		ORDER BY started_at DESC
		LIMIT 1
	`, rootFolderID, entities.ScanStatusRunning, entities.ScanStatusInterrupted).Scan(
		&state.ID, &state.RootFolderID, &scanMode, &state.CurrentFolderPath,
		&startedAt, &updatedAt, &completedAt, &status,
		&state.FilesProcessed, &state.FoldersProcessed,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No active scan found, not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query scan state: %w", err)
	}

	// Convert string to typed enum
	state.ScanMode = entities.ScanMode(scanMode)
	state.Status = entities.ScanStatus(status)

	// Convert Unix timestamps to time.Time
	if startedAt != nil {
		t := time.Unix(*startedAt, 0)
		state.StartedAt = t
	}
	if updatedAt != nil {
		t := time.Unix(*updatedAt, 0)
		state.UpdatedAt = t
	}
	if completedAt != nil {
		t := time.Unix(*completedAt, 0)
		state.CompletedAt = &t
	}

	return &state, nil
}

// UpdateCheckpoint updates the scan progress checkpoint
func UpdateCheckpoint(id int, folderPath string, filesCount, foldersCount int) error {
	db, err := GetDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	updatedAt := time.Now().Unix()

	_, err = db.Exec(`
		UPDATE scan_state
		SET current_folder_path = ?, files_processed = ?, folders_processed = ?, updated_at = ?
		WHERE id = ?
	`, folderPath, filesCount, foldersCount, updatedAt, id)

	if err != nil {
		return fmt.Errorf("failed to update checkpoint: %w", err)
	}

	return nil
}

// CompleteScanState marks a scan as completed
func CompleteScanState(id int) error {
	db, err := GetDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	now := time.Now().Unix()

	_, err = db.Exec(`
		UPDATE scan_state
		SET status = ?, completed_at = ?, updated_at = ?
		WHERE id = ?
	`, entities.ScanStatusCompleted, now, now, id)

	if err != nil {
		return fmt.Errorf("failed to complete scan state: %w", err)
	}

	return nil
}

// MarkInterrupted marks a scan as interrupted
func MarkInterrupted(id int) error {
	db, err := GetDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	updatedAt := time.Now().Unix()

	_, err = db.Exec(`
		UPDATE scan_state
		SET status = ?, updated_at = ?
		WHERE id = ?
	`, entities.ScanStatusInterrupted, updatedAt, id)

	if err != nil {
		return fmt.Errorf("failed to mark scan interrupted: %w", err)
	}

	return nil
}

// DeleteScanState deletes a scan state record (used with --rescan flag)
func DeleteScanState(rootFolderID int) error {
	db, err := GetDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		DELETE FROM scan_state
		WHERE root_folder_id = ?
	`, rootFolderID)

	if err != nil {
		return fmt.Errorf("failed to delete scan state: %w", err)
	}

	return nil
}
