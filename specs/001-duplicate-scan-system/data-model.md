# Data Model: Duplicate Scan System

**Feature**: 001-duplicate-scan-system  
**Date**: December 23, 2025  
**Purpose**: Database schema and entity definitions

## Entity Relationships

```
┌─────────────┐
│ RootFolder  │ (existing infrastructure table)
│ id          │
│ name        │
│ host_id     │
│ owner_id    │
│ agent_id    │
└─────┬───────┘
      │ 1
      │
      │ N
┌─────▼───────┐
│ Folder      │ (NEW)
│ id          │
│ root_folder │─────┐
│ parent_id   │◄────┘ (self-reference for hierarchy)
│ path        │
│ name        │
│ scanned_at  │
└─────┬───────┘
      │ 1
      │
      │ N
┌─────▼───────┐
│ File        │ (NEW)
│ id          │
│ root_folder │
│ folder_id   │
│ path        │
│ name        │
│ size        │
│ mtime       │
│ hash_value  │
│ hash_algo   │
│ error_status│
│ scanned_at  │
└─────────────┘

┌─────────────┐
│ ScanState   │ (NEW - checkpoint tracking)
│ id          │
│ root_folder │
│ scan_mode   │
│ current_path│
│ started_at  │
│ updated_at  │
│ completed_at│
│ status      │
│ files_count │
│ folders_count│
└─────────────┘
```

## Database Schema

### Files Table

**Purpose**: Store file metadata and hash values for duplicate detection

```sql
CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    root_folder_id INTEGER NOT NULL,
    folder_id INTEGER,                    -- NULL if folder not tracked separately
    path TEXT NOT NULL UNIQUE,            -- Absolute path (unique constraint prevents duplicates)
    name TEXT NOT NULL,                   -- Filename only (for queries and display)
    size INTEGER NOT NULL,                -- File size in bytes
    mtime INTEGER NOT NULL,               -- Modification time (Unix timestamp)
    hash_value TEXT,                      -- NULL until hashed, hex-encoded hash
    hash_algorithm TEXT,                  -- 'sha256', 'sha512', or 'sha3-256'
    error_status TEXT,                    -- NULL if no error, error message if access denied
    scanned_at INTEGER,                   -- Unix timestamp of last successful scan
    
    FOREIGN KEY (root_folder_id) REFERENCES root_folders(id) ON DELETE CASCADE,
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE SET NULL
);

-- Performance indexes
CREATE INDEX idx_files_hash_size ON files(hash_value, size);  -- Duplicate detection query
CREATE INDEX idx_files_root ON files(root_folder_id);         -- Filter by root
CREATE INDEX idx_files_folder ON files(folder_id);            -- Folder contents query
CREATE INDEX idx_files_error ON files(error_status) WHERE error_status IS NOT NULL;  -- Error tracking
```

**Fields**:
- `id`: Auto-incrementing primary key
- `root_folder_id`: Links to registered root folder (infrastructure table)
- `folder_id`: Links to parent folder (NULL if folder tracking not used)
- `path`: Absolute filesystem path, platform-specific separators, UNIQUE prevents duplicate entries
- `name`: Filename extracted from path for display and queries
- `size`: File size in bytes - used for pre-filtering duplicates before hash comparison
- `mtime`: Modification timestamp - used for partial duplicate analysis (name match, date differ)
- `hash_value`: Cryptographic hash as hex string - NULL until file is hashed
- `hash_algorithm`: Algorithm used ('sha256', 'sha512', 'sha3-256') - enables future migrations
- `error_status`: NULL on success, error message string if file unreadable (permission denied, etc.)
- `scanned_at`: Last successful scan timestamp - enables incremental re-scanning

**Constraints**:
- UNIQUE(path) - One record per file path prevents duplicates
- NOT NULL for root_folder_id, path, name, size, mtime - Required fields
- Foreign keys use ON DELETE CASCADE/SET NULL for referential integrity

---

### Folders Table

**Purpose**: Track folder hierarchy for duplicate folder detection and scan organization

```sql
CREATE TABLE IF NOT EXISTS folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    root_folder_id INTEGER NOT NULL,
    parent_folder_id INTEGER,              -- NULL for root folder itself
    path TEXT NOT NULL UNIQUE,             -- Absolute path (unique constraint)
    name TEXT NOT NULL,                    -- Folder name only
    scanned_at INTEGER,                    -- Unix timestamp of last scan
    
    FOREIGN KEY (root_folder_id) REFERENCES root_folders(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_folder_id) REFERENCES folders(id) ON DELETE CASCADE
);

-- Performance indexes
CREATE INDEX idx_folders_root ON folders(root_folder_id);      -- Filter by root
CREATE INDEX idx_folders_parent ON folders(parent_folder_id);  -- Child folder queries
CREATE INDEX idx_folders_path ON folders(path);                -- Path lookups for checkpoint
```

**Fields**:
- `id`: Auto-incrementing primary key
- `root_folder_id`: Links to registered root folder
- `parent_folder_id`: Self-referencing FK for hierarchy - NULL for root folder
- `path`: Absolute filesystem path with UNIQUE constraint
- `name`: Folder name extracted from path
- `scanned_at`: Last scan timestamp

**Constraints**:
- UNIQUE(path) - One record per folder path
- Self-referencing FK enables tree structure queries
- ON DELETE CASCADE - Deleting root cascades to all folders

---

### Scan State Table

**Purpose**: Checkpoint tracking for scan interruption and resumption

```sql
CREATE TABLE IF NOT EXISTS scan_state (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    root_folder_id INTEGER NOT NULL,
    scan_mode TEXT NOT NULL,               -- 'all', 'folders', 'files'
    current_folder_path TEXT,              -- Last folder being processed (checkpoint)
    started_at INTEGER NOT NULL,           -- Scan start timestamp
    updated_at INTEGER NOT NULL,           -- Last checkpoint update
    completed_at INTEGER,                  -- NULL if incomplete, timestamp when done
    status TEXT NOT NULL,                  -- 'running', 'completed', 'interrupted'
    files_processed INTEGER DEFAULT 0,
    folders_processed INTEGER DEFAULT 0,
    
    FOREIGN KEY (root_folder_id) REFERENCES root_folders(id) ON DELETE CASCADE
);

-- Performance indexes
CREATE INDEX idx_scan_state_root_status ON scan_state(root_folder_id, status);  -- Active scan lookup
CREATE INDEX idx_scan_state_started ON scan_state(started_at);                  -- Recent scans query
```

**Fields**:
- `id`: Auto-incrementing primary key
- `root_folder_id`: Which root folder is being scanned
- `scan_mode`: 'all', 'folders', or 'files' - determines what to process on resume
- `current_folder_path`: Last folder checkpoint - resume starts here
- `started_at`: When scan began
- `updated_at`: Last checkpoint save time
- `completed_at`: NULL while running, timestamp when finished
- `status`: 'running' (active), 'completed' (done), 'interrupted' (crashed/cancelled)
- `files_processed`: Counter for progress reporting
- `folders_processed`: Counter for progress reporting

**Status Transitions**:
- NULL → 'running': Scan starts
- 'running' → 'completed': Scan finishes normally
- 'running' → 'interrupted': Process killed/crashed
- 'interrupted' → 'running': Resume from checkpoint

---

## Go Entity Definitions

### File Entity

```go
// pkg/entities/file.go
package entities

import "time"

// File represents a file in the filesystem with hash and metadata
type File struct {
    ID           int       `json:"id"`
    RootFolderID int       `json:"root_folder_id"`
    FolderID     *int      `json:"folder_id,omitempty"` // Pointer for NULL
    Path         string    `json:"path"`
    Name         string    `json:"name"`
    Size         int64     `json:"size"`
    Mtime        time.Time `json:"mtime"`
    HashValue    *string   `json:"hash_value,omitempty"` // Pointer for NULL
    HashAlgorithm *string  `json:"hash_algorithm,omitempty"`
    ErrorStatus  *string   `json:"error_status,omitempty"`
    ScannedAt    *time.Time `json:"scanned_at,omitempty"`
}

// HasError returns true if file has an error status (permission denied, etc.)
func (f *File) HasError() bool {
    return f.ErrorStatus != nil && *f.ErrorStatus != ""
}

// IsHashed returns true if file has been hashed
func (f *File) IsHashed() bool {
    return f.HashValue != nil && *f.HashValue != ""
}
```

### Folder Entity

```go
// pkg/entities/folder.go
package entities

import "time"

// Folder represents a directory in the filesystem hierarchy
type Folder struct {
    ID             int       `json:"id"`
    RootFolderID   int       `json:"root_folder_id"`
    ParentFolderID *int      `json:"parent_folder_id,omitempty"` // Pointer for NULL (root)
    Path           string    `json:"path"`
    Name           string    `json:"name"`
    ScannedAt      *time.Time `json:"scanned_at,omitempty"`
}

// IsRoot returns true if this is the root folder (no parent)
func (f *Folder) IsRoot() bool {
    return f.ParentFolderID == nil
}
```

### ScanState Entity

```go
// pkg/entities/scan_state.go
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
    ScanModeAll     ScanMode = "all"      // Scan folders and files
    ScanModeFolders ScanMode = "folders"  // Scan folders only
    ScanModeFiles   ScanMode = "files"    // Scan files only
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
```

### Duplicate File Set

```go
// pkg/entities/duplicate.go
package entities

// DuplicateFileSet represents a group of files with identical size and hash
type DuplicateFileSet struct {
    Size          int64    `json:"size"`
    HashValue     string   `json:"hash_value"`
    HashAlgorithm string   `json:"hash_algorithm"`
    Files         []File   `json:"files"`
    FileCount     int      `json:"file_count"`
}

// DuplicateFolderSet represents folders with identical content
type DuplicateFolderSet struct {
    Folders      []Folder `json:"folders"`
    FolderCount  int      `json:"folder_count"`
    TotalFiles   int      `json:"total_files"`
}

// PartialDuplicate represents two folders with partial content overlap
type PartialDuplicate struct {
    Folder1          Folder   `json:"folder1"`
    Folder2          Folder   `json:"folder2"`
    Similarity       float64  `json:"similarity"`        // 0.0 to 100.0
    MatchingFiles    []string `json:"matching_files"`
    MissingInFolder1 []string `json:"missing_in_folder1"`
    MissingInFolder2 []string `json:"missing_in_folder2"`
    NameMismatches   []NameMismatch `json:"name_mismatches"`
}

// NameMismatch represents files with same name but different mtime
type NameMismatch struct {
    Name     string    `json:"name"`
    Path1    string    `json:"path1"`
    Path2    string    `json:"path2"`
    Mtime1   time.Time `json:"mtime1"`
    Mtime2   time.Time `json:"mtime2"`
}
```

---

## Database Migrations

### Migration 001: Create Scan Tables

```go
// pkg/datastore/migrations/001_create_scan_tables.go
package migrations

const Migration001CreateScanTables = `
-- Files table
CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    root_folder_id INTEGER NOT NULL,
    folder_id INTEGER,
    path TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    size INTEGER NOT NULL,
    mtime INTEGER NOT NULL,
    hash_value TEXT,
    hash_algorithm TEXT,
    error_status TEXT,
    scanned_at INTEGER,
    FOREIGN KEY (root_folder_id) REFERENCES root_folders(id) ON DELETE CASCADE,
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE SET NULL
);

CREATE INDEX idx_files_hash_size ON files(hash_value, size);
CREATE INDEX idx_files_root ON files(root_folder_id);
CREATE INDEX idx_files_folder ON files(folder_id);
CREATE INDEX idx_files_error ON files(error_status) WHERE error_status IS NOT NULL;

-- Folders table
CREATE TABLE IF NOT EXISTS folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    root_folder_id INTEGER NOT NULL,
    parent_folder_id INTEGER,
    path TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    scanned_at INTEGER,
    FOREIGN KEY (root_folder_id) REFERENCES root_folders(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_folder_id) REFERENCES folders(id) ON DELETE CASCADE
);

CREATE INDEX idx_folders_root ON folders(root_folder_id);
CREATE INDEX idx_folders_parent ON folders(parent_folder_id);
CREATE INDEX idx_folders_path ON folders(path);

-- Scan state table
CREATE TABLE IF NOT EXISTS scan_state (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    root_folder_id INTEGER NOT NULL,
    scan_mode TEXT NOT NULL,
    current_folder_path TEXT,
    started_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    completed_at INTEGER,
    status TEXT NOT NULL,
    files_processed INTEGER DEFAULT 0,
    folders_processed INTEGER DEFAULT 0,
    FOREIGN KEY (root_folder_id) REFERENCES root_folders(id) ON DELETE CASCADE
);

CREATE INDEX idx_scan_state_root_status ON scan_state(root_folder_id, status);
CREATE INDEX idx_scan_state_started ON scan_state(started_at);
`
```

---

## Query Examples

### Find Duplicate Files

```sql
-- Step 1: Find sizes with multiple files
WITH duplicate_sizes AS (
    SELECT size
    FROM files
    WHERE hash_value IS NOT NULL
    GROUP BY size
    HAVING COUNT(*) > 1
)

-- Step 2: Group files by size and hash
SELECT 
    f.size,
    f.hash_value,
    f.hash_algorithm,
    COUNT(*) as file_count,
    GROUP_CONCAT(f.path, '|') as file_paths
FROM files f
INNER JOIN duplicate_sizes ds ON f.size = ds.size
WHERE f.hash_value IS NOT NULL
GROUP BY f.size, f.hash_value, f.hash_algorithm
HAVING COUNT(*) > 1
ORDER BY f.size DESC;
```

### Find Files with Errors

```sql
SELECT 
    path,
    error_status,
    scanned_at
FROM files
WHERE error_status IS NOT NULL
ORDER BY scanned_at DESC;
```

### Get Active Scan State

```sql
SELECT *
FROM scan_state
WHERE root_folder_id = ?
AND status IN ('running', 'interrupted')
ORDER BY started_at DESC
LIMIT 1;
```

### Get Folder Contents

```sql
SELECT 
    id,
    path,
    name,
    size,
    hash_value
FROM files
WHERE folder_id = ?
ORDER BY name;
```

---

## Data Integrity Rules

1. **Path Uniqueness**: UNIQUE constraint on path fields prevents duplicate entries for same file/folder
2. **Referential Integrity**: Foreign keys ensure files/folders link to valid roots, cascading deletes clean up orphans
3. **NULL Semantics**: 
   - hash_value NULL = not yet hashed
   - error_status NULL = no error
   - completed_at NULL = scan in progress
4. **Index Strategy**: Composite index on (hash_value, size) optimizes duplicate detection O(log n)
5. **Transactional Writes**: All file/folder inserts use transactions (batch size 1000) for consistency
6. **Checkpoint Atomicity**: scan_state updates are atomic - checkpoint saves don't corrupt on crash

---

## Storage Estimates

**Assumptions**:
- Average path length: 100 characters
- Average filename: 20 characters
- Hash value: 64 characters (SHA-256 hex)

**Per File Record**: ~250 bytes
**Per Folder Record**: ~150 bytes

**Example: 100,000 files in 10,000 folders**:
- Files table: 100k × 250 bytes = 25 MB
- Folders table: 10k × 150 bytes = 1.5 MB
- Indexes: ~10 MB
- **Total: ~37 MB**

SQLite with WAL mode handles this efficiently with <50 MB memory footprint.
