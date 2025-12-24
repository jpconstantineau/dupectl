package datastore

import (
	"database/sql"
	"fmt"

	"github.com/jpconstantineau/dupectl/pkg/logger"
)

// Migration represents a database schema version
type Migration struct {
	Version     int
	Description string
	Up          func(*sql.DB) error
	Down        func(*sql.DB) error
}

var migrations = []Migration{
	{
		Version:     1,
		Description: "Create scan tables (root_folders, files, folders, scan_state)",
		Up:          migrationV1Up,
		Down:        migrationV1Down,
	},
}

// RunMigrations runs all pending migrations
func RunMigrations(db *sql.DB) error {
	// Create schema_version table if not exists
	err := createSchemaVersionTable(db)
	if err != nil {
		return fmt.Errorf("failed to create schema_version table: %w", err)
	}

	// Get current version
	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	logger.Info("Current database schema version: %d", currentVersion)

	// Run pending migrations
	for _, migration := range migrations {
		if migration.Version > currentVersion {
			logger.Info("Running migration %d: %s", migration.Version, migration.Description)

			tx, err := db.Begin()
			if err != nil {
				return fmt.Errorf("failed to start transaction for migration %d: %w", migration.Version, err)
			}

			err = migration.Up(db)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("migration %d failed: %w", migration.Version, err)
			}

			err = setVersion(db, migration.Version)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update schema version: %w", err)
			}

			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
			}

			logger.Info("Migration %d completed successfully", migration.Version)
		}
	}

	return nil
}

func createSchemaVersionTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_version (
		version INTEGER PRIMARY KEY,
		applied_at INTEGER NOT NULL
	);`
	_, err := db.Exec(query)
	return err
}

func getCurrentVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

func setVersion(db *sql.DB, version int) error {
	_, err := db.Exec("INSERT INTO schema_version (version, applied_at) VALUES (?, strftime('%s', 'now'))", version)
	return err
}

// Migration V1: Create scan tables
func migrationV1Up(db *sql.DB) error {
	// Create tables first
	tables := []string{
		CreateRootFoldersTableSQL,
		CreateFoldersTableSQL,
		CreateFilesTableSQL,
		CreateScanStateTableSQL,
	}

	for _, query := range tables {
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	// Create indexes separately
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_root_folders_path ON root_folders(path)",
		"CREATE INDEX IF NOT EXISTS idx_root_folders_agent ON root_folders(agent_id)",
		"CREATE INDEX IF NOT EXISTS idx_folders_parent ON folders(parent_folder_id)",
		"CREATE INDEX IF NOT EXISTS idx_folders_root ON folders(root_folder_id)",
		"CREATE INDEX IF NOT EXISTS idx_folders_path ON folders(path)",
		"CREATE INDEX IF NOT EXISTS idx_folders_removed ON folders(removed)",
		"CREATE INDEX IF NOT EXISTS idx_files_folder ON files(folder_id)",
		"CREATE INDEX IF NOT EXISTS idx_files_root ON files(root_folder_id)",
		"CREATE INDEX IF NOT EXISTS idx_files_removed ON files(removed)",
		"CREATE INDEX IF NOT EXISTS idx_files_path ON files(path)",
		"CREATE INDEX IF NOT EXISTS idx_files_hash ON files(hash_value, size) WHERE removed = 0 AND error_status IS NULL AND hash_value IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_scan_state_root ON scan_state(root_folder_id)",
		"CREATE INDEX IF NOT EXISTS idx_scan_state_completed ON scan_state(completed)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_scan_state_unique_active ON scan_state(root_folder_id, completed) WHERE completed = 0",
	}

	for _, query := range indexes {
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

func migrationV1Down(db *sql.DB) error {
	queries := []string{
		"DROP TABLE IF EXISTS scan_state",
		"DROP TABLE IF EXISTS files",
		"DROP TABLE IF EXISTS folders",
		"DROP TABLE IF EXISTS root_folders",
	}
	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}
