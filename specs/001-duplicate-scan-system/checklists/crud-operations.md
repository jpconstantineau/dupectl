# Checklist: Entity CRUD Operations & State Management

**Purpose**: Validate all semi-managed entities have complete lifecycle operations to prevent unmanageable data states  
**Created**: December 24, 2025  
**Domain**: Entity Lifecycle Management (root_folders, scan_state, files, folders)  
**Scope**: CLI commands + cascading operations  
**Coverage**: Orphaned records, corrupted metadata, lifecycle gaps  
**Audience**: DBA/Architect + QA/Test Team

## Root Folder Entity - Create Operations

- [ ] CHK001 - Is a CREATE operation defined for root_folders (add root command documented)? [Completeness, Spec §FR-001.2, FR-001.5]
- [ ] CHK002 - Does CREATE include all required attributes (path mandatory, traverse_links optional)? [Completeness, Spec §FR-001.5]
- [ ] CHK003 - Is path validation documented (relative→absolute conversion, duplicate detection)? [Completeness, Spec §FR-001.3]
- [ ] CHK004 - Is CREATE behavior specified when root already exists (prompt, restore removed flag, error)? [Clarity, Edge Case, Spec §Open Questions]
- [ ] CHK005 - Is user confirmation documented for auto-registration during scan? [Completeness, Spec §FR-001.2]
- [ ] CHK006 - Are default values documented for optional fields (traverse_links=false)? [Completeness, Spec §A-020]
- [ ] CHK007 - Is the relationship to existing agent/host/owner entities documented on CREATE? [Completeness, Spec §FR-004.1]

## Root Folder Entity - Read Operations

- [ ] CHK008 - Is a READ operation defined for root_folders (get root command documented)? [Completeness, Spec §FR-001.6]
- [ ] CHK009 - Does READ return all relevant attributes (path, folder_count, file_count, total_size, last_scan_date)? [Completeness, Spec §FR-001.6]
- [ ] CHK010 - Is display format for never-scanned roots specified ("N/A" or "Never scanned")? [Clarity, Spec §FR-001.7]
- [ ] CHK011 - Is READ behavior for removed roots specified (show with flag, hide, separate query)? [Gap, Edge Case]
- [ ] CHK012 - Is single-root READ operation available (query specific root by path or ID)? [Gap]
- [ ] CHK013 - Are READ results sortable (by path, size, last_scan_date)? [Gap, Usability]

## Root Folder Entity - Update Operations

- [ ] CHK014 - Is an UPDATE operation defined for root_folders configuration (modify traverse_links)? [Gap]
- [ ] CHK015 - Is statistics UPDATE triggered automatically after scan completion? [Completeness, Spec §FR-001.7, §root_folders Extensions]
- [ ] CHK016 - Is statistics UPDATE isolated from user-triggered scan operations (no manual refresh needed)? [Clarity, Spec §A-021]
- [ ] CHK017 - Is UPDATE behavior specified for changing root path (allowed, rejected, requires re-scan)? [Gap, Edge Case]
- [ ] CHK018 - Is restore operation documented for removed roots (unflag removed, preserve scan data)? [Gap, Spec §Open Questions]

## Root Folder Entity - Delete Operations

- [ ] CHK019 - Is a DELETE operation defined for root_folders (delete root command documented)? [Completeness, Spec §FR-022, SC-010]
- [ ] CHK020 - Is CASCADE delete behavior documented (removes all folders, files, scan_state)? [Completeness, Spec §Foreign Key Cascade Rules]
- [ ] CHK021 - Is user confirmation required before destructive DELETE? [Gap, Safety]
- [ ] CHK022 - Is DELETE behavior specified when scan is in progress (reject, stop scan first)? [Clarity, Spec §Open Questions]
- [ ] CHK023 - Is soft delete (removed flag) vs hard delete (physical removal) distinguished? [Clarity, Spec §FR-021.2]
- [ ] CHK024 - Can removed roots be permanently purged (hard delete after soft delete)? [Gap, Lifecycle]

## Scan State Entity - Create Operations

- [ ] CHK025 - Is CREATE operation for scan_state automatic during scan initiation? [Completeness, Spec §FR-014.1]
- [ ] CHK026 - Does CREATE include all checkpoint attributes (root_folder_id, scan_mode, current_folder_path, last_processed_file)? [Completeness, Spec §scan_state table]
- [ ] CHK027 - Is unique constraint enforced (one active scan per root_folder_id)? [Completeness, Spec §idx_scan_state_active]
- [ ] CHK028 - Is CREATE behavior specified when checkpoint already exists (upsert/update)? [Clarity, Spec §Checkpoint Save query]
- [ ] CHK029 - Are timestamps initialized correctly on CREATE (started_at, updated_at)? [Completeness, Spec §scan_state table]

## Scan State Entity - Read Operations

- [ ] CHK030 - Is READ operation for scan_state automatic on application startup? [Completeness, Spec §FR-014.2]
- [ ] CHK031 - Does READ detect incomplete scans (completed=0) for resume? [Completeness, Spec §FR-014.2, §Checkpoint Resume query]
- [ ] CHK032 - Is READ behavior specified for multiple incomplete scans (use most recent updated_at)? [Clarity, Spec §Checkpoint Resume query]
- [ ] CHK033 - Is user-facing READ operation available (show current scan status)? [Gap, Observability]
- [ ] CHK034 - Is stale checkpoint detection documented (threshold for updated_at age)? [Clarity, Spec §scan_state Design Rationale]

## Scan State Entity - Update Operations

- [ ] CHK035 - Is UPDATE operation for scan_state periodic during scan execution? [Completeness, Spec §FR-014.1, A-023]
- [ ] CHK036 - Is UPDATE frequency documented (after each folder, time-based, on shutdown)? [Clarity, Spec §A-023]
- [ ] CHK037 - Are updated fields documented (current_folder_path, last_processed_file, updated_at)? [Completeness, Spec §Checkpoint Save query]
- [ ] CHK038 - Is UPDATE atomic (transaction-safe for concurrent read during shutdown)? [Gap, Integrity]
- [ ] CHK039 - Is UPDATE behavior on scan completion specified (set completed=1)? [Gap, Spec §scan_state.completed]

## Scan State Entity - Delete Operations

- [ ] CHK040 - Is DELETE operation for scan_state triggered by --restart flag? [Completeness, Spec §FR-014.3]
- [ ] CHK041 - Is DELETE behavior documented for completed scans (cleanup strategy, retention policy)? [Gap, Spec §scan_state Design Rationale]
- [ ] CHK042 - Is CASCADE delete automatic when root_folder is deleted? [Completeness, Spec §Foreign Key Cascade Rules]
- [ ] CHK043 - Can users manually delete stale checkpoints (abandoned scans)? [Gap, Lifecycle]
- [ ] CHK044 - Is DELETE behavior specified during active scan (reject operation, force stop)? [Gap, Edge Case]

## Files Entity - Implicit Create Operations

- [ ] CHK045 - Is CREATE operation for files automatic during scan files or scan all? [Completeness, Spec §FR-004, FR-011]
- [ ] CHK046 - Are all file attributes populated on CREATE (path, size, mtime, hash_value, etc.)? [Completeness, Spec §files table]
- [ ] CHK047 - Is CREATE behavior specified when file already exists (update vs skip vs error)? [Gap, Spec §FR-020]
- [ ] CHK048 - Are timestamps set correctly (first_scanned_at on initial CREATE)? [Completeness, Spec §FR-019]
- [ ] CHK049 - Is error_status field populated on CREATE failure (permission denied)? [Completeness, Spec §FR-004.2, NFR-006]
- [ ] CHK050 - Are foreign keys validated on CREATE (folder_id, root_folder_id exist)? [Integrity, Spec §files Foreign Keys]

## Files Entity - Read Operations

- [ ] CHK051 - Is READ operation for duplicate files defined (get duplicates command)? [Completeness, Spec §FR-017]
- [ ] CHK052 - Does READ filter removed files (removed=0 in queries)? [Completeness, Spec §Duplicate File Detection query]
- [ ] CHK053 - Does READ filter error files (error_status IS NULL in queries)? [Completeness, Spec §Duplicate File Detection query]
- [ ] CHK054 - Is READ operation available for individual file by path? [Gap, Troubleshooting]
- [ ] CHK055 - Is READ operation available for all files in folder (hierarchy query)? [Gap, Spec §Duplicate Folder Detection]
- [ ] CHK056 - Can users query files by hash value (reverse lookup)? [Gap, Usability]
- [ ] CHK057 - Can users query removed files separately (audit deleted files)? [Gap, Spec §FR-021.1]

## Files Entity - Update Operations

- [ ] CHK058 - Is UPDATE operation for files automatic on rescan (hash recalculation)? [Completeness, Spec §FR-020]
- [ ] CHK059 - Is last_scanned_at updated on every scan encounter? [Completeness, Spec §FR-020]
- [ ] CHK060 - Is hash_value UPDATE behavior specified when hash_algorithm changes? [Clarity, Spec §FR-016]
- [ ] CHK061 - Is removed flag UPDATE triggered when file missing on filesystem? [Completeness, Spec §FR-021]
- [ ] CHK062 - Is removed flag reversible (file reappears at same path)? [Gap, Spec §FR-021.1]
- [ ] CHK063 - Can users manually UPDATE error_status (clear permission errors)? [Gap, Troubleshooting]
- [ ] CHK064 - Is UPDATE behavior specified for mtime changes (trigger rehash or skip)? [Gap, Optimization]

## Files Entity - Delete Operations

- [ ] CHK065 - Is CASCADE delete automatic when parent folder deleted? [Completeness, Spec §Foreign Key Cascade Rules]
- [ ] CHK066 - Is CASCADE delete automatic when root_folder deleted? [Completeness, Spec §Foreign Key Cascade Rules]
- [ ] CHK067 - Is soft delete (removed=1) vs hard delete (physical removal) distinguished? [Clarity, Spec §FR-021.1]
- [ ] CHK068 - Can users permanently purge removed files (cleanup old history)? [Gap, Lifecycle]
- [ ] CHK069 - Is bulk delete operation available (purge all files in root, all removed files)? [Gap, Maintenance]
- [ ] CHK070 - Is DELETE behavior specified for files with error_status (keep for audit or purge)? [Gap, Lifecycle]

## Folders Entity - Implicit Create Operations

- [ ] CHK071 - Is CREATE operation for folders automatic during scan folders or scan all? [Completeness, Spec §FR-010]
- [ ] CHK072 - Are all folder attributes populated on CREATE (path, parent_folder_id, root_folder_id)? [Completeness, Spec §folders table]
- [ ] CHK073 - Is parent-child hierarchy established correctly on CREATE (parent_folder_id FK)? [Completeness, Spec §folders table]
- [ ] CHK074 - Is CREATE behavior specified when folder already exists (update vs skip)? [Gap, Spec §FR-010]
- [ ] CHK075 - Is error_status field populated on CREATE failure (permission denied)? [Completeness, Spec §FR-004.3]
- [ ] CHK076 - Are self-referential root folders handled (parent_folder_id=NULL)? [Completeness, Spec §folders.parent_folder_id]

## Folders Entity - Read Operations

- [ ] CHK077 - Is READ operation for duplicate folders defined (exact match query)? [Completeness, Spec §FR-006, FR-018]
- [ ] CHK078 - Is READ operation for partial duplicate folders defined (similarity query)? [Completeness, Spec §FR-007, FR-018]
- [ ] CHK079 - Does READ filter removed folders (removed=0 in queries)? [Gap, Spec §folders.removed]
- [ ] CHK080 - Is READ operation available for folder hierarchy (parent→children traversal)? [Gap, Spec §Duplicate Folder Detection]
- [ ] CHK081 - Can users query folders by path pattern (LIKE queries)? [Gap, Usability]
- [ ] CHK082 - Can users query removed folders separately (audit deleted folders)? [Gap, Spec §FR-021.2]

## Folders Entity - Update Operations

- [ ] CHK083 - Is UPDATE operation for folders automatic on rescan (timestamp update)? [Completeness, Spec §FR-020 analogy]
- [ ] CHK084 - Is last_scanned_at updated on every scan encounter? [Gap, Consistency with files]
- [ ] CHK085 - Is removed flag UPDATE triggered when folder missing on filesystem? [Completeness, Spec §FR-021.2]
- [ ] CHK086 - Is cascade removed flag to children automatic (all subfolders, all files)? [Completeness, Spec §FR-021.3, A-022]
- [ ] CHK087 - Is removed flag reversible (folder reappears at same path)? [Gap, Edge Case]
- [ ] CHK088 - Can users manually UPDATE error_status (clear permission errors)? [Gap, Troubleshooting]

## Folders Entity - Delete Operations

- [ ] CHK089 - Is CASCADE delete to child folders automatic (recursive delete)? [Completeness, Spec §Foreign Key Cascade Rules]
- [ ] CHK090 - Is CASCADE delete to contained files automatic? [Completeness, Spec §Foreign Key Cascade Rules]
- [ ] CHK091 - Is CASCADE delete automatic when root_folder deleted? [Completeness, Spec §Foreign Key Cascade Rules]
- [ ] CHK092 - Is soft delete (removed=1) vs hard delete (physical removal) distinguished? [Clarity, Spec §FR-021.2]
- [ ] CHK093 - Can users permanently purge removed folders (cleanup old history)? [Gap, Lifecycle]
- [ ] CHK094 - Is DELETE behavior specified for folders with error_status (keep or purge)? [Gap, Lifecycle]

## Orphaned Records Prevention

- [ ] CHK095 - Can files exist without parent folder_id in database? [Integrity, Spec §files.folder_id FK]
- [ ] CHK096 - Can files exist without root_folder_id in database? [Integrity, Spec §files.root_folder_id FK]
- [ ] CHK097 - Can folders exist without root_folder_id in database? [Integrity, Spec §folders.root_folder_id FK]
- [ ] CHK098 - Can scan_state exist without root_folder_id in database? [Integrity, Spec §scan_state.root_folder_id FK]
- [ ] CHK099 - Are orphaned records prevented by foreign key constraints (NOT NULL + FK)? [Completeness, Spec §Foreign Keys]
- [ ] CHK100 - Are orphaned records automatically deleted on parent removal (CASCADE)? [Completeness, Spec §Foreign Key Cascade Rules]
- [ ] CHK101 - Is orphan detection available via query (files with invalid folder_id)? [Gap, Troubleshooting]

## Corrupted Metadata Handling

- [ ] CHK102 - Can files with NULL hash_value be identified and reprocessed? [Completeness, Spec §files.hash_value, §Design Rationale]
- [ ] CHK103 - Can files with error_status be retried (clear error, trigger rescan)? [Gap, Spec §FR-004.2]
- [ ] CHK104 - Can folders with error_status be retried (clear error, trigger rescan)? [Gap, Spec §FR-004.3]
- [ ] CHK105 - Is stale checkpoint detection automated (updated_at threshold check)? [Clarity, Spec §scan_state Design Rationale]
- [ ] CHK106 - Can users manually clear stale checkpoints (force delete scan_state)? [Gap, Lifecycle]
- [ ] CHK107 - Is hash algorithm migration strategy documented (rehash with new algorithm)? [Gap, Spec §FR-016]
- [ ] CHK108 - Can users query inconsistent states (first_scanned_at > last_scanned_at)? [Gap, Data Quality]
- [ ] CHK109 - Can users query files with mismatched root_folder_id between file and folder? [Gap, Data Quality]

## Lifecycle Gaps - Removed Flag Management

- [ ] CHK110 - Is transition documented: present (removed=0) → removed (removed=1)? [Completeness, Spec §FR-021]
- [ ] CHK111 - Is reverse transition documented: removed (removed=1) → present (removed=0)? [Gap, Spec §FR-021.1]
- [ ] CHK112 - Can users permanently delete removed files (hard delete after soft delete)? [Gap, Lifecycle]
- [ ] CHK113 - Can users permanently delete removed folders (hard delete after soft delete)? [Gap, Lifecycle]
- [ ] CHK114 - Is bulk purge available for all removed entities in root? [Gap, Maintenance]
- [ ] CHK115 - Is retention policy documented for removed entities (how long to keep)? [Gap, Policy]
- [ ] CHK116 - Are removed entities excluded from duplicate detection queries? [Completeness, Spec §Duplicate File Detection query]
- [ ] CHK117 - Can users query removed entity history (audit trail)? [Gap, Spec §FR-021.1]

## Lifecycle Gaps - Scan State Management

- [ ] CHK118 - Is transition documented: not started → in progress (completed=0)? [Completeness, Spec §scan_state.completed]
- [ ] CHK119 - Is transition documented: in progress → completed (completed=1)? [Gap, Spec §scan_state.completed]
- [ ] CHK120 - Is transition documented: completed → purged (delete record)? [Gap, Cleanup]
- [ ] CHK121 - Is transition documented: in progress → abandoned (stale detection)? [Gap, Spec §scan_state Design Rationale]
- [ ] CHK122 - Can users restart abandoned scans (clear stale checkpoint)? [Gap, Spec §FR-014.3]
- [ ] CHK123 - Can users view scan history (list of completed scans)? [Gap, Observability]
- [ ] CHK124 - Is automatic cleanup of completed scans documented (prevent table bloat)? [Gap, Maintenance]

## Lifecycle Gaps - Error Recovery

- [ ] CHK125 - Can users retry files with permission errors after fixing permissions? [Gap, Spec §FR-004.2]
- [ ] CHK126 - Can users retry folders with permission errors after fixing permissions? [Gap, Spec §FR-004.3]
- [ ] CHK127 - Is bulk retry operation available (all error_status entries in root)? [Gap, Efficiency]
- [ ] CHK128 - Can users manually clear error_status without rescan (false positive errors)? [Gap, Troubleshooting]
- [ ] CHK129 - Is error escalation documented (repeated errors → permanent skip)? [Gap, Spec §FR-004.2]

## Cross-Entity Consistency

- [ ] CHK130 - Are root statistics (folder_count, file_count) recalculated after deletes? [Gap, Spec §root_folders Extensions]
- [ ] CHK131 - Are root statistics recalculated after removed flag cascades? [Gap, Data Integrity]
- [ ] CHK132 - Is scan_state cleared when corresponding files/folders are deleted? [Completeness, Spec §Foreign Key Cascade Rules]
- [ ] CHK133 - Can removed flag propagation be verified (folder → subfolders → files)? [Gap, Testability]
- [ ] CHK134 - Is consistency check available (verify all FKs resolve, no orphans)? [Gap, Maintenance]

## Unmanageable States Prevention

- [ ] CHK135 - Can system reach state: root_folder deleted but files remain? [Integrity, Should be impossible via CASCADE]
- [ ] CHK136 - Can system reach state: folder deleted but files remain? [Integrity, Should be impossible via CASCADE]
- [ ] CHK137 - Can system reach state: active scan_state but root_folder deleted? [Integrity, Should be impossible via CASCADE]
- [ ] CHK138 - Can system reach state: files with NULL hash_value and no way to rehash? [Recovery, CHK102 must enable]
- [ ] CHK139 - Can system reach state: removed entities with no purge operation? [Recovery, CHK112-113 must enable]
- [ ] CHK140 - Can system reach state: stale checkpoint blocking new scans? [Recovery, CHK106 or CHK040 must enable]
- [ ] CHK141 - Can system reach state: error_status preventing rescan with no clear operation? [Recovery, CHK103-104 must enable]
- [ ] CHK142 - Can system reach state: inconsistent statistics with no recalc operation? [Recovery, CHK130-131 must enable]

## Command Coverage Completeness

- [ ] CHK143 - Is `add root` command fully specified (CREATE root_folder)? [Completeness, Spec §FR-001.5]
- [ ] CHK144 - Is `get root` command fully specified (READ root_folders)? [Completeness, Spec §FR-001.6]
- [ ] CHK145 - Is `delete root` command fully specified (DELETE root_folder + CASCADE)? [Completeness, Spec §FR-022]
- [ ] CHK146 - Is `scan all` command fully specified (CREATE/UPDATE files + folders)? [Completeness, Spec §FR-001]
- [ ] CHK147 - Is `scan folders` command fully specified (CREATE/UPDATE folders only)? [Completeness, Spec §FR-010]
- [ ] CHK148 - Is `scan files` command fully specified (CREATE/UPDATE files only)? [Completeness, Spec §FR-011]
- [ ] CHK149 - Is `get duplicates` command fully specified (READ files with filters)? [Completeness, Spec §FR-017]
- [ ] CHK150 - Is `get duplicate-folders` command fully specified (READ folders with similarity)? [Completeness, Spec §FR-018]
- [ ] CHK151 - Is `update root` command specified (UPDATE root_folder configuration)? [Gap, CHK014]
- [ ] CHK152 - Is `purge removed` command specified (DELETE removed entities)? [Gap, CHK112-113]
- [ ] CHK153 - Is `clear errors` command specified (UPDATE error_status to NULL)? [Gap, CHK103-104]
- [ ] CHK154 - Is `reset checkpoint` command specified (DELETE scan_state)? [Completeness, Spec §FR-014.3 --restart flag]

## Testability Requirements

- [ ] CHK155 - Can all CREATE operations be tested with valid inputs? [Measurability, Spec §Schema Tests]
- [ ] CHK156 - Can all READ operations be tested with known data? [Measurability, Spec §Schema Tests]
- [ ] CHK157 - Can all UPDATE operations be tested with before/after states? [Measurability, Spec §Schema Tests]
- [ ] CHK158 - Can all DELETE operations be tested with CASCADE verification? [Measurability, Spec §Schema Tests]
- [ ] CHK159 - Can orphan prevention be tested (attempt FK violation)? [Measurability, Spec §Schema Tests]
- [ ] CHK160 - Can corrupted metadata scenarios be created and recovered in tests? [Measurability, Gap]
- [ ] CHK161 - Can lifecycle transitions be tested systematically? [Measurability, Gap]
- [ ] CHK162 - Can all unmanageable states be prevented in tests (CHK135-142)? [Measurability, Critical]

---

**Summary**: 162 requirement quality checks for Entity CRUD Operations & State Management covering root_folders and scan_state lifecycle completeness. Focus on preventing unmanageable data states through comprehensive operation coverage.

**Notes**:
- Items marked [Gap] indicate missing operations that leave data in unmanageable states
- Items marked [Completeness] reference existing documented operations
- Items marked [Integrity] validate foreign key enforcement prevents orphans
- Items marked [Recovery] validate ability to fix corrupted/stale states
- CHK135-142 are critical: system MUST provide operations to recover from these states

**Coverage Summary**:
- Root Folder CRUD: CHK001-024 (24 items)
- Scan State CRUD: CHK025-044 (20 items)
- Files Entity CRUD: CHK045-070 (26 items)
- Folders Entity CRUD: CHK071-094 (24 items)
- Orphan Prevention: CHK095-101 (7 items)
- Corrupted Metadata: CHK102-109 (8 items)
- Lifecycle Gaps: CHK110-129 (20 items)
- Cross-Entity Consistency: CHK130-134 (5 items)
- Unmanageable States: CHK135-142 (8 items - CRITICAL)
- Command Coverage: CHK143-154 (12 items)
- Testability: CHK155-162 (8 items)

**Critical Findings** (operations missing to prevent unmanageable states):
- No documented UPDATE operation for root_folder configuration (CHK014)
- No documented purge operation for removed entities (CHK112-113)
- No documented clear operation for error_status (CHK103-104)
- No documented manual checkpoint clear (CHK106)
- No documented statistics recalculation (CHK130-131)
- No documented consistency check/repair operations (CHK134)
