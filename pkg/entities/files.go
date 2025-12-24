package entities

import (
	"time"
)

type Host struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Owner struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Policy struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Purpose struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Agent struct {
	Id      int        `json:"id"`
	Name    string     `json:"name"`
	Guid    string     `json:"uid"`
	Enabled bool       `json:"enabled"`
	Updated time.Time  `json:"updated"`
	Status  StatusName `json:"status"`
}

type RootFolder struct {
	Id      int        `json:"id"`
	Name    string     `json:"name"`
	Host    Host       `json:"host"`
	Owner   Owner      `json:"owner"`
	Agent   Agent      `json:"agent"`
	Purpose Purpose    `json:"purpose"`
	Status  StatusName `json:"status"`
}

type Folder struct {
	Id      int        `json:"id"`
	Name    string     `json:"name"`
	Owner   Owner      `json:"owner"`
	Agent   Agent      `json:"agent"`
	Purpose Purpose    `json:"purpose"`
	Status  StatusName `json:"status"`
}

type foldermsg struct {
	Host     string    `json:"host"`
	Fullname string    `json:"fullname"`
	Path     string    `json:"path"`
	Name     string    `json:"name"`
	Atime    time.Time `json:"atime"`
	Mtime    time.Time `json:"mtime"`
	Ctime    time.Time `json:"ctime"`
	Btime    time.Time `json:"btime"`
}

type filemsg struct {
	Host     string    `json:"host"`
	Fullname string    `json:"fullname"`
	Path     string    `json:"path"`
	Name     string    `json:"name"`
	Ext      string    `json:"ext"`
	Size     int64     `json:"size"`
	Atime    time.Time `json:"atime"`
	Mtime    time.Time `json:"mtime"`
	Ctime    time.Time `json:"ctime"`
	Btime    time.Time `json:"btime"`
	Hash     string    `json:"hash"`
}

// File represents a file in the filesystem with hash and metadata for duplicate detection
type File struct {
	ID            int        `json:"id"`
	RootFolderID  int        `json:"root_folder_id"`
	FolderID      *int       `json:"folder_id,omitempty"` // Pointer for NULL
	Path          string     `json:"path"`
	Name          string     `json:"name"`
	Size          int64      `json:"size"`
	Mtime         time.Time  `json:"mtime"`
	HashValue     *string    `json:"hash_value,omitempty"` // Pointer for NULL
	HashAlgorithm *string    `json:"hash_algorithm,omitempty"`
	ErrorStatus   *string    `json:"error_status,omitempty"`
	ScannedAt     *time.Time `json:"scanned_at,omitempty"`
}

// HasError returns true if file has an error status (permission denied, etc.)
func (f *File) HasError() bool {
	return f.ErrorStatus != nil && *f.ErrorStatus != ""
}

// IsHashed returns true if file has been hashed
func (f *File) IsHashed() bool {
	return f.HashValue != nil && *f.HashValue != ""
}

// ScanFolder represents a directory in the filesystem hierarchy for duplicate detection
type ScanFolder struct {
	ID             int        `json:"id"`
	RootFolderID   int        `json:"root_folder_id"`
	ParentFolderID *int       `json:"parent_folder_id,omitempty"` // Pointer for NULL (root)
	Path           string     `json:"path"`
	Name           string     `json:"name"`
	ScannedAt      *time.Time `json:"scanned_at,omitempty"`
}

// IsRoot returns true if this is the root folder (no parent)
func (sf *ScanFolder) IsRoot() bool {
	return sf.ParentFolderID == nil
}

// DuplicateFileSet represents a group of files with identical size and hash
type DuplicateFileSet struct {
	Size          int64  `json:"size"`
	HashValue     string `json:"hash_value"`
	HashAlgorithm string `json:"hash_algorithm"`
	Files         []File `json:"files"`
	FileCount     int    `json:"file_count"`
}

// DuplicateFolderSet represents folders with identical content
type DuplicateFolderSet struct {
	Folders     []ScanFolder `json:"folders"`
	FolderCount int          `json:"folder_count"`
	TotalFiles  int          `json:"total_files"`
}

// PartialDuplicate represents two folders with partial content overlap
type PartialDuplicate struct {
	Folder1          ScanFolder     `json:"folder1"`
	Folder2          ScanFolder     `json:"folder2"`
	Similarity       float64        `json:"similarity"` // 0.0 to 100.0
	MatchingFiles    []string       `json:"matching_files"`
	MissingInFolder1 []string       `json:"missing_in_folder1"`
	MissingInFolder2 []string       `json:"missing_in_folder2"`
	NameMismatches   []NameMismatch `json:"name_mismatches"`
}

// NameMismatch represents files with same name but different mtime
type NameMismatch struct {
	Name   string    `json:"name"`
	Path1  string    `json:"path1"`
	Path2  string    `json:"path2"`
	Mtime1 time.Time `json:"mtime1"`
	Mtime2 time.Time `json:"mtime2"`
}
