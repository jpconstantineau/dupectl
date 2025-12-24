/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var getRootJSON bool

type RootFolderInfo struct {
	Path           string  `json:"path"`
	FolderCount    int64   `json:"folder_count"`
	FileCount      int64   `json:"file_count"`
	TotalSizeBytes int64   `json:"total_size_bytes"`
	LastScanDate   *string `json:"last_scan_date"`
}

// getRootCmd represents the getRoot command
var getRootCmd = &cobra.Command{
	Use:   "root",
	Short: "Get list of root folders",
	Long: `List all registered root folders with scan statistics.

Displays path, folder count, file count, total size, and last scan date
for all registered root folders.

Example:
  dupectl get root
  dupectl get root --json`,
	Run: func(cmd *cobra.Command, args []string) {
		// Open database connection
		db, err := openDatabaseForGetRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// Query all root folders
		rows, err := db.Query(`
			SELECT path, folder_count, file_count, total_size, last_scan_date
			FROM root_folders
			ORDER BY path
		`)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to query root folders: %v\n", err)
			os.Exit(1)
		}
		defer rows.Close()

		// Collect results
		var rootFolders []RootFolderInfo

		for rows.Next() {
			var rf RootFolderInfo
			var lastScan sql.NullString

			err := rows.Scan(&rf.Path, &rf.FolderCount, &rf.FileCount, &rf.TotalSizeBytes, &lastScan)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to scan row: %v\n", err)
				os.Exit(1)
			}

			if lastScan.Valid && lastScan.String != "" {
				// Parse the timestamp and convert to RFC3339
				t, err := time.Parse("2006-01-02 15:04:05", lastScan.String)
				if err == nil {
					scanTime := t.UTC().Format(time.RFC3339)
					rf.LastScanDate = &scanTime
				} else {
					// Try with timezone
					t, err = time.Parse("2006-01-02 15:04:05-07:00", lastScan.String)
					if err == nil {
						scanTime := t.UTC().Format(time.RFC3339)
						rf.LastScanDate = &scanTime
					}
				}
			}

			rootFolders = append(rootFolders, rf)
		}

		// Output results
		if getRootJSON {
			outputRootJSON(rootFolders)
		} else {
			outputRootTable(rootFolders)
		}
	},
}

func openDatabaseForGetRoot() (*sql.DB, error) {
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

	return db, nil
}

func outputRootTable(rootFolders []RootFolderInfo) {
	if len(rootFolders) == 0 {
		fmt.Println("No root folders registered")
		return
	}

	fmt.Println("Root Folders")
	fmt.Println("════════════")
	fmt.Println()
	fmt.Printf("%-40s  %-10s  %-10s  %-12s  %s\n", "Path", "Folders", "Files", "Total Size", "Last Scan")
	fmt.Println(strings.Repeat("─", 120))

	for _, rf := range rootFolders {
		path := rf.Path
		if len(path) > 40 {
			path = "..." + path[len(path)-37:]
		}

		folderCountStr := formatNumberForTable(rf.FolderCount)
		fileCountStr := formatNumberForTable(rf.FileCount)
		sizeStr := formatBytesForTable(rf.TotalSizeBytes)

		lastScanStr := "Never scanned"
		if rf.LastScanDate != nil {
			t, err := time.Parse(time.RFC3339, *rf.LastScanDate)
			if err == nil {
				lastScanStr = t.Format("2006-01-02 15:04:05 MST")
			}
		}

		fmt.Printf("%-40s  %-10s  %-10s  %-12s  %s\n", path, folderCountStr, fileCountStr, sizeStr, lastScanStr)
	}

	fmt.Println()
	fmt.Printf("Total: %d root folder", len(rootFolders))
	if len(rootFolders) != 1 {
		fmt.Print("s")
	}
	fmt.Println()
}

func outputRootJSON(rootFolders []RootFolderInfo) {
	result := map[string]interface{}{
		"root_folders": rootFolders,
		"summary": map[string]int{
			"total_roots": len(rootFolders),
		},
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func formatNumberForTable(n int64) string {
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

func formatBytesForTable(bytes int64) string {
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
	getCmd.AddCommand(getRootCmd)
	getRootCmd.Flags().BoolVar(&getRootJSON, "json", false, "Output in JSON format")
}
