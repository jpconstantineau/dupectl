/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var addRootTraverseLinks bool

// addRootCmd represents the addRoot command
var addRootCmd = &cobra.Command{
	Use:   "root <root-folder-path>",
	Short: "Add new root where files are stored",
	Long: `Register a new root folder for monitoring.

This operation validates the path exists, checks if it's already registered,
and adds the root folder record to the database with the specified configuration.

Example:
  dupectl add root /home/user/documents
  dupectl add root "C:\Users\user\Documents" --traverse-links`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rootPath := args[0]

		// Convert relative path to absolute path
		absPath, err := filepath.Abs(rootPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid path: %v\n", err)
			os.Exit(1)
		}

		// Validate path exists on filesystem
		fileInfo, err := os.Stat(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Error: Root folder does not exist: %s\n", absPath)
				os.Exit(1)
			}
			if os.IsPermission(err) {
				fmt.Fprintf(os.Stderr, "Error: Permission denied accessing root folder: %s\n", absPath)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Error: failed to access path: %v\n", err)
			os.Exit(1)
		}

		// Verify it's a directory
		if !fileInfo.IsDir() {
			fmt.Fprintf(os.Stderr, "Error: Path is not a directory: %s\n", absPath)
			os.Exit(1)
		}

		// Open database connection
		db, err := openDatabaseForAddRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// Check if root folder already registered
		var existingID int64
		err = db.QueryRow("SELECT id FROM root_folders WHERE path = ?", absPath).Scan(&existingID)
		if err == nil {
			fmt.Fprintf(os.Stderr, "Error: Root folder already registered: %s\n", absPath)
			os.Exit(1)
		} else if err != sql.ErrNoRows {
			fmt.Fprintf(os.Stderr, "Error: failed to check for existing root folder: %v\n", err)
			os.Exit(1)
		}

		// Insert root folder record
		result, err := db.Exec(`
			INSERT INTO root_folders (
				path, 
				traverse_links, 
				folder_count, 
				file_count, 
				total_size
			) VALUES (?, ?, 0, 0, 0)
		`, absPath, addRootTraverseLinks)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to register root folder: %v\n", err)
			os.Exit(1)
		}

		rootID, _ := result.LastInsertId()

		// Display confirmation
		fmt.Printf("Root folder registered: %s\n", absPath)
		fmt.Println("Configuration:")
		fmt.Printf("  Traverse Links: %v\n", addRootTraverseLinks)
		fmt.Println()
		fmt.Printf("Run 'dupectl scan all %s' to start scanning.\n", absPath)

		if rootID > 0 {
			fmt.Printf("(Root folder ID: %d)\n", rootID)
		}
	},
}

func openDatabaseForAddRoot() (*sql.DB, error) {
	dbType := viper.GetString("server.database.type")
	if dbType != "sqlite" {
		return nil, fmt.Errorf("only SQLite databases are currently supported")
	}

	dbPath := viper.GetString("server.database.sqlite.name")
	if dbPath == "" {
		return nil, fmt.Errorf("database path not configured")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func init() {
	addCmd.AddCommand(addRootCmd)
	addRootCmd.Flags().BoolVar(&addRootTraverseLinks, "traverse-links", false, "Follow symbolic links during scans")
}
