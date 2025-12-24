# Data Model: Duplicate Scan System

**Phase**: 1 (Design & Contracts)  
**Date**: December 23, 2025  
**Feature**: 001-duplicate-scan-system

## Overview

Database schema design for duplicate file/folder scanning system. Uses SQLite with WAL mode, focuses on foreign key integrity, efficient indexing for duplicate detection queries, and support for checkpoint/resume functionality.

## Entity Relationship Diagram

```text
┌─────────────────┐
│  root_folders   │ (existing table, extended)
│─────────────────│
│ id (PK)         │
│ path            │
│ agent_id (FK)   │
│ traverse_links  │◄────┐
│ last_scan_date  │     │
│ folder_count    │     │
│ file_count      │     │
│ total_size      │     │
└─────────────────┘     │
         ▲              │
         │              │
         │              │
┌─────────────────┐     │
│    folders      │     │
│─────────────────│     │
│ id (PK)         │     │
│ path (UNIQUE)   │     │
│ parent_folder_id│────►│ (self-referential)
│ root_folder_id  │─────┘
│ error_status    │
│ first_scanned_at│
│ last_scanned_at │
│ removed         │
└─────────────────┘
         ▲
         │
         │
┌─────────────────┐
│     files       │
│─────────────────│
│ id (PK)         │
│ path (UNIQUE)   │
│ size            │
│ mtime           │
│ hash_value      │
│ hash_algorithm  │
│ error_status    │
│ first_scanned_at│
│ last_scanned_at │
│ removed         │
│ folder_id (FK)  │─────┘
│ root_folder_id  │─────┐
└─────────────────┘     │
                        │
┌─────────────────┐     │
│   scan_state    │     │
│─────────────────│     │
│ id (PK)         │     │
│ root_folder_id  │─────┘
│ scan_mode       │
│ current_folder  │
│ last_file       │
│ started_at      │
│ updated_at      │
│ completed       │
└─────────────────┘
```

## Table Definitions

### files

**Purpose**: Store file metadata, hash values, and scan history for duplicate detection

**Columns**:

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PRIMARY KEY AUTOINCREMENT | Unique file identifier |
| path | TEXT | NOT NULL UNIQUE | Absolute file path (platform-specific separators) |
| size | INTEGER | NOT NULL | File size in bytes |
| mtime | INTEGER | NOT NULL | File modification timestamp (Unix epoch seconds) |
| hash_value | TEXT | NULL | Hex-encoded hash digest (NULL until hashed) |
| hash_algorithm | TEXT | NULL | Hash algorithm used: 'sha256', 'sha512', 'sha3-256' |
| error_status | TEXT | NULL | Error message if scan failed (e.g., "permission denied") |
| first_scanned_at | INTEGER | NOT NULL | UTC timestamp when file first entered database (Unix epoch) |
| last_scanned_at | INTEGER | NOT NULL | UTC timestamp of most recent scan (Unix epoch) |
| removed | INTEGER | NOT NULL DEFAULT 0 | Boolean flag: 0=present on filesystem, 1=removed |
| folder_id | INTEGER | NOT NULL | Foreign key to folders.id |
| root_folder_id | INTEGER | NOT NULL | Foreign key to root_folders.id |

**Foreign Keys**:
- `folder_id` → `folders(id)` ON DELETE CASCADE
- `root_folder_id` → `root_folders(id)` ON DELETE CASCADE

**Indexes**:
```sql
CREATE INDEX idx_files_hash ON files(hash_value, size) 
    WHERE removed = 0 AND error_status IS NULL AND hash_value IS NOT NULL;
CREATE INDEX idx_files_folder ON files(folder_id);
CREATE INDEX idx_files_root ON files(root_folder_id);
CREATE INDEX idx_files_removed ON files(removed);
CREATE INDEX idx_files_path ON files(path);
```

**Design Rationale**:
- **Composite index on (hash_value, size)**: Optimizes duplicate detection query (FR-005: both must match). Partial index excludes removed files, files with errors, and unhashed files to reduce index size.
- **Separate folder_id and root_folder_id**: Enables direct root queries without traversing folder hierarchy. Foreign key to folder provides navigation up tree.
- **Nullable hash_value**: Supports incremental scanning - register file metadata first, hash later (scan folders command followed by scan files).
- **hash_algorithm field**: Enables algorithm migration - can re-hash with different algorithm without losing original hash (FR-016, NFR-009).
- **error_status field**: Tracks permission errors and scan failures - prevents repeated failed access attempts (FR-004.2).
- **removed flag**: Tracks file lifecycle - enables detection of file movements (FR-021).

**Example Row**:
```sql
INSERT INTO files VALUES (
    1,                                    -- id
    'C:\\Users\\user\\docs\\report.pdf', -- path (Windows format)
    1048576,                              -- size (1 MB)
    1703334000,                           -- mtime (2025-12-23 10:00:00 UTC)
    'a3f5d8...',                          -- hash_value (SHA-512, truncated)
    'sha512',                             -- hash_algorithm
    NULL,                                 -- error_status (no errors)
    1703334100,                           -- first_scanned_at
    1703334100,                           -- last_scanned_at
    0,                                    -- removed (present)
    42,                                   -- folder_id
    5                                     -- root_folder_id
);
```

### folders

**Purpose**: Store folder hierarchy, scan status, and error tracking

**Columns**:

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PRIMARY KEY AUTOINCREMENT | Unique folder identifier |
| path | TEXT | NOT NULL UNIQUE | Absolute folder path (platform-specific separators) |
| parent_folder_id | INTEGER | NULL | Foreign key to folders.id (NULL for root folder) |
| root_folder_id | INTEGER | NOT NULL | Foreign key to root_folders.id |
| error_status | TEXT | NULL | Error message if scan failed (e.g., "permission denied") |
| first_scanned_at | INTEGER | NOT NULL | UTC timestamp when folder first entered database (Unix epoch) |
| last_scanned_at | INTEGER | NOT NULL | UTC timestamp of most recent scan (Unix epoch) |
| removed | INTEGER | NOT NULL DEFAULT 0 | Boolean flag: 0=present on filesystem, 1=removed |

**Foreign Keys**:
- `parent_folder_id` → `folders(id)` ON DELETE CASCADE
- `root_folder_id` → `root_folders(id)` ON DELETE CASCADE

**Indexes**:
```sql
CREATE INDEX idx_folders_parent ON folders(parent_folder_id);
CREATE INDEX idx_folders_root ON folders(root_folder_id);
CREATE INDEX idx_folders_path ON folders(path);
CREATE INDEX idx_folders_removed ON folders(removed);
```

**Design Rationale**:
- **Self-referential parent_folder_id**: Represents folder tree structure. NULL for root folder itself (top of monitored tree).
- **Path index**: Supports LIKE queries for cascading removed flag (FR-021.3: when folder removed, all subfolders removed).
- **Separate tracking from files**: Folder can be registered without files being hashed (scan folders command).
- **error_status field**: Tracks permission errors on folder access - entire folder contents skipped if denied (FR-004.3).

**Example Row**:
```sql
INSERT INTO folders VALUES (
    42,                           -- id
    'C:\\Users\\user\\docs',      -- path
    41,                           -- parent_folder_id (C:\\Users\\user)
    5,                            -- root_folder_id
    NULL,                         -- error_status (no errors)
    1703334000,                   -- first_scanned_at
    1703334000,                   -- last_scanned_at
    0                             -- removed (present)
);
```

### scan_state

**Purpose**: Store checkpoint data for scan resume capability (FR-014)

**Columns**:

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PRIMARY KEY AUTOINCREMENT | Unique checkpoint identifier |
| root_folder_id | INTEGER | NOT NULL | Foreign key to root_folders.id |
| scan_mode | TEXT | NOT NULL | Scan type: 'all', 'folders', 'files' |
| current_folder_path | TEXT | NULL | Last folder being processed (NULL if not started) |
| last_processed_file | TEXT | NULL | Last file completed (NULL for folder-only scans) |
| started_at | INTEGER | NOT NULL | UTC timestamp when scan started (Unix epoch) |
| updated_at | INTEGER | NOT NULL | UTC timestamp of last checkpoint save (Unix epoch) |
| completed | INTEGER | NOT NULL DEFAULT 0 | Boolean flag: 0=in-progress, 1=completed |

**Foreign Keys**:
- `root_folder_id` → `root_folders(id)` ON DELETE CASCADE

**Indexes**:
```sql
CREATE UNIQUE INDEX idx_scan_state_active ON scan_state(root_folder_id) 
    WHERE completed = 0;
```

**Design Rationale**:
- **Unique index on root_folder_id WHERE completed = 0**: Prevents concurrent scans of same root folder (FR-031). Multiple completed scan records allowed for history.
- **scan_mode field**: Enables correct resume logic - folders-only scan resumes differently than files-only.
- **Nullable current_folder_path and last_processed_file**: Handles edge case where scan interrupted before first folder/file processed.
- **updated_at timestamp**: Detects stale checkpoints - if very old, may indicate corruption or forceful termination (SIGKILL).

**Example Row**:
```sql
INSERT INTO scan_state VALUES (
    1,                           -- id
    5,                           -- root_folder_id
    'all',                       -- scan_mode (scanning both folders and files)
    'C:\\Users\\user\\docs',     -- current_folder_path
    'C:\\Users\\user\\docs\\report.pdf',  -- last_processed_file
    1703334000,                  -- started_at
    1703334100,                  -- updated_at (100 seconds into scan)
    0                            -- completed (still in progress)
);
```

### root_folders (Extensions)

**Purpose**: Extend existing table with scan-specific configuration and statistics

**New Columns** (added via ALTER TABLE):

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| traverse_links | INTEGER | NOT NULL DEFAULT 0 | Boolean flag: 0=skip symlinks, 1=follow symlinks (FR-001.5) |
| last_scan_date | INTEGER | NULL | UTC timestamp of most recent scan completion (Unix epoch) |
| folder_count | INTEGER | NOT NULL DEFAULT 0 | Total folders discovered in most recent scan (FR-001.7) |
| file_count | INTEGER | NOT NULL DEFAULT 0 | Total files discovered in most recent scan (FR-001.7) |
| total_size | INTEGER | NOT NULL DEFAULT 0 | Total bytes of all files in most recent scan (FR-001.7) |

**Migration SQL**:
```sql
ALTER TABLE root_folders ADD COLUMN traverse_links INTEGER NOT NULL DEFAULT 0;
ALTER TABLE root_folders ADD COLUMN last_scan_date INTEGER;
ALTER TABLE root_folders ADD COLUMN folder_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE root_folders ADD COLUMN file_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE root_folders ADD COLUMN total_size INTEGER NOT NULL DEFAULT 0;
```

**Design Rationale**:
- **traverse_links configuration**: Per-root setting allows different behavior for different monitored trees (FR-001.5, A-020: default false).
- **Summary statistics**: Cached from database for fast display in `get root` command (FR-001.7, A-021). Updated at scan completion.
- **last_scan_date**: Distinguishes never-scanned roots from recently scanned (display "Never scanned" vs timestamp).

**Example Row** (extended fields only):
```sql
-- Existing root_folders row with new columns
UPDATE root_folders SET 
    traverse_links = 0,           -- Don't follow symlinks
    last_scan_date = 1703334200,  -- Scanned 2025-12-23 10:03:20 UTC
    folder_count = 1543,          -- 1543 folders found
    file_count = 12847,           -- 12847 files found
    total_size = 5368709120       -- 5 GB total
WHERE id = 5;
```

## Queries

### Duplicate File Detection

**Query**: Find all duplicate file sets (2+ files with identical size and hash)

```sql
SELECT 
    f.hash_value,
    f.size,
    f.hash_algorithm,
    GROUP_CONCAT(f.path, '|') AS file_paths,
    COUNT(*) AS duplicate_count
FROM files f
WHERE f.removed = 0 
  AND f.error_status IS NULL
  AND f.hash_value IS NOT NULL
GROUP BY f.hash_value, f.size
HAVING COUNT(*) >= 2
ORDER BY f.size DESC, duplicate_count DESC;
```

**Optimization**: Uses idx_files_hash partial index, filters out removed/errored files, returns largest duplicates first (most storage savings).

**With --min-count Filter** (FR-017.2):
```sql
-- Same as above but with HAVING COUNT(*) >= ?
-- Parameterized: bind --min-count value to placeholder
```

### Duplicate Folder Detection (Exact Match)

**Query**: Find folders where all files have identical matches in another folder (FR-006)

**Approach**: Multi-step process (too complex for single SQL query, implemented in application logic)
1. Get all folder IDs
2. For each folder, get set of (size, hash) for all contained files
3. Compare sets between folders - exact match = duplicate folder

**SQL Component** (get files for folder):
```sql
SELECT f.size, f.hash_value
FROM files f
WHERE f.folder_id = ?
  AND f.removed = 0
  AND f.error_status IS NULL
  AND f.hash_value IS NOT NULL
ORDER BY f.size, f.hash_value;
```

### Partial Folder Duplicate Detection

**Query**: Find folders with ≥50% file overlap (FR-007, FR-007.1)

**Approach**: Application logic with similarity calculation
1. For each folder pair, get file sets (size, hash)
2. Calculate intersection size
3. Calculate similarity: (intersection / total_unique_files) * 100
4. Filter by threshold (default 50%, configurable)
5. Return matches with key differences (FR-008)

**SQL Component** (same as duplicate folder detection above)

### Cascading Removed Flag

**Query**: Mark folder and all contained files/subfolders as removed (FR-021.3)

```sql
-- Mark folder as removed
UPDATE folders 
SET removed = 1, last_scanned_at = ?
WHERE path = ? OR path LIKE ? || '%';

-- Mark all files in folder as removed
UPDATE files
SET removed = 1, last_scanned_at = ?
WHERE path LIKE ? || '%';
```

**Optimization**: Path index enables efficient LIKE queries. Single UPDATE with LIKE covers entire hierarchy.

**Example**: Removing `C:\Users\user\docs`
```sql
UPDATE folders SET removed = 1, last_scanned_at = 1703334300
WHERE path = 'C:\\Users\\user\\docs' OR path LIKE 'C:\\Users\\user\\docs\\%';

UPDATE files SET removed = 1, last_scanned_at = 1703334300
WHERE path LIKE 'C:\\Users\\user\\docs\\%';
```

### Checkpoint Save

**Query**: Persist current scan state (FR-014.1)

```sql
INSERT INTO scan_state (
    root_folder_id, scan_mode, current_folder_path, last_processed_file, 
    started_at, updated_at, completed
) VALUES (?, ?, ?, ?, ?, ?, 0)
ON CONFLICT (root_folder_id) DO UPDATE SET
    current_folder_path = excluded.current_folder_path,
    last_processed_file = excluded.last_processed_file,
    updated_at = excluded.updated_at;
```

**Note**: Uses UPSERT (INSERT ... ON CONFLICT) to handle first checkpoint vs subsequent updates.

### Checkpoint Resume

**Query**: Retrieve scan state on startup (FR-014.2)

```sql
SELECT scan_mode, current_folder_path, last_processed_file, started_at
FROM scan_state
WHERE root_folder_id = ?
  AND completed = 0
ORDER BY updated_at DESC
LIMIT 1;
```

**Note**: Most recent incomplete scan for given root.

### Root Folder Statistics

**Query**: Display summary information for `get root` command (FR-001.6)

```sql
SELECT 
    path,
    folder_count,
    file_count,
    total_size,
    datetime(last_scan_date, 'unixepoch') AS last_scan_date_utc,
    CASE 
        WHEN last_scan_date IS NULL THEN 'Never scanned'
        ELSE strftime('%Y-%m-%d %H:%M:%S UTC', last_scan_date, 'unixepoch')
    END AS last_scan_display
FROM root_folders
ORDER BY path;
```

**Optimization**: Statistics are cached, no joins required. Fast query even with thousands of roots.

### Update Root Statistics

**Query**: Refresh cached statistics after scan completion (FR-001.7)

```sql
UPDATE root_folders
SET 
    folder_count = (SELECT COUNT(*) FROM folders WHERE root_folder_id = ? AND removed = 0),
    file_count = (SELECT COUNT(*) FROM files WHERE root_folder_id = ? AND removed = 0),
    total_size = (SELECT COALESCE(SUM(size), 0) FROM files WHERE root_folder_id = ? AND removed = 0 AND error_status IS NULL),
    last_scan_date = ?
WHERE id = ?;
```

**Note**: Executed once at scan completion, not on every file.

## Migration Strategy

### Initial Migration (Version 1)

**File**: `pkg/datastore/migrations.go` - `RunMigrations()` function

**Steps**:
1. Create `files` table with indexes
2. Create `folders` table with indexes
3. Create `scan_state` table with unique index
4. Alter `root_folders` table (add 5 new columns)
5. Insert migration version record: `INSERT INTO schema_version VALUES (1, 'duplicate_scan_system', ?)`

**Idempotency**: All `CREATE TABLE IF NOT EXISTS` and check `schema_version` before running.

**Rollback**: Not supported for initial release (breaking change requires data export/import).

### Future Migrations (Version 2+)

**Schema Versioning** (NFR-009):
```sql
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at INTEGER NOT NULL
);
```

**Example Future Migration** (adding index):
```sql
-- Version 2: Add index on files.mtime for temporal queries
IF NOT EXISTS (SELECT 1 FROM schema_version WHERE version = 2) THEN
    CREATE INDEX idx_files_mtime ON files(mtime);
    INSERT INTO schema_version VALUES (2, 'add_mtime_index', unixepoch());
END IF;
```

## Data Integrity Rules

### Foreign Key Cascade Rules

**Deleting Root Folder**:
- CASCADE DELETE to `folders` (all folders in hierarchy deleted)
- CASCADE DELETE to `files` (all files in hierarchy deleted)
- CASCADE DELETE to `scan_state` (checkpoint deleted)

**Deleting Folder**:
- CASCADE DELETE to child `folders` (subfolders deleted)
- CASCADE DELETE to `files` (files in folder deleted)

**Why CASCADE**: Orphaned records serve no purpose. If root is unregistered, all scan data should be removed (FR-022).

### Constraints

**UNIQUE Constraints**:
- `files.path` - prevents duplicate file entries
- `folders.path` - prevents duplicate folder entries
- `scan_state.root_folder_id WHERE completed = 0` - prevents concurrent scans

**NOT NULL Constraints**:
- All timestamp fields (first_scanned_at, last_scanned_at, started_at, updated_at)
- All foreign keys (folder_id, root_folder_id)
- Size field (files.size)

**DEFAULT Values**:
- `removed` = 0 (default to present)
- `completed` = 0 (default to in-progress)
- `traverse_links` = 0 (default to disabled)

## Performance Characteristics

### Table Sizes (Estimated)

**For 100,000 files across 10,000 folders**:
- `files` table: ~100,000 rows × 200 bytes/row = 20 MB
- `folders` table: ~10,000 rows × 150 bytes/row = 1.5 MB
- `scan_state` table: ~10 rows × 100 bytes/row = 1 KB (minimal)
- Indexes: ~15 MB total

**Total**: ~37 MB for 100K files (well under NFR-002: <500 MB RAM)

### Query Performance

**Duplicate detection query** (with idx_files_hash):
- 100K files → <50ms (meets constitution: <50ms for common queries)
- Index scan + grouping, no table scan

**Cascade removed flag** (with idx_folders_path):
- 1000 subfolders → <100ms
- LIKE query uses path index prefix match

**Checkpoint save/resume**:
- <5ms (single-row UPSERT or SELECT)
- Meets NFR-008: <5 second shutdown (checkpoint is tiny fraction)

## Test Validation

### Schema Tests

**Test Suite**: `tests/integration/database_test.go`

**Cases**:
- [ ] Create tables with all indexes (no errors)
- [ ] Insert file with all fields → verify retrieval
- [ ] Insert folder with parent_folder_id → verify hierarchy
- [ ] Delete root folder → verify CASCADE to folders and files
- [ ] Delete folder → verify CASCADE to child folders and files
- [ ] Insert duplicate file path → verify UNIQUE constraint error
- [ ] Insert scan_state for same root twice → verify UNIQUE constraint error
- [ ] Query duplicate files with known test data → verify correct grouping

### Migration Tests

**Test Suite**: `tests/integration/migration_test.go`

**Cases**:
- [ ] Run migration on fresh database → verify all tables created
- [ ] Run migration twice → verify idempotency (no errors)
- [ ] Query schema_version table → verify version recorded

### Query Performance Tests

**Test Suite**: `tests/integration/performance_test.go`

**Cases**:
- [ ] Insert 10,000 files → measure insert time (should batch)
- [ ] Query duplicates on 10,000 files → measure query time (should use index)
- [ ] Cascade removed flag on 1,000 subfolders → measure update time

## References

- Feature Specification: `specs/001-duplicate-scan-system/spec.md`
- Research Document: `specs/001-duplicate-scan-system-01/research.md`
- DupeCTL Constitution: `.specify/memory/constitution.md`
- SQLite WAL Mode: https://www.sqlite.org/wal.html
- SQLite Foreign Keys: https://www.sqlite.org/foreignkeys.html
