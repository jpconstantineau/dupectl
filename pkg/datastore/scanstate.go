package datastore

import (
	"database/sql"
)

const CreateScanStateTableSQL = `
CREATE TABLE IF NOT EXISTS scan_state (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    root_folder_id INTEGER NOT NULL,
    scan_mode TEXT NOT NULL,
    current_folder_path TEXT,
    last_processed_file TEXT,
    started_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    completed INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (root_folder_id) REFERENCES root_folders(id) ON DELETE CASCADE
);`

const CreateScanStateIndexesSQL = `
CREATE INDEX IF NOT EXISTS idx_scan_state_root ON scan_state(root_folder_id);
CREATE INDEX IF NOT EXISTS idx_scan_state_completed ON scan_state(completed);
CREATE UNIQUE INDEX IF NOT EXISTS idx_scan_state_unique_active ON scan_state(root_folder_id, completed) WHERE completed = 0;
`

// ScanState represents a scan checkpoint
type ScanState struct {
	ID                int64
	RootFolderID      int64
	ScanMode          string
	CurrentFolderPath *string
	LastProcessedFile *string
	StartedAt         int64
	UpdatedAt         int64
	Completed         bool
}

// InsertScanState creates a new scan checkpoint
func InsertScanState(db *sql.DB, state *ScanState) (int64, error) {
	query := `
	INSERT INTO scan_state (root_folder_id, scan_mode, current_folder_path, 
	                        last_processed_file, started_at, updated_at, completed)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	completed := 0
	if state.Completed {
		completed = 1
	}

	result, err := db.Exec(query, state.RootFolderID, state.ScanMode, state.CurrentFolderPath,
		state.LastProcessedFile, state.StartedAt, state.UpdatedAt, completed)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// UpdateScanState updates checkpoint progress
func UpdateScanState(db *sql.DB, state *ScanState) error {
	query := `
	UPDATE scan_state 
	SET current_folder_path = ?, last_processed_file = ?, updated_at = ?
	WHERE id = ?
	`
	_, err := db.Exec(query, state.CurrentFolderPath, state.LastProcessedFile,
		state.UpdatedAt, state.ID)
	return err
}

// CompleteScanState marks a scan as completed
func CompleteScanState(db *sql.DB, stateID int64) error {
	query := `UPDATE scan_state SET completed = 1, updated_at = strftime('%s', 'now') WHERE id = ?`
	_, err := db.Exec(query, stateID)
	return err
}

// GetActiveScanState retrieves active scan for root folder
func GetActiveScanState(db *sql.DB, rootFolderID int64) (*ScanState, error) {
	query := `
	SELECT id, root_folder_id, scan_mode, current_folder_path, last_processed_file,
	       started_at, updated_at, completed
	FROM scan_state
	WHERE root_folder_id = ? AND completed = 0
	ORDER BY started_at DESC
	LIMIT 1
	`

	state := &ScanState{}
	var completed int
	err := db.QueryRow(query, rootFolderID).Scan(&state.ID, &state.RootFolderID, &state.ScanMode,
		&state.CurrentFolderPath, &state.LastProcessedFile, &state.StartedAt,
		&state.UpdatedAt, &completed)
	if err != nil {
		return nil, err
	}
	state.Completed = completed != 0
	return state, nil
}

// DeleteScanState removes a checkpoint (for restart)
func DeleteScanState(db *sql.DB, stateID int64) error {
	query := `DELETE FROM scan_state WHERE id = ?`
	_, err := db.Exec(query, stateID)
	return err
}
