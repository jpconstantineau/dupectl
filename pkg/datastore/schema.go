package datastore

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Migration represents a database schema migration
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// migrations contains all database schema migrations in order
var migrations = []Migration{
	{
		Version: 1,
		Name:    "create_scan_tables",
		SQL:     migration001CreateScanTables,
	},
}

// migration001CreateScanTables creates the files, folders, and scan_state tables
const migration001CreateScanTables = `
-- Root folders table (minimal implementation for MVP)
CREATE TABLE IF NOT EXISTS root_folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    created_at INTEGER NOT NULL
);

-- Files table
CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    root_folder_id INTEGER NOT NULL,
    folder_id INTEGER,
    path TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    size INTEGER NOT NULL,
    mtime INTEGER NOT NULL,
    hash_value TEXT,
    hash_algorithm TEXT,
    error_status TEXT,
    scanned_at INTEGER,
    FOREIGN KEY (root_folder_id) REFERENCES root_folders(id) ON DELETE CASCADE,
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_files_hash_size ON files(hash_value, size);
CREATE INDEX IF NOT EXISTS idx_files_root ON files(root_folder_id);
CREATE INDEX IF NOT EXISTS idx_files_folder ON files(folder_id);
CREATE INDEX IF NOT EXISTS idx_files_error ON files(error_status) WHERE error_status IS NOT NULL;

-- Folders table
CREATE TABLE IF NOT EXISTS folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    root_folder_id INTEGER NOT NULL,
    parent_folder_id INTEGER,
    path TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    scanned_at INTEGER,
    FOREIGN KEY (root_folder_id) REFERENCES root_folders(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_folder_id) REFERENCES folders(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_folders_root ON folders(root_folder_id);
CREATE INDEX IF NOT EXISTS idx_folders_parent ON folders(parent_folder_id);
CREATE INDEX IF NOT EXISTS idx_folders_path ON folders(path);

-- Scan state table
CREATE TABLE IF NOT EXISTS scan_state (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    root_folder_id INTEGER NOT NULL,
    scan_mode TEXT NOT NULL,
    current_folder_path TEXT,
    started_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    completed_at INTEGER,
    status TEXT NOT NULL,
    files_processed INTEGER DEFAULT 0,
    folders_processed INTEGER DEFAULT 0,
    FOREIGN KEY (root_folder_id) REFERENCES root_folders(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_scan_state_root_status ON scan_state(root_folder_id, status);
CREATE INDEX IF NOT EXISTS idx_scan_state_started ON scan_state(started_at);
`

// ApplyMigrations executes all pending migrations on the database
func ApplyMigrations(db *sql.DB) error {
	// Create migrations tracking table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current schema version
	var currentVersion int
	err = db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if migration.Version <= currentVersion {
			continue // Skip already applied migrations
		}

		// Begin transaction for this migration
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %d: %w", migration.Version, err)
		}

		// Split SQL by semicolons and execute each statement
		// Note: This is a simple split and doesn't handle semicolons within strings
		statements := splitSQL(migration.SQL)
		for i, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			_, err = tx.Exec(stmt)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to execute migration %d (%s) statement %d: %w", migration.Version, migration.Name, i+1, err)
			}
		}

		// Record migration as applied
		_, err = tx.Exec(
			"INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)",
			migration.Version,
			migration.Name,
			getCurrentTimestamp(),
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		// Commit transaction
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
		}
	}

	return nil
}

// getCurrentTimestamp returns current Unix timestamp
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// splitSQL splits a SQL string into individual statements
// This is a simple implementation that splits on semicolons
// It doesn't handle semicolons within strings (good enough for migrations)
func splitSQL(sql string) []string {
	return strings.Split(sql, ";")
}
