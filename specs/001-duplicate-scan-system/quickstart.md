# Quickstart Guide: Duplicate Scan System

**Feature**: 001-duplicate-scan-system  
**Date**: December 23, 2025  
**Purpose**: User guide for getting started with duplicate file detection

## Overview

DupeCTL's duplicate scan system helps you identify and analyze duplicate files across your filesystem. The system uses cryptographic hashing (SHA-256 by default) to detect files with identical content, even if they have different names or locations.

**Key Features**:
- ✅ Three scan modes: complete, folders-only, or files-only
- ✅ Automatic scan resumption after interruptions
- ✅ Progress indication during long-running scans
- ✅ Cross-platform support (Windows, Linux, macOS)
- ✅ Flexible output formats (human-readable table or JSON)
- ✅ Configurable hash algorithms (SHA-256, SHA-512, SHA3-256)

---

## Quick Start (5 Minutes)

### Step 1: Scan Your Files

Perform a complete scan of a folder to identify duplicates:

```bash
dupectl scan all /path/to/your/documents
```

**What happens**:
- System recursively scans folder structure
- Calculates cryptographic hash for each file
- Stores results in database for later analysis
- Shows progress every 10 seconds

**Example Output**:
```
Scanning: /home/user/documents
Progress: 1,234 files, 156 folders processed (10s elapsed)
Progress: 2,890 files, 312 folders processed (20s elapsed)
Scan complete: 5,678 files, 423 folders in 1m 45s
Duplicates found: 89 files in 12 duplicate sets
```

### Step 2: View Duplicate Files

Display the duplicates that were found:

```bash
dupectl get duplicates
```

**Example Output**:
```
Duplicate Files Report
======================

Duplicate Set #1 (Size: 10.5 MB, Hash: sha256:abc123...)
  File Count: 3
  ├─ /home/user/documents/photo1.jpg
  ├─ /home/user/backup/photo1.jpg
  └─ /media/archive/photos/photo1.jpg

Duplicate Set #2 (Size: 2.3 KB, Hash: sha256:def456...)
  File Count: 2
  ├─ /home/user/docs/readme.md
  └─ /backup/docs/readme.md

Summary:
  Total Duplicate Sets: 2
  Total Duplicate Files: 5
  Storage Recoverable: 22.1 MB
```

**Done!** You've identified all duplicate files in your folder.

---

## Common Workflows

### Workflow 1: Complete Scan (Recommended)

**Use Case**: First-time scan or periodic full analysis

```bash
# Single command scans everything
dupectl scan all /my/data

# View results
dupectl get duplicates
```

**Time Estimate**: ~1-2 minutes per 10,000 files (varies by file size and disk speed)

---

### Workflow 2: Large Archives (Deferred File Hashing)

**Use Case**: Very large folder trees where you want to map structure first

```bash
# Step 1: Quick folder structure scan (very fast)
dupectl scan folders /large/archive
# Time: Seconds to minutes depending on folder count

# Step 2: Hash files later (or during off-hours)
dupectl scan files /large/archive
# Time: Most time-consuming step

# Step 3: View duplicates
dupectl get duplicates
```

**Why split the scan?**
- Folder scan is very fast (5000+ folders/sec)
- File hashing takes longer (depends on file size)
- Allows you to plan when to run expensive file hashing

---

### Workflow 3: Incremental Updates

**Use Case**: You've already scanned files, want to re-scan after changes

```bash
# Initial full scan
dupectl scan all /documents

# ... time passes, files are added/modified ...

# Re-hash files only (skip folder traversal)
dupectl scan files /documents --rescan

# View updated duplicates
dupectl get duplicates
```

**Benefits**:
- Faster than full re-scan
- Only processes files (skips folder traversal)
- Updates hashes for modified files

---

### Workflow 4: Recover from Interruption

**Use Case**: Scan was interrupted (Ctrl+C, system crash, etc.)

```bash
# Start a scan (gets interrupted)
dupectl scan all /huge/dataset
# ... after processing 50,000 files ...
^C
Interrupt received. Saving checkpoint...
Checkpoint saved at: /huge/dataset/subfolder/path

# Resume automatically - just run same command again
dupectl scan all /huge/dataset
# Output: Resuming interrupted scan from checkpoint...
```

**Key Points**:
- Checkpoints saved at folder boundaries
- Already-scanned files are skipped
- No need to start over from beginning

---

## Advanced Usage

### Filter Duplicates by File Count

Show only files with many duplicates (e.g., 5+ copies):

```bash
dupectl get duplicates --min-count 5
```

**Use Case**: Focus on most duplicated files first

---

### Filter Duplicates by Size

Show only large file duplicates (e.g., >10 MB):

```bash
dupectl get duplicates --min-size 10485760
```

**Use Case**: Prioritize large files for maximum space recovery

---

### Export Results to JSON

Generate machine-readable report for scripting:

```bash
dupectl get duplicates --json > report.json
```

**Use Case**: Automate duplicate analysis, integrate with other tools

**Example Processing with jq**:
```bash
# Find files with 10+ duplicates
cat report.json | jq '.duplicate_sets[] | select(.file_count >= 10)'

# Calculate total wasted space
cat report.json | jq '.summary.total_wasted_bytes'
```

---

### Verbose Logging

See detailed operation logs during scan:

```bash
dupectl scan all /data --verbose
```

**Output Includes**:
```
[VERBOSE] Processing folder: /data/photos
[VERBOSE] Hashing file: /data/photos/img1.jpg (5.2 MB)
[VERBOSE] Hash: sha256:abc123... (completed in 52ms)
[WARNING] Permission denied: /data/.ssh/id_rsa
```

**Use Case**: Debugging, understanding what's happening, troubleshooting issues

---

## Configuration

### Change Hash Algorithm

Edit `~/.dupectl.yaml` (or create if doesn't exist):

```yaml
scan:
  hash_algorithm: "sha512"        # Options: sha256, sha512, sha3-256
  progress_interval: "10s"        # How often to show progress
  batch_size: 1000                # Files per database transaction
```

**Hash Algorithm Comparison**:
- **sha256** (default): Fast, widely used, excellent security
- **sha512**: Stronger security, slightly slower
- **sha3-256**: Latest SHA standard, future-proof, moderate speed

**Note**: Changing algorithm requires re-scanning files

---

### Adjust Progress Update Frequency

Show progress more/less frequently:

```yaml
scan:
  progress_interval: "5s"   # Update every 5 seconds (more frequent)
  # or
  progress_interval: "30s"  # Update every 30 seconds (less frequent)
```

---

## Understanding the Output

### Scan Progress Format

```
Progress: <files> files, <folders> folders processed (<time> elapsed)
```

**Example**:
```
Progress: 5,432 files, 234 folders processed (1m 30s elapsed)
```

**Fields**:
- `files`: Number of files processed (hashed or registered)
- `folders`: Number of folders traversed
- `time`: Time since scan started

---

### Duplicate Set Format

```
Duplicate Set #<number> (Size: <size>, Hash: <algorithm>:<hash>)
  File Count: <count>
  ├─ <path1>
  ├─ <path2>
  └─ <pathN>
```

**Example**:
```
Duplicate Set #1 (Size: 10.5 MB, Hash: sha256:abc123...)
  File Count: 3
  ├─ /home/user/photo.jpg
  ├─ /backup/photo.jpg
  └─ /archive/photo.jpg
```

**Key Information**:
- **Size**: How large each duplicate file is
- **Hash**: Cryptographic fingerprint (files with same hash are identical)
- **File Count**: How many copies exist
- **Paths**: Where each duplicate is located

---

### Summary Statistics

```
Summary:
  Total Duplicate Sets: <N>          # Number of unique files with duplicates
  Total Duplicate Files: <M>         # Total duplicate file records
  Total Wasted Space: <bytes>        # Sum of all duplicate file sizes
  Storage Recoverable: <bytes>       # Space if keeping 1 copy per set
```

**Example**:
```
Summary:
  Total Duplicate Sets: 12
  Total Duplicate Files: 45
  Total Wasted Space: 234.5 MB
  Storage Recoverable: 187.6 MB
```

**Calculation**:
- Total Wasted = Sum of all duplicate file sizes
- Recoverable = Total Wasted - (1 copy per unique file)

---

## Troubleshooting

### Problem: "Root folder not registered" Prompt

**Symptom**:
```
Root folder '/path/to/scan' is not registered in the database.
Register this root folder now? (y/n):
```

**Solution**: Type `y` to register the folder and proceed with scan

**Alternative**: Pre-register with `dupectl add root /path/to/scan`

---

### Problem: Permission Denied Errors

**Symptom**:
```
[WARNING] Permission denied: /home/user/.ssh/id_rsa
```

**Behavior**:
- Warning shown in console
- File marked with error status in database
- Scan continues with remaining files

**Solution Options**:
1. Run scan with elevated permissions: `sudo dupectl scan all /path`
2. Accept warnings and skip inaccessible files
3. Change file permissions: `chmod +r <file>`

**Note**: Permission errors don't stop the scan

---

### Problem: Scan Taking Too Long

**Symptoms**:
- Scan running for hours
- Progress seems slow

**Solutions**:

1. **Use folder-only scan first**:
   ```bash
   dupectl scan folders /path  # Fast structure mapping
   # Review folder count, then decide if full scan needed
   ```

2. **Check file count**:
   - 10,000 files: ~2-5 minutes expected
   - 100,000 files: ~20-50 minutes expected
   - 1,000,000 files: ~3-8 hours expected

3. **Verify disk isn't bottleneck**:
   - Network drives are much slower
   - Local SSD is fastest

4. **Use faster hash algorithm**:
   - SHA-256 fastest (default, recommended)

---

### Problem: "No duplicates found"

**Symptom**:
```
No duplicate files found. All files are unique.
```

**Possible Reasons**:
1. **Actually no duplicates** - all files are truly unique ✓
2. **Files haven't been hashed yet** - ran `scan folders` but not `scan files`
3. **Files modified** - duplicates existed but files were changed

**Solutions**:
- Ensure full scan completed: `dupectl scan all /path`
- Check database has hashed files: `dupectl get duplicates --json | jq '.summary'`

---

### Problem: Scan Interrupted, Can't Resume

**Symptom**:
- Scan interrupted
- Resume doesn't work as expected

**Solution**:
1. Check scan state:
   ```bash
   # Look for 'interrupted' scans in database
   sqlite3 dupedb.db "SELECT * FROM scan_state WHERE status='interrupted';"
   ```

2. Manually reset if needed:
   ```bash
   # Delete interrupted scan state to start fresh
   sqlite3 dupedb.db "DELETE FROM scan_state WHERE status='interrupted';"
   dupectl scan all /path
   ```

---

## Performance Tips

### Tip 1: Scan During Off-Hours

For very large scans, run during low-activity times:
```bash
# Linux/macOS: Schedule for 2 AM
echo "dupectl scan all /archive" | at 02:00

# Windows: Use Task Scheduler
schtasks /create /tn "DupeScan" /tr "dupectl scan all C:\Archive" /sc once /st 02:00
```

### Tip 2: Scan External Drives When Locally Attached

- Network shares: 10-100 MB/sec (slow)
- USB 3.0: 100-500 MB/sec (medium)
- Local SSD: 500-3000 MB/sec (fast)

**Recommendation**: For external drives, attach directly via USB for fastest scan

### Tip 3: Split Large Archives

Scan subdirectories separately:
```bash
dupectl scan all /archive/2020
dupectl scan all /archive/2021
dupectl scan all /archive/2022
# Then query all duplicates together
dupectl get duplicates
```

**Benefits**:
- Shorter individual scan times
- Can prioritize certain folders
- Easier to resume if interrupted

---

## Next Steps

### Explore More Commands

```bash
# View help
dupectl scan --help
dupectl get duplicates --help

# List registered root folders
dupectl get root

# View configuration
cat ~/.dupectl.yaml
```

### Automate Regular Scans

Create a weekly duplicate detection job:

```bash
# Linux/macOS: Add to crontab
0 2 * * 0 /usr/local/bin/dupectl scan all /home/user/documents

# Windows: Task Scheduler with weekly trigger
```

### Integration with Scripts

Use JSON output for automated processing:

```bash
#!/bin/bash
# Example: Report duplicates via email

dupectl scan all /data
dupectl get duplicates --json > /tmp/report.json

# Parse and email if duplicates found
COUNT=$(cat /tmp/report.json | jq '.summary.total_sets')
if [ "$COUNT" -gt 0 ]; then
    mail -s "Duplicates Found: $COUNT sets" admin@example.com < /tmp/report.json
fi
```

---

## FAQ

**Q: Does scanning modify my files?**  
A: No, scanning is read-only. Files are only read for hashing, never modified.

**Q: How much disk space does the database use?**  
A: ~250 bytes per file. For 100,000 files: ~25 MB database size.

**Q: Can I scan multiple root folders?**  
A: Yes, scan each root separately. `get duplicates` will show duplicates across all scanned roots.

**Q: What if I have files with same content but different names?**  
A: Duplicate detection is based on content (hash), not names. Files with same content but different names are correctly identified as duplicates.

**Q: How accurate is duplicate detection?**  
A: Cryptographic hashing (SHA-256) provides effectively 100% accuracy. Collision probability is negligible (< 10^-60).

**Q: Can I delete duplicates automatically?**  
A: Not in v1. This version focuses on detection only. Future versions may add remediation features (delete, symlink, etc.).

**Q: What happens if files change between scans?**  
A: Re-scan updates the hash. Old hash is replaced with new one. Changed files will no longer match as duplicates.

---

## Getting Help

**View Command Help**:
```bash
dupectl --help
dupectl scan --help
dupectl scan all --help
dupectl get duplicates --help
```

**Report Issues**:
- Check verbose output: `dupectl scan all /path --verbose`
- Review logs and error messages
- Check database: `sqlite3 dupedb.db ".tables"`

---

## Summary

**Essential Commands**:
- `dupectl scan all <path>` - Complete scan (folders + files)
- `dupectl get duplicates` - View duplicates
- `dupectl scan folders <path>` - Folder structure only
- `dupectl scan files <path>` - File hashing only

**Best Practices**:
- Start with small folder for testing
- Use `scan all` for most use cases
- Split scan modes only for very large archives
- Export JSON for automation
- Configure hash algorithm before first scan

Happy duplicate hunting! 🔍
