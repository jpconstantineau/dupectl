package entities

// ScanFile represents a file entity for scanning
type ScanFile struct {
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
}

// ScanFolder represents a folder entity for scanning
type ScanFolder struct {
	ID             int64
	Path           string
	ParentFolderID *int64
	RootFolderID   int64
	ErrorStatus    *string
	FirstScannedAt int64
	LastScannedAt  int64
	Removed        bool
}

// ScanState represents a scan checkpoint
type ScanState struct {
	ID                int64
	RootFolderID      int64
	ScanMode          string // "all", "folders", "files"
	CurrentFolderPath *string
	LastProcessedFile *string
	StartedAt         int64
	UpdatedAt         int64
	Completed         bool
}
