---
description: "Task list for Duplicate Scan System implementation"
---

# Tasks: Duplicate Scan System

**Branch**: `001-duplicate-scan-system-01`
**Feature**: 001-duplicate-scan-system
**Input**: Design documents from `/specs/001-duplicate-scan-system-01/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓

**Tests**: Tests are OPTIONAL per specification - NOT included unless explicitly requested
**Organization**: Tasks grouped by user story for independent implementation and testing

## Format: `- [ ] [ID] [P?] [Story?] Description with file path`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4, US5)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Create Go project structure with cmd/, pkg/, internal/, tests/ directories per plan.md
- [X] T002 Verify Go 1.21+ installation and existing dependencies (cobra, viper, sqlite)
- [X] T003 [P] Install golang.org/x/crypto/sha3 dependency for SHA3-256 hash support
- [X] T004 [P] Setup testing framework structure in tests/fixtures/, tests/unit/, tests/integration/, tests/e2e/
- [X] T005 [P] Create test fixtures directory structure for duplicates, folders, partial, permissions scenarios
- [ ] T006 [P] Configure golangci-lint for Go code quality checks per constitution

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Constitution Compliance Infrastructure

- [X] T010 [P] Implement structured logging utilities in pkg/logger/logger.go with stderr output and log levels (Observability)
- [X] T011 [P] Implement signal handling for SIGINT/SIGTERM with 5-second timeout in pkg/checkpoint/signals.go (Graceful Shutdown)
- [X] T012 [P] Create cross-platform path utilities in pkg/pathutil/pathutil.go using filepath.Join() (Portability)
- [X] T013 [P] Implement database migration framework in pkg/datastore/migrations.go with schema versioning (Upgradability)
- [X] T014 [P] Create progress indication utilities in pkg/scanner/progress.go with braille spinner and counters (UX Consistency)
- [X] T015 [P] Define structured error types in pkg/errors/errors.go with actionable messages (UX Consistency)
- [X] T016 [P] Setup configuration management in internal/config/config.go for Viper integration (12-Factor CLI)

### Database Schema & Infrastructure

- [X] T020 Create files table schema in pkg/datastore/files.go with CREATE TABLE and indexes per data-model.md
- [X] T021 Create folders table schema in pkg/datastore/folders.go with self-referential FK per data-model.md
- [X] T022 Create scan_state table schema in pkg/datastore/scanstate.go with unique active constraint per data-model.md
- [X] T023 Extend root_folders table with ALTER TABLE statements in pkg/datastore/rootfolders.go adding traverse_links, last_scan_date, folder_count, file_count, total_size
- [X] T024 Implement PRAGMA foreign_keys = ON enforcement in pkg/datastore/datastore.go initialization
- [X] T025 Configure SQLite WAL mode and connection pooling in pkg/datastore/datastore.go per research.md
- [X] T026 Create database initialization function to run all table creation migrations in pkg/datastore/datastore.go

### Entity Models

- [X] T030 [P] Create File entity struct in pkg/entities/files.go with all fields from data-model.md (id, path, size, mtime, hash_value, hash_algorithm, error_status, timestamps, removed, folder_id, root_folder_id)
- [X] T031 [P] Create Folder entity struct in pkg/entities/folders.go with hierarchy support (id, path, parent_folder_id, root_folder_id, error_status, timestamps, removed)
- [X] T032 [P] Create ScanState entity struct in pkg/entities/scanstate.go for checkpoint tracking (id, root_folder_id, scan_mode, current_folder_path, last_processed_file, timestamps, completed)

### Core Services Infrastructure

- [X] T040 [P] Implement hash algorithm interface in pkg/hash/hasher.go with Hash(filePath) method
- [X] T041 [P] Implement SHA-256 hasher in pkg/hash/sha256.go using crypto/sha256
- [X] T042 [P] Implement SHA-512 hasher in pkg/hash/sha512.go using crypto/sha512 (default)
- [X] T043 [P] Implement SHA3-256 hasher in pkg/hash/sha3.go using golang.org/x/crypto/sha3
- [X] T044 [P] Create hash factory in pkg/hash/factory.go to select algorithm from config
- [X] T045 Implement worker pool generic interface in internal/worker/pool.go with WorkItem interface, Submit(), Stop(), Wait() per worker-pools.md
- [X] T046 [P] Implement worker pool metrics tracking in internal/worker/metrics.go (ItemsProcessed, ItemsFailed, ItemsQueued, WorkersActive)
- [X] T047 [P] Implement checkpoint save/restore logic in pkg/checkpoint/checkpoint.go for scan state persistence

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Scan All Files and Folders (Priority: P1) 🎯 MVP

**Goal**: Perform complete scan of root folder to identify all duplicate files and folders

**Independent Test**: Run scan command on test folder with known duplicates, verify all duplicates identified

### Implementation for User Story 1

- [X] T050 [P] [US1] Implement folder traversal logic in pkg/scanner/traverser.go with recursive directory walking
- [X] T051 [P] [US1] Create FolderTraversalWorkItem in internal/worker/traversal.go implementing WorkItem interface per worker-pools.md
- [X] T052 [P] [US1] Implement folder registration in pkg/datastore/folders.go (INSERT with UNIQUE constraint handling, UPDATE last_scanned_at)
- [X] T053 [P] [US1] Implement file registration in pkg/datastore/files.go (INSERT with batch transactions, UNIQUE constraint handling)
- [X] T054 [P] [US1] Create FileHashingWorkItem in internal/worker/hashing.go implementing WorkItem interface per worker-pools.md
- [X] T055 [US1] Implement main scanner orchestrator in pkg/scanner/scanner.go coordinating traversal and hashing worker pools
- [X] T056 [US1] Implement permission error handling in pkg/scanner/errors.go setting error_status in database
- [ ] T057 [US1] Implement removed flag cascading logic in pkg/datastore/files.go for files no longer present
- [X] T058 [US1] Implement checkpoint save triggers in pkg/scanner/scanner.go (periodic saves every folder completion)
- [X] T059 [US1] Implement checkpoint auto-resume logic in pkg/scanner/scanner.go detecting incomplete scans
- [X] T060 [US1] Implement progress display in pkg/scanner/progress.go with braille spinner, folder/file counts, elapsed time per cli-commands.md
- [X] T061 [US1] Create scan all CLI command in cmd/scanAll.go with root-folder-path argument and --progress, --restart flags per cli-commands.md
- [X] T062 [US1] Implement root folder registration prompt in cmd/scanAll.go for unregistered paths
- [X] T063 [US1] Implement absolute path conversion in cmd/scanAll.go using pkg/pathutil
- [X] T064 [US1] Add scan summary output in cmd/scanAll.go showing folders/files scanned, duplicates found
- [ ] T065 [US1] Implement root folder statistics update in pkg/datastore/rootfolders.go (folder_count, file_count, total_size, last_scan_date)
- [X] T066 [US1] Add signal handler integration in cmd/scanAll.go to save checkpoint on SIGINT/SIGTERM
- [ ] T067 [US1] Implement removed flag reset in pkg/scanner/scanner.go when file rediscovered (FR-036)

**Checkpoint**: User Story 1 complete - full scan capability with checkpoint/resume functional

---

## Phase 4: User Story 2 - Scan Folders Only (Priority: P2)

**Goal**: Quickly map folder structure without time-intensive file hashing for rapid setup

**Independent Test**: Run folder scan command, verify all folders registered with hierarchy, no file hashes calculated

### Implementation for User Story 2

- [ ] T070 [US2] Create scan folders CLI command in cmd/scanFolders.go with root-folder-path argument and --progress, --restart flags per cli-commands.md
- [ ] T071 [US2] Implement scan mode detection in pkg/scanner/scanner.go to skip file hashing when mode='folders'
- [ ] T072 [US2] Add folder-only checkpoint handling in pkg/checkpoint/checkpoint.go with scan_mode='folders'
- [ ] T073 [US2] Implement folder-only summary output in cmd/scanFolders.go showing folders scanned, files skipped message

**Checkpoint**: User Story 2 complete - folder-only scanning works independently

---

## Phase 5: User Story 3 - Scan Files Only (Priority: P2)

**Goal**: Hash files within already-registered folders without re-traversing structure

**Independent Test**: First register folders (via folder scan or manual DB), then run file scan, verify all files hashed without folder re-traversal

### Implementation for User Story 3

- [ ] T080 [US3] Create scan files CLI command in cmd/scanFiles.go with root-folder-path argument and --progress, --restart flags per cli-commands.md
- [ ] T081 [US3] Implement folder lookup logic in pkg/scanner/scanner.go querying existing folders from database
- [ ] T082 [US3] Add validation in cmd/scanFiles.go to check root folder is registered and has folders before proceeding
- [ ] T083 [US3] Implement file-only scan mode in pkg/scanner/scanner.go using database folder records as input
- [ ] T084 [US3] Add file-only checkpoint handling in pkg/checkpoint/checkpoint.go with scan_mode='files'
- [ ] T085 [US3] Implement file-only summary output in cmd/scanFiles.go showing files scanned, duplicates found
- [ ] T086 [US3] Implement hash algorithm migration detection in pkg/scanner/scanner.go with rehash and logging (FR-037)

**Checkpoint**: User Story 3 complete - file-only scanning works independently

---

## Phase 6: User Story 4 - Identify Partial Folder Duplicates (Priority: P3)

**Goal**: Find folders that are mostly similar but not identical with minimum 50% similarity threshold

**Independent Test**: Create folder pairs with overlapping but not identical file sets, run scan, query for partial matches, verify overlap percentages correct

### Implementation for User Story 4

- [ ] T090 [P] [US4] Implement folder file set query in pkg/duplicate/detector.go to get all files for a folder
- [ ] T091 [P] [US4] Implement set comparison logic in pkg/duplicate/partial.go calculating overlap percentage
- [ ] T092 [US4] Implement difference detection in pkg/duplicate/partial.go identifying missing files from each side
- [ ] T093 [US4] Implement same-name-different-date detection in pkg/duplicate/partial.go comparing file names and mtimes
- [ ] T094 [US4] Create partial duplicate query in pkg/datastore/queries.go with similarity threshold parameter (default 50%)
- [ ] T095 [US4] Implement get partial duplicates CLI command in cmd/getDuplicates.go with --partial flag and --min-similarity parameter

**Checkpoint**: User Story 4 complete - partial folder duplicate detection works independently

---

## Phase 7: User Story 5 - Query and View Duplicate Files (Priority: P1)

**Goal**: View duplicate files identified during scans with configurable output formats

**Independent Test**: After scans performed, query duplicates with various filters and formats, verify correct grouping and output

### Implementation for User Story 5

- [X] T100 [P] [US5] Implement duplicate file query in pkg/datastore/queries.go using idx_files_hash index per data-model.md
- [X] T101 [P] [US5] Implement duplicate file detector in pkg/duplicate/detector.go grouping files by size+hash
- [X] T102 [P] [US5] Implement table format renderer in pkg/duplicate/formatter.go with duplicate sets per cli-commands.md
- [X] T103 [P] [US5] Implement JSON format renderer in pkg/duplicate/formatter.go with structured output per cli-commands.md
- [X] T104 [US5] Create get duplicates CLI command in cmd/getDuplicates.go with --json and --min-count flags per cli-commands.md
- [X] T105 [US5] Add min-count filtering in pkg/duplicate/detector.go to filter by duplicate set size
- [ ] T106 [US5] Implement duplicate folder query in pkg/datastore/queries.go finding folders with identical subtrees
- [ ] T107 [US5] Add duplicate folder detection in pkg/duplicate/detector.go comparing complete folder hash sets

**Checkpoint**: User Story 5 complete - duplicate querying with multiple formats works independently

---

## Phase 8: Root Folder Management (Supporting Functionality)

**Goal**: Provide root folder CRUD operations to support scan workflows

**Independent Test**: Register, list, and delete root folders, verify database state correct

- [ ] T110 [P] Implement root folder registration in pkg/datastore/rootfolders.go (INSERT with absolute path conversion)
- [ ] T111 [P] Implement root folder listing query in pkg/datastore/rootfolders.go with statistics per FR-001.6
- [ ] T112 [P] Implement root folder deletion with CASCADE in pkg/datastore/rootfolders.go removing all scan data
- [ ] T113 [P] Create add root CLI command in cmd/addRoot.go with path and --traverse-links flag per cli-commands.md
- [ ] T114 [P] Create get root CLI command in cmd/getRoot.go displaying table with path, counts, size, last scan date per cli-commands.md
- [ ] T115 [P] Create delete root CLI command in cmd/deleteRoot.go with confirmation prompt per cli-commands.md
- [ ] T116 [P] Create purge CLI command in cmd/purgeRoot.go with type arg [files|folders|all] and --before flag per cli-commands.md
- [ ] T117 [P] Create refresh CLI command in cmd/refreshRoot.go to recalculate statistics per cli-commands.md
- [ ] T118 [P] Create verify CLI command in cmd/verifyRoot.go with consistency checks and --repair flag per cli-commands.md

**Checkpoint**: Root folder management complete

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

### Constitution Compliance Validation

- [ ] T120 Verify test coverage ≥70% across all new code, 80% for core logic (Testing)
- [ ] T121 Verify all functions <50 lines, complexity ≤10 using gocyclo (Clean Code)
- [ ] T122 Run golangci-lint with zero warnings (Code Quality)
- [ ] T123 Verify all public APIs have GoDoc comments (Maintainability)
- [ ] T124 Test on Windows, Linux, macOS platforms for path handling (Portability)
- [ ] T125 Verify graceful shutdown <5 seconds on SIGINT/SIGTERM in all scan commands (Graceful Shutdown)
- [ ] T126 Load test with 100K files verifying <500 MB RAM usage (Performance - NFR-002)
- [ ] T127 Verify hash throughput ≥50 MB/sec on standard hardware (Performance - NFR-001)
- [ ] T128 Verify scan speed 10K files <5 minutes (Performance - SC-001)
- [ ] T129 Security audit: SQL injection prevention with parameterized queries, path traversal validation (Security)
- [ ] T130 Verify database file permissions 0600 in production (Security)
- [ ] T131 Test database migration rollback scenario (Upgradability)
- [ ] T132 Verify all dependencies justified: cobra, viper, sqlite, sha3 (Minimal Dependencies)

### Test Suite Implementation (OPTIONAL - if requested)

- [ ] T140 [P] Create unit tests for hash algorithms in tests/unit/hash_test.go verifying SHA-256, SHA-512, SHA3-256
- [ ] T141 [P] Create unit tests for folder traversal in tests/unit/traverser_test.go with permission error cases
- [ ] T142 [P] Create unit tests for duplicate detection in tests/unit/detector_test.go with known duplicate sets
- [ ] T143 [P] Create unit tests for checkpoint save/restore in tests/unit/checkpoint_test.go
- [ ] T144 [P] Create integration test for complete scan workflow in tests/integration/scan_test.go
- [ ] T145 [P] Create integration test for checkpoint resume in tests/integration/checkpoint_test.go with simulated interruption
- [ ] T146 [P] Create integration test for removed flag cascading in tests/integration/removed_test.go
- [ ] T147 [P] Create database operations test in tests/integration/database_test.go validating FK integrity and indexes
- [ ] T148 [P] Create e2e CLI test in tests/e2e/cli_test.go running full command workflows
- [ ] T149 [P] Create e2e container test in tests/e2e/container_test.go validating checkpoint persistence across restarts
- [ ] T150 [P] Create worker pool concurrency test in tests/integration/worker_test.go validating thread safety

### Documentation & Code Quality

- [ ] T160 [P] Update README.md with scan command examples and quickstart guide
- [ ] T161 [P] Create user guide documentation in docs/scanning-guide.md
- [ ] T162 [P] Document configuration options in docs/configuration.md (hash algorithm, worker count, progress interval)
- [ ] T163 [P] Add architecture documentation in docs/design/scanning-architecture.md
- [ ] T164 [P] Code cleanup and refactoring for cyclomatic complexity <10
- [ ] T165 [P] Add inline code comments and GoDoc for all exported functions
- [ ] T166 Run quickstart.md validation ensuring all setup steps work

### Performance Optimization

- [ ] T170 [P] Profile hash performance and optimize buffer sizes in pkg/hash/
- [ ] T171 [P] Optimize database batch insert sizes in pkg/datastore/files.go
- [ ] T172 [P] Tune worker pool sizes for different hardware profiles
- [ ] T173 [P] Optimize duplicate detection queries with EXPLAIN QUERY PLAN analysis

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational phase completion
  - User Story 1 (P1): Can start after Foundational - No dependencies on other stories
  - User Story 2 (P2): Can start after Foundational - No dependencies (standalone)
  - User Story 3 (P2): Can start after Foundational - No dependencies (standalone)
  - User Story 4 (P3): Can start after Foundational - No dependencies (standalone)
  - User Story 5 (P1): Can start after Foundational - No dependencies (standalone)
- **Root Management (Phase 8)**: Can run in parallel with user stories
- **Polish (Phase 9)**: Depends on all desired user stories being complete

### User Story Independence

All user stories are independently implementable and testable:

- **US1 (Scan All)**: Standalone full scanning capability
- **US2 (Scan Folders)**: Standalone folder registration capability
- **US3 (Scan Files)**: Standalone file hashing capability (requires folders exist but independent implementation)
- **US4 (Partial Duplicates)**: Standalone partial matching capability
- **US5 (Query Duplicates)**: Standalone query capability

### Within Each User Story

- Constitution infrastructure from Phase 2 must be complete
- Database schema from Phase 2 must be created
- Entity models from Phase 2 must exist
- Core services (hash, worker pool, checkpoint) from Phase 2 must be implemented
- Tasks within a story follow: models → services → orchestration → CLI command

### Parallel Opportunities

**Phase 1 (Setup)**: Tasks T003-T006 can run in parallel

**Phase 2 (Foundational)**: 
- Constitution tasks T010-T016 can run in parallel
- Database schema tasks T020-T023 can run in parallel
- Entity models T030-T032 can run in parallel
- Core services T040-T044 (hash implementations) can run in parallel
- Checkpoint T047 can run in parallel with hash and worker pool

**Phase 3-7 (User Stories)**: Once Foundational complete, ALL user stories can be implemented in parallel by different developers

**Within User Story 1**:
- T050-T054 (traversal, work items, database operations) can run in parallel
- T060-T063 (CLI implementation tasks) can run in parallel

**Within User Story 4**:
- T090-T093 (detection logic components) can run in parallel

**Within User Story 5**:
- T100-T103 (query, detection, formatting) can run in parallel

**Phase 8 (Root Management)**: All tasks T110-T115 can run in parallel

**Phase 9 (Polish)**: 
- All test tasks T140-T150 can run in parallel
- All documentation tasks T160-T166 can run in parallel
- All performance tasks T170-T173 can run in parallel

---

## Parallel Execution Example: User Story 1

```bash
# After Foundational phase complete, launch User Story 1 components in parallel:

# Terminal 1: Folder traversal
Task T050: Implement folder traversal logic in pkg/scanner/traverser.go
Task T051: Create FolderTraversalWorkItem in internal/worker/traversal.go
Task T052: Implement folder registration in pkg/datastore/folders.go

# Terminal 2: File hashing  
Task T053: Implement file registration in pkg/datastore/files.go
Task T054: Create FileHashingWorkItem in internal/worker/hashing.go

# Terminal 3: CLI and display
Task T060: Implement progress display in pkg/scanner/progress.go
Task T061: Create scan all CLI command in cmd/scanAll.go

# Then integrate with sequential tasks T055-T059, T062-T066
```

---

## Implementation Strategy

### MVP First (User Story 1 + User Story 5 Only)

This delivers immediate value: users can scan and see duplicates

1. **Complete Phase 1: Setup** (T001-T006)
2. **Complete Phase 2: Foundational** (T010-T047) ⚠️ CRITICAL - blocks all stories
3. **Complete Phase 3: User Story 1 - Scan All** (T050-T066) → Full scan capability
4. **Complete Phase 7: User Story 5 - Query Duplicates** (T100-T107) → View results
5. **Complete Phase 8: Root Management** (T110-T115) → Basic CRUD
6. **STOP and VALIDATE**: Test MVP with real data
7. Deploy/demo MVP if ready

**MVP Task Count**: ~55 tasks
**MVP Deliverable**: Users can register roots, scan all files/folders, query duplicates in table/JSON format

### Incremental Delivery (All User Stories)

1. **Foundation** (Phases 1-2) → ~30 tasks → Infrastructure ready
2. **MVP** (US1 + US5 + Root Mgmt) → +55 tasks → Core value delivery ✅ Ship v1.0
3. **Quick Scan** (US2) → +4 tasks → Efficient folder mapping ✅ Ship v1.1
4. **Incremental Scan** (US3) → +6 tasks → Flexible workflows ✅ Ship v1.2
5. **Advanced Detection** (US4) → +6 tasks → Partial duplicates ✅ Ship v1.3
6. **Polish** (Phase 9) → +50 tasks → Production-ready ✅ Ship v2.0

**Total Task Count**: ~151 tasks (101 implementation + 50 polish/tests)

### Parallel Team Strategy

With 3 developers after Foundational phase:

- **Developer A**: User Story 1 (Scan All) - 17 tasks
- **Developer B**: User Story 5 (Query Duplicates) - 8 tasks  
- **Developer C**: User Story 2 + User Story 3 (Folder/File Scans) - 10 tasks

Then merge and test integration, followed by:

- **Developer A**: User Story 4 (Partial Duplicates) - 6 tasks
- **Developer B**: Root Management - 6 tasks
- **Developer C**: Polish & Testing - 50 tasks

**Timeline Estimate**: 
- Foundation: 1 week
- MVP (parallel): 2 weeks
- Additional stories (parallel): 1 week
- Polish: 1 week
- **Total**: ~5 weeks with 3 developers

---

## Task Statistics

**Total Tasks**: 151

**By Phase**:
- Phase 1 (Setup): 6 tasks
- Phase 2 (Foundational): 24 tasks (17 infrastructure + 7 database + 3 entities)
- Phase 3 (US1 - Scan All): 17 tasks
- Phase 4 (US2 - Scan Folders): 4 tasks
- Phase 5 (US3 - Scan Files): 6 tasks
- Phase 6 (US4 - Partial Duplicates): 6 tasks
- Phase 7 (US5 - Query Duplicates): 8 tasks
- Phase 8 (Root Management): 6 tasks
- Phase 9 (Polish): 24 constitution checks + 11 tests + 7 docs + 4 performance = 46 tasks

**By User Story**:
- US1 (Scan All): 17 tasks
- US2 (Scan Folders): 4 tasks
- US3 (Scan Files): 6 tasks
- US4 (Partial Duplicates): 6 tasks
- US5 (Query Duplicates): 8 tasks
- Infrastructure: 30 tasks
- Root Management: 6 tasks
- Polish/Tests: 46 tasks

**Parallel Opportunities**: 
- Phase 1: 4 tasks can run in parallel
- Phase 2: 20+ tasks can run in parallel across different modules
- Phases 3-8: All 5 user stories can develop in parallel (47 tasks)
- Phase 9: 46 polish tasks can largely run in parallel

**MVP Scope** (Recommended first delivery):
- Phase 1: Setup (6 tasks)
- Phase 2: Foundational (24 tasks)
- Phase 3: US1 - Scan All (17 tasks)
- Phase 7: US5 - Query Duplicates (8 tasks)
- Phase 8: Root Management (6 tasks)
- **Total: 61 tasks for MVP** → delivers complete scan and query workflow

**Critical Path**: Setup → Foundational (BLOCKS everything) → User Stories (parallel) → Polish

---

## Notes

- **[P] marker** indicates tasks that can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story] label** maps task to specific user story for traceability and independent testing
- **Tests are OPTIONAL**: Only include test tasks (T140-T150) if explicitly requested or TDD approach required
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- **Critical**: Phase 2 (Foundational) MUST complete before any user story work begins
- **MVP Focus**: Implement US1 (Scan All) + US5 (Query Duplicates) first for maximum user value
- All file paths use Go conventions: `pkg/` for application logic, `internal/` for private code, `cmd/` for CLI commands
