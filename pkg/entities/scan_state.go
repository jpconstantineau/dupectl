package entities

import "time"

// ScanStatus represents the current state of a scan operation
type ScanStatus string

const (
	ScanStatusRunning     ScanStatus = "running"
	ScanStatusCompleted   ScanStatus = "completed"
	ScanStatusInterrupted ScanStatus = "interrupted"
)

// ScanMode represents the type of scan being performed
type ScanMode string

const (
	ScanModeAll     ScanMode = "all"     // Scan folders and files
	ScanModeFolders ScanMode = "folders" // Scan folders only
	ScanModeFiles   ScanMode = "files"   // Scan files only
)

// ScanState tracks the progress and checkpoint of a scan operation
type ScanState struct {
	ID                int        `json:"id"`
	RootFolderID      int        `json:"root_folder_id"`
	ScanMode          ScanMode   `json:"scan_mode"`
	CurrentFolderPath *string    `json:"current_folder_path,omitempty"`
	StartedAt         time.Time  `json:"started_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	Status            ScanStatus `json:"status"`
	FilesProcessed    int        `json:"files_processed"`
	FoldersProcessed  int        `json:"folders_processed"`
}

// IsActive returns true if scan is currently running or interrupted (can be resumed)
func (s *ScanState) IsActive() bool {
	return s.Status == ScanStatusRunning || s.Status == ScanStatusInterrupted
}

// IsCompleted returns true if scan finished successfully
func (s *ScanState) IsCompleted() bool {
	return s.Status == ScanStatusCompleted
}

// CanResume returns true if scan can be resumed (interrupted status)
func (s *ScanState) CanResume() bool {
	return s.Status == ScanStatusInterrupted
}
