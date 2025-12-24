# CLI Command Contracts

**Phase**: 1 (Design & Contracts)  
**Date**: December 23, 2025  
**Feature**: 001-duplicate-scan-system

## Overview

Command-line interface contracts for duplicate scan system. Defines command structure, arguments, flags, output formats, and exit codes for all user-facing CLI commands.

## Command Structure

All commands follow Cobra/Viper pattern with consistent verb-noun structure (Constitution IV: UX Consistency).

### Root Command

**Command**: `dupectl`

**Description**: DupeCTL - Duplicate file and folder detection system

**Global Flags**:
- `--config string` - Config file path (default: `~/.dupectl.yaml`)
- `--verbose` or `-v` - Enable verbose logging
- `--help` or `-h` - Display help information

**Exit Codes**:
- `0` - Success
- `1` - General error
- `2` - Usage error (invalid arguments/flags)

## Scan Commands

### scan all

**Command**: `dupectl scan all <root-folder-path> [flags]`

**Description**: Scan all files and folders recursively, calculating hashes for duplicate detection

**Arguments**:
- `<root-folder-path>` (required) - Absolute or relative path to root folder

**Flags**:
- `--progress` - Display real-time progress (spinner, folder count, file count, elapsed time)
- `--restart` - Restart scan from beginning (default: resume from checkpoint)
- `--help` or `-h` - Display command help

**Behavior**:
1. Convert relative path to absolute path
2. Check if root folder path is registered in database
3. If not registered, prompt: "Root folder not registered. Register now? (y/n)"
   - If yes → register with default config (traverse_links=false), proceed with scan
   - If no → exit with code 2
4. Check for existing incomplete scan checkpoint
   - If found and `--restart` not specified → resume from checkpoint
   - If `--restart` specified → clear checkpoint, start fresh
5. Start scan: traverse folder tree + hash all files
6. Display progress if `--progress` specified (update every 10 seconds)
7. Save periodic checkpoints (every folder completion)
8. On completion: display summary statistics
9. On SIGINT/SIGTERM: save checkpoint, exit cleanly

**Output Format** (without `--progress`):
```text
Scanning root folder: /path/to/root
Scan completed in 2m 34s
Folders scanned: 1,543
Files scanned: 12,847
Duplicates found: 456 files in 123 sets
```

**Output Format** (with `--progress`):
```text
Scanning root folder: /path/to/root
⠋ Folders: 823 | Files: 6,420 | Elapsed: 1m 15s
⠙ Folders: 1,102 | Files: 8,934 | Elapsed: 1m 25s
...
✓ Scan completed in 2m 34s
Folders scanned: 1,543
Files scanned: 12,847
Duplicates found: 456 files in 123 sets
```

**Error Cases**:
- Root folder path does not exist → Exit 1 with "Error: Root folder does not exist: /path/to/root"
- Permission denied on root folder → Exit 1 with "Error: Permission denied accessing root folder: /path/to/root"
- Active scan already in progress for this root → Exit 1 with "Error: Scan already in progress for this root folder. Wait for completion or use --restart to cancel and restart."

**Example Usage**:
```bash
dupectl scan all /home/user/documents --progress
dupectl scan all "C:\Users\user\Documents" --restart
dupectl scan all ../relative/path
```

### scan folders

**Command**: `dupectl scan folders <root-folder-path> [flags]`

**Description**: Scan folder structure only (no file hashing) for quick hierarchy registration

**Arguments**:
- `<root-folder-path>` (required) - Absolute or relative path to root folder

**Flags**:
- `--progress` - Display real-time progress (spinner, folder count, elapsed time)
- `--restart` - Restart scan from beginning (default: resume from checkpoint)
- `--help` or `-h` - Display command help

**Behavior**: Same as `scan all` but skips file hashing step

**Output Format** (without `--progress`):
```text
Scanning folder structure: /path/to/root
Scan completed in 12s
Folders scanned: 1,543
Files skipped: 12,847 (use 'dupectl scan files' to hash files)
```

**Error Cases**: Same as `scan all`

**Example Usage**:
```bash
dupectl scan folders /home/user/documents
dupectl scan folders "C:\Users\user\Documents" --progress --restart
```

### scan files

**Command**: `dupectl scan files <root-folder-path> [flags]`

**Description**: Hash files only (assumes folders already registered) for incremental scanning

**Arguments**:
- `<root-folder-path>` (required) - Absolute or relative path to root folder

**Flags**:
- `--progress` - Display real-time progress (spinner, file count, elapsed time)
- `--restart` - Restart scan from beginning (default: resume from checkpoint)
- `--help` or `-h` - Display command help

**Behavior**: Same as `scan all` but skips folder traversal step (uses registered folders from database)

**Output Format** (without `--progress`):
```text
Hashing files in: /path/to/root
Scan completed in 2m 22s
Files scanned: 12,847
Duplicates found: 456 files in 123 sets
```

**Error Cases**:
- Root folder not registered → Exit 1 with "Error: Root folder not registered. Run 'dupectl scan folders' first or use 'dupectl scan all'."
- No folders found in database for root → Exit 1 with "Error: No folders registered for this root. Run 'dupectl scan folders' first."

**Example Usage**:
```bash
dupectl scan files /home/user/documents --progress
dupectl scan files "C:\Users\user\Documents"
```

## Query Commands

### get duplicates

**Command**: `dupectl get duplicates [flags]`

**Description**: Query and display duplicate files detected across all scanned root folders

**Arguments**: None (queries all roots)

**Flags**:
- `--json` - Output in JSON format (default: human-readable table)
- `--min-count int` - Minimum duplicates per set (default: 2, i.e., 2+ files with same hash)
- `--help` or `-h` - Display command help

**Behavior**:
1. Query database for duplicate file sets (size + hash match)
2. Filter by `--min-count` if specified
3. Format output according to `--json` flag
4. Display results or "No duplicates found"

**Output Format** (table, default):
```text
Duplicate Files
═══════════════

Set 1 (3 files, 1.5 MB each):
  Hash: a3f5d8... (SHA-512)
  Files:
    /home/user/docs/report.pdf
    /home/user/backup/report.pdf
    /media/external/reports/report.pdf

Set 2 (2 files, 524 KB each):
  Hash: b7e2c1... (SHA-512)
  Files:
    /home/user/photos/IMG_001.jpg
    /home/user/archive/photos/IMG_001.jpg

Total: 5 duplicate files in 2 sets
Potential space savings: 3.5 MB
```

**Output Format** (JSON, `--json`):
```json
{
  "duplicate_sets": [
    {
      "hash": "a3f5d8...",
      "algorithm": "sha512",
      "size": 1572864,
      "count": 3,
      "files": [
        "/home/user/docs/report.pdf",
        "/home/user/backup/report.pdf",
        "/media/external/reports/report.pdf"
      ]
    },
    {
      "hash": "b7e2c1...",
      "algorithm": "sha512",
      "size": 536576,
      "count": 2,
      "files": [
        "/home/user/photos/IMG_001.jpg",
        "/home/user/archive/photos/IMG_001.jpg"
      ]
    }
  ],
  "summary": {
    "total_duplicate_files": 5,
    "total_duplicate_sets": 2,
    "potential_space_savings_bytes": 3670016
  }
}
```

**Error Cases**:
- No scans completed → Exit 1 with "Error: No scans found. Run 'dupectl scan all' first."
- Invalid --min-count value → Exit 2 with "Error: --min-count must be >= 2"

**Example Usage**:
```bash
dupectl get duplicates
dupectl get duplicates --json
dupectl get duplicates --min-count=5  # Only show sets with 5+ files
```

### get root

**Command**: `dupectl get root [flags]`

**Description**: List all registered root folders with scan statistics

**Arguments**: None

**Flags**:
- `--json` - Output in JSON format (default: human-readable table)
- `--help` or `-h` - Display command help

**Behavior**:
1. Query database for all root folders
2. Retrieve cached summary statistics (folder_count, file_count, total_size, last_scan_date)
3. Format output according to `--json` flag
4. Display results or "No root folders registered"

**Output Format** (table, default):
```text
Root Folders
════════════

Path                    Folders    Files      Total Size    Last Scan
──────────────────────────────────────────────────────────────────────
/home/user/documents      1,543     12,847     5.0 GB       2025-12-23 10:03:20 UTC
/media/external           423       3,241      1.2 GB       2025-12-22 14:15:00 UTC
C:\Users\user\Downloads   89        452        512 MB       Never scanned

Total: 3 root folders
```

**Output Format** (JSON, `--json`):
```json
{
  "root_folders": [
    {
      "path": "/home/user/documents",
      "folder_count": 1543,
      "file_count": 12847,
      "total_size_bytes": 5368709120,
      "last_scan_date": "2025-12-23T10:03:20Z"
    },
    {
      "path": "/media/external",
      "folder_count": 423,
      "file_count": 3241,
      "total_size_bytes": 1288490188,
      "last_scan_date": "2025-12-22T14:15:00Z"
    },
    {
      "path": "C:\\Users\\user\\Downloads",
      "folder_count": 89,
      "file_count": 452,
      "total_size_bytes": 536870912,
      "last_scan_date": null
    }
  ],
  "summary": {
    "total_roots": 3
  }
}
```

**Error Cases**: None (returns empty list if no roots)

**Example Usage**:
```bash
dupectl get root
dupectl get root --json
```

## Management Commands

### add root

**Command**: `dupectl add root <root-folder-path> [flags]`

**Description**: Register a new root folder for monitoring

**Arguments**:
- `<root-folder-path>` (required) - Absolute or relative path to root folder

**Flags**:
- `--traverse-links` - Follow symbolic links during scans (default: false)
- `--help` or `-h` - Display command help

**Behavior**:
1. Convert relative path to absolute path
2. Validate path exists on filesystem
3. Check if path already registered
4. Insert root folder record with configuration
5. Display confirmation

**Output Format**:
```text
Root folder registered: /home/user/documents
Configuration:
  Traverse Links: false
  
Run 'dupectl scan all /home/user/documents' to start scanning.
```

**Error Cases**:
- Root folder path does not exist → Exit 1 with "Error: Root folder does not exist: /path/to/root"
- Root folder already registered → Exit 1 with "Error: Root folder already registered: /path/to/root"
- Permission denied → Exit 1 with "Error: Permission denied accessing root folder: /path/to/root"

**Example Usage**:
```bash
dupectl add root /home/user/documents
dupectl add root "C:\Users\user\Documents" --traverse-links
```

### delete root

**Command**: `dupectl delete root <root-folder-path> [flags]`

**Description**: Remove registered root folder and delete all associated scan data

**Arguments**:
- `<root-folder-path>` (required) - Absolute or relative path to root folder

**Flags**:
- `--yes` or `-y` - Skip confirmation prompt
- `--help` or `-h` - Display command help

**Behavior**:
1. Convert relative path to absolute path
2. Check if root folder is registered
3. Prompt for confirmation (unless `--yes` specified): "Delete root folder and all scan data? (y/n)"
4. If confirmed → delete root folder record (CASCADE deletes folders, files, scan_state)
5. Display confirmation

**Output Format**:
```text
Root folder deleted: /home/user/documents
Removed: 1,543 folders, 12,847 files, 5.0 GB of scan data
```

**Error Cases**:
- Root folder not registered → Exit 1 with "Error: Root folder not registered: /path/to/root"
- User cancels confirmation → Exit 0 with "Deletion cancelled."

**Example Usage**:
```bash
dupectl delete root /home/user/documents
dupectl delete root "C:\Users\user\Documents" --yes
```

## Help Text

All commands must include comprehensive `--help` output with:
- Command description
- Usage syntax
- Argument descriptions
- Flag descriptions with default values
- Examples (minimum 2)
- Related commands

**Example** (`dupectl scan all --help`):
```text
Scan all files and folders recursively, calculating hashes for duplicate detection.

This command traverses the folder tree starting from the specified root folder,
registers all folders and files in the database, and calculates hash values for
each file to enable duplicate detection.

Usage:
  dupectl scan all <root-folder-path> [flags]

Arguments:
  root-folder-path    Path to root folder (absolute or relative)

Flags:
  --progress          Display real-time progress (spinner, counts, elapsed time)
  --restart           Restart scan from beginning instead of resuming from checkpoint
  -h, --help          Display this help message

Examples:
  # Scan with progress display
  dupectl scan all /home/user/documents --progress

  # Restart scan from beginning
  dupectl scan all "C:\Users\user\Documents" --restart

  # Scan with both flags
  dupectl scan all ../relative/path --progress --restart

Related Commands:
  dupectl scan folders    Scan folder structure only (faster)
  dupectl scan files      Hash files only (assumes folders registered)
  dupectl get duplicates  Query duplicate files
```

## Progress Display Format

**Braille Spinner Animation** (8 frames, cycles every 800ms):
```text
⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧
```

**Progress Line Format**:
```text
{spinner} Folders: {count} | Files: {count} | Elapsed: {duration}
```

**Update Frequency**: Every 10 seconds (configurable via `progress_interval` in config file)

**Example**:
```text
⠋ Folders: 823 | Files: 6,420 | Elapsed: 1m 15s
```

**On Completion**: Replace spinner with checkmark
```text
✓ Scan completed in 2m 34s
```

**On Error**: Replace spinner with X
```text
✗ Scan failed: permission denied on /path/to/folder
```

## Signal Handling

**SIGINT (Ctrl+C)** and **SIGTERM**:
1. Display message: "Shutting down gracefully, saving checkpoint..."
2. Cancel context → stop all workers
3. Wait for in-flight operations to complete (5 second timeout)
4. Save checkpoint to database
5. Display message: "Checkpoint saved. Run 'dupectl scan all <path>' to resume."
6. Exit with code 0

**Example**:
```text
⠋ Folders: 823 | Files: 6,420 | Elapsed: 1m 15s
^C
Shutting down gracefully, saving checkpoint...
Checkpoint saved at folder: /home/user/documents/subfolder/path
Run 'dupectl scan all /home/user/documents' to resume.
```

## Exit Codes

Consistent across all commands (Constitution IV: UX Consistency):

| Code | Meaning | Examples |
|------|---------|----------|
| 0 | Success | Command completed successfully, user cancelled confirmation |
| 1 | General error | File not found, permission denied, database error |
| 2 | Usage error | Invalid arguments, invalid flag values, missing required argument |

## Color Coding

**Colors** (must work without colors if terminal doesn't support):
- Green: Success messages, checkmark (✓)
- Red: Error messages, X (✗)
- Yellow: Warning messages, prompts
- Cyan: Progress spinner
- White: Normal output

**Example**:
```text
✓ Scan completed in 2m 34s          [GREEN]
Folders scanned: 1,543               [WHITE]
Error: Permission denied             [RED]
Register root folder now? (y/n)     [YELLOW]
```

## Configuration File

**File**: `~/.dupectl.yaml` (or path specified by `--config`)

**Scan Configuration**:
```yaml
scan:
  hash_algorithm: sha512          # sha256, sha512, sha3-256
  worker_count: 8                 # Number of parallel workers (default: CPU cores)
  progress_interval: 10           # Seconds between progress updates (default: 10)
  checkpoint_interval: 60         # Seconds between periodic checkpoints (default: 60)
```

**Validation**:
- `hash_algorithm` must be one of: sha256, sha512, sha3-256
- `worker_count` must be >= 1 and <= 100
- `progress_interval` must be >= 1
- `checkpoint_interval` must be >= 10

## References

- Feature Specification: `specs/001-duplicate-scan-system/spec.md` (FR-023, FR-024)
- DupeCTL Constitution: `.specify/memory/constitution.md` (Principle IV: UX Consistency)
- Cobra CLI Documentation: https://github.com/spf13/cobra
