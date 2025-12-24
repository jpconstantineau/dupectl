/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// deleteRootCmd represents the deleteRoot command
var deleteRootCmd = &cobra.Command{
	Use:   "root <root-folder-path>",
	Short: "Delete root folder from database",
	Long: `Remove registered root folder and delete all associated scan data.

This operation deletes the root folder record from the database, which CASCADE deletes
all associated folders, files, and scan state. This action cannot be undone.

Example:
  dupectl delete root /path/to/root
  dupectl delete root "C:\Users\user\Documents" --yes`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rootPath := args[0]
		absPath, err := filepath.Abs(rootPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid path: %v\n", err)
			os.Exit(1)
		}

		// Open database connection
		db, err := openDatabaseForDeleteRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// Check if root folder exists and get its statistics
		var rootID int64
		var folderCount, fileCount, totalSize int64
		err = db.QueryRow(`
			SELECT id, folder_count, file_count, total_size 
			FROM root_folders 
			WHERE path = ?
		`, absPath).Scan(&rootID, &folderCount, &fileCount, &totalSize)

		if err == sql.ErrNoRows {
			fmt.Fprintf(os.Stderr, "Error: Root folder not registered: %s\n", absPath)
			os.Exit(1)
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to query root folder: %v\n", err)
			os.Exit(1)
		}

		// Prompt for confirmation unless --yes flag is set
		if !rootYes {
			fmt.Printf("Delete root folder and all scan data? (y/n): ")
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
				os.Exit(1)
			}

			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Deletion cancelled.")
				os.Exit(0)
			}
		}

		// Delete root folder (CASCADE will delete folders, files, scan_state)
		_, err = db.Exec("DELETE FROM root_folders WHERE id = ?", rootID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to delete root folder: %v\n", err)
			os.Exit(1)
		}

		// Display confirmation
		fmt.Printf("Root folder deleted: %s\n", absPath)
		fmt.Printf("Removed: %s folders, %s files, %s of scan data\n",
			formatNumberForDelete(folderCount),
			formatNumberForDelete(fileCount),
			formatBytesForDelete(totalSize))
	},
}

func openDatabaseForDeleteRoot() (*sql.DB, error) {
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

	// Enable foreign keys to ensure CASCADE works
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func formatNumberForDelete(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	str := fmt.Sprintf("%d", n)
	result := ""
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result
}

func formatBytesForDelete(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

func init() {
	deleteCmd.AddCommand(deleteRootCmd)
}
