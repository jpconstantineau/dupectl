# Research: Duplicate Scan System

**Feature**: 001-duplicate-scan-system  
**Date**: December 23, 2025  
**Purpose**: Technical research and decision documentation for implementation planning

## Hash Algorithm Selection

### Decision: Go stdlib crypto packages (SHA-256, SHA-512, SHA3-256)

**Rationale**:
- All three algorithms available in Go standard library (crypto/sha256, crypto/sha512, golang.org/x/crypto/sha3)
- SHA-256: Excellent performance (~500 MB/s), widely adopted, sufficient security (256-bit)
- SHA-512: Better security (512-bit), good performance (~400 MB/s), recommended for sensitive data
- SHA3-256: Latest SHA standard, future-proof, slightly slower (~300 MB/s) but acceptable
- Zero collision probability for file duplicate detection at these key sizes
- No external dependencies (constitution compliance - Principle VI)

**Implementation Approach**:
```go
// pkg/hash/hasher.go interface
type Hasher interface {
    Hash(reader io.Reader) (string, error)
    Algorithm() string
}

// Factory pattern based on config
func NewHasher(algorithm string) (Hasher, error) {
    switch algorithm {
    case "sha256": return &SHA256Hasher{}, nil
    case "sha512": return &SHA512Hasher{}, nil
    case "sha3-256": return &SHA3Hasher{}, nil
    }
}
```

**Performance Target**: Achieve ≥50 MB/sec as specified in requirements by reading files in 64KB chunks.

**Alternatives Considered**:
- MD5/SHA-1: Rejected due to known collision vulnerabilities
- BLAKE2/BLAKE3: Rejected - not in stdlib, adds external dependency
- xxHash: Rejected - non-cryptographic, unsuitable for duplicate detection

---

## Database Schema Design

### Decision: Dedicated scan tables with foreign keys to existing infrastructure

**Files Table**:
```sql
CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    root_folder_id INTEGER NOT NULL,
    folder_id INTEGER,
    path TEXT NOT NULL,          -- Absolute path
    name TEXT NOT NULL,
    size INTEGER NOT NULL,
    mtime INTEGER NOT NULL,      -- Unix timestamp
    hash_value TEXT,             -- NULL if not yet hashed
    hash_algorithm TEXT,         -- sha256, sha512, sha3-256
    error_status TEXT,           -- NULL if no error, error message if permission denied
    scanned_at INTEGER,          -- Unix timestamp of last scan
    FOREIGN KEY (root_folder_id) REFERENCES root_folders(id),
    FOREIGN KEY (folder_id) REFERENCES folders(id),
    UNIQUE(path)                 -- Prevent duplicate entries
);

CREATE INDEX idx_files_hash ON files(hash_value, size);  -- For duplicate detection
CREATE INDEX idx_files_root ON files(root_folder_id);
CREATE INDEX idx_files_folder ON files(folder_id);
```

**Folders Table**:
```sql
CREATE TABLE IF NOT EXISTS folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    root_folder_id INTEGER NOT NULL,
    parent_folder_id INTEGER,    -- NULL for root
    path TEXT NOT NULL,          -- Absolute path
    name TEXT NOT NULL,
    scanned_at INTEGER,          -- Unix timestamp of last scan
    FOREIGN KEY (root_folder_id) REFERENCES root_folders(id),
    FOREIGN KEY (parent_folder_id) REFERENCES folders(id),
    UNIQUE(path)                 -- Prevent duplicate entries
);

CREATE INDEX idx_folders_root ON folders(root_folder_id);
CREATE INDEX idx_folders_parent ON folders(parent_folder_id);
```

**Scan State Table** (for checkpoint/resume):
```sql
CREATE TABLE IF NOT EXISTS scan_state (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    root_folder_id INTEGER NOT NULL,
    scan_mode TEXT NOT NULL,      -- 'all', 'folders', 'files'
    current_folder_path TEXT,     -- Last folder being processed
    started_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    completed_at INTEGER,         -- NULL if in progress
    status TEXT NOT NULL,         -- 'running', 'completed', 'interrupted'
    files_processed INTEGER DEFAULT 0,
    folders_processed INTEGER DEFAULT 0,
    FOREIGN KEY (root_folder_id) REFERENCES root_folders(id)
);

CREATE INDEX idx_scan_state_root ON scan_state(root_folder_id, status);
```

**Rationale**:
- Separate tables maintain clean separation (Constitution Principle II)
- Foreign keys ensure referential integrity
- Indexes optimize duplicate queries (size + hash lookup is O(log n))
- Absolute paths stored (clarification decision #8)
- Error status field enables permission error tracking (clarification #9)
- Hash algorithm stored for future migrations (clarification #1)

---

## Folder Traversal Pattern

### Decision: Recursive depth-first traversal with os.Walk

**Implementation**:
```go
// pkg/scanner/folder.go
func (s *FolderScanner) ScanDirectory(rootPath string) error {
    return filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            // Log permission error, continue with siblings
            s.logPermissionError(path, err)
            return filepath.SkipDir
        }
        
        if info.IsDir() {
            s.processFolder(path, info)
            s.saveCheckpoint(path) // Checkpoint at folder boundaries
        } else {
            s.processFile(path, info)
        }
        
        s.reportProgress() // Check if 10s elapsed
        return nil
    })
}
```

**Rationale**:
- filepath.Walk handles cross-platform paths (Constitution Principle X)
- Automatic recursion simplifies implementation
- Error handling continues scan on permission denied (NFR-006)
- Checkpoint at folder boundaries (clarification #3, assumption A-010)

**Alternatives Considered**:
- filepath.WalkDir (Go 1.16+): Better performance but requires custom logic for checkpoints
- Manual recursion: More control but higher complexity

---

## Progress Reporting Mechanism

### Decision: Time-based ticker with atomic counters

**Implementation**:
```go
// pkg/scanner/progress.go
type ProgressTracker struct {
    filesProcessed   atomic.Int64
    foldersProcessed atomic.Int64
    lastReport       atomic.Value // time.Time
    interval         time.Duration
}

func (pt *ProgressTracker) ReportIfDue() {
    last := pt.lastReport.Load().(time.Time)
    if time.Since(last) >= pt.interval {
        fmt.Printf("Progress: %d files, %d folders processed\n", 
            pt.filesProcessed.Load(), pt.foldersProcessed.Load())
        pt.lastReport.Store(time.Now())
    }
}
```

**Rationale**:
- Time-based (10s configurable) per clarification #2
- Atomic operations enable goroutine-safe concurrent updates
- Console output only (no file logging for progress) per clarification #9
- Minimal performance overhead (<1ms per check)

---

## Duplicate Detection Algorithm

### Decision: Two-phase approach - Size pre-filter then hash comparison

**File Duplicates**:
```sql
-- Phase 1: Group by size (fast)
SELECT size, COUNT(*) as count
FROM files
WHERE hash_value IS NOT NULL
GROUP BY size
HAVING count > 1;

-- Phase 2: Group by hash within same size (definitive)
SELECT size, hash_value, hash_algorithm, GROUP_CONCAT(path) as paths
FROM files
WHERE size IN (size_list_from_phase1)
AND hash_value IS NOT NULL
GROUP BY size, hash_value, hash_algorithm
HAVING COUNT(*) > 1;
```

**Folder Duplicates** (exact match):
```go
// pkg/detector/folders.go
func DetectExactFolderDuplicates() []DuplicateSet {
    // 1. Compute folder signature: sorted concat of all file hashes
    // 2. Group folders by signature
    // 3. Folders with identical signatures are exact duplicates
}
```

**Partial Folder Duplicates**:
```go
// pkg/detector/partial.go
func DetectPartialDuplicates(minSimilarity float64) []PartialMatch {
    // 1. For each folder pair, compute set intersection of file hashes
    // 2. Similarity = (intersection size / union size) * 100
    // 3. Filter by minSimilarity threshold (default 50% per clarification #4)
    // 4. Identify missing files and name mismatches
}
```

**Rationale**:
- Size pre-filter reduces hash comparisons by ~90%
- Database indexes make lookups efficient
- Folder signature approach avoids O(n²) file comparisons
- Partial matching uses set operations (well-understood algorithm)

---

## Checkpoint and Resume Strategy

### Decision: Folder-level checkpoints with scan_state table

**Checkpoint Logic**:
1. After completing each folder, update scan_state.current_folder_path
2. Batch database writes (1000 file records per transaction) for performance
3. On interruption (SIGINT), complete current transaction and update scan_state

**Resume Logic**:
1. On scan start, check scan_state for existing 'running' or 'interrupted' status
2. If found, resume from current_folder_path
3. Skip folders/files already processed (check folders/files tables)
4. Continue traversal from checkpoint

**Rationale**:
- Folder-level granularity balances checkpoint frequency vs performance (assumption A-010)
- Database persists checkpoint automatically (transactional consistency)
- Resume logic simple: skip processed items, continue from checkpoint
- Meets clarification #3 requirement for checkpoint-based resume

---

## Configuration Management

### Decision: Viper config with sensible defaults

**New Configuration Keys**:
```yaml
# .dupectl.yaml
scan:
  hash_algorithm: "sha256"        # sha256, sha512, sha3-256
  progress_interval: "10s"        # Time between progress updates
  batch_size: 1000                # Files per database transaction
  concurrent_hashers: 4           # Parallel file hashing goroutines
```

**Implementation**:
```go
// cmd/root.go - setDefaults()
viper.SetDefault("scan.hash_algorithm", "sha256")
viper.SetDefault("scan.progress_interval", "10s")
viper.SetDefault("scan.batch_size", 1000)
viper.SetDefault("scan.concurrent_hashers", 4)
```

**Rationale**:
- Global config per clarification #1
- Configurable progress interval per clarification #2
- Follows existing DupeCTL config pattern
- Defaults chosen for typical use case

---

## Output Formatting Strategy

### Decision: Formatter package with table and JSON implementations

**Table Format** (default):
```
Duplicate Set #1 (Size: 10.5 MB, Hash: sha256:abc123...)
  /path/to/file1.txt
  /path/to/file2.txt
  /backup/file1.txt

Duplicate Set #2 (Size: 2.3 KB, Hash: sha256:def456...)
  /docs/readme.md
  /archive/readme.md

Total: 2 duplicate sets, 5 files
```

**JSON Format** (--json flag):
```json
{
  "duplicate_sets": [
    {
      "size": 11010048,
      "hash": "sha256:abc123...",
      "algorithm": "sha256",
      "files": [
        "/path/to/file1.txt",
        "/path/to/file2.txt",
        "/backup/file1.txt"
      ]
    }
  ],
  "summary": {
    "duplicate_sets": 2,
    "total_files": 5
  }
}
```

**Rationale**:
- Table format human-readable (Constitution Principle IV)
- JSON format machine-parseable for scripting
- Command-line flag pattern standard in CLI tools
- Meets clarification #10 requirement

---

## Testing Strategy

### Unit Tests (70% of test effort):
- Hash algorithm correctness (known input/output pairs)
- Duplicate detection logic (various scenarios)
- Path normalization (relative to absolute conversion)
- Similarity calculation (partial duplicate matching)
- Progress tracking (time-based intervals)

### Integration Tests (20%):
- Database operations (CRUD for files/folders/scan_state)
- Folder traversal with permission errors
- Checkpoint save and resume flow
- Configuration loading from file

### End-to-End Tests (10%):
- Full scan workflow: scan all, scan folders, scan files
- Get duplicates with filtering and formats
- Root folder registration with confirmation prompts
- Scan interruption and resumption

**Test Coverage Target**: ≥70% per Constitution Principle III

---

## Performance Optimization Strategies

1. **Batch Database Writes**: Insert files in batches of 1000 to reduce transaction overhead
2. **Concurrent Hashing**: Use goroutine pool (4 workers) for parallel file hashing
3. **Read Buffering**: Read files in 64KB chunks for optimal disk I/O
4. **Index Usage**: Composite index on (hash_value, size) for fast duplicate lookups
5. **Streaming Hash Calculation**: Use io.Copy to avoid loading entire files in memory
6. **Early Exit**: Skip files with error_status set (permission denied) on re-scans

**Expected Performance**:
- Hash calculation: 100-500 MB/s depending on algorithm and disk speed
- Folder scan: 5000+ files/sec for metadata-only operations
- Memory: <200 MB for typical workloads, <500 MB for 100k files

---

## Cross-Platform Considerations

### Path Handling:
```go
// Always use filepath package
filepath.Join(root, "subdir", "file.txt")  // Cross-platform
filepath.Abs(relativePath)                  // Convert to absolute
filepath.ToSlash(path)                      // For display/JSON
```

### Separator Handling:
```go
// Use os.PathSeparator for dynamic separator
// Store absolute paths with platform-native separators in database
// Convert to forward slashes for JSON output (consistency)
```

### Case Sensitivity:
```go
// Treat paths case-sensitively (even on case-insensitive filesystems)
// Duplicate detection compares exact paths
// Document limitation: case-insensitive FS may have case-variant paths treated as different
```

---

## Summary of Technical Decisions

| Decision Area | Choice | Key Reason |
|---------------|--------|------------|
| Hash Algorithms | Go stdlib crypto (SHA-256/512, SHA3-256) | Performance, security, zero dependencies |
| Database Schema | Dedicated tables with FKs to infrastructure | Clean separation, referential integrity |
| Traversal | filepath.Walk with depth-first | Cross-platform, built-in recursion |
| Progress | Time-based atomic counters | Configurable intervals, goroutine-safe |
| Duplicate Detection | Size pre-filter + hash comparison | Efficient O(n log n) with indexes |
| Checkpoints | Folder-level with scan_state table | Balance frequency vs performance |
| Configuration | Viper with YAML defaults | Existing pattern, easy user configuration |
| Output Format | Table (default) + JSON (flag) | Human + machine readable |
| Testing | 70% unit, 20% integration, 10% e2e | Constitution compliance |
| Performance | Concurrent hashing, batched writes, buffered I/O | Meet 50 MB/s target |

All decisions align with DupeCTL Constitution principles and address clarifications from specification phase.
