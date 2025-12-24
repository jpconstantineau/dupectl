/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/jpconstantineau/dupectl/internal/config"
	"github.com/jpconstantineau/dupectl/pkg/pathutil"
	"github.com/spf13/cobra"

	_ "modernc.org/sqlite"
)

var (
	purgeBefore string
)

// purgeCmd represents the purge command
var purgeCmd = &cobra.Command{
	Use:   "purge [files|folders|all] <root-folder-path>",
	Short: "Permanently delete removed entities from database",
	Long: `Permanently delete entities marked as removed=1 from the database to free storage.

This operation permanently deletes file and/or folder records from the database.
This cannot be undone. Use --before to limit purge to entities removed before a specific date.

Examples:
  dupectl purge files /home/user/documents
  dupectl purge all /home/user/documents --before 2024-01-01
  dupectl purge folders "C:\Users\user\Documents" --yes`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		entityType := args[0]
		rootPath := args[1]
		runPurge(entityType, rootPath)
	},
}

func init() {
	rootCmd.AddCommand(purgeCmd)

	purgeCmd.Flags().StringVar(&purgeBefore, "before", "", "Only purge entities removed before date (YYYY-MM-DD)")
}

func runPurge(entityType, rootPath string) {
	// Validate entity type
	if entityType != "files" && entityType != "folders" && entityType != "all" {
		fmt.Fprintf(os.Stderr, "Error: Invalid entity type '%s'. Must be 'files', 'folders', or 'all'\n", entityType)
		os.Exit(2)
	}

	// Convert to absolute path
	absPath, err := pathutil.ToAbsolute(rootPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to resolve path: %v\n", err)
		os.Exit(2)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to load configuration: %v\n", err)
		os.Exit(2)
	}

	// Open database
	db, err := sql.Open("sqlite", cfg.DatabasePath+"?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to open database: %v\n", err)
		os.Exit(2)
	}
	defer db.Close()

	// Check if root folder is registered
	var rootID int64
	err = db.QueryRow("SELECT id FROM root_folders WHERE path = ?", absPath).Scan(&rootID)
	if err == sql.ErrNoRows {
		fmt.Fprintf(os.Stderr, "Error: Root folder not registered: %s\n", absPath)
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to query root folder: %v\n", err)
		os.Exit(2)
	}

	// Parse --before date if specified
	var beforeTimestamp int64
	if purgeBefore != "" {
		beforeDate, err := time.Parse("2006-01-02", purgeBefore)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid --before date format. Use YYYY-MM-DD: %v\n", err)
			os.Exit(2)
		}
		beforeTimestamp = beforeDate.Unix()
	}

	// Count removed entities
	fileCount, folderCount := countRemovedEntities(db, rootID, beforeTimestamp)

	if entityType == "files" && fileCount == 0 {
		fmt.Println("No removed files to purge")
		return
	}
	if entityType == "folders" && folderCount == 0 {
		fmt.Println("No removed folders to purge")
		return
	}
	if entityType == "all" && fileCount == 0 && folderCount == 0 {
		fmt.Println("No removed entities to purge")
		return
	}

	// Display what will be purged
	fmt.Printf("Purging removed %s from: %s\n", entityType, absPath)
	if entityType == "files" {
		fmt.Printf("Found %d removed files\n", fileCount)
	} else if entityType == "folders" {
		fmt.Printf("Found %d removed folders\n", folderCount)
	} else {
		fmt.Printf("Found %d removed files and %d removed folders\n", fileCount, folderCount)
	}

	// Prompt for confirmation unless --yes flag is set
	if !rootYes {
		var response string
		totalCount := fileCount + folderCount
		fmt.Printf("Permanently delete %d %s? This cannot be undone. (y/n): ", totalCount, entityType)
		fmt.Scanln(&response)

		if response != "y" && response != "Y" && response != "yes" {
			fmt.Println("Purge cancelled.")
			return
		}
	}

	// Perform purge
	purgedFiles, purgedFolders := performPurge(db, rootID, entityType, beforeTimestamp)

	// Display results
	if entityType == "files" {
		fmt.Printf("Purged %d files from database\n", purgedFiles)
	} else if entityType == "folders" {
		fmt.Printf("Purged %d folders from database\n", purgedFolders)
	} else {
		fmt.Printf("Purged %d files and %d folders from database\n", purgedFiles, purgedFolders)
	}
}

func countRemovedEntities(db *sql.DB, rootID int64, beforeTimestamp int64) (fileCount, folderCount int) {
	// Count removed files
	fileQuery := "SELECT COUNT(*) FROM files WHERE root_folder_id = ? AND removed = 1"
	fileArgs := []interface{}{rootID}
	if beforeTimestamp > 0 {
		fileQuery += " AND last_scanned_at < ?"
		fileArgs = append(fileArgs, beforeTimestamp)
	}
	db.QueryRow(fileQuery, fileArgs...).Scan(&fileCount)

	// Count removed folders
	folderQuery := "SELECT COUNT(*) FROM folders WHERE root_folder_id = ? AND removed = 1"
	folderArgs := []interface{}{rootID}
	if beforeTimestamp > 0 {
		folderQuery += " AND last_scanned_at < ?"
		folderArgs = append(folderArgs, beforeTimestamp)
	}
	db.QueryRow(folderQuery, folderArgs...).Scan(&folderCount)

	return fileCount, folderCount
}

func performPurge(db *sql.DB, rootID int64, entityType string, beforeTimestamp int64) (purgedFiles, purgedFolders int64) {
	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to begin transaction: %v\n", err)
		os.Exit(2)
	}
	defer tx.Rollback()

	// Purge files
	if entityType == "files" || entityType == "all" {
		fileQuery := "DELETE FROM files WHERE root_folder_id = ? AND removed = 1"
		fileArgs := []interface{}{rootID}
		if beforeTimestamp > 0 {
			fileQuery += " AND last_scanned_at < ?"
			fileArgs = append(fileArgs, beforeTimestamp)
		}
		result, err := tx.Exec(fileQuery, fileArgs...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to purge files: %v\n", err)
			os.Exit(2)
		}
		purgedFiles, _ = result.RowsAffected()
	}

	// Purge folders
	if entityType == "folders" || entityType == "all" {
		folderQuery := "DELETE FROM folders WHERE root_folder_id = ? AND removed = 1"
		folderArgs := []interface{}{rootID}
		if beforeTimestamp > 0 {
			folderQuery += " AND last_scanned_at < ?"
			folderArgs = append(folderArgs, beforeTimestamp)
		}
		result, err := tx.Exec(folderQuery, folderArgs...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to purge folders: %v\n", err)
			os.Exit(2)
		}
		purgedFolders, _ = result.RowsAffected()
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to commit purge: %v\n", err)
		os.Exit(2)
	}

	return purgedFiles, purgedFolders
}
