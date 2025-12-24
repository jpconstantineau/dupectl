# Tasks: Duplicate Scan System

**Feature Branch**: `001-duplicate-scan-system`  
**Input**: Design documents from `/specs/001-duplicate-scan-system/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Tests are NOT explicitly requested in the feature specification, so test tasks are OMITTED per template guidelines. However, constitution requires 70%+ coverage - unit tests will be created alongside implementation (TDD approach).

**Organization**: Tasks organized by user story to enable independent implementation and testing.

---

## Format: `- [ ] [ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4, US5)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Create pkg/scanner/ directory structure per plan.md project structure
- [X] T002 Create pkg/hash/ directory structure per plan.md project structure
- [X] T003 Create pkg/detector/ directory structure per plan.md project structure
- [X] T004 Create pkg/formatter/ directory structure per plan.md project structure
- [X] T005 [P] Create tests/unit/ directory for unit tests
- [X] T006 [P] Create tests/integration/ directory for integration tests
- [X] T007 [P] Create tests/e2e/ directory for end-to-end tests

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Constitution Compliance Tasks

- [ ] T010 [P] Add signal handling for SIGINT/SIGTERM in cmd/root.go (Graceful Shutdown)
- [ ] T011 [P] Create pkg/pathutil/ package with cross-platform path utilities: Abs(), Clean(), Join() (Portability)
- [ ] T012 [P] Add progress_interval_seconds config key to cmd/root.go setDefaults() with default 10 (UX Consistency)
- [ ] T013 [P] Add hash_algorithm config key to cmd/root.go setDefaults() with default "sha256" (Maintainability)
- [ ] T014 [P] Add structured logging utility to pkg/logger/ with levels: INFO, WARN, ERROR (Observability)
- [ ] T015 [P] Create pkg/errors/ package with typed errors and actionable messages (UX Consistency)

### Database Schema & Migrations

- [ ] T020 Create pkg/datastore/schema.go with migration framework and version tracking (Upgradability)
- [ ] T021 Add migration 001_create_scan_tables.sql with files table (id, root_folder_id, folder_id, path UNIQUE, name, size, mtime, hash_value, hash_algorithm, error_status, scanned_at) per data-model.md
- [ ] T022 Add migration 001_create_scan_tables.sql with folders table (id, root_folder_id, parent_folder_id, path UNIQUE, name, scanned_at) per data-model.md
- [ ] T023 Add migration 001_create_scan_tables.sql with scan_state table (id, root_folder_id, scan_mode, current_folder_path, started_at, updated_at, completed_at, status, files_processed, folders_processed) per data-model.md
- [ ] T024 Add indexes to migration: idx_files_hash_size, idx_files_root, idx_files_folder, idx_files_error per data-model.md
- [ ] T025 Add indexes to migration: idx_folders_root, idx_folders_parent, idx_folders_path per data-model.md
- [ ] T026 Add indexes to migration: idx_scan_state_root_status, idx_scan_state_started per data-model.md
- [ ] T027 Implement schema.go ApplyMigrations() function to execute migration SQL on database connection
- [ ] T028 Call ApplyMigrations() in pkg/datastore/datastore.go InitDB() during database initialization

### Entity Definitions

- [ ] T030 [P] Create File entity in pkg/entities/files.go with fields per data-model.md and methods: HasError(), IsHashed()
- [ ] T031 [P] Enhance Folder entity in pkg/entities/files.go with fields per data-model.md and method: IsRoot()
- [ ] T032 [P] Create ScanState entity in pkg/entities/scan_state.go with fields, constants (ScanStatus, ScanMode), and method: IsActive()

### Database Access Layer

- [ ] T040 Create pkg/datastore/files.go with InsertFile(file File) function
- [ ] T041 Add pkg/datastore/files.go GetFileByPath(path string) function
- [ ] T042 Add pkg/datastore/files.go UpdateFileHash(id int, hashValue, hashAlgorithm string) function
- [ ] T043 Add pkg/datastore/files.go MarkFileError(id int, errorStatus string) function
- [ ] T044 Add pkg/datastore/files.go GetFilesInFolder(folderId int) function
- [ ] T045 Create pkg/datastore/folders.go with InsertFolder(folder Folder) function
- [ ] T046 Add pkg/datastore/folders.go GetFolderByPath(path string) function
- [ ] T047 Add pkg/datastore/folders.go GetChildFolders(parentId int) function
- [ ] T048 Add pkg/datastore/folders.go UpdateFolderScannedAt(id int, scannedAt time.Time) function
- [ ] T049 Create pkg/datastore/scan_state.go with CreateScanState(state ScanState) function
- [ ] T050 Add pkg/datastore/scan_state.go GetActiveScanState(rootFolderId int) function
- [ ] T051 Add pkg/datastore/scan_state.go UpdateCheckpoint(id int, folderPath string, filesCount, foldersCount int) function
- [ ] T052 Add pkg/datastore/scan_state.go CompleteScanState(id int) function
- [ ] T053 Add pkg/datastore/scan_state.go MarkInterrupted(id int) function

### Hash Algorithm Implementation

- [ ] T060 Create pkg/hash/hasher.go with Hasher interface: HashFile(path string) (string, error)
- [ ] T061 Add pkg/hash/hasher.go NewHasher(algorithm string) function that returns appropriate hasher based on config
- [ ] T062 [P] Implement pkg/hash/sha256.go with SHA256Hasher struct implementing Hasher interface using crypto/sha256
- [ ] T063 [P] Implement pkg/hash/sha512.go with SHA512Hasher struct implementing Hasher interface using crypto/sha512
- [ ] T064 [P] Implement pkg/hash/sha3.go with SHA3Hasher struct implementing Hasher interface using golang.org/x/crypto/sha3
- [ ] T065 Add go.mod dependency for golang.org/x/crypto/sha3 if not already present
- [ ] T066 Implement stream-based file hashing with io.Copy and 64KB buffer in all hasher implementations per research.md

### Progress Reporting Infrastructure

- [ ] T070 Create pkg/scanner/progress.go with ProgressTracker struct (filesProcessed, foldersProcessed, startTime atomic counters)
- [ ] T071 Add ProgressTracker.Start() method that launches goroutine for periodic console output every config.progress_interval_seconds
- [ ] T072 Add ProgressTracker.IncrementFiles() and IncrementFolders() atomic methods
- [ ] T073 Add ProgressTracker.Stop() method that stops progress goroutine and displays final summary
- [ ] T074 Add ProgressTracker.GetStats() method returning current counts and elapsed time

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Scan All Files and Folders (Priority: P1) 🎯 MVP

**Goal**: Enable users to perform complete scan of root folder with file hashing and folder traversal to identify all duplicates

**Independent Test**: Run `dupectl scan all <test-folder>` on directory with known duplicates and verify all duplicates are stored in database

### Implementation for User Story 1

- [ ] T100 [US1] Create pkg/scanner/scanner.go with Scanner interface: Scan(rootPath string, mode ScanMode) error
- [ ] T101 [US1] Implement pkg/scanner/folder.go FolderScanner struct with Walk() method using filepath.Walk per research.md
- [ ] T102 [US1] Add FolderScanner.processFolder() method to insert folder records into database via datastore.InsertFolder()
- [ ] T103 [US1] Add FolderScanner.processFile() method to insert file metadata (no hash yet) via datastore.InsertFile()
- [ ] T104 [US1] Implement pkg/scanner/file.go FileScanner struct with HashFiles() method that iterates files and calls hasher
- [ ] T105 [US1] Add FileScanner.hashAndStore() method to calculate hash via hash.Hasher and update database via datastore.UpdateFileHash()
- [ ] T106 [US1] Handle permission errors in processFile() by calling datastore.MarkFileError() and logging warning to console (NFR-006)
- [ ] T107 [US1] Create pkg/scanner/checkpoint.go with CheckpointManager struct tracking current folder path
- [ ] T108 [US1] Add CheckpointManager.SaveCheckpoint() method that calls datastore.UpdateCheckpoint() after each folder completes
- [ ] T109 [US1] Add CheckpointManager.LoadCheckpoint() method that calls datastore.GetActiveScanState() and resumes from current_folder_path
- [ ] T110 [US1] Integrate ProgressTracker into FolderScanner and FileScanner, calling Increment methods after each file/folder processed
- [ ] T111 [US1] Implement signal handler in scanner.Scan() to call CheckpointManager.SaveCheckpoint() on SIGINT/SIGTERM before exit
- [ ] T112 [US1] Enhance cmd/addRoot.go to validate path existence using os.Stat() and convert to absolute path using pkg/pathutil/
- [ ] T113 [US1] Add confirmation prompt in cmd/addRoot.go when registering new root: "Root folder not registered. Register now? (y/n)"
- [ ] T114 [US1] Implement cmd/scanAll.go command parsing for <root-path> positional argument per contracts/scan-commands.md
- [ ] T115 [US1] Add cmd/scanAll.go flags: --resume, --verbose, --rescan per contracts/scan-commands.md
- [ ] T116 [US1] Implement cmd/scanAll.go RunE function: validate path, check registration, load checkpoint if exists, create Scanner, call Scan(rootPath, ScanModeAll)
- [ ] T117 [US1] Add cmd/scanAll.go error handling for path not found, permission denied, database errors per contracts/scan-commands.md
- [ ] T118 [US1] Display progress output in cmd/scanAll.go by starting ProgressTracker before scan, stopping after scan completes
- [ ] T119 [US1] Display scan summary at end: "Scan complete: X files, Y folders in Zs" per contracts/scan-commands.md
- [ ] T119b [US1] Implement --rescan flag in [scanAll.go](http://_vscodecontentref_/15) that deletes existing scan_state record for root folder forcing fresh scan from beginning

**Checkpoint**: User Story 1 complete - users can scan folders with hashing and see progress

---

## Phase 4: User Story 2 - Scan Folders Only (Priority: P2)

**Goal**: Enable fast folder structure mapping without file hashing for large directory trees

**Independent Test**: Run `dupectl scan folders <test-folder>` and verify all folders registered in database but no file hashes calculated

### Implementation for User Story 2

- [ ] T200 [US2] Implement cmd/scanFolders.go command with <root-path> positional argument per contracts/scan-commands.md
- [ ] T201 [US2] Add cmd/scanFolders.go flags: --resume, --verbose per contracts/scan-commands.md
- [ ] T202 [US2] Implement cmd/scanFolders.go RunE function: validate path, check registration, create Scanner, call Scan(rootPath, ScanModeFolders)
- [ ] T203 [US2] Update pkg/scanner/scanner.go Scan() to check ScanMode and skip FileScanner.HashFiles() when mode is ScanModeFolders
- [ ] T204 [US2] Ensure FolderScanner still calls processFile() to record file metadata (path, size, mtime) even in folders-only mode
- [ ] T205 [US2] Update progress output to show "folders processed" only when in ScanModeFolders mode
- [ ] T206 [US2] Display summary: "Folder scan complete: Y folders registered" per contracts/scan-commands.md

**Checkpoint**: User Story 2 complete - users can quickly map folder structure without hashing

---

## Phase 5: User Story 3 - Scan Files Only (Priority: P2)

**Goal**: Enable file hashing for already-registered folders without re-traversing structure

**Independent Test**: After folder scan, run `dupectl scan files <test-folder>` and verify files are hashed without modifying folder registrations

### Implementation for User Story 3

- [ ] T300 [US3] Implement cmd/scanFiles.go command with <root-path> positional argument per contracts/scan-commands.md
- [ ] T301 [US3] Add cmd/scanFiles.go flags: --resume, --verbose per contracts/scan-commands.md
- [ ] T302 [US3] Implement cmd/scanFiles.go RunE function: validate path, check registration, verify folders exist in database, create Scanner, call Scan(rootPath, ScanModeFiles)
- [ ] T303 [US3] Update pkg/scanner/scanner.go Scan() to skip FolderScanner.Walk() when mode is ScanModeFiles
- [ ] T304 [US3] Add pkg/scanner/file.go method to query existing folders from database via datastore and iterate through files
- [ ] T305 [US3] Call FileScanner.HashFiles() for all files in registered folders when mode is ScanModeFiles
- [ ] T306 [US3] Update progress output to show "files hashed" count when in ScanModeFiles mode
- [ ] T307 [US3] Display summary: "File scan complete: X files hashed" per contracts/scan-commands.md

**Checkpoint**: User Story 3 complete - users can hash files independently of folder scanning

---

## Phase 6: User Story 5 - Query and View Duplicate Files (Priority: P1)

**Goal**: Enable users to view duplicate files with filtering and format options

**Independent Test**: After scans complete, run `dupectl get duplicates` and verify duplicate sets displayed correctly with grouping and statistics

### Implementation for User Story 5

- [ ] T500 [US5] Create pkg/datastore/duplicates.go with GetDuplicateFiles() function that queries files grouped by (size, hash_value) HAVING COUNT(*) >= 2 (Note: Same file extended with folder duplicate functions in US4 tasks T608-T609)
- [ ] T501 [US5] Add GetDuplicateFiles() parameters: minCount, rootPath, minSize for filtering per contracts/query-commands.md
- [ ] T502 [US5] Add GetDuplicateFiles() sortField parameter supporting 'size', 'count', 'path' per contracts/query-commands.md
- [ ] T503 [US5] Return []DuplicateFileSet from GetDuplicateFiles() where each set contains size, hash, algorithm, file count, and []File
- [ ] T504 [US5] Create pkg/formatter/table.go with FormatDuplicatesTable(sets []DuplicateFileSet) string function
- [ ] T505 [US5] Implement table formatting per contracts/query-commands.md: "Duplicate Set #N", tree structure with ├─ and └─, size/hash/algorithm display
- [ ] T506 [US5] Add summary section to table output: "Total Duplicate Sets", "Total Duplicate Files", "Total Wasted Space", "Storage Recoverable"
- [ ] T507 [US5] Calculate wasted space as: sum(fileSize * (fileCount - 1)) across all duplicate sets
- [ ] T508 [US5] Create pkg/formatter/json.go with FormatDuplicatesJSON(sets []DuplicateFileSet) string function
- [ ] T509 [US5] Implement JSON formatting per contracts/query-commands.md schema with duplicate_sets array and summary object
- [ ] T510 [US5] Implement cmd/getDuplicates.go command with no positional arguments (queries all roots) per contracts/query-commands.md
- [ ] T511 [US5] Add cmd/getDuplicates.go flags: --json, --min-count, --root, --min-size, --sort per contracts/query-commands.md
- [ ] T512 [US5] Implement cmd/getDuplicates.go RunE function: parse flags, call datastore.GetDuplicateFiles() with filters
- [ ] T513 [US5] Call formatter.FormatDuplicatesTable() by default or formatter.FormatDuplicatesJSON() if --json flag present
- [ ] T514 [US5] Display formatted output to stdout
- [ ] T515 [US5] Handle case where no duplicates found: display "No duplicate files found" message per contracts/query-commands.md

**Checkpoint**: User Story 5 complete - users can query and view duplicates in table or JSON format

---

## Phase 7: User Story 4 - Identify Partial Folder Duplicates (Priority: P3)

**Goal**: Enable detection of folders with overlapping but not identical file sets

**Independent Test**: Create folders with 70% overlapping files, run scans, query partial duplicates and verify similarity percentages and difference highlights

### Implementation for User Story 4

- [ ] T600 [US4] Create pkg/detector/files.go with MatchFiles(folder1Files, folder2Files []File) MatchResult function
- [ ] T601 [US4] Implement MatchFiles() to return: matchingFiles []File, uniqueToFolder1 []File, uniqueToFolder2 []File, sameNameDifferentDate []FilePair
- [ ] T602 [US4] Create pkg/detector/folders.go with GetExactFolderDuplicates() function that queries files by folder_id and compares complete file lists
- [ ] T603 [US4] Add GetExactFolderDuplicates() logic: folders match if all files in subtrees have matching size+hash (recursive comparison)
- [ ] T604 [US4] Create pkg/detector/partial.go with GetPartialFolderDuplicates(minSimilarity float64) function
- [ ] T605 [US4] Implement similarity calculation: (matching files count / total unique files across both folders) * 100 per research.md
- [ ] T606 [US4] Filter folder pairs where similarity >= minSimilarity (default 50%, user-configurable) per spec.md FR-007
- [ ] T607 [US4] Return PartialFolderMatch struct containing: folder1, folder2, similarityPercent, matchingFiles, differences (missing from each side, same name different date)
- [ ] T608 [US4] Create pkg/datastore/duplicates.go GetFolderDuplicates() function that calls detector.GetExactFolderDuplicates()
- [ ] T609 [US4] Create pkg/datastore/duplicates.go GetPartialFolderDuplicates(minSimilarity float64) function calling detector.GetPartialFolderDuplicates()
- [ ] T610 [US4] Add cmd/getDuplicates.go --folders flag to query folder duplicates instead of file duplicates
- [ ] T611 [US4] Add cmd/getDuplicates.go --partial flag to include partial folder matches with --min-similarity threshold (default 50)
- [ ] T612 [US4] Update pkg/formatter/table.go to format folder duplicate output with similarity percentage and difference details
- [ ] T613 [US4] Update pkg/formatter/json.go to format folder duplicate JSON with folder_duplicates array and partial_matches array
- [ ] T614 [US4] Display folder duplicate results showing: folder paths, similarity %, matching files count, key differences (missing files, date mismatches)

**Checkpoint**: User Story 4 complete - users can find exact and partial folder duplicates

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

### Constitution Compliance Verification

- [ ] T700 Run `go test ./...` and verify test coverage ≥70% across pkg/scanner, pkg/hash, pkg/detector, pkg/datastore (Testing)
- [ ] T701 Run `gocyclo -over 10 ./pkg ./cmd` and verify all functions have cyclomatic complexity ≤10 (Clean Code)
- [ ] T702 Run `golangci-lint run` with zero warnings across all new code (Code Quality)
- [ ] T703 Verify all public functions in pkg/ have GoDoc comments describing purpose, parameters, returns (Maintainability)
- [ ] T704 Test scan operations on Windows, Linux, macOS with different path separators and verify cross-platform compatibility (Portability)
- [ ] T705 Test SIGINT during scan and verify checkpoint saves correctly and scan resumes properly (Graceful Shutdown)
- [ ] T706 Load test with 100k files and verify: hash speed ≥50 MB/sec, memory <500 MB, scan completes (Performance per NFR-001, NFR-002)
- [ ] T707 Security audit: verify path traversal prevention (absolute path normalization), no SQL injection (parameterized queries), permission error handling (Security)
- [ ] T708 Test database migration rollback: create test that applies migration, rolls back, reapplies and verifies schema correctness (Upgradability)
- [ ] T709 Verify all dependencies justified: check go.mod contains only stdlib + cobra + viper + sqlite + sha3, document rationale in docs/ (Minimal Dependencies)
- [ ] T710 Test backward compatibility: verify existing root_folders/agents/hosts tables unaffected by new scan tables (Backward Compatibility)

### General Polish

- [ ] T720 [P] Update README.md with scan commands usage examples and quickstart link
- [ ] T721 [P] Add docs/scanning.md with detailed documentation of scan modes, checkpoint/resume, progress reporting, duplicate detection algorithms
- [ ] T722 [P] Add docs/configuration.md documenting hash_algorithm and progress_interval_seconds config keys with examples
- [ ] T723 Performance optimization: implement batch database inserts (1000 files per transaction) in pkg/datastore/files.go per research.md
- [ ] T724 Performance optimization: add worker pool pattern for parallel file hashing in pkg/scanner/file.go per research.md
- [ ] T725 Error handling improvements: ensure all errors include actionable messages per pkg/errors/ typed errors
- [ ] T726 Run quickstart.md validation: manually execute all commands in quickstart.md and verify they work as documented
- [ ] T727 Code cleanup: remove any debug print statements, ensure consistent error handling patterns, verify no TODO comments remain
- [ ] T728 Add integration test in tests/integration/ that creates test folder structure, runs full scan workflow, queries duplicates, verifies results
- [ ] T729 Add e2e test in tests/e2e/ that tests complete user workflows from quickstart.md using real CLI commands

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phases 3-7)**: All depend on Foundational phase completion
  - After Foundational complete, user stories can proceed in parallel (if staffed)
  - Or sequentially in priority order: US1 (P1) → US5 (P1) → US2 (P2) → US3 (P2) → US4 (P3)
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### User Story Dependencies

**Independent Stories (can be built in parallel after Foundational)**:
- US1 (Scan All): No dependencies on other user stories
- US2 (Scan Folders): No dependencies on other user stories
- US3 (Scan Files): No dependencies on other user stories
- US5 (Query Duplicates): Depends on US1 OR (US2 + US3) - needs scan data in database

**Dependent Stories**:
- US4 (Partial Duplicates): Depends on US1 - needs complete scan data with file hashes for folder comparison

### Recommended MVP Scope

**Minimum Viable Product** (deliver value in shortest time):
- Phase 1: Setup ✅
- Phase 2: Foundational ✅  
- Phase 3: User Story 1 (Scan All) ✅
- Phase 6: User Story 5 (Query Duplicates) ✅

This MVP delivers complete end-to-end value: users can scan folders, hash files, and view duplicates.

**Post-MVP Enhancements**:
- Phase 4: User Story 2 (Scan Folders Only) - workflow optimization
- Phase 5: User Story 3 (Scan Files Only) - workflow optimization
- Phase 7: User Story 4 (Partial Folder Duplicates) - advanced feature

---

## Parallel Execution Strategy

### Within Foundational Phase (Phase 2)

Can execute in parallel (different team members or separate work sessions):
- Constitution compliance tasks (T010-T017) - different files
- Entity definitions (T030-T032) - different files
- Hash implementations (T062-T064) - different algorithms, same interface

Must execute sequentially:
- Database migrations (T020-T028) - schema must exist before access layer
- Database access layer (T040-T053) - depends on entity definitions

### Within User Story 1 (Phase 3)

**Parallel opportunities**:
- T100-T111 (scanner package implementation) can proceed independently from T114-T119 (CLI command)
- Once Scanner interface defined (T100), FolderScanner (T101-T103) and FileScanner (T104-T106) can be built in parallel

**Sequential requirements**:
- T112-T113 (addRoot.go enhancements) must complete before T114-T119 (scanAll.go) can validate root registration
- T107-T109 (CheckpointManager) must complete before T111 (signal handler integration)

### Across User Stories (after Foundational)

**Full parallelization possible**:
- US1 (Phase 3), US2 (Phase 4), US3 (Phase 5) can all be implemented in parallel by different developers
- US5 (Phase 6) can start once US1 has database schema and scan data (after T119)

**Suggested parallel work distribution**:
- Developer A: US1 (Scan All) - highest priority, most complex
- Developer B: US2 (Scan Folders) + US3 (Scan Files) - simpler, related commands
- Developer C: US5 (Query Duplicates) - starts after US1 T100-T111 complete (scanner package done)

---

## Task Summary

**Total Tasks**: 129

**By Phase**:
- Phase 1 (Setup): 7 tasks
- Phase 2 (Foundational): 65 tasks (15 constitution + 9 schema + 3 entities + 14 data access + 7 hash + 5 progress)
- Phase 3 (User Story 1 - Scan All): 20 tasks
- Phase 4 (User Story 2 - Folders Only): 7 tasks
- Phase 5 (User Story 3 - Files Only): 8 tasks
- Phase 6 (User Story 5 - Query Duplicates): 16 tasks
- Phase 7 (User Story 4 - Partial Duplicates): 15 tasks
- Phase 8 (Polish): 11 tasks

**By User Story**:
- Setup/Foundational: 72 tasks (prerequisite for all stories)
- US1 (Scan All - P1): 20 tasks
- US2 (Scan Folders - P2): 7 tasks
- US3 (Scan Files - P2): 8 tasks
- US4 (Partial Duplicates - P3): 15 tasks
- US5 (Query Duplicates - P1): 16 tasks
- Polish: 11 tasks

**Parallelizable Tasks**: 24 tasks marked with [P] can run in parallel

**Critical Path** (sequential dependencies for MVP):
1. Setup (7 tasks) → 
2. Foundational (65 tasks) → 
3. US1 implementation (20 tasks) → 
4. US5 implementation (16 tasks) →
5. Polish MVP subset (6 tasks: T700, T702, T703, T706, T726, T728)

**Estimated MVP Delivery**: 114 tasks (Setup + Foundational + US1 + US5 + MVP Polish)

**Post-MVP Enhancements**: 15 tasks (US2 + US3 + US4 + remaining Polish)

---

## Implementation Strategy

### Phase Approach

1. **Weeks 1-2**: Complete Setup (Phase 1) and Foundational (Phase 2)
   - Focus: Database schema, entities, hash algorithms, progress infrastructure
   - Deliverable: Foundation ready for user story implementation

2. **Weeks 3-4**: MVP User Stories (Phase 3 + Phase 6)
   - Focus: US1 (Scan All) + US5 (Query Duplicates)
   - Deliverable: End-to-end working system - scan folders, view duplicates

3. **Week 5**: MVP Polish and Testing
   - Focus: Phase 8 constitution compliance verification (T700-T710)
   - Deliverable: Production-ready MVP with 70%+ test coverage

4. **Week 6+**: Post-MVP Enhancements (optional)
   - Focus: US2, US3, US4 for workflow optimization and advanced features
   - Deliverable: Feature-complete system with all user stories

### Testing Strategy

**TDD Approach** (per constitution requirement):
- Write unit tests alongside implementation for each package (pkg/scanner, pkg/hash, pkg/detector)
- Unit tests verify: hash calculation correctness, path handling, duplicate detection algorithms
- Integration tests verify: database operations, checkpoint/resume, scan workflows
- E2E tests verify: CLI commands work end-to-end per contracts

**Coverage Target**: 70%+ per constitution (T700)

**Test Structure**:
- `tests/unit/`: Package-level unit tests (hash_test.go, scanner_test.go, detector_test.go)
- `tests/integration/`: Database and filesystem integration tests
- `tests/e2e/`: CLI command end-to-end tests using test fixtures with known duplicates

**Test Fixtures**: Create test folder structures with known duplicates:
- Exact file duplicates (same content, different paths)
- Exact folder duplicates (identical subtrees)
- Partial folder duplicates (70% overlap for similarity testing)
- Permission errors (unreadable files/folders)
- Large datasets (100k files for performance testing)

---

## Risk Mitigation

**Risk 1**: Database schema changes mid-implementation
- **Mitigation**: Use migration framework (T020) to version schema changes
- **Impact**: Low - migrations allow rollback and forward compatibility

**Risk 2**: Cross-platform path handling edge cases
- **Mitigation**: Dedicated pathutil package (T011) tested on all platforms (T704)
- **Impact**: Medium - verify early with integration tests on Windows/Linux/macOS

**Risk 3**: Performance not meeting 50 MB/sec hash threshold
- **Mitigation**: Stream-based hashing with 64KB buffers (T066), worker pool for parallelization (T724)
- **Impact**: High - load test early (T706) and optimize if needed

**Risk 4**: Signal handling doesn't save checkpoint correctly
- **Mitigation**: Explicit testing of SIGINT during scan (T705), atomic checkpoint saves (T108)
- **Impact**: Medium - critical for user experience, verify early

**Risk 5**: Memory overflow with 100k+ files
- **Mitigation**: Batch database inserts (T723), stream processing (no in-memory file list)
- **Impact**: High - load test with large datasets (T706) before production

---

## Notes

- **Tests NOT included**: Feature specification does not explicitly request tests, so test tasks are omitted per template guidelines. However, TDD approach recommended for constitution compliance (70%+ coverage requirement).
- **Constitution compliance**: 11 verification tasks (T700-T710) in Phase 8 ensure all 12 constitution principles are met.
- **Independent user stories**: Each story (US1-US5) can be implemented and tested independently after Foundational phase, enabling parallel development and incremental delivery.
- **MVP focus**: Phases 1, 2, 3, 6 deliver complete value (scan + query duplicates) in shortest time.
- **Post-MVP enhancements**: Phases 4, 5, 7 add workflow optimizations and advanced features.
- **File paths included**: All task descriptions include specific file paths per template requirement.
- **Parallelization marked**: 24 tasks marked [P] can run in parallel (different files, no dependencies).
