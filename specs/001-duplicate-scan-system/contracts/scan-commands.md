# CLI Contract: Scan Commands

**Feature**: 001-duplicate-scan-system  
**Date**: December 23, 2025  
**Purpose**: Command-line interface contracts for scan operations

## Command: `dupectl scan all <root-path>`

**Purpose**: Perform complete scan of folder structure and file hashing for duplicate detection

**Syntax**:
```bash
dupectl scan all <root-path> [flags]
```

**Arguments**:
- `<root-path>` (required): Absolute or relative path to root folder to scan
  - Relative paths converted to absolute before processing
  - Must exist on filesystem
  - Will prompt for registration if not in database

**Flags**:
- `--resume` (optional): Resume interrupted scan from last checkpoint (default: auto-detect)
- `--restart` (optional): Restart interrupted scan from root folder
- `--verbose` (optional): Enable detailed logging to console

**Behavior**:
1. Validate root-path exists on filesystem
2. Convert relative path to absolute path
3. Check if root-path registered in database:
   - If not registered: Prompt "Root folder not registered. Register now? (y/n)"
   - If user confirms: Register root folder, then proceed
   - If user declines: Exit with error code 2
4. Check for existing scan_state with 'running' or 'interrupted' status:
   - If found and --resume not explicitly disabled: Resume from checkpoint
   - Otherwise: Start new scan
5. Recursively traverse folder tree:
   - Register each folder in folders table
   - Hash each file and store in files table
   - Save checkpoint every folder
   - Report progress every 10 seconds (or configured interval)
6. On completion: Update scan_state to 'completed', display summary

**Output** (progress):
```
Scanning: /path/to/root
Progress: 1,234 files, 156 folders processed (10s elapsed)
Progress: 2,890 files, 312 folders processed (20s elapsed)
...
Scan complete: 10,456 files, 1,023 folders in 2m 15s
Duplicates found: 234 files in 45 duplicate sets
```

**Error Handling**:
- Path doesn't exist: "Error: Path /invalid/path does not exist"
- Permission denied on root: "Error: Permission denied accessing /path/to/root"
- Permission denied on subfolder: Warning logged, folder skipped, scan continues
- Permission denied on file: Error status recorded in database, console warning, scan continues
- Database error: "Error: Database operation failed: <details>"

**Exit Codes**:
- 0: Success (scan completed)
- 1: Error (filesystem/database error)
- 2: User declined registration prompt

**Examples**:
```bash
# Scan with relative path
dupectl scan all ./my-documents

# Scan with absolute path
dupectl scan all /home/user/documents

# Resume interrupted scan
dupectl scan all /data/archive --resume

# Verbose output
dupectl scan all /backup --verbose
```

---

## Command: `dupectl scan folders <root-path>`

**Purpose**: Scan folder structure only without file hashing (fast structure mapping)

**Syntax**:
```bash
dupectl scan folders <root-path> [flags]
```

**Arguments**:
- `<root-path>` (required): Absolute or relative path to root folder

**Flags**:
- `--resume` (optional): Resume from checkpoint (default: auto-detect)
- `--restart` (optional): Restart fron root path
- `--verbose` (optional): Detailed logging

**Behavior**:
1-4. Same validation and registration as `scan all`
5. Recursively traverse folder tree:
   - Register each folder in folders table
   - Skip file hashing (no file records created)
   - Save checkpoint every folder
   - Report progress every 10 seconds
6. On completion: Update scan_state to 'completed'

**Output**:
```
Scanning folders: /path/to/root
Progress: 156 folders processed (10s elapsed)
Progress: 312 folders processed (20s elapsed)
...
Folder scan complete: 1,023 folders in 45s
Note: Run 'dupectl scan files <root-path>' to hash file contents
```

**Error Handling**: Same as `scan all`

**Exit Codes**: Same as `scan all`

**Examples**:
```bash
# Quick folder structure mapping
dupectl scan folders /large/archive

# Resume folder scan
dupectl scan folders /data --resume
```

---

## Command: `dupectl scan files <root-path>`

**Purpose**: Hash files in previously scanned folders (incremental file scanning)

**Syntax**:
```bash
dupectl scan files <root-path> [flags]
```

**Arguments**:
- `<root-path>` (required): Root folder path (must have folders already scanned)

**Flags**:
- `--resume` (optional): Resume from checkpoint (default: auto-detect)
- `--restart` (optional): Restart from root folder
- `--verbose` (optional): Detailed logging
- `--rescan` (optional): Re-hash files even if already hashed (update hashes)

**Behavior**:
1-4. Same validation as other scan commands
5. Query folders table for all folders under root:
   - If no folders found: Error "No folders found. Run 'dupectl scan folders <root-path>' first"
6. For each folder:
   - List files in folder
   - Hash each file (or skip if already hashed and --rescan not set)
   - Store/update file record in database
   - Save checkpoint every folder
   - Report progress every 10 seconds
7. On completion: Update scan_state to 'completed'

**Output**:
```
Scanning files: /path/to/root
Progress: 1,234 files processed, 89 duplicates found (10s elapsed)
Progress: 2,890 files processed, 234 duplicates found (20s elapsed)
...
File scan complete: 10,456 files hashed in 2m 10s
Duplicates found: 234 files in 45 duplicate sets
```

**Error Handling**:
- No folders scanned: "Error: No folders found for /path. Run 'dupectl scan folders' first"
- Permission denied on file: Warning logged, error status set, continue
- Hash calculation error: Error status set, continue

**Exit Codes**: Same as `scan all`

**Examples**:
```bash
# Hash files after folder scan
dupectl scan files /data/archive

# Re-hash all files (update existing hashes)
dupectl scan files /backup --rescan

# Resume file hashing
dupectl scan files /documents --resume
```

---

## Command: `dupectl scan` (parent command)

**Purpose**: Display scan subcommands and usage

**Syntax**:
```bash
dupectl scan
```

**Behavior**: Display help text with available subcommands

**Output**:
```
Scan folders and files for duplicate detection

Usage:
  dupectl scan [command]

Available Commands:
  all         Scan folders and files completely
  folders     Scan folder structure only (no file hashing)
  files       Scan files in previously registered folders

Flags:
  -h, --help   help for scan

Use "dupectl scan [command] --help" for more information about a command.
```

---

## Scan Workflow Patterns

### Pattern 1: Complete Scan (Most Common)
```bash
# Single command does everything
dupectl scan all /my/data
dupectl get duplicates
```

### Pattern 2: Deferred File Hashing (Large Archives)
```bash
# Step 1: Quick folder structure scan
dupectl scan folders /large/archive

# Step 2: Hash files later (or in batches)
dupectl scan files /large/archive
dupectl get duplicates
```

### Pattern 3: Incremental Updates
```bash
# Initial full scan
dupectl scan all /documents

# Later: Re-scan files only (after file changes)
dupectl scan files /documents --rescan
```

### Pattern 4: Resume Interrupted Scan
```bash
# Start scan (interrupted by Ctrl+C or crash)
dupectl scan all /huge/dataset
^C

# Resume automatically
dupectl scan all /huge/dataset
# Output: "Resuming interrupted scan from checkpoint: /huge/dataset/subfolder"
```

---

## Configuration Options

**Configurable via `.dupectl.yaml` or environment variables**:

```yaml
scan:
  hash_algorithm: "sha256"        # Hash algorithm (sha256, sha512, sha3-256)
  progress_interval: "10s"        # Progress update frequency
  batch_size: 1000                # Files per database transaction
  concurrent_hashers: 4           # Parallel hashing goroutines
```

**Environment Variables**:
```bash
DUPECTL_SCAN_HASH_ALGORITHM=sha256
DUPECTL_SCAN_PROGRESS_INTERVAL=10s
DUPECTL_SCAN_BATCH_SIZE=1000
DUPECTL_SCAN_CONCURRENT_HASHERS=4
```

---

## Progress Reporting Format

**Console Output Structure**:
```
Status: <root-path>
Progress: <files> files, <folders> folders processed (<elapsed> elapsed)
[Optional] Duplicates found: <count> files in <sets> duplicate sets
[Optional] Errors: <count> files with permission denied
```

**Example**:
```
Scanning: /home/user/documents
Progress: 5,432 files, 234 folders processed (1m 30s elapsed)
Duplicates found: 89 files in 12 duplicate sets
Errors: 3 files with permission denied
```

**Verbose Mode Additional Output**:
```
[VERBOSE] Processing folder: /home/user/documents/photos
[VERBOSE] Hashing file: /home/user/documents/photo1.jpg (5.2 MB)
[VERBOSE] Hash: sha256:abc123... (completed in 52ms)
[WARNING] Permission denied: /home/user/.ssh/id_rsa
```

---

## Signal Handling

**SIGINT (Ctrl+C)**:
```
^C
Interrupt received. Saving checkpoint...
Checkpoint saved at: /current/folder/path
Scan state: 5,432 files, 234 folders processed
Resume with: dupectl scan all /path/to/root
```

**SIGTERM**:
- Save checkpoint
- Update scan_state status to 'interrupted'
- Exit gracefully with code 0

---

## Registration Prompt Flow

**When root path not registered**:
```
Root folder '/path/to/scan' is not registered in the database.
Register this root folder now? (y/n): _
```

**User enters 'y'**:
```
Registering root folder: /path/to/scan
Root folder registered successfully.
Starting scan...
```

**User enters 'n'**:
```
Scan cancelled. Register root folder with: dupectl add root /path/to/scan
```

---

## Performance Expectations

**Typical Performance** (SSD, modern hardware):
- Folder scan: 5,000+ folders/sec
- File hashing (SHA-256): 100-500 MB/sec (depends on file size, disk speed)
- Database writes: 1,000+ records/sec (batched transactions)

**Progress Estimate**:
```
Scanning: /data/archive (estimated: 50,000 files)
Progress: 10,000 files processed (20% complete, 40s elapsed, ~2m 40s remaining)
```

Note: Time estimates become available after processing at least 1,000 files for statistical accuracy.
