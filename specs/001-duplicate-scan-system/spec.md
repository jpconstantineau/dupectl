# Feature Specification: Duplicate Scan System

**Feature Branch**: `001-duplicate-scan-system`  
**Created**: December 23, 2025  
**Status**: Draft  
**Input**: User description: "the scan set of cli commands either scan all files and folders, scan all folders or scan all files. The folder command of the cli will add the folder to be monitored and start the recursive scan of the folder tree but not start the scanning of the files in each folder. Scanning of the files includes hashing each file and add their hash to the database entries. duplicates files are identified when both their sizes and hashes match. duplicate folders are found when all files within their subtree are identical according to the file matching logic. Partial folder duplicates can be found when only a subset of the files in a folder are found to be identical. The system highlights potential matches when partial folder matches but the key differences are either for missing files from one set to another and/or when files of the same name dont match but have a different date."

## Clarifications

### Session 2025-12-23

- Q: How do users configure hash algorithm choice (SHA-256, SHA-512, SHA3-256)? → A: Global configuration in config file only - all scans use same algorithm
- Q: What is the progress indication update frequency and mechanism? → A: Console output every N seconds (by default 10 seconds), configurable
- Q: When scans are interrupted, should system resume from checkpoint or restart from beginning? → A: Resume from last checkpoint - track progress in database and continue where stopped
- Q: What is the minimum similarity threshold for detecting partial folder duplicates? → A: 50% minimum similarity threshold
- Q: Should scan data (files/folders/hashes) use existing agent tables or new dedicated tables? → A: Hybrid - use existing agent/host relationships but separate tables for file/folder/hash data
- Q: How do users specify which root folder to scan when multiple roots are registered? → A: Scan commands require mandatory positional argument for root folder path
- Q: What happens when user scans a root folder path that isn't registered in database? → A: Prompt user for confirmation to register before scanning
- Q: How should system handle relative paths vs absolute paths for root folders? → A: Convert to absolute, store absolute
- Q: How should permission errors be logged during scans? → A: Console only and mark in database for future scans
- Q: What output format should the get duplicates command use? → A: Command line options for JSON or table format

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
2. **Given** folders have been scanned, **When** user later requests duplicate folder detection, **Then** the system indicates that file scanning is required before folder duplicates can be identified
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

- What happens when a scan is interrupted mid-process (system crash, user cancellation)? System resumes from last checkpoint using database-tracked progress
- What happens when user provides a root folder path that doesn't exist or isn't registered in the database? If path exists on filesystem but not in database, prompt for registration confirmation; if path doesn't exist on filesystem, reject with clear error message
- What happens when user provides a relative path vs absolute path for root folder? System converts relative paths to absolute paths using current working directory before validation and storage
- How does system handle files that cannot be read due to permissions? Display error to console, mark file in database with error status, and continue scanning remaining files
- What happens when file sizes are identical but hashes differ?
- How does system handle symbolic links, shortcuts, or hard links?
- What happens when folders contain millions of small files versus few very large files?
- How does system handle files that are modified during the scanning process?
- What happens when two folders have identical structure but zero files (empty folder trees)?
- How does system handle files with special characters or very long paths?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide three distinct scan modes: scan all (folders + files), scan folders only, scan files only
- **FR-001.1**: All scan commands MUST accept a mandatory positional argument specifying the root folder path to scan (format: `dupectl scan <mode> <root-folder-path>`)
- **FR-001.2**: If provided root folder path is not registered in database, system MUST prompt user for confirmation to register it before proceeding with scan, allowing user to cancel if path was incorrect
- **FR-001.3**: System MUST convert relative root folder paths to absolute paths before storage and processing to ensure path consistency across different working directories
- **FR-002**: System MUST recursively traverse folder trees starting from registered root folders
- **FR-003**: System MUST calculate cryptographic hash for each file during file scanning using configurable hash algorithm with options limited to secure algorithms with low collision rates (SHA-256, SHA-512, SHA3-256)
- **FR-004**: System MUST store folder structure, file metadata (path, size, modification date), hash values, and hash algorithm type in database
- **FR-004.1**: System MUST use dedicated database tables for folders and files (separate from existing agent/host/owner/policy tables) while maintaining foreign key relationships to existing RootFolder, Agent, and Host entities for organizational tracking
- **FR-004.2**: System MUST record file access errors (e.g., permission denied) in database with error status and timestamp to prevent repeated failed access attempts in subsequent scans
- **FR-005**: System MUST identify duplicate files by matching both file size AND hash value
- **FR-006**: System MUST identify duplicate folders when all files within their entire subtrees match according to file duplicate logic
- **FR-007**: System MUST detect partial folder duplicates when a subset of files match with minimum similarity threshold of 50%
- **FR-007.1**: Partial folder duplicate queries MAY allow user to specify higher similarity thresholds to filter results (e.g., 70%, 80%), but default minimum is 50%
- **FR-008**: System MUST highlight key differences in partial matches: missing files from either side and files with same name but different dates
- **FR-009**: System MUST distinguish between folder registration (folder scan) and file content analysis (file scan)
- **FR-010**: Folder scan MUST register folder hierarchy without calculating file hashes
- **FR-011**: File scan MUST process files within already-registered folders without re-traversing folder structure
- **FR-012**: System MUST persist all scan results to database for later query and analysis
- **FR-013**: System MUST provide progress indication during scan operations showing folders/files processed
- **FR-013.1**: Progress updates MUST be output to console at configurable time intervals with default of 10 seconds
- **FR-014**: System MUST handle scan interruptions gracefully by tracking progress in database and allowing resumption from last successfully processed folder/file checkpoint
- **FR-014.1**: System MUST persist scan state including current root folder, current subfolder path, and last processed file to enable reliable resumption after interruption
- **FR-015**: System MUST allow users to configure hash algorithm selection in configuration file with options: SHA-256 (default), SHA-512, SHA3-256, applied globally to all scan operations
- **FR-016**: System MUST store hash algorithm type with each file record to enable future algorithm migrations and maintain data integrity across configuration changes
- **FR-017**: System MUST provide command to query duplicate files with configurable output format: human-readable table format (default) or JSON format (via --json flag) for scripting
- **FR-017.1**: Table format output MUST group files by duplicate set, showing all files with identical size and hash together
- **FR-017.2**: Duplicate query command MUST support optional --min-count filter to return only duplicate sets with at least N files (e.g., --min-count=3 shows only files with 3+ duplicates)
- **FR-018**: System MUST provide command to query duplicate folders (exact matches) and partial folder duplicates with similarity percentage

### Key Entities

- **Root Folder**: A top-level directory registered for monitoring, serves as the starting point for recursive scans, links to existing Host, Owner, Agent, and Purpose entities from infrastructure tables
- **Folder**: A directory within the monitored tree, stored in dedicated folders table, has hierarchical relationship to parent folder and root, contains zero or more files and subfolders
- **File**: A file within a monitored folder, stored in dedicated files table, has attributes: path, size, modification date, hash value, hash algorithm type, error status (for access failures), relationship to containing folder
- **Duplicate File Set**: A group of 2+ files with identical size and hash values
- **Duplicate Folder Set**: A group of 2+ folders where all files in their complete subtrees match
- **Partial Duplicate Folder Pair**: Two folders with overlapping but not identical file sets, includes similarity percentage and difference details
- **Scan State**: Progress tracking record stored in database to enable scan resumption, includes current root folder ID, current folder path, and last processed file

### Non-Functional Requirements

- **NFR-001 Performance**: File hashing should achieve minimum throughput of 50 MB/sec on standard hardware to enable reasonable scan times for large datasets
- **NFR-002 Performance**: System should handle scanning of at least 100,000 files without memory overflow or excessive memory consumption (stay under 500 MB RAM)
- **NFR-003 Portability**: Scan operations must work identically on Windows, Linux, and macOS without platform-specific behavior
- **NFR-004 Observability**: Provide clear progress indication with counts of folders/files processed and estimated time remaining, updated at configurable intervals (default 10 seconds) to avoid excessive console output
- **NFR-005 Observability**: Log all scan operations including start/end times, files processed, errors encountered, and results summary
- **NFR-006 Security**: Handle file access permissions gracefully - display permission errors to console during scan, mark affected files in database with error status to avoid repeated attempts in future scans, and continue with remaining files without crashing
- **NFR-007 Maintainability**: Separate concerns: folder traversal logic, file hashing logic, duplicate detection logic, and database operations should be in distinct modules
- **NFR-008 Graceful Shutdown**: Handle SIGINT/SIGTERM during scans - save partial progress and provide clean exit
- **NFR-009 Upgradability**: Database schema for scan results should support versioning to enable future enhancements without data loss
- **NFR-010 Dependencies**: Justify hash algorithm library choice - prefer standard library implementations over external dependencies

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can successfully complete a full scan of a folder structure containing 10,000 files in under 5 minutes on standard hardware
- **SC-002**: System correctly identifies 100% of duplicate files in test scenarios with known duplicates (zero false positives or false negatives)
- **SC-003**: System correctly identifies 100% of duplicate folders (identical subtrees) in test scenarios
- **SC-004**: Users can distinguish between folder-only scan (fast structure mapping) and full scan (with hashing) and select appropriate mode for their needs
- **SC-004.1**: Users can query duplicate results in both human-readable table format and JSON format for different use cases (interactive vs scripting)
- **SC-005**: System detects at least 80% of meaningful partial folder duplicates in test scenarios with similarity threshold of 50% or higher
- **SC-006**: Scan operations provide visible progress indication allowing users to monitor operation without anxiety about system hang
- **SC-007**: Interrupted scans can be resumed from last checkpoint without corrupting database or losing data integrity, avoiding need to reprocess already-scanned files

## Assumptions

- **A-001**: Secure cryptographic hashing (SHA-256 or stronger) provides sufficient collision resistance for duplicate detection - probability of hash collision for different file contents is negligible
- **A-002**: SHA-256 is default hash algorithm, providing good balance of security and performance for most use cases
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
- No database tables exist for files with columns: id, path, size, mtime, hash_value, hash_algorithm, error_status, folder_id, root_folder_id
- No database tables exist for folders with columns: id, path, parent_folder_id, root_folder_id, scan_status
- No scan_state table for checkpoint tracking: root_folder_id, current_folder_path, last_processed_file, scan_mode, started_at
- Existing Folder entity in entities/files.go lacks required fields (parent_folder_id, full path, scan timestamp)
- Existing filemsg struct has hash field but no hash_algorithm or error_status fields

**Scanning Logic (Priority: Must Have)**
- No folder traversal implementation - cmd/scanFolders.go is stub only
- No file hashing implementation - no pkg/scanner or pkg/hash module exists
- No hash algorithm configuration reading from config file
- No progress reporting mechanism with configurable time intervals
- No checkpoint save/restore logic for scan interruption handling
- No permission error handling with database marking

**Duplicate Detection (Priority: Must Have)**
- No logic to identify duplicate files by matching size AND hash
- No logic to identify duplicate folders (identical subtrees)
- No logic to detect partial folder duplicates with similarity calculation
- cmd/getDuplicates.go is stub with no implementation

**CLI Command Implementation (Priority: Must Have)**
- cmd/scanAll.go: No root folder path argument parsing, no call to scan logic
- cmd/scanFolders.go: No implementation, needs folder-only traversal
- cmd/scanFiles.go: No implementation, needs file-only hashing
- cmd/getDuplicates.go: No output formatting (table/JSON), no --min-count filter
- cmd/addRoot.go: No validation for path existence, no absolute path conversion, no registration prompt

**Configuration (Priority: Must Have)**
- No hash algorithm config setting in viper defaults (setDefaults in root.go)
- No progress interval config setting
- No configuration documentation for scan settings

### Moderate Gaps

**Path Handling (Priority: Should Have)**
- No absolute path conversion logic for relative paths
- No path validation and normalization across platforms
- No handling of special characters or long paths

**Error Handling (Priority: Should Have)**
- No console warning output for permission errors during scan
- No graceful handling of unreadable files with continuation
- No validation that provided root path exists on filesystem

**Data Integrity (Priority: Should Have)**
- No foreign key relationships defined between new scan tables and existing infrastructure tables
- No database migration/versioning strategy for schema changes

**User Experience (Priority: Nice to Have)**
- No interactive confirmation prompts for unregistered root folders
- No progress estimation (time remaining) in progress output
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
