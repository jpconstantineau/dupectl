# Worker Pool Contracts

**Phase**: 1 (Design & Contracts)  
**Date**: December 23, 2025  
**Feature**: 001-duplicate-scan-system

## Overview

Internal contracts for worker pool implementations used in parallel folder traversal and file hashing. Defines interfaces, behavior contracts, error handling, and synchronization guarantees.

## Generic Worker Pool Interface

### WorkItem Interface

**Package**: `internal/worker`

**Definition**:
```go
// WorkItem represents a unit of work that can be processed by a worker.
// Implementations must be thread-safe and idempotent where possible.
type WorkItem interface {
    // Process executes the work item and returns an error if processing fails.
    // Must be safe to call concurrently from multiple goroutines.
    // Should check context cancellation and return early if ctx.Done().
    Process(ctx context.Context) error
    
    // ID returns a unique identifier for this work item (for logging/debugging).
    ID() string
}
```

**Contract**:
- `Process()` must be **thread-safe** - can be called concurrently
- `Process()` must **check context** - return early if `ctx.Done()` is closed
- `Process()` must **not panic** - return error instead (worker will recover panics but proper error handling is preferred)
- `ID()` must return **unique identifier** within work queue (enables duplicate detection and logging)

### WorkerPool Interface

**Package**: `internal/worker`

**Definition**:
```go
// WorkerPool manages a pool of goroutines that process WorkItems concurrently.
type WorkerPool struct {
    workers   int                    // Number of worker goroutines
    workQueue chan WorkItem          // Buffered channel of work items
    wg        sync.WaitGroup         // Waits for all workers to complete
    ctx       context.Context        // Cancellation context
    cancel    context.CancelFunc     // Cancellation function
    errors    chan error             // Error collection channel
    metrics   *PoolMetrics           // Performance metrics
}

// PoolMetrics tracks worker pool performance statistics.
type PoolMetrics struct {
    ItemsProcessed  int64  // Atomic counter
    ItemsFailed     int64  // Atomic counter
    ItemsQueued     int64  // Atomic counter
    WorkersActive   int32  // Atomic counter
}

// NewWorkerPool creates a new worker pool with specified worker count.
// Workers start immediately upon creation.
func NewWorkerPool(ctx context.Context, workers int) *WorkerPool

// Submit adds a work item to the queue. Blocks if queue is full.
// Returns error if context is cancelled or pool is stopped.
func (wp *WorkerPool) Submit(item WorkItem) error

// Start begins processing work items (called automatically by NewWorkerPool).
func (wp *WorkerPool) Start()

// Stop stops accepting new work and cancels context.
// Does not wait for in-flight work to complete (use Wait for that).
func (wp *WorkerPool) Stop()

// Wait blocks until all workers have completed and returns collected errors.
// Should be called after Stop() to ensure graceful shutdown.
func (wp *WorkerPool) Wait() []error

// Metrics returns current pool performance statistics.
func (wp *WorkerPool) Metrics() PoolMetrics
```

**Contracts**:

**NewWorkerPool**:
- MUST create `workers` goroutines immediately
- MUST create buffered work queue with capacity = `workers * 10` (prevents blocking producers)
- MUST start processing immediately (goroutines wait on work queue)
- MUST validate `workers >= 1` (return error if invalid)

**Submit**:
- MUST block if work queue is full (backpressure mechanism)
- MUST return error if context is cancelled (prevent submitting to stopped pool)
- MUST return error if pool is stopped
- MUST increment `ItemsQueued` metric
- MUST be **thread-safe** (multiple goroutines can call simultaneously)

**Stop**:
- MUST cancel context (signals all workers to stop)
- MUST close work queue (prevents new submissions)
- MUST be **idempotent** (safe to call multiple times)
- MUST NOT wait for workers (that's Wait's job)

**Wait**:
- MUST wait for all workers to finish using WaitGroup
- MUST close error channel
- MUST return all collected errors (slice may be empty)
- MUST be called **after Stop()** for graceful shutdown

**Worker Behavior** (internal goroutine):
- MUST listen on both work queue and context.Done() (select statement)
- MUST process work items until queue closed or context cancelled
- MUST **recover from panics** - log panic, increment ItemsFailed, continue
- MUST collect errors to error channel (non-blocking send)
- MUST decrement WaitGroup on exit

## Folder Traversal Worker

### FolderTraversalWorkItem

**Package**: `internal/worker`

**Definition**:
```go
// FolderTraversalWorkItem represents a folder to traverse and register.
type FolderTraversalWorkItem struct {
    folderPath     string
    rootFolderID   int
    parentFolderID *int  // nil for root folder
    db             *sql.DB
    resultChan     chan<- *FolderResult
}

// FolderResult contains the result of traversing a folder.
type FolderResult struct {
    FolderID       int
    FolderPath     string
    SubfolderPaths []string
    FilePaths      []string
    Error          error
}

func (ft *FolderTraversalWorkItem) Process(ctx context.Context) error
func (ft *FolderTraversalWorkItem) ID() string  // returns folderPath
```

**Process() Contract**:
1. Check `ctx.Done()` - return early if cancelled
2. Register folder in database:
   - INSERT INTO folders (path, parent_folder_id, root_folder_id, first_scanned_at, last_scanned_at, removed)
   - Use parameterized query (SQL injection protection)
   - Handle UNIQUE constraint error (folder already registered → update last_scanned_at)
3. Read folder contents using `os.ReadDir()`
   - Handle permission errors → set error_status in database, return non-fatal error
   - Collect subfolder paths and file paths separately
4. For each file: INSERT INTO files (path, size, mtime, folder_id, root_folder_id, first_scanned_at, last_scanned_at)
   - Use batch INSERT with transaction (1000 files per transaction)
   - Handle UNIQUE constraint error (file already registered → update last_scanned_at)
5. Send `FolderResult` to result channel (non-blocking)
   - If channel full, log warning and drop result (backpressure handling)
6. Return error if database operations fail (fatal error)
7. Return nil on success

**Concurrency Guarantees**:
- Multiple workers can process different folders simultaneously
- Database connection pool handles concurrent writes (SQLite WAL mode)
- No shared state between work items (thread-safe)

**Error Handling**:
- Permission errors: Non-fatal, logged, folder marked with error_status
- Database errors: Fatal, returned to caller, increments ItemsFailed
- File read errors: Non-fatal, file skipped, logged

## File Hashing Worker

### FileHashingWorkItem

**Package**: `internal/worker`

**Definition**:
```go
// FileHashingWorkItem represents a file to hash.
type FileHashingWorkItem struct {
    filePath       string
    fileID         int
    algorithm      string  // "sha256", "sha512", "sha3-256"
    db             *sql.DB
    resultChan     chan<- *HashResult
}

// HashResult contains the result of hashing a file.
type HashResult struct {
    FileID         int
    FilePath       string
    HashValue      string  // Hex-encoded
    HashAlgorithm  string
    Size           int64
    Error          error
}

func (fh *FileHashingWorkItem) Process(ctx context.Context) error
func (fh *FileHashingWorkItem) ID() string  // returns filePath
```

**Process() Contract**:
1. Check `ctx.Done()` - return early if cancelled
2. Open file for reading
   - Handle permission errors → set error_status in database, return non-fatal error
   - Handle file not found → set removed flag in database, return non-fatal error
3. Create hash instance based on algorithm
   - `crypto/sha256.New()` for sha256
   - `crypto/sha512.New()` for sha512
   - `sha3.New256()` for sha3-256
4. Read file in 64KB chunks, streaming to hash
   - Check `ctx.Done()` periodically (every chunk)
   - Handle I/O errors → return fatal error
5. Finalize hash and encode as hex string
6. UPDATE files SET hash_value=?, hash_algorithm=?, last_scanned_at=? WHERE id=?
   - Use parameterized query
   - Handle database errors → return fatal error
7. Send `HashResult` to result channel (non-blocking)
8. Return nil on success

**Concurrency Guarantees**:
- Multiple workers can hash different files simultaneously
- Each worker has dedicated file handle (no shared file descriptors)
- Database connection pool handles concurrent updates
- No shared state between work items (thread-safe)

**Error Handling**:
- Permission errors: Non-fatal, logged, file marked with error_status
- File not found: Non-fatal, logged, file marked as removed
- I/O errors: Fatal, returned to caller, increments ItemsFailed
- Database errors: Fatal, returned to caller

**Performance Optimization**:
- 64KB chunk size balances memory usage and hash performance
- Only one 64KB buffer per worker (memory-efficient)
- No buffering of entire file (supports files larger than available RAM)

## Work Distribution Pattern

### Breadth-First Traversal

**Pattern**: Producer-Consumer with Work Queue

**Implementation**:
```go
// Producer: Scan orchestrator
pool := worker.NewWorkerPool(ctx, workerCount)
defer pool.Stop()

// Submit root folder as first work item
rootItem := &FolderTraversalWorkItem{
    folderPath:     rootPath,
    rootFolderID:   rootID,
    parentFolderID: nil,
    db:             db,
    resultChan:     resultChan,
}
pool.Submit(rootItem)

// Consumer: Process results and submit subfolder work items
go func() {
    for result := range resultChan {
        if result.Error != nil {
            log.Error("Folder traversal error", "path", result.FolderPath, "error", result.Error)
            continue
        }
        
        // Submit subfolder work items
        for _, subfolderPath := range result.SubfolderPaths {
            subfolderItem := &FolderTraversalWorkItem{
                folderPath:     subfolderPath,
                rootFolderID:   rootFolderID,
                parentFolderID: &result.FolderID,
                db:             db,
                resultChan:     resultChan,
            }
            if err := pool.Submit(subfolderItem); err != nil {
                log.Error("Failed to submit subfolder", "path", subfolderPath, "error", err)
                break  // Context cancelled or pool stopped
            }
        }
        
        // Submit file hashing work items (if scan mode is 'all' or 'files')
        for _, filePath := range result.FilePaths {
            fileItem := &FileHashingWorkItem{
                filePath:   filePath,
                fileID:     getFileID(filePath),  // Query from database
                algorithm:  hashAlgorithm,
                db:         db,
                resultChan: hashResultChan,
            }
            if err := hashPool.Submit(fileItem); err != nil {
                log.Error("Failed to submit file", "path", filePath, "error", err)
                break
            }
        }
    }
}()

// Wait for completion
pool.Wait()
```

**Guarantees**:
- All subfolders eventually submitted (breadth-first order)
- No subfolder processed before parent (parent creates work items for children)
- Bounded work queue prevents memory overflow
- Context cancellation stops all workers gracefully

## Error Collection Pattern

### Non-Fatal Errors

**Definition**: Errors that should be logged but don't stop the scan

**Examples**:
- Permission denied on file/folder
- File not found (removed during scan)
- Symbolic link loop detected

**Handling**:
1. Log error with context (path, error message)
2. Update database record with error_status
3. Return nil from `Process()` (success with error recorded)
4. Continue processing other items

### Fatal Errors

**Definition**: Errors that indicate serious problems and should stop the scan

**Examples**:
- Database connection lost
- Database transaction failed
- Disk full (can't write to database)
- Context cancelled (graceful shutdown)

**Handling**:
1. Log error with context
2. Return error from `Process()` (increments ItemsFailed)
3. Error collected in error channel
4. Scan orchestrator checks error count at end
5. If fatal errors > threshold, display error summary and exit

## Panic Recovery

**Worker Panic Handler**:
```go
defer func() {
    if r := recover(); r != nil {
        log.Error("Worker panic recovered", "panic", r, "stack", debug.Stack())
        atomic.AddInt64(&wp.metrics.ItemsFailed, 1)
        // Worker continues - panic doesn't crash entire pool
    }
    wp.wg.Done()
}()
```

**Contract**:
- Worker recovers panic and logs with stack trace
- Other workers unaffected (NFR-007.3: worker failure isolation)
- Failed work item counted in metrics
- Worker goroutine exits gracefully

## Shutdown Sequence

### Graceful Shutdown

**Sequence** (triggered by SIGINT/SIGTERM):
1. Signal handler calls `pool.Stop()` → cancels context
2. Workers check `ctx.Done()` on next iteration → exit loop
3. In-flight work items complete (workers finish current `Process()` call)
4. Main goroutine calls `pool.Wait()` → waits for WaitGroup
5. Checkpoint saved with current state
6. Application exits

**Timeout** (5 seconds per NFR-008):
```go
// Wait with timeout
done := make(chan struct{})
go func() {
    pool.Wait()
    close(done)
}()

select {
case <-done:
    log.Info("Workers stopped gracefully")
case <-time.After(5 * time.Second):
    log.Warn("Workers did not stop within timeout, forcing exit")
    // Checkpoint may be incomplete but application must exit
}
```

**Guarantees**:
- Workers stop within 5 seconds (or forced exit)
- Completed work is checkpointed
- In-flight work may be lost (re-processed on resume)
- No data corruption (database transactions are atomic)

## Performance Metrics

### Metrics Tracked

**Real-Time Counters** (atomic operations):
```go
type PoolMetrics struct {
    ItemsProcessed  int64  // Successful work items
    ItemsFailed     int64  // Failed work items (errors or panics)
    ItemsQueued     int64  // Total items submitted to queue
    WorkersActive   int32  // Current active workers
}
```

**Usage**:
```go
metrics := pool.Metrics()
log.Info("Worker pool status",
    "processed", metrics.ItemsProcessed,
    "failed", metrics.ItemsFailed,
    "queued", metrics.ItemsQueued,
    "active", metrics.WorkersActive,
)
```

**Display in Progress** (if --progress flag):
```text
⠋ Folders: 823 | Files: 6,420 | Workers: 8/8 | Elapsed: 1m 15s
```

### Performance Characteristics

**Folder Traversal**:
- **Throughput**: ~1000 folders/second (I/O bound)
- **Bottleneck**: Filesystem metadata access, database INSERT
- **Optimal Workers**: CPU cores × 2 (more concurrency helps with I/O wait)

**File Hashing**:
- **Throughput**: ~50 MB/sec per worker (CPU bound for large files)
- **Bottleneck**: Hash calculation, disk read
- **Optimal Workers**: CPU cores (more workers = context switching overhead)

**Recommendation**: Single worker pool sized to CPU cores (good balance for mixed workload per A-026)

## Testing Contracts

### Unit Tests

**Test**: `internal/worker/pool_test.go`

**Cases**:
- [ ] NewWorkerPool creates specified number of workers
- [ ] Submit blocks when queue full
- [ ] Submit returns error when context cancelled
- [ ] Stop cancels context and closes queue
- [ ] Wait returns all collected errors
- [ ] Worker recovers from panic
- [ ] Metrics counters updated correctly

### Integration Tests

**Test**: `tests/integration/worker_test.go`

**Cases**:
- [ ] Process 1000 work items concurrently → verify all completed
- [ ] Process work items with errors → verify non-fatal errors don't stop pool
- [ ] Cancel context during processing → verify graceful shutdown within 5 seconds
- [ ] Submit work items from multiple goroutines → verify thread-safety
- [ ] Worker panics on specific item → verify other workers continue

### Race Condition Tests

**Command**: `go test -race ./internal/worker`

**Cases**:
- [ ] Concurrent Submit calls → no data races
- [ ] Concurrent metrics reads → no data races
- [ ] Context cancellation during Submit → no data races

## References

- Feature Specification: `specs/001-duplicate-scan-system/spec.md` (FR-015.3-015.8, NFR-007.1-007.3)
- Research Document: `specs/001-duplicate-scan-system-01/research.md` (Concurrency Patterns section)
- DupeCTL Constitution: `.specify/memory/constitution.md` (Principle VII: Concurrency)
- Go Concurrency Patterns: https://go.dev/blog/pipelines
