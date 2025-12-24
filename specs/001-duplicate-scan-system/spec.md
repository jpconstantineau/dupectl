# Feature Specification: Duplicate Scan System

**Feature Branch**: `001-duplicate-scan-system`  
**Created**: December 23, 2025  
**Status**: Draft  
**Input**: User description: "the scan set of cli commands either scan all files and folders, scan all folders or scan all files. The folder command of the cli will add the folder to be monitored and start the recursive scan of the folder tree but not start the scanning of the files in each folder. Scanning of the files includes hashing each file and add their hash to the database entries. duplicates files are identified when both their sizes and hashes match. duplicate folders are found when all files within their subtree are identical according to the file matching logic. Partial folder duplicates can be found when only a subset of the files in a folder are found to be identical. The system highlights potential matches when partial folder matches but the key differences are either for missing files from one set to another and/or when files of the same name dont match but have a different date."

## Clarifications

### Session 2025-12-23

- Q: How do users configure hash algorithm choice (SHA-256, SHA-512, SHA3-256)? → A: Global configuration in config file only - all scans use same algorithm
- Q: What is the progress indication update frequency and mechanism? → A: Real-time progress via --progress flag showing spinner, folder/file counts, and elapsed time; updates at configurable intervals (default 10 seconds)
- Q: When scans are interrupted, should system resume from checkpoint or restart from beginning? → A: Resume from last checkpoint - track progress in database and continue where stopped
- Q: What is the minimum similarity threshold for detecting partial folder duplicates? → A: 50% minimum similarity threshold
- Q: Should scan data (files/folders/hashes) use existing agent tables or new dedicated tables? → A: Hybrid - use existing agent/host relationships but separate tables for file/folder/hash data
- Q: How do users specify which root folder to scan when multiple roots are registered? → A: Scan commands require mandatory positional argument for root folder path
- Q: What happens when user scans a root folder path that isn't registered in database? → A: Prompt user for confirmation to register before scanning
- Q: How should system handle relative paths vs absolute paths for root folders? → A: Convert to absolute, store absolute
- Q: How should permission errors be logged during scans? → A: Console only and mark in database for future scans
- Q: What output format should the get duplicates command use? → A: Command line options for JSON or table format

### Session 2025-12-24 (Checklist Review)

**CRITICAL LIFECYCLE GAPS - Require Resolution Before Implementation**

- Q: How can users rehash files that have NULL hash_value (failed initial hash, permission fixed later)? → A: NEEDS CLARIFICATION - Propose: leverage `dupectl scan files <root-path>` command to recalculate hashes for files with NULL hash_value or matching specific criteria (error status, hash algorithm mismatch)
- Q: How can users permanently delete removed entities (files/folders flagged removed=1) from database? → A: NEEDS CLARIFICATION - Propose: Add `dupectl purge files <root-path>` `dupectl purge folders <root-path>` and `dupectl purge all <root-path>` command with optional --before=date filter to cleanup old removed records and free storage
- Q: How can users manually clear stale scan checkpoints (abandoned scans with old updated_at timestamps)? → A: NEEDS CLARIFICATION - Existing --restart flag clears checkpoint for new scan, but propose: Add threshold detection (e.g., checkpoint older than 24 hours) with automatic prompt or `dupectl delete checkpoint <root-path>` command
- Q: How can users retry files/folders with error_status after fixing permissions? → A: NEEDS CLARIFICATION - Propose: leverage `dupectl scan files <root-path>` command to reset error_status to NULL, or add --retry-errors flag to scan commands to force reprocessing of error entries
- Q: How are root folder statistics (folder_count, file_count, total_size) recalculated after manual deletes or removed flag cascades? → A: NEEDS CLARIFICATION - Propose: Automatic recalculation on scan completion (already planned) + manual `dupectl refresh all <root-path>` command for on-demand updates
- Q: What consistency check operations are available to detect orphaned records, invalid foreign keys, or data corruption? → A: NEEDS CLARIFICATION - Propose: Add `dupectl verify all <root-path>` command to check FK integrity, detect orphans, identify inconsistent timestamps, validate removed flag cascades, with optional --repair flag
- Q: How does system handle files that reappear at same path after being marked removed=1? → A: NEEDS CLARIFICATION - Propose: During scan, if file exists and removed=1, automatically set removed=0 and update last_scanned_at (restore file to active state)
- Q: What is the strategy for migrating to a different hash algorithm (rehash all files with new algorithm)? → A: NEEDS CLARIFICATION - leverage scan command to recalculate hashes when configuration is detected to have changed, overwriting old hash_value and updating hash type

**PRODUCTION READINESS GAPS**

- Q: What exit codes should CLI commands return (success, user error, system error, cancelled)? → A: NEEDS CLARIFICATION - Propose: Exit code 0 (success), 1 (user error: invalid args, not found), 2 (system error: database, filesystem, permission), 130 (SIGINT/user cancelled)
- Q: What CHECK constraints should enforce data integrity (hash_algorithm values, scan_mode values, boolean flags, non-negative counts)? → A: NEEDS CLARIFICATION - Propose: Add SQL CHECK constraints for hash_algorithm IN ('sha256', 'sha512', 'sha3-256'), scan_mode IN ('all', 'folders', 'files'), removed/completed IN (0,1), size/counts >= 0
- Q: What SQLite PRAGMA settings are required (foreign_keys, synchronous, journal_mode)? → A: NEEDS CLARIFICATION - Propose: `PRAGMA foreign_keys = ON` (critical for CASCADE), `PRAGMA journal_mode = WAL` (already planned), `PRAGMA synchronous = FULL` (safety over speed), document in quickstart.md
- Q: Where is configuration file located and can path be overridden? → A: NEEDS CLARIFICATION - Propose: Default ~/.dupectl.yaml (Windows: %USERPROFILE%\.dupectl.yaml), override with --config flag or DUPECTL_CONFIG env var
- Q: What is complete JSON output schema for `get duplicates --json`? → A: NEEDS CLARIFICATION - Propose: Document JSON structure with field names (snake_case), types, nesting, timestamp format (ISO 8601), example output
- Q: What table rendering library/style is used for human-readable output? → A: NEEDS CLARIFICATION - Propose: Use Go library like `olekukonko/tablewriter` or `rodaine/table`, specify ASCII vs Unicode box-drawing, column alignment rules
- Q: Is non-interactive mode supported for scripting (--yes flag to auto-accept confirmations)? → A: NEEDS CLARIFICATION - Propose: Add --yes/-y flag to skip all confirmation prompts (root registration, destructive deletes), exit with error if confirmation required but non-interactive

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Scan All Files and Folders (Priority: P1)

A user wants to perform a complete scan of their root folder to identify all duplicate files and folders across their entire monitored directory tree. This is the most common use case when first setting up duplicate detection or when performing periodic comprehensive scans.

**Why this priority**: This is the primary value proposition of the system - finding duplicates. Without this capability, no other features matter. It delivers immediate value by showing all duplicates in a single operation.

**Independent Test**: Can be fully tested by running the scan command on a test folder structure with known duplicates (files with identical content and duplicate folder trees) and verifying that all duplicates are correctly identified and stored in the database.

**Acceptance Scenarios**:

1. **Given** a root folder containing multiple files and subfolders with some duplicate content, **When** user executes the scan all command with root folder path (e.g., `dupectl scan all /path/to/root`), **Then** the system recursively traverses the folder tree, hashes all files, stores hash values in database, and identifies all duplicate files (matching size and hash) and duplicate folders (identical subtrees)
2. **Given** a scan is in progress, **When** user monitors the operation, **Then** the system provides progress indication showing folders and files processed
3. **Given** a scan has completed, **When** user retrieves results, **Then** the system reports total files scanned, duplicate files found, and duplicate folders identified
4. **Given** user provides a root folder path that exists on filesystem but is not registered in database, **When** scan command is executed, **Then** system prompts for confirmation to register the root folder and proceeds with scan only if user confirms

---

### User Story 2 - Scan Folders Only (Priority: P2)

A user wants to quickly map out the folder structure and establish monitoring on a large directory tree without the time-intensive process of hashing all files. This allows rapid initial setup and deferred file scanning for later.

**Why this priority**: This enables efficient workflow for large directory trees where file scanning might take hours. Users can establish monitoring coverage quickly and perform file scanning selectively or during off-hours.

**Independent Test**: Can be tested independently by running the folder scan command on a test structure and verifying that all folders are registered in the database with correct hierarchy relationships, but no file hashes are calculated.

**Acceptance Scenarios**:

1. **Given** a root folder with deep nested folder structure, **When** user executes the scan folders command with root folder path (e.g., `dupectl scan folders /path/to/root`), **Then** the system recursively traverses and registers all folders in the database without processing file contents
2. **Given** folders have been scanned without files, **When** user later requests duplicate folder detection, **Then** the system indicates that file scanning is required before folder duplicates can be identified
3. **Given** a folder scan is complete, **When** user views registered folders, **Then** the system shows the complete folder hierarchy with folder counts and registration timestamps

---

### User Story 3 - Scan Files Only (Priority: P2)

A user wants to scan and hash all files within already-registered folders without re-traversing the folder structure. This is useful for updating file hashes after initial folder registration or when files have changed.

**Why this priority**: Enables efficient incremental scanning where folder structure is already known. This supports workflows where folders are monitored first, then files are scanned in batches or on-demand.

**Independent Test**: Can be tested independently by first registering folders (manually or via folder scan), then running file scan command and verifying that all files are hashed and duplicates identified without modifying folder registrations.

**Acceptance Scenarios**:

1. **Given** folders have been previously registered in the database, **When** user executes the scan files command with root folder path (e.g., `dupectl scan files /path/to/root`), **Then** the system processes all files within registered folders, calculates hash values, stores them in database, and identifies duplicate files
2. **Given** some files have been modified since last scan, **When** user runs file scan again, **Then** the system updates hash values for changed files and recalculates duplicate matches
3. **Given** file scanning is complete, **When** user queries duplicates, **Then** the system returns accurate duplicate file listings based on size and hash matching

---

### User Story 4 - Identify Partial Folder Duplicates (Priority: P3)

A user wants to find folders that are mostly similar but not identical - where a significant subset of files match but some files are missing or different. This helps identify near-duplicates that might represent different versions or incomplete copies.

**Why this priority**: Adds sophisticated duplicate detection beyond exact matches. While valuable, it's not essential for basic duplicate detection and can be implemented after core scanning is working.

**Independent Test**: Can be tested independently by creating folder pairs with overlapping but not identical file sets, running complete scan, then querying for partial matches and verifying the system correctly identifies overlap percentages and key differences.

**Acceptance Scenarios**:

1. **Given** two folders where 70% of files are identical, **When** user queries for partial folder duplicates, **Then** the system identifies both folders as potential matches with similarity percentage and lists matching/non-matching files
2. **Given** partial folder matches exist, **When** user views details, **Then** the system highlights key differences: files missing from each side and files with same name but different modification dates
3. **Given** user wants to filter partial matches, **When** user specifies minimum similarity threshold (e.g., 80%), **Then** the system returns only folder pairs meeting or exceeding that threshold (default minimum is 50% if not specified)

---

### User Story 5 - Query and View Duplicate Files (Priority: P1)

A user wants to view the duplicate files that have been identified during scans, with the ability to filter results and choose between human-readable or machine-parseable output formats.

**Why this priority**: This is essential for the primary use case - users need to see the results of duplicate detection. Without query capability, scan results are trapped in database with no visibility.

**Independent Test**: Can be tested independently after scans have been performed by querying duplicates and verifying output format, grouping, and filtering work correctly.

**Acceptance Scenarios**:

1. **Given** duplicate files have been identified during scans, **When** user executes get duplicates command with default options (e.g., `dupectl get duplicates`), **Then** system displays results in human-readable table format with files grouped by duplicate set showing size, hash, and file paths
2. **Given** user needs machine-readable output for scripting, **When** user executes get duplicates command with --json flag (e.g., `dupectl get duplicates --json`), **Then** system outputs structured JSON with duplicate sets and file metadata
3. **Given** user wants to filter high-duplication files, **When** user executes get duplicates with --min-count filter (e.g., `dupectl get duplicates --min-count=5`), **Then** system returns only duplicate sets containing 5 or more files

---

### Edge Cases

- What happens when a scan is interrupted mid-process (system crash, user cancellation)? System resumes from last checkpoint using database-tracked progress by default; user can use --restart flag to start fresh
- What happens when user provides a root folder path that doesn't exist or isn't registered in the database? If path exists on filesystem but not in database, prompt for registration confirmation; if path doesn't exist on filesystem, reject with clear error message
- What happens when user provides a relative path vs absolute path for root folder? System converts relative paths to absolute paths using current working directory before validation and storage
- How does system handle files and folders that cannot be accessed due to permissions? Display error to console, mark file/folder in database with error status flag (permission denied), continue scanning remaining items without crash
- What happens when a folder has permission denied? Folder is recorded in database with permission error flag, contents are not scanned, scan continues with sibling folders
- What happens when file sizes are identical but hashes differ? Files are considered different (not duplicates) - both size AND hash must match for duplicate detection
- What happens when file hashes are identical but sizes differ? Files are considered different (not duplicates), though this scenario is theoretically impossible with cryptographic hash functions
- What happens when files have identical size and hash but different names or paths? Files are considered duplicates - filenames and paths do not affect duplicate detection logic
- How does system handle symbolic links, shortcuts, or hard links?
- What happens when folders contain millions of small files versus few very large files?
- How does system handle files that are modified during the scanning process?
- What happens when two folders have identical structure but zero files (empty folder trees)?
- How does system handle files with special characters or very long paths?
- What happens when a file is moved from one location to another within monitored paths? Previous location shows file as removed, new location shows file as newly scanned, matching hash enables correlation
- What happens when user removes a path with --remove-path but that path is currently being scanned? Operation should be rejected with error message or scan should be stopped first
- What happens when --progress flag is used but output is redirected to a file? Spinner and progress updates should still be written to stderr/console for user visibility
- What happens when configuration file contains invalid values (e.g., unsupported hash algorithm, negative progress interval)? System fails at startup with clear error message indicating the problem and valid options
- What happens when configuration file is missing or unreadable? System uses default values and logs warning about missing configuration
- What happens when traverse_links is enabled and a circular symbolic link reference is encountered? System detects the cycle using path tracking and skips the circular link with warning message
- What happens when getting root folder list and a root folder no longer exists on filesystem? Display root in table with warning indicator but retain database records for historical data
- What happens when total size calculation overflows (petabytes of data)? Use appropriate units (PB, EB) and ensure proper numeric handling for large values
- What happens when a sub-folder is deleted but its parent folder still exists? Only the deleted sub-folder and its contents are flagged as removed, parent folder and siblings remain active
- What happens when a root folder is flagged as removed but later re-added at the same path? System should detect existing records and offer to restore them (unflag removed) or start fresh with new scan
- What happens when cascading removed flag to folder contents and the folder contains thousands of files? Operation should be efficient (single database query with WHERE clause) rather than individual updates per file
- What happens when application is killed forcefully (SIGKILL) without chance to save checkpoint? Scan resumes from last periodic checkpoint saved during scan; some progress may be lost but data integrity is maintained
- What happens when multiple scan processes start simultaneously trying to resume the same checkpoint? First process to acquire checkpoint lock proceeds with scan; other processes detect active scan and exit or wait
- What happens when checkpoint data is corrupted in database? System detects corruption on startup, logs error, and requires --restart flag to begin fresh scan
- What happens in containerized environment when database is on shared volume but container is replaced? New container instance detects incomplete scan checkpoint and automatically resumes from where previous container stopped
- What happens when multiple workers try to hash the same file simultaneously? System uses work queue with atomic dequeue operations ensuring each file is processed by exactly one worker
- What happens when a worker crashes or panics during hashing/traversal? Other workers continue unaffected; failed work item is either retried by another worker or skipped with error logged
- What happens when worker pool size is configured larger than number of available CPU cores? System allows configuration but may see diminishing returns or performance degradation due to context switching overhead
- What happens when shutdown signal is received while workers are processing files? System stops accepting new work, waits for in-flight operations to complete (with 5 second timeout), saves checkpoint of completed work, then exits

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide three distinct scan modes: scan all (folders + files), scan folders only, scan files only
- **FR-001.1**: All scan commands MUST accept a mandatory positional argument specifying the root folder path to scan (format: `dupectl scan <mode> <root-folder-path>`)
- **FR-001.2**: If provided root folder path is not registered in database, system MUST prompt user for confirmation to register it before proceeding with scan, allowing user to cancel if path was incorrect
- **FR-001.3**: System MUST convert relative root folder paths to absolute paths before storage and processing to ensure path consistency across different working directories
- **FR-001.4**: System MUST support management of multiple root folders simultaneously, allowing users to register, configure, and scan different directory trees independently
- **FR-001.5**: When adding a root folder, system MUST allow users to configure options including: path (mandatory), traverse links flag (optional, default: false) to control symbolic link traversal behavior
- **FR-001.6**: System MUST provide a command to list all registered root folders (e.g., `dupectl get root`) displaying results in table format with columns: path, number of folders, number of files, total size (in human-readable units), and last scan date (UTC)
- **FR-001.7**: Root folder listing MUST calculate summary statistics from database (folder count, file count, total size) based on most recent scan data, showing "N/A" or "Never scanned" for roots that have not been scanned yet
- **FR-002**: System MUST recursively traverse folder trees starting from registered root folders
- **FR-003**: System MUST calculate cryptographic hash for each file during file scanning using configurable hash algorithm with options limited to secure algorithms with low collision rates (SHA-256, SHA-512, SHA3-256)
- **FR-004**: System MUST store folder structure, file metadata (path, size, modification date), hash values, and hash algorithm type in database
- **FR-004.1**: System MUST use dedicated database tables for folders and files (separate from existing agent/host/owner/policy tables) while maintaining foreign key relationships to existing RootFolder, Agent, and Host entities for organizational tracking
- **FR-004.2**: System MUST record access errors for both files and folders (e.g., permission denied) in database with error status flag and timestamp to prevent repeated failed access attempts in subsequent scans
- **FR-004.3**: Files and folders with no access permissions MUST be saved in database with a flag indicating permission denial prevented deep scanning, preserving record of attempted scan while marking as inaccessible
- **FR-005**: System MUST identify duplicate files by matching both file size AND hash value - files are duplicates ONLY when both conditions are satisfied
- **FR-005.1**: Files with identical size but different hash values MUST be considered different files (not duplicates)
- **FR-005.2**: Files with identical hash values but different sizes MUST be considered different files (not duplicates), though this scenario should be impossible with cryptographic hash functions
- **FR-005.3**: Filenames MUST NOT be used in duplicate detection logic - files with different names but identical size and hash are considered duplicates
- **FR-006**: System MUST identify duplicate folders when all files within their entire subtrees match according to file duplicate logic
- **FR-007**: System MUST detect partial folder duplicates when a subset of files match with minimum similarity threshold of 50%
- **FR-007.1**: Partial folder duplicate queries MAY allow user to specify higher similarity thresholds to filter results (e.g., 70%, 80%), but default minimum is 50%
- **FR-008**: System MUST highlight key differences in partial matches: missing files from either side and files with same name but different dates
- **FR-009**: System MUST distinguish between folder registration (folder scan) and file content analysis (file scan)
- **FR-010**: Folder scan MUST register folder hierarchy without calculating file hashes
- **FR-011**: File scan MUST process files within already-registered folders without re-traversing folder structure
- **FR-012**: System MUST persist all scan results to database for later query and analysis
- **FR-013**: System MUST provide progress indication during scan operations showing folders/files processed
- **FR-013.1**: System MUST support a --progress command-line option that displays real-time progress information including: braille spinner animation, number of folders scanned, number of files scanned, and elapsed time since scan start
- **FR-013.2**: Progress updates MUST be output to console at configurable time intervals with default of 10 seconds
- **FR-014**: System MUST handle scan interruptions gracefully by saving checkpoint to database before shutdown and automatically resuming from last checkpoint when application restarts (default behavior)
- **FR-014.1**: System MUST persist scan state including current root folder, current subfolder path, and last processed file to database before responding to shutdown signals (SIGINT, SIGTERM) or application termination
- **FR-014.2**: System MUST detect incomplete scans on application startup and automatically resume from last saved checkpoint without user intervention (default behavior)
- **FR-014.3**: System MUST provide a command-line option (e.g., --restart) to restart a scan from the beginning rather than resuming from checkpoint, clearing previous scan state for that root folder
- **FR-014.4**: Checkpoint mechanism MUST support containerized deployments where container instances can be stopped and restarted without losing scan progress
- **FR-015**: System MUST allow users to configure hash algorithm selection in configuration file with options: SHA-512 (default), SHA-256, SHA3-256, applied globally to all scan operations
- **FR-015.1**: System MUST validate all configuration options at startup, rejecting invalid values with clear error messages indicating acceptable values and current configuration source (file path)
- **FR-015.2**: Each configuration option MUST have automated tests that validate: default value behavior, valid value acceptance, invalid value rejection, and runtime behavior with configured values
- **FR-015.3**: System MUST allow users to configure number of worker threads/goroutines for parallel operations in configuration file with default value appropriate for typical hardware (e.g., number of CPU cores)
- **FR-015.4**: System MUST support parallel folder tree traversal using configurable number of workers to optimize performance for directory structures with many small files
- **FR-015.5**: System MUST support parallel file hashing using configurable number of workers to optimize performance for large files where hashing is the bottleneck
- **FR-015.6**: Parallel operations MUST be implemented with proper synchronization mechanisms to prevent race conditions when accessing shared resources (database connections, scan state, checkpoint data)
- **FR-015.7**: Parallel operations MUST be designed to prevent deadlocks through proper lock ordering and timeout mechanisms
- **FR-015.8**: Worker pool configuration MUST be validated at startup, rejecting values less than 1 or greater than reasonable maximum (e.g., 100) with clear error messages
- **FR-016**: System MUST store hash algorithm type with each file record to enable future algorithm migrations and maintain data integrity across configuration changes
- **FR-019**: System MUST record the timestamp (UTC) of when a file was first entered in the database during the initial scan of that file
- **FR-020**: System MUST update the "last scanned" timestamp field (UTC) when a file is encountered during subsequent scans of the same path
- **FR-021**: System MUST flag files as "removed" in the database when they are no longer present in their previously scanned location, enabling tracking of file movements between locations
- **FR-021.1**: Removed files MUST remain in database with removed flag set to true rather than being deleted, preserving history for file movement analysis
- **FR-021.2**: System MUST flag folders as "removed" in the database when they are no longer present on the filesystem (applies to both root folders and sub-folders)
- **FR-021.3**: When a folder is flagged as removed, system MUST automatically flag all files and sub-folders contained within that folder hierarchy as removed in a cascading manner
- **FR-017**: System MUST provide command to query duplicate files with configurable output format: human-readable table format (default) or JSON format (via --json flag) for scripting
- **FR-017.1**: Table format output MUST group files by duplicate set, showing all files with identical size and hash together
- **FR-017.2**: Duplicate query command MUST support optional --min-count filter to return only duplicate sets with at least N files (e.g., --min-count=3 shows only files with 3+ duplicates)
- **FR-018**: System MUST provide command to query duplicate folders (exact matches) and partial folder duplicates with similarity percentage
- **FR-022**: System MUST provide a command-line option to remove a registered path and delete all associated scan data (folders, files, hashes) for that path from the database
- **FR-023**: Each command-line option (--progress, --restart, --json, --min-count, path removal, etc.) MUST have automated tests that validate the option performs its intended function
- **FR-024**: Each command-line option MUST be properly documented in the CLI interface help text (accessible via --help or -h flags), including option syntax, description, and usage examples
- **FR-025**: Repository MUST include test fixtures (test folders and files) with known characteristics for automated testing, including: exact duplicate files, duplicate folder structures, partial folder duplicates, files with permission issues, and edge cases
- **FR-025.1**: Test fixtures MUST be organized under a dedicated test data directory (e.g., tests/fixtures/ or testdata/) with documented structure and expected test outcomes
- **FR-025.2**: Test fixtures MUST include files with known hash values, sizes, and relationships to enable deterministic validation of duplicate detection logic
- **FR-026**: System MUST include integration tests covering complete workflows: root registration → scan → query → results verification
- **FR-026.1**: Integration tests MUST validate checkpoint save/resume functionality by intentionally interrupting scans and verifying successful resumption
- **FR-026.2**: Integration tests MUST validate error handling paths including permission-denied files/folders, invalid paths, and corrupted checkpoints
- **FR-027**: System MUST include database operation tests validating: schema creation, data persistence, query correctness, foreign key relationships, and cascading operations (removed flag propagation)
- **FR-030**: System MUST include signal handling tests validating graceful shutdown on SIGINT/SIGTERM with checkpoint save and clean exit within 5 seconds
- **FR-031**: System MUST include concurrent operation tests validating checkpoint locking prevents multiple scan processes from running simultaneously on the same root folder
- **FR-032**: System MUST include containerized deployment tests validating checkpoint persistence and resume across container stop/start cycles

### Key Entities

- **Root Folder**: A top-level directory registered for monitoring, serves as the starting point for recursive scans, has configuration options (path, traverse_links flag), stores summary statistics (folder count, file count, total size, last scan timestamp), links to existing Host, Owner, Agent, and Purpose entities from infrastructure tables
- **Folder**: A directory within the monitored tree, stored in dedicated folders table, has hierarchical relationship to parent folder and root, contains zero or more files and subfolders, has attributes: path, first scanned timestamp (UTC), last scanned timestamp (UTC), removed flag (boolean), error status (for access failures including permission denied)
- **File**: A file within a monitored folder, stored in dedicated files table, has attributes: path, size, modification date, hash value, hash algorithm type, error status (for access failures including permission denied), first scanned timestamp (UTC), last scanned timestamp (UTC), removed flag (boolean), relationship to containing folder
- **Duplicate File Set**: A group of 2+ files with identical size AND hash values - both criteria must match; filename and path are not considered in duplicate detection
- **Duplicate Folder Set**: A group of 2+ folders where all files in their complete subtrees match
- **Partial Duplicate Folder Pair**: Two folders with overlapping but not identical file sets, includes similarity percentage and difference details
- **Scan State**: Progress tracking record stored in database to enable scan resumption, includes current root folder ID, current folder path, and last processed file
- **Test Fixtures**: Predefined test data (folders and files) stored in repository with known characteristics and expected outcomes, used for automated testing of scanning and duplicate detection logic

### Non-Functional Requirements

- **NFR-001 Performance**: File hashing should achieve minimum throughput of 50 MB/sec on standard hardware to enable reasonable scan times for large datasets
- **NFR-002 Performance**: System should handle scanning of at least 100,000 files without memory overflow or excessive memory consumption (stay under 500 MB RAM)
- **NFR-003 Portability**: Scan operations must work identically on Windows, Linux, and macOS without platform-specific behavior
- **NFR-004 Observability**: Provide clear progress indication with counts of folders/files processed and elapsed time, updated at configurable intervals (default 10 seconds) to avoid excessive console output
- **NFR-005 Observability**: Log all scan operations including start/end times, files processed, errors encountered, and results summary
- **NFR-006 Security**: Handle file and folder access permissions gracefully - display permission errors to console during scan, mark affected files and folders in database with error status flag to avoid repeated attempts in future scans, and continue with remaining items without crashing
- **NFR-007 Maintainability**: Separate concerns: folder traversal logic, file hashing logic, duplicate detection logic, and database operations should be in distinct modules
- **NFR-007.1 Concurrency**: Parallel operations (folder traversal, file hashing) must use thread-safe/goroutine-safe data structures and synchronization primitives to ensure data integrity
- **NFR-007.2 Concurrency**: Database operations from multiple workers must use connection pooling and proper transaction isolation to prevent race conditions and ensure consistency
- **NFR-007.3 Concurrency**: Worker pool implementation must gracefully handle worker failures without affecting other workers or causing system-wide crashes
- **NFR-008 Graceful Shutdown**: Handle SIGINT/SIGTERM during scans by immediately saving checkpoint state to database, flushing all pending writes, and providing clean exit within 5 seconds to support container orchestration systems
- **NFR-009 Upgradability**: Database schema for scan results should support versioning to enable future enhancements without data loss
- **NFR-010 Dependencies**: Justify hash algorithm library choice - prefer standard library implementations over external dependencies
- **NFR-011 Testability**: All command-line options and configuration options must have corresponding automated tests validating correct behavior, error handling, and validation logic
- **NFR-012 Usability**: All command-line options must include comprehensive help documentation accessible through standard --help flag with clear descriptions and usage examples
- **NFR-013 Testability**: Test fixtures must be version-controlled in repository and structured to support both unit tests and integration tests, with clear separation between different test scenarios
- **NFR-014 Testability**: Test suite must include unit tests (individual functions/modules), integration tests (complete workflows), and end-to-end tests (full system scenarios) with minimum 80% code coverage for core scanning and duplicate detection logic

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can successfully complete a full scan of a folder structure containing 10,000 files in under 5 minutes on standard hardware
- **SC-002**: System correctly identifies 100% of duplicate files in test scenarios with known duplicates (zero false positives or false negatives)
- **SC-003**: System correctly identifies 100% of duplicate folders (identical subtrees) in test scenarios
- **SC-004**: Users can distinguish between folder-only scan (fast structure mapping) and full scan (with hashing) and select appropriate mode for their needs
- **SC-004.1**: Users can query duplicate results in both human-readable table format and JSON format for different use cases (interactive vs scripting)
- **SC-005**: System detects at least 80% of meaningful partial folder duplicates in test scenarios with similarity threshold of 50% or higher
- **SC-006**: Scan operations provide visible progress indication allowing users to monitor operation without anxiety about system hang
- **SC-007**: Interrupted scans save checkpoint before shutdown and automatically resume from last checkpoint on application restart without user intervention, with --restart option available for fresh scans when needed
- **SC-007.1**: System supports containerized deployments with checkpoint persistence across container stop/start cycles, enabling reliable operation in orchestrated environments
- **SC-008**: Users can monitor scan progress in real-time using --progress option, seeing current status with spinner, folder count, file count, and elapsed time
- **SC-009**: System accurately tracks file lifecycle with first scanned timestamp, last scanned timestamp, and removal flag, enabling detection of file movements between scans
- **SC-009.1**: System accurately tracks folder lifecycle including removal detection, automatically cascading removed flag to all contained files and subfolders when parent folder is removed
- **SC-010**: Users can successfully remove a registered path and all its scan data using a dedicated command-line option
- **SC-011**: All command-line options have help documentation and automated tests ensuring they work as documented
- **SC-012**: Repository includes comprehensive test fixtures enabling 100% automated validation of duplicate detection logic without requiring external test data
- **SC-013**: All configuration options are validated at startup and have automated tests covering default values, valid inputs, invalid inputs, and runtime behavior
- **SC-014**: Users can manage multiple root folders with get root command providing clear overview of all registered roots, their scan status, and storage usage in table format
- **SC-015**: Test suite achieves minimum 80% code coverage for core logic (scanning, hashing, duplicate detection, database operations) with passing unit, integration, and end-to-end tests
- **SC-016**: All edge cases documented in specification have corresponding automated tests validating correct behavior
- **SC-017**: Parallel scanning operations complete without race conditions, deadlocks, or data corruption when tested with multiple concurrent workers
- **SC-018**: Users can configure worker pool size to optimize scanning performance based on their workload characteristics (many small files vs few large files)

## Assumptions

- **A-001**: Secure cryptographic hashing (SHA-256 or stronger) provides sufficient collision resistance for duplicate detection - probability of hash collision for different file contents is negligible
- **A-002**: SHA-512 is default hash algorithm, providing the lowest collision probability among the supported algorithms due to its 512-bit output, ensuring maximum reliability for duplicate detection
- **A-002.1**: All scans within a deployment use the same hash algorithm configured in the application config file - mixed algorithms across concurrent operations are not supported
- **A-003**: File modification dates from filesystem are sufficiently reliable for highlighting potential version differences in partial matches
- **A-004**: Users have read permissions for files they want to scan - permission-denied files will be logged and skipped
- **A-005**: Folder structure changes during scan are rare edge case - current scan operates on snapshot of structure at scan start time
- **A-006**: Symbolic links and hard links are treated as regular files/folders - no special handling for resolving link targets (avoids circular reference complexity)
- **A-007**: Empty files (size 0) are considered duplicates if multiple exist - size and hash matching still applies
- **A-008**: Partial folder duplicate similarity is calculated as: (matching files count / total unique files across both folders) * 100
- **A-009**: System runs on filesystem that provides reliable file size and modification timestamp metadata
- **A-010**: Checkpoint granularity for scan resumption is at the folder level - entire folder is reprocessed if interruption occurs mid-folder to ensure consistency
- **A-011**: Dedicated scan tables (folders, files) remain loosely coupled to infrastructure tables - scan operations should function independently even if agent/host data is minimal or absent
- **A-012**: All paths stored in database are absolute paths with platform-appropriate separators (forward slash on Unix, backslash on Windows) to ensure consistency and avoid ambiguity
- **A-013**: Human-readable table format is default output for duplicate queries, assuming primary use case is interactive terminal usage; JSON format available via explicit flag for automation scenarios
- **A-014**: Resume from checkpoint is the default behavior on application restart after interruption, assuming users prefer not to waste already-completed work; explicit --restart flag available for cases where clean restart is needed; this design especially benefits containerized deployments where containers may be stopped and restarted frequently
- **A-015**: Progress display is optional (via --progress flag) rather than always-on, assuming some users prefer minimal output or are running scans in scripts
- **A-016**: Removed files are flagged rather than deleted from database, assuming value in tracking file history and movements over time exceeds storage cost of historical records
- **A-017**: Test fixtures are kept small (under 10 MB total) to maintain reasonable repository size while providing sufficient coverage for test scenarios
- **A-018**: Test fixtures use platform-agnostic content (text files, binary patterns) that work identically across Windows, Linux, and macOS for portable testing
- **A-019**: Configuration validation occurs at application startup (fail-fast approach) rather than at first use, preventing runtime errors from invalid configuration
- **A-020**: Traverse links is disabled by default (false) to prevent circular references and scanning of external linked directories, but can be enabled per root folder for use cases requiring link traversal
- **A-021**: Root folder summary statistics (folder count, file count, total size) are calculated from database records rather than live filesystem queries for performance, meaning they reflect state as of last scan
- **A-022**: Cascading removed flag from folder to all contained files/subfolders is implemented efficiently using database queries with hierarchical path matching rather than recursive individual updates
- **A-023**: Checkpoints are saved to database both periodically during scan (e.g., after each folder completion) and immediately on shutdown signal to minimize loss of progress while maintaining database performance
- **A-024**: Duplicate detection is purely content-based (size + hash) and completely independent of filename, file path, or file metadata like modification date - this enables detection of renamed or moved duplicates
- **A-025**: Test suite runs in CI/CD pipeline on every commit with automated pass/fail gates preventing merge of code that breaks tests or reduces coverage below threshold
- **A-026**: Default worker pool size is set to number of CPU cores, assuming this provides good balance for mixed workloads; users can tune based on their specific file size distribution and I/O characteristics
- **A-027**: Parallel folder traversal is most beneficial for directory structures with many folders and small files; parallel file hashing is most beneficial for fewer large files - configuration allows users to optimize for their use case
- **A-028**: SQLite with WAL mode provides sufficient concurrent read/write performance for multiple workers; database connection pool size should be at least equal to worker count to prevent bottlenecks

## Dependencies

- **D-001**: Requires database system already configured and accessible (based on existing dupedb.db file)
- **D-002**: Depends on root folder registration capability (appears to exist based on cmd/addRoot.go) which references existing Host, Owner, Agent, and Purpose entities
- **D-003**: Scan data tables (files, folders, hashes) will maintain foreign key relationships to existing infrastructure tables (agents, hosts, root_folders) for organizational context while keeping scan-specific data architecturally separate

## Out of Scope

- **OS-001**: Real-time monitoring or automatic re-scanning when files change - this spec covers on-demand scanning only
- **OS-002**: Duplicate file resolution actions (delete, move, symlink) - this spec focuses on detection, not remediation
- **OS-003**: Content-aware duplicate detection (similar images, similar documents) - relies purely on byte-for-byte hash matching
- **OS-004**: Network or cloud storage scanning - assumes local filesystem access
- **OS-005**: Deduplication or storage optimization - identification only, no storage changes
- **OS-006**: Graphical user interface for scan operations - CLI commands only

## Implementation Gaps (Code vs Spec Analysis)

### Critical Missing Components

**Database Schema (Priority: Must Have)**
- No database tables exist for files with columns: id, path, size, mtime, hash_value, hash_algorithm, error_status, first_scanned_at (UTC), last_scanned_at (UTC), removed (boolean), folder_id, root_folder_id
- No database tables exist for folders with columns: id, path, parent_folder_id, root_folder_id, scan_status, first_scanned_at (UTC), last_scanned_at (UTC), removed (boolean), error_status
- No scan_state table for checkpoint tracking: root_folder_id, current_folder_path, last_processed_file, scan_mode, started_at
- Existing RootFolder entity lacks configuration fields: traverse_links (boolean), last_scan_date (UTC)
- Existing RootFolder entity lacks summary statistic fields: folder_count, file_count, total_size
- Existing Folder entity in entities/files.go lacks required fields (parent_folder_id, full path, scan timestamp)
- Existing filemsg struct has hash field but no hash_algorithm or error_status fields

**Scanning Logic (Priority: Must Have)**
- No folder traversal implementation - cmd/scanFolders.go is stub only
- No file hashing implementation - no pkg/scanner or pkg/hash module exists
- No hash algorithm configuration reading from config file
- No worker pool implementation for parallel folder traversal
- No worker pool implementation for parallel file hashing
- No work queue with atomic operations for distributing work to workers
- No synchronization mechanisms (mutexes, channels) for thread-safe operations
- No database connection pooling for concurrent worker access
- No worker failure handling and retry logic
- No graceful worker pool shutdown on SIGINT/SIGTERM
- No progress reporting mechanism with configurable time intervals
- No checkpoint save/restore logic for scan interruption handling
- No signal handlers (SIGINT/SIGTERM) to save checkpoint before shutdown
- No automatic resume detection and execution on application startup
- No checkpoint locking mechanism to prevent concurrent scan processes
- No permission error handling with database marking for files and folders
- No folder removal detection logic during scans
- No cascading removed flag implementation when folder hierarchy is deleted

**Duplicate Detection (Priority: Must Have)**
- No logic to identify duplicate files by matching size AND hash
- No logic to identify duplicate folders (identical subtrees)
- No logic to detect partial folder duplicates with similarity calculation
- cmd/getDuplicates.go is stub with no implementation

**CLI Command Implementation (Priority: Must Have)**
- cmd/scanAll.go: No root folder path argument parsing, no call to scan logic, missing --progress and --restart flags
- cmd/scanFolders.go: No implementation, needs folder-only traversal, missing --progress and --restart flags
- cmd/scanFiles.go: No implementation, needs file-only hashing, missing --progress and --restart flags
- cmd/getDuplicates.go: No output formatting (table/JSON), no --min-count filter
- cmd/addRoot.go: No validation for path existence, no absolute path conversion, no registration prompt, no traverse_links flag implementation
- cmd/getRoot.go: Missing table format output with columns for path, folder count, file count, total size, last scan date
- cmd/deleteRoot.go: Missing implementation for removing path and deleting all associated scan data
- No help documentation (--help) for any command-line options
- No automated tests for command-line options

**Test Infrastructure (Priority: Must Have)**
- No test fixtures directory structure (e.g., tests/fixtures/ or testdata/)
- No test files with known hash values for duplicate detection validation
- No test folder structures for duplicate folder detection scenarios
- No test fixtures for partial folder duplicate scenarios
- No test fixtures for edge cases (empty files, permission errors, special characters)
- No documentation describing test fixture structure and expected outcomes
- No integration tests for complete workflows (register → scan → query → verify)
- No tests for checkpoint save/resume functionality
- No tests for error handling paths (permissions, invalid paths, corruption)
- No database operation tests (schema, persistence, queries, cascading)
- No signal handling tests (SIGINT/SIGTERM graceful shutdown)
- No concurrent operation tests (checkpoint locking)
- No tests for parallel worker operations (race conditions, deadlocks)
- No tests for worker pool shutdown and cleanup
- No tests for work distribution across multiple workers
- No containerized deployment tests (checkpoint persistence across restarts)
- No test coverage measurement or minimum coverage enforcement
- No CI/CD pipeline configuration for automated test execution

**Configuration (Priority: Must Have)**
- No hash algorithm config setting in viper defaults (setDefaults in root.go)
- No worker pool size config setting (for parallel folder traversal and file hashing)
- No progress interval config setting
- No configuration documentation for scan settings
- No configuration validation logic at startup
- No tests for configuration option validation (valid/invalid values)
- No tests for configuration option runtime behavior (hash algorithm selection, progress interval, worker pool size)

### Moderate Gaps

**Path Handling (Priority: Should Have)**
- No absolute path conversion logic for relative paths
- No path validation and normalization across platforms
- No handling of special characters or long paths

**Error Handling (Priority: Should Have)**
- No console warning output for permission errors during scan (files and folders)
- No graceful handling of unreadable files/folders with continuation
- No validation that provided root path exists on filesystem

**Data Integrity (Priority: Should Have)**
- No foreign key relationships defined between new scan tables and existing infrastructure tables
- No database migration/versioning strategy for schema changes

**User Experience (Priority: Nice to Have)**
- No interactive confirmation prompts for unregistered root folders
- No summary statistics at end of scan (files processed, duplicates found)

### Existing Code Assets (Can Leverage)

**Infrastructure (Already Exists)**
- Database connection via pkg/datastore/datastore.go (SQLite with WAL mode)
- Agent table creation pattern in pkg/datastore/agent.go (can use as template)
- Entity models in pkg/entities/files.go (Host, Owner, Agent, RootFolder, Folder)
- Configuration via Viper in cmd/root.go with defaults
- Cobra CLI command structure in cmd/ directory
- API infrastructure in pkg/api/ (may be useful for future extensions)

**Recommendations for Planning Phase**

1. **Start with database schema creation** - Foundation for everything else
2. **Implement basic folder traversal** - Core scanning capability  
3. **Add file hashing with configurable algorithm** - Duplicate detection prerequisite
4. **Build duplicate detection queries** - Deliver user value
5. **Add CLI argument parsing and validation** - User interface
6. **Implement progress reporting** - User feedback
7. **Add checkpoint/resume logic** - Reliability
8. **Build output formatting (table/JSON)** - Result presentation

All critical gaps are now documented and ready for task decomposition in planning phase.
