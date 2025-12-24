# Checklist: Database Schema & Data Model Requirements Quality

**Purpose**: Validate database schema requirements are consistent, complete, and follow best practices  
**Created**: December 24, 2025  
**Domain**: Data Model & Database Schema (SQLite tables, indexes, constraints, migrations)  
**Depth**: Standard level (thorough validation for design review)  
**Audience**: DBA/Architect (best practices) + QA/Test Team (testability)

## Requirement Completeness

- [ ] CHK001 - Are all entity tables documented with complete column specifications (name, type, constraints, description)? [Completeness, Spec §data-model.md]
- [ ] CHK002 - Are all foreign key relationships explicitly documented with source/target tables and columns? [Completeness, Spec §Foreign Keys]
- [ ] CHK003 - Are all indexes documented with columns, type (unique/partial), and performance rationale? [Completeness, Spec §Indexes]
- [ ] CHK004 - Are cascade rules (ON DELETE/ON UPDATE) specified for every foreign key relationship? [Completeness, Spec §Foreign Key Cascade Rules]
- [ ] CHK005 - Are all UNIQUE constraints documented with business rationale? [Completeness, Spec §Constraints]
- [ ] CHK006 - Are NOT NULL constraints explicitly specified for every column? [Completeness, Spec §Table Definitions]
- [ ] CHK007 - Are DEFAULT values documented for all columns where applicable? [Completeness, Spec §Table Definitions]
- [ ] CHK008 - Is the schema versioning strategy documented (version table, migration tracking)? [Completeness, Spec §Migration Strategy]
- [ ] CHK009 - Are all query patterns documented with example SQL and expected performance? [Completeness, Spec §Queries]
- [ ] CHK010 - Is the database file location and naming convention specified? [Gap]

## Data Type Appropriateness

- [ ] CHK011 - Are INTEGER types used appropriately for all numeric columns (id, size, timestamps, counts)? [Data Type, Spec §files, §folders, §scan_state]
- [ ] CHK012 - Are TEXT types used appropriately for all string columns (path, hash, algorithm, errors)? [Data Type, Spec §files, §folders]
- [ ] CHK013 - Is the timestamp storage format consistently specified (Unix epoch INTEGER) across all tables? [Consistency, Spec §files.first_scanned_at, §folders.first_scanned_at]
- [ ] CHK014 - Are boolean flags consistently represented as INTEGER (0/1) with DEFAULT values? [Consistency, Spec §files.removed, §scan_state.completed]
- [ ] CHK015 - Is the hash_value column type (TEXT) appropriate for hex-encoded hashes of varying lengths? [Data Type, Spec §files.hash_value]
- [ ] CHK016 - Is the size column type (INTEGER) sufficient for file sizes up to SQLite's 8-byte signed limit (9.2 exabytes)? [Data Type, Spec §files.size]
- [ ] CHK017 - Are path columns (TEXT) specified to handle maximum path lengths for target platforms (260 Windows, 4096 Linux)? [Clarity, Spec §files.path, §folders.path]
- [ ] CHK018 - Is the error_status column type (TEXT) appropriate for multi-line error messages? [Data Type, Spec §files.error_status]
- [ ] CHK019 - Are enumerated values (hash_algorithm, scan_mode) documented with valid values and CHECK constraints? [Completeness, Spec §files.hash_algorithm, §scan_state.scan_mode]

## Primary Key Requirements

- [ ] CHK020 - Does every table have a PRIMARY KEY constraint documented? [Completeness, Spec §files.id, §folders.id, §scan_state.id]
- [ ] CHK021 - Are all primary keys INTEGER AUTOINCREMENT with clear justification? [Consistency, Spec §Table Definitions]
- [ ] CHK022 - Is the use of surrogate keys (id) vs natural keys (path) justified for each table? [Clarity, Spec §Design Rationale]
- [ ] CHK023 - Are primary key naming conventions consistent across all tables (id field)? [Consistency, Spec §Table Definitions]

## Foreign Key Requirements

- [ ] CHK024 - Are all parent-child relationships represented with foreign key constraints? [Completeness, Spec §files.folder_id, §folders.parent_folder_id]
- [ ] CHK025 - Is the folders.parent_folder_id self-referential FK correctly documented with NULL handling for root? [Clarity, Spec §folders, §Design Rationale]
- [ ] CHK026 - Are cascade delete rules (ON DELETE CASCADE) appropriate for all FK relationships? [Integrity, Spec §Foreign Key Cascade Rules]
- [ ] CHK027 - Is the dual FK pattern (folder_id + root_folder_id) in files table justified with performance rationale? [Clarity, Spec §files, §Design Rationale]
- [ ] CHK028 - Are cascade update rules documented (or explicitly omitted with rationale)? [Completeness, Spec §Foreign Keys]
- [ ] CHK029 - Is foreign key enforcement explicitly enabled in SQLite configuration (PRAGMA foreign_keys = ON)? [Completeness, Gap]
- [ ] CHK030 - Are circular FK dependency scenarios addressed (e.g., folder ↔ files if bidirectional)? [Edge Case, Spec §Design Rationale]

## Constraint Requirements

- [ ] CHK031 - Is the UNIQUE constraint on files.path justified and documented? [Clarity, Spec §files]
- [ ] CHK032 - Is the UNIQUE constraint on folders.path justified and documented? [Clarity, Spec §folders]
- [ ] CHK033 - Is the partial UNIQUE constraint on scan_state.root_folder_id (WHERE completed = 0) correctly specified? [Clarity, Spec §scan_state, §Indexes]
- [ ] CHK034 - Are NOT NULL constraints on timestamp columns justified (ensuring data integrity)? [Clarity, Spec §files.first_scanned_at]
- [ ] CHK035 - Are nullable columns (hash_value, error_status) justified with business logic rationale? [Clarity, Spec §files, §Design Rationale]
- [ ] CHK036 - Are CHECK constraints specified for enumerated values (hash_algorithm, scan_mode)? [Gap, Spec §files.hash_algorithm]
- [ ] CHK037 - Are CHECK constraints specified for boolean flags (removed IN (0,1), completed IN (0,1))? [Gap]
- [ ] CHK038 - Are CHECK constraints specified for non-negative values (size >= 0, counts >= 0)? [Gap]
- [ ] CHK039 - Are timestamp ordering constraints documented (first_scanned_at <= last_scanned_at)? [Gap, Integrity Rule]

## Index Requirements

- [ ] CHK040 - Is the composite index on (hash_value, size) justified for duplicate detection query performance? [Clarity, Spec §idx_files_hash]
- [ ] CHK041 - Is the partial index WHERE clause (removed=0, error_status IS NULL, hash_value IS NOT NULL) justified? [Clarity, Spec §idx_files_hash, §Design Rationale]
- [ ] CHK042 - Are indexes on foreign key columns (folder_id, root_folder_id) documented for join performance? [Completeness, Spec §idx_files_folder]
- [ ] CHK043 - Is the index on folders.path justified for LIKE query performance (cascade removed flag)? [Clarity, Spec §idx_folders_path, §Queries]
- [ ] CHK044 - Are index column orderings specified and justified (e.g., hash_value before size)? [Clarity, Spec §idx_files_hash]
- [ ] CHK045 - Are missing indexes identified for common query patterns (e.g., files.mtime for temporal queries)? [Coverage, Gap]
- [ ] CHK046 - Is index maintenance overhead documented (insert/update performance impact)? [Gap, Performance]
- [ ] CHK047 - Are covering indexes considered for frequently accessed column combinations? [Gap, Performance]

## Migration Requirements

- [ ] CHK048 - Is the initial migration SQL complete with all CREATE TABLE statements? [Completeness, Spec §Migration Strategy]
- [ ] CHK049 - Is migration idempotency guaranteed (CREATE TABLE IF NOT EXISTS, version checks)? [Clarity, Spec §Migration Strategy]
- [ ] CHK050 - Is the schema_version table structure documented with columns and purpose? [Completeness, Spec §Future Migrations]
- [ ] CHK051 - Are ALTER TABLE migrations for root_folders documented with all 5 new columns? [Completeness, Spec §root_folders Extensions]
- [ ] CHK052 - Is the migration execution order specified (dependencies between migrations)? [Clarity, Spec §Migration Strategy]
- [ ] CHK053 - Is the migration rollback strategy documented (or explicitly not supported with rationale)? [Completeness, Spec §Migration Strategy]
- [ ] CHK054 - Are data migration requirements specified for schema changes (e.g., populating new columns)? [Gap]
- [ ] CHK055 - Is the migration versioning scheme documented (integer sequence, semantic versioning)? [Clarity, Spec §schema_version]
- [ ] CHK056 - Are migration failure handling requirements specified (transaction rollback, partial completion)? [Gap, Edge Case]

## Data Integrity Rules

- [ ] CHK057 - Are referential integrity rules documented for orphaned records (handled by CASCADE)? [Clarity, Spec §Foreign Key Cascade Rules]
- [ ] CHK058 - Is the removed flag lifecycle documented (when set, how queried, cascading logic)? [Clarity, Spec §files.removed, §Queries]
- [ ] CHK059 - Is the hash_value NULL state handling documented (pre-hashing, errors, null queries)? [Clarity, Spec §files.hash_value, §Design Rationale]
- [ ] CHK060 - Are timestamp consistency rules documented (first <= last, created <= updated)? [Gap, Integrity Rule]
- [ ] CHK061 - Is the root_folder_id consistency between files and folders enforced (matching root)? [Integrity, Spec §files.root_folder_id]
- [ ] CHK062 - Are statistics cache consistency rules documented (when to update folder_count, file_count, total_size)? [Clarity, Spec §root_folders, §Queries]
- [ ] CHK063 - Is the scan_state cleanup strategy documented (when to delete completed scans, retention policy)? [Gap, Spec §scan_state.completed]
- [ ] CHK064 - Are path format consistency rules documented (absolute paths, separator normalization)? [Gap, Spec §files.path]

## Performance Requirements

- [ ] CHK065 - Are query performance targets quantified for duplicate detection (<50ms per constitution)? [Measurability, Spec §Query Performance]
- [ ] CHK066 - Are table size estimates documented for capacity planning (e.g., 100K files = 20MB)? [Completeness, Spec §Table Sizes]
- [ ] CHK067 - Are index size estimates included in total storage calculations? [Completeness, Spec §Table Sizes]
- [ ] CHK068 - Is bulk insert performance documented (batching strategy, transaction size)? [Gap, Spec §Query Performance Tests]
- [ ] CHK069 - Is the cascade removed flag query performance quantified (1000 subfolders = <100ms)? [Measurability, Spec §Query Performance]
- [ ] CHK070 - Are checkpoint save/resume performance targets specified (<5ms per requirement)? [Measurability, Spec §Query Performance]
- [ ] CHK071 - Is WAL mode configuration documented with performance implications (concurrency, disk I/O)? [Gap, Spec §Overview]
- [ ] CHK072 - Are query plan optimization strategies documented (EXPLAIN QUERY PLAN usage)? [Gap]

## Schema Consistency

- [ ] CHK073 - Are table naming conventions consistent (lowercase, underscores, plural nouns)? [Consistency, Spec §Table Definitions]
- [ ] CHK074 - Are column naming conventions consistent across tables (id, path, removed, error_status)? [Consistency, Spec §Table Definitions]
- [ ] CHK075 - Are foreign key naming conventions consistent (table_id pattern for references)? [Consistency, Spec §files.folder_id]
- [ ] CHK076 - Are index naming conventions consistent (idx_table_column pattern)? [Consistency, Spec §Indexes]
- [ ] CHK077 - Are timestamp column naming consistent (first_scanned_at, last_scanned_at, started_at, updated_at)? [Consistency, Spec §Table Definitions]
- [ ] CHK078 - Are boolean flag representations consistent (INTEGER 0/1 with DEFAULT 0)? [Consistency, Spec §files.removed]
- [ ] CHK079 - Are nullable vs NOT NULL patterns consistent by column purpose (timestamps NOT NULL, hashes NULL)? [Consistency, Spec §Table Definitions]
- [ ] CHK080 - Are enumerated value formats consistent (lowercase strings: 'sha256', 'all')? [Consistency, Spec §files.hash_algorithm, §scan_state.scan_mode]

## Query Patterns & Optimization

- [ ] CHK081 - Are all functional requirements mapped to specific SQL queries? [Traceability, Spec §Queries]
- [ ] CHK082 - Is the duplicate file detection query optimized with proper index usage? [Performance, Spec §Duplicate File Detection]
- [ ] CHK083 - Is the GROUP_CONCAT usage in duplicate query appropriate for result size limits? [Edge Case, Spec §Duplicate File Detection]
- [ ] CHK084 - Are parameterized queries used for all user inputs (SQL injection prevention)? [Security, Spec §Queries]
- [ ] CHK085 - Is the folder hierarchy traversal strategy documented (recursive CTE, application logic)? [Clarity, Spec §Duplicate Folder Detection]
- [ ] CHK086 - Are LIKE query patterns optimized with proper index prefix matching? [Performance, Spec §Cascading Removed Flag]
- [ ] CHK087 - Is the UPSERT pattern (INSERT ... ON CONFLICT) correctly specified for checkpoint saves? [Clarity, Spec §Checkpoint Save]
- [ ] CHK088 - Are transaction boundaries documented for multi-statement operations? [Gap, Integrity]
- [ ] CHK089 - Is the ORDER BY clause justified for all queries returning multiple rows? [Clarity, Spec §Duplicate File Detection]

## SQLite-Specific Best Practices

- [ ] CHK090 - Is WAL (Write-Ahead Logging) mode explicitly enabled with rationale? [Completeness, Spec §Overview]
- [ ] CHK091 - Is the SQLite page size configured appropriately for workload (default 4KB vs optimized)? [Gap]
- [ ] CHK092 - Is the cache size configured for expected dataset size? [Gap, Performance]
- [ ] CHK093 - Is synchronous mode documented (FULL, NORMAL, OFF) with durability trade-offs? [Gap]
- [ ] CHK094 - Is foreign key enforcement PRAGMA explicitly set (not relying on defaults)? [Completeness, Gap]
- [ ] CHK095 - Are SQLite version requirements documented (minimum version for features used)? [Gap]
- [ ] CHK096 - Is the temp_store configuration documented (memory vs disk for temporary tables)? [Gap, Performance]
- [ ] CHK097 - Is the journal_mode documented beyond WAL (DELETE, TRUNCATE, PERSIST alternatives)? [Completeness, Spec §Overview]
- [ ] CHK098 - Is the auto_vacuum setting documented (NONE, FULL, INCREMENTAL)? [Gap, Storage]

## Edge Case Coverage

- [ ] CHK099 - Is NULL handling documented for all nullable columns in query WHERE clauses? [Clarity, Spec §Queries]
- [ ] CHK100 - Are empty table scenarios addressed (no files, no folders, no roots)? [Coverage, Gap]
- [ ] CHK101 - Is the maximum path length handling documented (truncation, error, validation)? [Edge Case, Gap]
- [ ] CHK102 - Are duplicate path scenarios prevented (UNIQUE constraint + error handling)? [Coverage, Spec §files.path]
- [ ] CHK103 - Is the maximum hash_value length documented (SHA-512 = 128 hex chars)? [Clarity, Spec §files.hash_value]
- [ ] CHK104 - Are concurrent access scenarios addressed (WAL mode + locking strategy)? [Coverage, Spec §Overview]
- [ ] CHK105 - Is the stale checkpoint detection strategy documented (updated_at timestamp threshold)? [Clarity, Spec §scan_state, §Design Rationale]
- [ ] CHK106 - Are zero-size file handling rules documented (include in duplicates or exclude)? [Edge Case, Gap]
- [ ] CHK107 - Is the root folder self-reference scenario addressed (parent_folder_id NULL)? [Clarity, Spec §folders.parent_folder_id]

## Testability Requirements

- [ ] CHK108 - Can all table creation statements be validated in tests (no syntax errors)? [Measurability, Spec §Schema Tests]
- [ ] CHK109 - Can foreign key cascade behavior be validated with test data? [Measurability, Spec §Schema Tests]
- [ ] CHK110 - Can UNIQUE constraint violations be tested systematically? [Measurability, Spec §Schema Tests]
- [ ] CHK111 - Can query performance be measured against documented targets? [Measurability, Spec §Query Performance Tests]
- [ ] CHK112 - Can migration idempotency be validated (run twice without errors)? [Measurability, Spec §Migration Tests]
- [ ] CHK113 - Can index usage be verified with EXPLAIN QUERY PLAN? [Measurability, Gap]
- [ ] CHK114 - Can data integrity rules be validated with constraint violations? [Measurability, Spec §Schema Tests]
- [ ] CHK115 - Can schema version tracking be tested (migration sequence validation)? [Measurability, Spec §Migration Tests]

## Documentation Completeness

- [ ] CHK116 - Is each table documented with business purpose and design rationale? [Completeness, Spec §Table Definitions]
- [ ] CHK117 - Is each index documented with query pattern justification? [Completeness, Spec §Indexes, §Design Rationale]
- [ ] CHK118 - Are all design decisions documented with trade-off analysis? [Clarity, Spec §Design Rationale]
- [ ] CHK119 - Are example rows provided for each table to illustrate data format? [Completeness, Spec §Example Row]
- [ ] CHK120 - Are all queries documented with example SQL and expected results? [Completeness, Spec §Queries]
- [ ] CHK121 - Is the ERD (Entity Relationship Diagram) complete and consistent with table definitions? [Consistency, Spec §Entity Relationship Diagram]
- [ ] CHK122 - Are external references documented (SQLite docs, best practices guides)? [Completeness, Spec §References]

## Security & Access Control

- [ ] CHK123 - Is SQL injection prevention strategy documented (parameterized queries)? [Gap, Security]
- [ ] CHK124 - Are database file permissions requirements specified? [Gap, Security]
- [ ] CHK125 - Is database encryption at rest addressed (SQLite Encryption Extension considerations)? [Gap, Security]
- [ ] CHK126 - Are sensitive data fields identified (if any exist beyond file paths)? [Gap, Security]

## Cross-Table Consistency

- [ ] CHK127 - Are root_folder_id references consistent between files and folders tables? [Consistency, Spec §files.root_folder_id, §folders.root_folder_id]
- [ ] CHK128 - Are removed flag semantics consistent across files and folders tables? [Consistency, Spec §files.removed, §folders.removed]
- [ ] CHK129 - Are timestamp semantics consistent (Unix epoch INTEGER) across all tables? [Consistency, Spec §Table Definitions]
- [ ] CHK130 - Are error_status field semantics consistent between files and folders? [Consistency, Spec §files.error_status, §folders.error_status]
- [ ] CHK131 - Is the hash_algorithm enumeration consistent with configuration requirements? [Consistency, Spec §files.hash_algorithm]

---

**Summary**: 131 requirement quality checks for Database Schema & Data Model covering structure, integrity, performance, migrations, and best practices. Target audience: DBA/Architect + QA/Test Team for design validation and test planning.

**Notes**:
- Items marked [Gap] indicate missing requirements that should be added to data-model.md
- Items marked [Ambiguity] or [Clarity] indicate existing requirements needing more detail
- Items marked [Spec §section] reference specific data-model.md sections for validation
- Items marked [Consistency] check for patterns across multiple tables/columns
- Items marked [Measurability] validate that requirements include quantifiable targets

**Coverage Summary**:
- Schema Structure: CHK001-CHK019 (19 items)
- Keys & Relationships: CHK020-CHK030 (11 items)
- Constraints: CHK031-CHK039 (9 items)
- Indexes: CHK040-CHK047 (8 items)
- Migrations: CHK048-CHK056 (9 items)
- Data Integrity: CHK057-CHK064 (8 items)
- Performance: CHK065-CHK072 (8 items)
- Consistency: CHK073-CHK080 (8 items)
- Query Patterns: CHK081-CHK089 (9 items)
- SQLite Best Practices: CHK090-CHK098 (9 items)
- Edge Cases: CHK099-CHK107 (9 items)
- Testability: CHK108-CHK115 (8 items)
- Documentation: CHK116-CHK122 (7 items)
- Security: CHK123-CHK126 (4 items)
- Cross-Table: CHK127-CHK131 (5 items)
