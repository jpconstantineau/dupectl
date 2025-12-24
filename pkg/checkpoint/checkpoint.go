package checkpoint

import (
	"database/sql"
	"time"

	"github.com/jpconstantineau/dupectl/pkg/datastore"
	"github.com/jpconstantineau/dupectl/pkg/logger"
)

// Manager handles scan checkpoint operations
type Manager struct {
	db           *sql.DB
	rootFolderID int64
	scanMode     string
	stateID      *int64
}

// NewManager creates a checkpoint manager
func NewManager(db *sql.DB, rootFolderID int64, scanMode string) *Manager {
	return &Manager{
		db:           db,
		rootFolderID: rootFolderID,
		scanMode:     scanMode,
	}
}

// Start creates a new scan checkpoint
func (m *Manager) Start() error {
	now := time.Now().Unix()
	state := &datastore.ScanState{
		RootFolderID: m.rootFolderID,
		ScanMode:     m.scanMode,
		StartedAt:    now,
		UpdatedAt:    now,
		Completed:    false,
	}

	id, err := datastore.InsertScanState(m.db, state)
	if err != nil {
		return err
	}

	m.stateID = &id
	logger.Info("Checkpoint started for root folder %d, scan mode: %s", m.rootFolderID, m.scanMode)
	return nil
}

// Save updates checkpoint with current progress
func (m *Manager) Save(currentFolder, lastFile *string) error {
	if m.stateID == nil {
		return nil // No checkpoint active
	}

	now := time.Now().Unix()
	state := &datastore.ScanState{
		ID:                *m.stateID,
		CurrentFolderPath: currentFolder,
		LastProcessedFile: lastFile,
		UpdatedAt:         now,
	}

	err := datastore.UpdateScanState(m.db, state)
	if err != nil {
		logger.Error("Failed to save checkpoint: %v", err)
		return err
	}

	logger.Debug("Checkpoint saved: folder=%v, file=%v", currentFolder, lastFile)
	return nil
}

// Complete marks the scan as completed
func (m *Manager) Complete() error {
	if m.stateID == nil {
		return nil
	}

	err := datastore.CompleteScanState(m.db, *m.stateID)
	if err != nil {
		logger.Error("Failed to complete checkpoint: %v", err)
		return err
	}

	logger.Info("Scan completed successfully")
	return nil
}

// Resume retrieves existing checkpoint
func (m *Manager) Resume() (*datastore.ScanState, error) {
	state, err := datastore.GetActiveScanState(m.db, m.rootFolderID)
	if err == sql.ErrNoRows {
		return nil, nil // No active checkpoint
	}
	if err != nil {
		return nil, err
	}

	m.stateID = &state.ID
	logger.Info("Resuming scan from checkpoint: folder=%v", state.CurrentFolderPath)
	return state, nil
}

// Clear removes checkpoint (for restart)
func (m *Manager) Clear() error {
	if m.stateID != nil {
		err := datastore.DeleteScanState(m.db, *m.stateID)
		if err != nil {
			return err
		}
		m.stateID = nil
	}

	// Also clear any active checkpoints for this root
	state, err := datastore.GetActiveScanState(m.db, m.rootFolderID)
	if err == nil && state != nil {
		return datastore.DeleteScanState(m.db, state.ID)
	}
	if err != sql.ErrNoRows {
		return err
	}

	logger.Info("Checkpoint cleared")
	return nil
}
