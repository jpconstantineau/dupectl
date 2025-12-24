# Quickstart Guide: Duplicate Scan System Development

**Phase**: 1 (Design & Contracts)  
**Date**: December 23, 2025  
**Feature**: 001-duplicate-scan-system

## Overview

Developer setup guide for implementing duplicate file/folder scanning system in DupeCTL. Covers environment setup, building, testing, and development workflow for Golang on Windows with multi-platform release via GoReleaser.

## Prerequisites

### Required Software

**Go 1.21+**:
- Download: https://go.dev/dl/
- Install: Run installer, verify with `go version`
- Windows: Add Go bin directory to PATH if not done by installer

**Git**:
- Already installed (repository cloned)
- Verify: `git --version`

**GoReleaser** (for releases, optional for development):
- Install: `go install github.com/goreleaser/goreleaser@latest`
- Verify: `goreleaser --version`
- Alternative: Download binary from https://github.com/goreleaser/goreleaser/releases

**SQLite** (optional, for manual database inspection):
- Download: https://www.sqlite.org/download.html
- Windows: Download `sqlite-tools-win32-x86-*.zip`
- Extract and add to PATH

### Platform Support

**Primary Development**: Windows (your machine)
**Target Platforms**: Windows, Linux, macOS (via GoReleaser)

**Cross-Platform Notes**:
- Use `filepath.Join()` for paths (never hardcode `/` or `\`)
- Test path handling on Windows (backslashes)
- CI/CD will test Linux and macOS builds

## Initial Setup

### Clone Repository

```powershell
# Already done, but for reference:
git clone https://github.com/yourusername/dupectl.git
cd dupectl
```

### Install Dependencies

```powershell
# Download all Go module dependencies
go mod download

# Verify dependencies
go mod verify
```

**Expected Output**:
```
all modules verified
```

### Verify Existing Code

```powershell
# Build existing code to ensure setup is correct
go build -o dupectl.exe

# Run to see current commands
.\dupectl.exe --help
```

**Expected Output**: List of existing commands (add, get, delete, apply, etc.)

## Project Structure

### Key Directories

```text
dupectl/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go             # Root command + Viper config
│   ├── scan*.go            # Scan commands (to be implemented)
│   ├── get*.go             # Query commands
│   └── add*.go             # Management commands
├── pkg/                    # Application logic
│   ├── entities/           # Domain models
│   ├── scanner/            # Scanning services (to be created)
│   ├── hash/               # Hashing services (to be created)
│   ├── duplicate/          # Duplicate detection (to be created)
│   ├── datastore/          # Database layer
│   └── checkpoint/         # Checkpoint management (to be created)
├── internal/               # Private code
│   └── worker/             # Worker pools (to be created)
├── tests/                  # Test suites
│   ├── fixtures/           # Test data (to be created)
│   ├── unit/               # Unit tests
│   ├── integration/        # Integration tests
│   └── e2e/                # End-to-end tests
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
└── main.go                 # Entry point
```

## Development Workflow

### 1. Create Feature Branch

```powershell
# Checkout feature branch (already created by speckit.plan)
git checkout 001-duplicate-scan-system-01

# Verify branch
git branch --show-current
```

### 2. Implement a Feature

**Example**: Create files table in database

**Step 1**: Extend datastore (following `pkg/datastore/agent.go` pattern)

**File**: `pkg/datastore/files.go`

```go
package datastore

import "database/sql"

// CreateFilesTable creates the files table for scan data
func CreateFilesTable(db *sql.DB) error {
    query := `
    CREATE TABLE IF NOT EXISTS files (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        path TEXT NOT NULL UNIQUE,
        size INTEGER NOT NULL,
        mtime INTEGER NOT NULL,
        hash_value TEXT,
        hash_algorithm TEXT,
        error_status TEXT,
        first_scanned_at INTEGER NOT NULL,
        last_scanned_at INTEGER NOT NULL,
        removed INTEGER DEFAULT 0,
        folder_id INTEGER NOT NULL,
        root_folder_id INTEGER NOT NULL,
        FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE CASCADE,
        FOREIGN KEY (root_folder_id) REFERENCES root_folders(id) ON DELETE CASCADE
    );
    
    CREATE INDEX IF NOT EXISTS idx_files_hash 
        ON files(hash_value, size) 
        WHERE removed = 0 AND error_status IS NULL AND hash_value IS NOT NULL;
    `
    
    _, err := db.Exec(query)
    return err
}
```

**Step 2**: Write unit test

**File**: `tests/unit/files_test.go`

```go
package unit

import (
    "testing"
    "dupectl/pkg/datastore"
)

func TestCreateFilesTable(t *testing.T) {
    db, err := datastore.GetInMemoryDB()  // Test helper
    if err != nil {
        t.Fatalf("Failed to create test database: %v", err)
    }
    defer db.Close()
    
    err = datastore.CreateFilesTable(db)
    if err != nil {
        t.Errorf("CreateFilesTable failed: %v", err)
    }
    
    // Verify table exists
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='files'").Scan(&count)
    if err != nil {
        t.Errorf("Query failed: %v", err)
    }
    if count != 1 {
        t.Errorf("Expected 1 table, got %d", count)
    }
}
```

**Step 3**: Run test

```powershell
go test ./tests/unit/files_test.go -v
```

**Step 4**: Commit

```powershell
git add pkg/datastore/files.go tests/unit/files_test.go
git commit -m "Add files table creation with indexes"
```

### 3. Run All Tests

```powershell
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run with verbose output
go test ./... -v

# Run specific package
go test ./pkg/datastore/... -v

# Run with race detection
go test ./... -race
```

**Coverage Target**: ≥70% overall, ≥80% for core logic (scanner, hash, duplicate)

### 4. Build Application

```powershell
# Development build
go build -o dupectl.exe

# Build with race detector (debug)
go build -race -o dupectl-race.exe

# Build for Linux (cross-compile from Windows)
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o dupectl-linux
```

### 5. Run Locally

```powershell
# Run directly with go run
go run main.go scan all C:\path\to\test\folder --progress

# Or build and run
.\dupectl.exe scan all C:\path\to\test\folder --progress
```

### 6. Debug

**VSCode Launch Configuration** (`.vscode/launch.json`):

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug dupectl",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/main.go",
            "args": ["scan", "all", "C:\\test\\folder", "--progress"],
            "env": {},
            "showLog": true
        }
    ]
}
```

**Set Breakpoint**: Click left margin in VSCode, press F5 to start debugging

### 7. Manual Database Inspection

```powershell
# Open database with SQLite CLI
sqlite3 ~/.dupectl/dupedb.db

# View tables
.tables

# View files table schema
.schema files

# Query files
SELECT * FROM files LIMIT 10;

# Exit
.quit
```

## Testing Strategy

### Unit Tests (70% of tests)

**Location**: `tests/unit/`

**Scope**: Individual functions, single package

**Example**: Hash algorithm correctness

**File**: `tests/unit/hash_test.go`

```go
func TestSHA512Hash(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"Hello World", "2c74fd17edafd80e8447b0d46741ee243b7eb74dd2149a0ab1b9246fb30382f27e853d8585719e0e67cbda0daa8f51671064615d645ae27acb15bfb1447f459b"},
        {"", "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"},
    }
    
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            hasher := sha512.New()
            hasher.Write([]byte(tt.input))
            result := hex.EncodeToString(hasher.Sum(nil))
            
            if result != tt.expected {
                t.Errorf("SHA-512(%q) = %s, want %s", tt.input, result, tt.expected)
            }
        })
    }
}
```

**Run**: `go test ./tests/unit/hash_test.go -v`

### Integration Tests (20% of tests)

**Location**: `tests/integration/`

**Scope**: Multiple packages, database operations

**Example**: Complete scan workflow

**File**: `tests/integration/scan_test.go`

```go
func TestScanAllWorkflow(t *testing.T) {
    // Setup: Create test database and test folder
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    testDir := setupTestFixtures(t)  // Uses tests/fixtures/
    defer os.RemoveAll(testDir)
    
    // Execute: Run scan
    scanner := scanner.NewScanner(db, testDir, "sha512")
    err := scanner.ScanAll(context.Background())
    if err != nil {
        t.Fatalf("ScanAll failed: %v", err)
    }
    
    // Verify: Check database records
    var fileCount int
    db.QueryRow("SELECT COUNT(*) FROM files WHERE root_folder_id = ?", scanner.RootID).Scan(&fileCount)
    
    expectedFiles := countFilesInFixtures(testDir)
    if fileCount != expectedFiles {
        t.Errorf("Expected %d files, got %d", expectedFiles, fileCount)
    }
}
```

**Run**: `go test ./tests/integration/... -v`

### End-to-End Tests (10% of tests)

**Location**: `tests/e2e/`

**Scope**: Full CLI commands, complete workflows

**Example**: CLI command execution

**File**: `tests/e2e/cli_test.go`

```go
func TestScanAllCommand(t *testing.T) {
    // Build test binary
    buildTestBinary(t)
    defer removeTestBinary(t)
    
    // Setup test folder
    testDir := setupTestFixtures(t)
    defer os.RemoveAll(testDir)
    
    // Execute CLI command
    cmd := exec.Command("./dupectl-test", "scan", "all", testDir, "--progress")
    output, err := cmd.CombinedOutput()
    if err != nil {
        t.Fatalf("Command failed: %v\nOutput: %s", err, output)
    }
    
    // Verify output contains expected text
    if !strings.Contains(string(output), "Scan completed") {
        t.Errorf("Expected 'Scan completed' in output, got: %s", output)
    }
}
```

**Run**: `go test ./tests/e2e/... -v`

## Test Fixtures

### Creating Test Data

**Location**: `tests/fixtures/`

**Structure** (from contracts):
```text
tests/fixtures/
├── duplicates/              # Exact duplicate files
│   ├── file1.txt           # "Hello World"
│   ├── file1_copy.txt      # "Hello World" (same content)
│   ├── file2.bin           # Random 1MB binary
│   └── file2_copy.bin      # Identical 1MB binary
├── folders/                # Duplicate folder structures
│   ├── folder_a/
│   │   ├── doc1.txt
│   │   └── doc2.txt
│   └── folder_b/           # Identical structure
│       ├── doc1.txt
│       └── doc2.txt
├── partial/                # Partial duplicates (50%+ match)
│   ├── partial_a/
│   │   ├── common1.txt
│   │   ├── common2.txt
│   │   └── unique_a.txt
│   └── partial_b/
│       ├── common1.txt
│       ├── common2.txt
│       └── unique_b.txt
└── permissions/            # Permission test cases
    ├── readable.txt
    └── restricted/         # Set chmod 000 in test setup
        └── hidden.txt
```

**Create Fixtures**:

```powershell
# Script to create test fixtures
cd tests\fixtures

# Create duplicates
mkdir duplicates
"Hello World" | Out-File -FilePath duplicates\file1.txt -NoNewline
"Hello World" | Out-File -FilePath duplicates\file1_copy.txt -NoNewline

# Create binary (1MB random data)
$bytes = New-Object byte[] 1048576
(New-Object Random).NextBytes($bytes)
[System.IO.File]::WriteAllBytes("duplicates\file2.bin", $bytes)
Copy-Item duplicates\file2.bin duplicates\file2_copy.bin

# Create folder duplicates
mkdir folders\folder_a, folders\folder_b
"Document 1" | Out-File -FilePath folders\folder_a\doc1.txt
"Document 2" | Out-File -FilePath folders\folder_a\doc2.txt
Copy-Item folders\folder_a\* folders\folder_b\

# Create partial duplicates
mkdir partial\partial_a, partial\partial_b
"Common content 1" | Out-File -FilePath partial\partial_a\common1.txt
"Common content 2" | Out-File -FilePath partial\partial_a\common2.txt
"Unique to A" | Out-File -FilePath partial\partial_a\unique_a.txt
Copy-Item partial\partial_a\common*.txt partial\partial_b\
"Unique to B" | Out-File -FilePath partial\partial_b\unique_b.txt
```

**Document Expected Hashes** (`tests/fixtures/README.md`):

```markdown
## Test Fixtures

### duplicates/
- `file1.txt`: SHA-512 = 2c74fd17edafd80e8447b0d46741ee243b7eb74dd2149a0ab1b9246fb30382f27e853d8585719e0e67cbda0daa8f51671064615d645ae27acb15bfb1447f459b
- `file1_copy.txt`: Identical hash (DUPLICATE)
- `file2.bin`: 1MB random binary
- `file2_copy.bin`: Identical hash (DUPLICATE)

### folders/
- `folder_a/`: 2 files
- `folder_b/`: Identical structure and content (DUPLICATE FOLDER)

### partial/
- `partial_a/`: 3 files (2 common, 1 unique)
- `partial_b/`: 3 files (2 common, 1 unique)
- Similarity: 66% (2 common / 3 total unique)
```

## Configuration

### Development Config

**File**: `~/.dupectl.yaml`

```yaml
scan:
  hash_algorithm: sha512      # Default for testing
  worker_count: 4             # Lower for debugging
  progress_interval: 5        # Faster updates for dev
  checkpoint_interval: 30     # More frequent checkpoints
  
database:
  path: ~/.dupectl/dupedb.db
  wal_mode: true
  
logging:
  level: debug                # Verbose logging for dev
  file: ~/.dupectl/dupectl.log
```

### Test Config

**File**: `tests/test-config.yaml`

```yaml
scan:
  hash_algorithm: sha256      # Faster for tests
  worker_count: 2             # Minimal workers
  progress_interval: 1        # Fast updates
  checkpoint_interval: 10     # Frequent checkpoints
  
database:
  path: ":memory:"            # In-memory database for tests
  wal_mode: false
  
logging:
  level: error                # Minimal logging in tests
```

## Release Process (GoReleaser)

### Configuration

**File**: `.goreleaser.yaml`

```yaml
project_name: dupectl

builds:
  - id: dupectl
    main: ./main.go
    binary: dupectl
    goos:
      - windows
      - linux
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}}

archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
    files:
      - LICENSE
      - README.md
      - docs/**/*

checksum:
  name_template: 'checksums.txt'

snapshot:
  name_template: "{{ incpatch .Version }}-next"

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
```

### Local Build (All Platforms)

```powershell
# Build snapshot (local testing, no Git tags required)
goreleaser build --snapshot --clean

# Output in dist/ folder:
# - dist/dupectl_windows_amd64/dupectl.exe
# - dist/dupectl_linux_amd64/dupectl
# - dist/dupectl_darwin_amd64/dupectl
```

### Release Build (Tagged)

```powershell
# Tag version
git tag -a v0.1.0 -m "Initial duplicate scan system release"
git push origin v0.1.0

# Build and publish release (requires GitHub token)
$env:GITHUB_TOKEN="your_token_here"
goreleaser release --clean
```

## Common Tasks

### Add New CLI Command

1. Create file: `cmd/newcommand.go`
2. Follow pattern from existing commands (e.g., `cmd/scanAll.go`)
3. Register with Cobra in `cmd/root.go` or parent command
4. Write tests: `tests/e2e/newcommand_test.go`
5. Update help text and documentation

### Add New Database Table

1. Create file: `pkg/datastore/newtable.go`
2. Follow pattern from `pkg/datastore/agent.go`
3. Add to migration function: `pkg/datastore/migrations.go`
4. Write tests: `tests/integration/newtable_test.go`
5. Update data model: `specs/001-duplicate-scan-system-01/data-model.md`

### Add New Hash Algorithm

1. Create file: `pkg/hash/newalgo.go`
2. Implement `Hasher` interface (defined in `pkg/hash/hasher.go`)
3. Register in config validation: `cmd/root.go`
4. Write tests: `tests/unit/newalgo_test.go`
5. Update documentation

### Run Specific Test

```powershell
# Run single test function
go test ./tests/unit/hash_test.go -run TestSHA512Hash -v

# Run all tests in file
go test ./tests/unit/hash_test.go -v

# Run tests matching pattern
go test ./... -run TestScan -v
```

## Troubleshooting

### Issue: "go: module not found"

**Solution**: Run `go mod download` to fetch dependencies

### Issue: "database locked" error

**Solution**: 
- Ensure no other process is using database
- Check WAL mode is enabled: `PRAGMA journal_mode=WAL;`
- Increase connection pool size in tests

### Issue: Tests fail on Windows with path errors

**Solution**: 
- Use `filepath.Join()` instead of hardcoded paths
- Convert test fixtures to use platform-agnostic paths
- Check for backslash vs forward slash issues

### Issue: High memory usage during tests

**Solution**:
- Use `:memory:` database for unit tests
- Clean up fixtures after each test with `defer`
- Reduce worker count in test config

### Issue: Race condition detected

**Solution**:
- Run with `-race` flag to identify issue: `go test ./... -race`
- Add mutex protection around shared state
- Use atomic operations for counters

## References

- Feature Specification: `specs/001-duplicate-scan-system/spec.md`
- Implementation Plan: `specs/001-duplicate-scan-system-01/plan.md`
- Data Model: `specs/001-duplicate-scan-system-01/data-model.md`
- CLI Contracts: `specs/001-duplicate-scan-system-01/contracts/cli-commands.md`
- Worker Contracts: `specs/001-duplicate-scan-system-01/contracts/worker-pools.md`
- Go Documentation: https://go.dev/doc/
- Cobra Documentation: https://cobra.dev/
- GoReleaser Documentation: https://goreleaser.com/
