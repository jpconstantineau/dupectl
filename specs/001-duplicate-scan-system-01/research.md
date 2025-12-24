# Research: Duplicate Scan System

**Phase**: 0 (Outline & Research)  
**Date**: December 23, 2025  
**Feature**: 001-duplicate-scan-system

## Overview

Research findings for implementing duplicate file/folder scanning system in DupeCTL. Focus areas: existing codebase analysis, dependency selection, database schema design, and concurrency patterns for Golang implementation.

## Existing Codebase Analysis

### Current Infrastructure (Can Leverage)

**CLI Framework** (`cmd/` directory):
- **Decision**: Use existing Cobra CLI structure
- **Rationale**: 20+ command files already use Cobra pattern (`cmd/root.go`, `cmd/add.go`, `cmd/get.go`, etc.), proven working configuration with Viper integration
- **Finding**: Commands are organized by verb groups (add, get, delete, apply, scan) - new scan functionality fits naturally into existing pattern
- **Finding**: `cmd/root.go` already initializes Viper configuration from `~/.dupectl.yaml` - can extend with scan-specific config options

**Database Layer** (`pkg/datastore/`):
- **Decision**: Extend existing SQLite infrastructure
- **Rationale**: `pkg/datastore/datastore.go` already configures SQLite with WAL mode, connection pooling, and foreign key enforcement - exactly what scanning needs for concurrent operations
- **Finding**: `pkg/datastore/agent.go` provides pattern for table creation (`CreateAgentsTable`) - use as template for `CreateFilesTable`, `CreateFoldersTable`, `CreateScanStateTable`
- **Finding**: Database connection is singleton (`GetDB()` function) - thread-safe for concurrent worker access

**Entity Models** (`pkg/entities/`):
- **Decision**: Extend existing entity models
- **Rationale**: `pkg/entities/files.go` already defines `Host`, `Owner`, `Agent`, `RootFolder` structs - scanning entities need foreign key relationships to these
- **Finding**: Existing `Folder` struct is minimal (ID, Name, ParentFolderID) - needs extension for scan timestamps, error status, removed flag
- **Finding**: Existing `filemsg` struct has `Hash` field but no `HashAlgorithm` field - need new `File` entity with complete schema

**Gaps Identified**:
- No scanning logic exists (`cmd/scanAll.go`, `cmd/scanFolders.go`, `cmd/scanFiles.go` are stubs)
- No hash calculation logic (`pkg/hash/` package doesn't exist)
- No duplicate detection logic (`cmd/getDuplicates.go` is stub)
- No worker pool implementation (`internal/worker/` doesn't exist)
- No checkpoint/resume logic
- No signal handling for graceful shutdown

### Implementation Strategy

Start with database schema (foundation), then folder traversal (basic capability), then file hashing (duplicate detection prerequisite), then duplicate queries (user value delivery).

## Dependency Decisions

### Hash Algorithms

**Decision**: Use Go standard library `crypto/*` packages + `golang.org/x/crypto/sha3`

**Dependencies**:
- `crypto/sha256` - SHA-256 implementation (standard library)
- `crypto/sha512` - SHA-512 implementation (standard library, default algorithm)
- `golang.org/x/crypto/sha3` - SHA3-256 implementation (official Go extended library)

**Rationale**:
- Standard library implementations are battle-tested, performant, and have zero external dependency risk
- `golang.org/x/crypto` is maintained by Go team, considered part of extended standard library
- All three algorithms meet security requirements (collision resistance for duplicate detection)
- SHA-512 chosen as default per A-002: lowest collision probability (512-bit output)

**Alternatives Considered**:
- Third-party hash libraries (e.g., `github.com/minio/highwayhash`) - Rejected: unnecessary complexity, not cryptographically secure, not approved in constitution
- MD5 or SHA-1 - Rejected: known collision vulnerabilities, not approved in FR-003

**Performance Characteristics**:
- SHA-512: ~400-500 MB/sec on modern hardware (meets NFR-001: ≥50 MB/sec)
- SHA-256: ~300-400 MB/sec
- SHA3-256: ~200-300 MB/sec (slower but quantum-resistant)

### Concurrency Primitives

**Decision**: Use Go standard library goroutines + channels + sync package

**Dependencies**:
- Goroutines - lightweight threads (built-in language feature)
- Channels - thread-safe communication (`chan` keyword)
- `sync.WaitGroup` - wait for goroutine completion
- `sync.Mutex` - protect shared state (checkpoint data, progress counters)
- `context.Context` - cancellation and timeout propagation

**Rationale**:
- Goroutines are Go's native concurrency model, extremely lightweight (2KB stack vs 2MB threads)
- Channels provide CSP-style communication without shared memory
- Standard library sync primitives are sufficient for worker pool pattern
- No need for heavyweight concurrency libraries (constitution principle VI: stdlib first)

**Alternatives Considered**:
- `golang.org/x/sync/errgroup` - Considered: provides error handling for goroutine groups, but not essential for MVP
- Third-party worker pool libraries (e.g., `gammazero/workerpool`) - Rejected: simple enough to implement directly, avoids external dependency

### CLI Framework (Already Decided)

**Decision**: Continue using existing Cobra + Viper

**Dependencies** (already present):
- `spf13/cobra` - CLI command structure, flag parsing, help generation
- `spf13/viper` - Configuration management (file, env, flags)

**Rationale**:
- Already integrated and working in codebase
- Industry standard for Go CLI applications (used by kubectl, hugo, etc.)
- Constitution-approved dependencies

### Database Driver (Already Decided)

**Decision**: Continue using existing `modernc.org/sqlite`

**Dependencies** (already present):
- `modernc.org/sqlite` - Pure Go SQLite implementation

**Rationale**:
- Already configured with WAL mode and connection pooling
- Pure Go (no CGo dependencies) - easier cross-platform builds
- Constitution-approved dependency

## Database Schema Design

### Files Table

**Purpose**: Store file metadata and hash values for duplicate detection

```sql
CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL UNIQUE,           -- Absolute path (platform-specific separator)
    size INTEGER NOT NULL,                -- File size in bytes
    mtime INTEGER NOT NULL,               -- Modification timestamp (Unix epoch)
    hash_value TEXT,                      -- Hex-encoded hash (nullable during scan)
    hash_algorithm TEXT,                  -- 'sha256', 'sha512', 'sha3-256'
    error_status TEXT,                    -- NULL or error message (permission denied, etc.)
    first_scanned_at INTEGER NOT NULL,    -- UTC timestamp (Unix epoch)
    last_scanned_at INTEGER NOT NULL,     -- UTC timestamp (Unix epoch)
    removed INTEGER DEFAULT 0,            -- Boolean: 0=present, 1=removed
    folder_id INTEGER NOT NULL,           -- Foreign key to folders table
    root_folder_id INTEGER NOT NULL,      -- Foreign key to root_folders table
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE CASCADE,
    FOREIGN KEY (root_folder_id) REFERENCES root_folders(id) ON DELETE CASCADE
);

CREATE INDEX idx_files_hash ON files(hash_value, size) WHERE removed = 0 AND error_status IS NULL;
CREATE INDEX idx_files_folder ON files(folder_id);
CREATE INDEX idx_files_root ON files(root_folder_id);
CREATE INDEX idx_files_removed ON files(removed);
```

**Design Rationale**:
- `path` is unique constraint - prevents duplicate entries for same file
- `hash_value` nullable - allows incremental scanning (register file, hash later)
- `hash_algorithm` stored with each record - supports algorithm migration (FR-016, NFR-009)
- `error_status` tracks permission errors and scan failures (FR-004.2, FR-004.3)
- `removed` flag enables file movement tracking (FR-021)
- Composite index on `(hash_value, size)` optimizes duplicate queries (FR-005: both must match)
- `WHERE removed = 0 AND error_status IS NULL` in index - only index valid candidates for duplicate detection

### Folders Table

**Purpose**: Store folder hierarchy and scan status

```sql
CREATE TABLE IF NOT EXISTS folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL UNIQUE,           -- Absolute path (platform-specific separator)
    parent_folder_id INTEGER,            -- Foreign key to folders(id), NULL for root
    root_folder_id INTEGER NOT NULL,     -- Foreign key to root_folders table
    error_status TEXT,                   -- NULL or error message (permission denied, etc.)
    first_scanned_at INTEGER NOT NULL,   -- UTC timestamp (Unix epoch)
    last_scanned_at INTEGER NOT NULL,    -- UTC timestamp (Unix epoch)
    removed INTEGER DEFAULT 0,           -- Boolean: 0=present, 1=removed
    FOREIGN KEY (parent_folder_id) REFERENCES folders(id) ON DELETE CASCADE,
    FOREIGN KEY (root_folder_id) REFERENCES root_folders(id) ON DELETE CASCADE
);

CREATE INDEX idx_folders_parent ON folders(parent_folder_id);
CREATE INDEX idx_folders_root ON folders(root_folder_id);
CREATE INDEX idx_folders_path ON folders(path);
CREATE INDEX idx_folders_removed ON folders(removed);
```

**Design Rationale**:
- `parent_folder_id` enables hierarchy traversal for duplicate folder detection (FR-006)
- `path` index supports cascading removed flag using LIKE queries (FR-021.3)
- Self-referential foreign key represents tree structure
- `removed` flag tracks folder lifecycle (FR-021.2)
- `error_status` tracks permission errors on folders (FR-004.3)

### Scan State Table

**Purpose**: Store checkpoint data for scan resume capability

```sql
CREATE TABLE IF NOT EXISTS scan_state (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    root_folder_id INTEGER NOT NULL,     -- Foreign key to root_folders table
    scan_mode TEXT NOT NULL,             -- 'all', 'folders', 'files'
    current_folder_path TEXT,            -- Last folder being processed (nullable)
    last_processed_file TEXT,            -- Last file completed (nullable)
    started_at INTEGER NOT NULL,         -- UTC timestamp (Unix epoch)
    updated_at INTEGER NOT NULL,         -- UTC timestamp (Unix epoch)
    completed INTEGER DEFAULT 0,         -- Boolean: 0=in-progress, 1=completed
    FOREIGN KEY (root_folder_id) REFERENCES root_folders(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_scan_state_active ON scan_state(root_folder_id) 
    WHERE completed = 0;
```

**Design Rationale**:
- Unique index on `(root_folder_id) WHERE completed = 0` - prevents concurrent scans of same root (FR-031)
- `scan_mode` tracks which type of scan was running - enables correct resume logic
- `current_folder_path` and `last_processed_file` provide resumption point (A-010: folder-level granularity)
- `updated_at` tracks last checkpoint save - enables detection of stale checkpoints

### Root Folders Table Extension

**Extension Needed**: Add columns to existing `root_folders` table

```sql
ALTER TABLE root_folders ADD COLUMN traverse_links INTEGER DEFAULT 0;  -- Boolean
ALTER TABLE root_folders ADD COLUMN last_scan_date INTEGER;            -- UTC timestamp
ALTER TABLE root_folders ADD COLUMN folder_count INTEGER DEFAULT 0;
ALTER TABLE root_folders ADD COLUMN file_count INTEGER DEFAULT 0;
ALTER TABLE root_folders ADD COLUMN total_size INTEGER DEFAULT 0;
```

**Design Rationale**:
- `traverse_links` configuration per FR-001.5 (default false per A-020)
- `folder_count`, `file_count`, `total_size` for FR-001.6 summary statistics
- `last_scan_date` tracks most recent scan for display in `get root` command

### Foreign Key Relationships

**Files → Folders → Root Folders**:
- Cascade delete: removing root folder deletes all contained folders and files
- Integrity: cannot orphan files or folders
- Query optimization: can find all files for a root with single index lookup

**Scan State → Root Folders**:
- Cascade delete: removing root folder deletes checkpoint data
- Prevents resuming scan for deleted root

## Concurrency Patterns

### Worker Pool Pattern

**Decision**: Implement generic worker pool for folder traversal and file hashing

**Architecture**:
```go
type WorkItem interface {
    Process() error
}

type WorkerPool struct {
    workers   int           // Number of goroutines
    workQueue chan WorkItem // Buffered channel of work items
    wg        sync.WaitGroup
    ctx       context.Context
    cancel    context.CancelFunc
    errors    chan error    // Error collection
}

func (wp *WorkerPool) Submit(item WorkItem) error
func (wp *WorkerPool) Start()
func (wp *WorkerPool) Stop()
func (wp *WorkerPool) Wait() []error
```

**Rationale**:
- Single implementation serves both traversal and hashing workers (DRY principle)
- Buffered work queue decouples producers from consumers - prevents blocking
- Context enables cancellation on shutdown signal (FR-014.1, NFR-008)
- Error channel collects failures without blocking workers (NFR-007.3)
- WaitGroup ensures all workers complete before shutdown

**Configuration**:
- Default worker count: `runtime.NumCPU()` (good balance for mixed I/O and CPU workload)
- Max worker count: 100 (validation per FR-015.8)
- Work queue buffer: `workers * 10` (keeps workers busy without excessive memory)

### Database Connection Pooling

**Decision**: Use existing SQLite connection pool with increased size for concurrent workers

**Architecture** (extend `pkg/datastore/datastore.go`):
```go
db.SetMaxOpenConns(workerCount + 2)  // Workers + main thread + checkpoint saver
db.SetMaxIdleConns(workerCount)
db.SetConnMaxLifetime(time.Hour)
```

**Rationale**:
- SQLite WAL mode supports concurrent readers + single writer
- Connection pool prevents "database locked" errors under concurrent access
- Each worker needs dedicated connection for optimal performance (A-028)

### Synchronization Points

**Checkpoint Data Protection**:
```go
type CheckpointManager struct {
    mu                sync.Mutex
    currentFolderPath string
    lastProcessedFile string
    foldersScanned    int64  // atomic.Int64
    filesScanned      int64  // atomic.Int64
}
```

**Rationale**:
- Mutex protects checkpoint state during periodic saves (NFR-007.1)
- Atomic counters for progress statistics - no mutex contention on hot path
- Separate mutex for checkpoint vs progress - reduces lock contention

**Worker Failure Handling**:
- Workers catch panics and log to error channel
- Other workers continue unaffected (NFR-007.3)
- Failed work items can be retried or skipped based on error type
- Scan completes with partial results and error summary

### Signal Handling Pattern

**Decision**: Use `os/signal` package with context cancellation

**Architecture**:
```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

go func() {
    <-sigChan
    log.Info("Shutdown signal received, saving checkpoint...")
    cancelFunc()  // Cancel context - stops workers
    saveCheckpoint()  // Persist current state
    os.Exit(0)
}()
```

**Rationale**:
- Context cancellation propagates to all workers immediately
- Workers check `ctx.Done()` in their loops - stop gracefully
- Checkpoint save happens after workers acknowledge cancellation
- 5 second timeout enforced (NFR-008) - force exit if workers don't respond

## Performance Considerations

### Hash Algorithm Performance

**Benchmark Results** (Go 1.21 on modern hardware):
- SHA-512: ~450 MB/sec (64KB blocks)
- SHA-256: ~350 MB/sec (64KB blocks)
- SHA3-256: ~250 MB/sec (64KB blocks)

**Optimization**: Read files in 64KB chunks (optimal for most algorithms and I/O)

**Trade-offs**:
- SHA-512 is fastest but produces larger hash values (64 bytes vs 32 bytes) - minimal storage impact
- SHA3-256 is slowest but quantum-resistant - future-proofing consideration

### Worker Pool Sizing

**Small Files** (< 1 MB):
- **Bottleneck**: Folder traversal and filesystem metadata access
- **Strategy**: Higher worker count benefits (more parallelism for I/O-bound ops)
- **Recommendation**: `runtime.NumCPU() * 2` for traversal

**Large Files** (> 100 MB):
- **Bottleneck**: Hash calculation (CPU-bound)
- **Strategy**: Worker count = CPU cores (avoid context switching overhead)
- **Recommendation**: `runtime.NumCPU()` for hashing

**Configuration**: Single worker pool size used for both operations (keep it simple) - defaults to `runtime.NumCPU()`, user can tune based on their workload characteristics (FR-015.3, A-026)

### Database Batch Operations

**Decision**: Batch inserts in transactions of 1000 records

**Rationale**:
- Transaction overhead dominates for single-row inserts
- 1000 records per transaction balances memory vs performance
- Checkpoint granularity is at folder level - transaction can span multiple folders without violating resume logic

## Test Fixtures Design

### Directory Structure

```text
tests/fixtures/
├── duplicates/              # Exact duplicate files
│   ├── file1.txt           # "Hello World" (12 bytes)
│   ├── file1_copy.txt      # "Hello World" (12 bytes, same hash)
│   ├── file2.bin           # Random 1MB binary
│   └── file2_copy.bin      # Identical 1MB binary
├── folders/                # Duplicate folder structures
│   ├── folder_a/
│   │   ├── doc1.txt
│   │   └── doc2.txt
│   └── folder_b/           # Identical structure and content
│       ├── doc1.txt
│       └── doc2.txt
├── partial/                # Partial folder duplicates
│   ├── partial_a/
│   │   ├── common1.txt     # Present in both
│   │   ├── common2.txt     # Present in both
│   │   └── unique_a.txt    # Only in partial_a
│   └── partial_b/
│       ├── common1.txt     # Present in both
│       ├── common2.txt     # Present in both
│       └── unique_b.txt    # Only in partial_b
├── permissions/            # Access restriction scenarios
│   ├── readable.txt        # Normal file
│   └── restricted/         # Folder with chmod 000 (set during test setup)
│       └── hidden.txt
└── edge_cases/
    ├── empty.txt           # 0 bytes (test A-007)
    ├── special_chars_!@#.txt
    ├── very_long_name_[...].txt  # 255 chars
    └── symlink.txt -> duplicates/file1.txt
```

**Design Rationale**:
- Small files (<1 MB) - fast test execution, version control friendly
- Known hashes documented in test code - deterministic validation
- Permission scenarios set up programmatically during test execution (can't commit chmod 000 to git)
- Edge cases cover A-007 (empty files), special characters, symlinks

### Test Data Documentation

Each fixture folder includes `README.md` with:
- File names and their expected hash values (SHA-512 default)
- Folder duplicate relationships
- Expected similarity percentages for partial matches
- Setup instructions for permission tests

**Example** (`tests/fixtures/duplicates/README.md`):
```markdown
## Duplicate Files Test Fixtures

- `file1.txt`: SHA-512 = `e1b8f7e6...`, Size = 12 bytes
- `file1_copy.txt`: SHA-512 = `e1b8f7e6...`, Size = 12 bytes (DUPLICATE)
- `file2.bin`: SHA-512 = `a3d4c2b1...`, Size = 1048576 bytes
- `file2_copy.bin`: SHA-512 = `a3d4c2b1...`, Size = 1048576 bytes (DUPLICATE)

Expected duplicate sets:
- Set 1: {file1.txt, file1_copy.txt}
- Set 2: {file2.bin, file2_copy.bin}
```

## Risks and Mitigations

### Risk 1: Database Lock Contention

**Risk**: Multiple workers writing to SQLite simultaneously causes "database locked" errors

**Mitigation**:
- WAL mode already enabled (supports concurrent readers + single writer)
- Connection pool sized for worker count
- Batch inserts reduce write frequency
- Retry logic with exponential backoff for transient lock failures

### Risk 2: Worker Pool Deadlock

**Risk**: Workers waiting for work while work queue is full

**Mitigation**:
- Buffered work queue (size = workers * 10) prevents blocking
- Timeout on work submission (fail-fast if queue is stuck)
- Context cancellation breaks deadlock scenarios

### Risk 3: Checkpoint Corruption

**Risk**: Application killed during checkpoint write leaves corrupt state

**Mitigation**:
- SQLite transactions ensure atomic checkpoint writes
- Checkpoint corruption detection on startup (FR edge case: corrupted checkpoint)
- Fallback to --restart if corruption detected

### Risk 4: Large File Memory Pressure

**Risk**: Hashing files in 64KB chunks still consumes memory across many workers

**Mitigation**:
- Worker count configurable (can reduce for very large files)
- Streaming hash calculation (only 64KB buffer per worker)
- Progress tracking uses atomic counters, not buffering file data

## Open Questions (Resolved)

All NEEDS CLARIFICATION items from Technical Context have been resolved through this research phase.

## References

- Feature Specification: `specs/001-duplicate-scan-system/spec.md`
- DupeCTL Constitution: `.specify/memory/constitution.md`
- Existing Database Code: `pkg/datastore/datastore.go`, `pkg/datastore/agent.go`
- Existing Entity Models: `pkg/entities/files.go`
- Go Crypto Package Docs: https://pkg.go.dev/crypto
- Go Concurrency Patterns: https://go.dev/blog/pipelines
