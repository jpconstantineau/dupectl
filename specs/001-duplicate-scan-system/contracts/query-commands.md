# CLI Contract: Query Commands

**Feature**: 001-duplicate-scan-system  
**Date**: December 23, 2025  
**Purpose**: Command-line interface contracts for querying duplicate detection results

## Command: `dupectl get duplicates`

**Purpose**: Query and display duplicate files identified during scans

**Syntax**:
```bash
dupectl get duplicates [flags]
```

**Arguments**: None (queries all duplicates across all root folders)

**Flags**:
- `--json` (optional): Output in JSON format for scripting (default: human-readable table)
- `--min-count <N>` (optional): Show only duplicate sets with at least N files (default: 2)
- `--root <path>` (optional): Filter duplicates to specific root folder path
- `--min-size <bytes>` (optional): Show only duplicates with file size >= bytes
- `--sort <field>` (optional): Sort by 'size' (default), 'count', or 'path'

**Behavior**:
1. Query database for duplicate files (size + hash match)
2. Group files by (size, hash_value, hash_algorithm)
3. Filter by --min-count (default 2+)
4. Apply optional filters (root, min-size)
5. Sort by specified field
6. Format output (table or JSON)
7. Display results with summary statistics

**Output** (default table format):
```
Duplicate Files Report
======================

Duplicate Set #1 (Size: 10.5 MB, Hash: sha256:abc123..., Algorithm: sha256)
  File Count: 3
  ├─ /home/user/documents/photo1.jpg
  ├─ /home/user/backup/photo1.jpg
  └─ /media/archive/photos/photo1.jpg

Duplicate Set #2 (Size: 2.3 KB, Hash: sha256:def456..., Algorithm: sha256)
  File Count: 2
  ├─ /home/user/docs/readme.md
  └─ /backup/docs/readme.md

Duplicate Set #3 (Size: 154.2 MB, Hash: sha512:789abc..., Algorithm: sha512)
  File Count: 5
  ├─ /data/videos/movie.mp4
  ├─ /backup1/videos/movie.mp4
  ├─ /backup2/videos/movie.mp4
  ├─ /archive/media/movie.mp4
  └─ /external/movie.mp4

────────────────────────────────────────────────────────────────
Summary:
  Total Duplicate Sets: 3
  Total Duplicate Files: 10
  Total Wasted Space: 477.2 MB (across all copies)
  Storage Recoverable: 322.2 MB (if keeping 1 copy per set)
```

**Output** (JSON format with `--json`):
```json
{
  "duplicate_sets": [
    {
      "size": 11010048,
      "hash_value": "sha256:abc123...",
      "hash_algorithm": "sha256",
      "file_count": 3,
      "files": [
        {
          "id": 123,
          "path": "/home/user/documents/photo1.jpg",
          "name": "photo1.jpg",
          "mtime": "2025-12-20T10:30:00Z",
          "scanned_at": "2025-12-23T08:15:00Z"
        },
        {
          "id": 456,
          "path": "/home/user/backup/photo1.jpg",
          "name": "photo1.jpg",
          "mtime": "2025-12-20T10:30:00Z",
          "scanned_at": "2025-12-23T08:16:00Z"
        },
        {
          "id": 789,
          "path": "/media/archive/photos/photo1.jpg",
          "name": "photo1.jpg",
          "mtime": "2025-12-20T10:30:00Z",
          "scanned_at": "2025-12-23T08:17:00Z"
        }
      ]
    },
    {
      "size": 2355,
      "hash_value": "sha256:def456...",
      "hash_algorithm": "sha256",
      "file_count": 2,
      "files": [
        {
          "id": 234,
          "path": "/home/user/docs/readme.md",
          "name": "readme.md",
          "mtime": "2025-12-15T14:22:00Z",
          "scanned_at": "2025-12-23T08:18:00Z"
        },
        {
          "id": 567,
          "path": "/backup/docs/readme.md",
          "name": "readme.md",
          "mtime": "2025-12-15T14:22:00Z",
          "scanned_at": "2025-12-23T08:19:00Z"
        }
      ]
    }
  ],
  "summary": {
    "total_sets": 2,
    "total_files": 5,
    "total_wasted_bytes": 33030499,
    "recoverable_bytes": 22020331
  }
}
```

**Error Handling**:
- No scans performed: "Error: No files scanned yet. Run 'dupectl scan all <path>' first."
- No duplicates found: "No duplicate files found. All files are unique."
- Invalid --root path: "Error: Root folder '/invalid/path' not found in database."
- Invalid --min-count: "Error: --min-count must be >= 2"
- Database error: "Error: Database query failed: <details>"

**Exit Codes**:
- 0: Success (duplicates found and displayed)
- 1: Error (database error, invalid arguments)
- 0: No duplicates found (not an error condition)

**Examples**:
```bash
# Show all duplicates
dupectl get duplicates

# Show only files with 5+ duplicates
dupectl get duplicates --min-count 5

# Show duplicates as JSON for scripting
dupectl get duplicates --json

# Show large file duplicates (>10 MB)
dupectl get duplicates --min-size 10485760

# Show duplicates for specific root folder
dupectl get duplicates --root /home/user/documents

# Sort by duplicate count (most duplicated first)
dupectl get duplicates --sort count

# Combine filters
dupectl get duplicates --min-count 3 --min-size 1048576 --json
```

---

## Command: `dupectl get folders` (future enhancement)

**Purpose**: Query duplicate folders (exact and partial matches)

**Syntax**:
```bash
dupectl get folders [flags]
```

**Flags**:
- `--json` (optional): JSON output
- `--exact` (optional): Show only exact folder duplicates (100% match)
- `--partial` (optional): Show partial duplicates (default: both exact and partial)
- `--min-similarity <percent>` (optional): Minimum similarity for partial matches (default: 50%)
- `--root <path>` (optional): Filter to specific root folder

**Behavior** (future implementation):
1. Query folders and their file lists
2. Compute folder signatures (exact matches)
3. Compute similarity scores (partial matches)
4. Filter by similarity threshold
5. Format output

**Output** (conceptual - not implemented in this phase):
```
Duplicate Folders Report
========================

Exact Folder Duplicates:
  Set #1 (Folders: 2, Files per folder: 1,234, Total size: 5.2 GB)
  ├─ /home/user/projects/app-v1
  └─ /backup/projects/app-v1

Partial Folder Duplicates:
  Match #1 (Similarity: 85%, Folder1: 100 files, Folder2: 95 files)
  ├─ Folder 1: /home/user/documents/2024
  └─ Folder 2: /backup/documents/2024
    Matching files: 85
    Missing in Folder 1: 5 files (list.txt, ...)
    Missing in Folder 2: 10 files (archive.zip, ...)

Summary:
  Exact duplicate folder sets: 1
  Partial duplicate pairs: 1
```

**Note**: Folder duplicate detection is marked as Priority P3 and may be implemented in future iterations. This contract is included for completeness but not required for initial implementation.

---

## Command: `dupectl get` (parent command)

**Purpose**: Display get subcommands and usage

**Syntax**:
```bash
dupectl get
```

**Behavior**: Display help text with available subcommands

**Output**:
```
Get information about scanned files and duplicates

Usage:
  dupectl get [command]

Available Commands:
  duplicates  Show duplicate files
  root        Get list of root folders (existing)
  agent       Get list of agents (existing)
  host        Get list of hosts (existing)
  owner       Get list of owners (existing)
  policy      Get list of policies (existing)
  purpose     Get list of purposes (existing)

Flags:
  -h, --help   help for get

Use "dupectl get [command] --help" for more information about a command.
```

---

## Query Workflow Patterns

### Pattern 1: Basic Duplicate Detection
```bash
# Scan and view all duplicates
dupectl scan all /my/data
dupectl get duplicates
```

### Pattern 2: Filter by Size (Find Large Duplicates)
```bash
# Find duplicate files >100 MB
dupectl get duplicates --min-size 104857600
```

### Pattern 3: Export for Scripting
```bash
# Generate JSON report for automated processing
dupectl get duplicates --json > duplicates-report.json

# Parse with jq
cat duplicates-report.json | jq '.duplicate_sets[] | select(.file_count > 5)'
```

### Pattern 4: Focus on High-Duplication Files
```bash
# Show only files with 10+ copies
dupectl get duplicates --min-count 10
```

### Pattern 5: Per-Root Analysis
```bash
# Analyze duplicates within specific folder
dupectl get duplicates --root /home/user/documents
```

---

## Output Formatting Guidelines

### Human-Readable Table Format

**Design Principles**:
- Clear hierarchy with indentation (tree-like structure)
- Size in human-readable units (KB, MB, GB)
- Truncated hash values (first 16 chars for readability)
- Color coding (if terminal supports):
  - Headers: Bold
  - File paths: Normal
  - Summary: Cyan
  - Warnings: Yellow

**Size Formatting**:
```
Bytes         Display
1,024         1.0 KB
1,048,576     1.0 MB
10,485,760    10.0 MB
1,073,741,824 1.0 GB
```

**Hash Display**:
```
Full: sha256:abc123def456789012345678901234567890123456789012345678901234
Display: sha256:abc123def456...
```

### JSON Format

**Schema**:
```typescript
interface DuplicateReport {
  duplicate_sets: DuplicateSet[];
  summary: Summary;
}

interface DuplicateSet {
  size: number;                 // Bytes
  hash_value: string;           // Full hash with algorithm prefix
  hash_algorithm: string;       // sha256, sha512, sha3-256
  file_count: number;
  files: File[];
}

interface File {
  id: number;
  path: string;                 // Absolute path
  name: string;
  mtime: string;                // ISO 8601 timestamp
  scanned_at: string;           // ISO 8601 timestamp
}

interface Summary {
  total_sets: number;
  total_files: number;
  total_wasted_bytes: number;   // Total duplicate file size
  recoverable_bytes: number;    // Space if keeping 1 copy per set
}
```

---

## Performance Considerations

**Query Optimization**:
- Use composite index on (hash_value, size) for duplicate detection
- Limit result sets to top N (e.g., 1000 duplicate sets max)
- Stream results for large datasets (avoid loading all into memory)

**Expected Performance**:
- 100k files in database: <1 second for duplicate query
- 1M files in database: <5 seconds for duplicate query
- JSON serialization: <100ms for typical result sets

**Memory Usage**:
- Table format: Minimal (streaming output)
- JSON format: ~1 KB per duplicate set × number of sets

---

## Future Enhancements (Out of Scope for v1)

### Potential Flags (Not Implemented):
- `--format csv`: CSV output for spreadsheet import
- `--group-by root`: Group duplicates by root folder
- `--since <date>`: Show duplicates scanned after date
- `--exclude <pattern>`: Exclude paths matching pattern

### Potential Commands (Not Implemented):
- `dupectl get folders`: Folder duplicate detection (P3 priority)
- `dupectl get stats`: Statistics dashboard (scan coverage, duplicate ratios)
- `dupectl get errors`: Files with permission denied errors
- `dupectl get recent`: Recently scanned files

These enhancements can be added in future iterations based on user feedback and priority reassessment.
