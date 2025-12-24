package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// refreshCmd represents the refresh command
var refreshCmd = &cobra.Command{
	Use:   "refresh all <root-folder-path>",
	Short: "Recalculate root folder statistics from database without full scan",
	Long: `Recalculate root folder statistics from database without full scan.

This operation updates folder_count, file_count, total_size, and last_scan_date
by querying the current database state, without performing a full filesystem scan.

Example:
  dupectl refresh all /path/to/root`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if args[0] != "all" {
			fmt.Fprintf(os.Stderr, "Error: first argument must be 'all'\n")
			os.Exit(1)
		}

		rootPath := args[1]
		absPath, err := filepath.Abs(rootPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid path: %v\n", err)
			os.Exit(1)
		}

		// Open database connection
		db, err := openDatabaseForRefresh()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// Verify root folder exists
		var rootID int64
		err = db.QueryRow("SELECT id FROM rootfolders WHERE path = ?", absPath).Scan(&rootID)
		if err == sql.ErrNoRows {
			fmt.Fprintf(os.Stderr, "Error: root folder not found in database: %s\n", absPath)
			fmt.Fprintf(os.Stderr, "Use 'dupectl add root %s' to register it first.\n", absPath)
			os.Exit(1)
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to query root folder: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Refreshing statistics for: %s\n", absPath)

		// Calculate statistics from database
		var folderCount, fileCount, totalSize int64

		err = db.QueryRow("SELECT COUNT(*) FROM folders WHERE root_folder_id = ? AND removed = 0", rootID).Scan(&folderCount)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to count folders: %v\n", err)
			os.Exit(1)
		}

		err = db.QueryRow("SELECT COUNT(*) FROM files WHERE root_folder_id = ? AND removed = 0", rootID).Scan(&fileCount)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to count files: %v\n", err)
			os.Exit(1)
		}

		err = db.QueryRow("SELECT COALESCE(SUM(size), 0) FROM files WHERE root_folder_id = ? AND removed = 0", rootID).Scan(&totalSize)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to calculate total size: %v\n", err)
			os.Exit(1)
		}

		// Update root folder statistics
		now := time.Now().UTC()
		_, err = db.Exec(`
			UPDATE rootfolders 
			SET folder_count = ?, 
				file_count = ?, 
				total_size = ?, 
				last_scan_date = ?
			WHERE id = ?
		`, folderCount, fileCount, totalSize, now, rootID)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to update statistics: %v\n", err)
			os.Exit(1)
		}

		// Display updated statistics
		fmt.Printf("Folder count: %s\n", formatNumber(folderCount))
		fmt.Printf("File count: %s\n", formatNumber(fileCount))
		fmt.Printf("Total size: %s\n", formatBytes(totalSize))
		fmt.Printf("Last updated: %s\n", now.Format("2006-01-02 15:04:05 MST"))
	},
}

func openDatabaseForRefresh() (*sql.DB, error) {
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

func formatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	// Add thousand separators
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

func formatBytes(bytes int64) string {
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
	rootCmd.AddCommand(refreshCmd)
}
