/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/jpconstantineau/dupectl/internal/config"
	"github.com/jpconstantineau/dupectl/pkg/duplicate"
	"github.com/spf13/cobra"

	_ "modernc.org/sqlite"
)

var (
	duplicatesJSON     bool
	duplicatesDetails  bool
	duplicatesMinCount int
	duplicatesMinSize  string
)

// getDuplicatesCmd represents the getDuplicates command
var getDuplicatesCmd = &cobra.Command{
	Use:   "duplicates",
	Short: "Query and display duplicate files",
	Long: `Find and display duplicate files identified during scans.
Supports summary and detailed views, with table and JSON output formats.

By default, shows a summary table grouped by root folder.
Use --details to see individual file paths.

Examples:
  dupectl get duplicates                      # Summary view (default)
  dupectl get duplicates --details            # Detailed view with file paths
  dupectl get duplicates --json               # JSON output
  dupectl get duplicates --min-count 3        # Only sets with 3+ files
  dupectl get duplicates --min-size 1M        # 1 megabyte minimum
  dupectl get duplicates --min-size 512K      # 512 kilobytes minimum
  dupectl get duplicates --min-size 1048576   # bytes also supported`,
	Run: func(cmd *cobra.Command, args []string) {
		runGetDuplicates()
	},
}

func init() {
	getCmd.AddCommand(getDuplicatesCmd)

	getDuplicatesCmd.Flags().BoolVar(&duplicatesJSON, "json", false, "Output in JSON format")
	getDuplicatesCmd.Flags().BoolVar(&duplicatesDetails, "details", false, "Show detailed view with individual file paths")
	getDuplicatesCmd.Flags().IntVar(&duplicatesMinCount, "min-count", 2, "Minimum number of duplicates in a set")
	getDuplicatesCmd.Flags().StringVar(&duplicatesMinSize, "min-size", "0", "Minimum file size (e.g., 1M, 512K, 1024) - 0 = no minimum")
}

func runGetDuplicates() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to load configuration: %v\n", err)
		os.Exit(2)
	}

	// Parse minimum size
	minSize, err := parseSize(duplicatesMinSize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid --min-size value '%s': %v\n", duplicatesMinSize, err)
		os.Exit(2)
	}

	// Open database
	db, err := sql.Open("sqlite", cfg.DatabasePath+"?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to open database: %v\n", err)
		os.Exit(2)
	}
	defer db.Close()

	// Create detector
	detector := duplicate.NewDetector(db)

	// Find duplicates
	sets, err := detector.FindDuplicateFiles(duplicatesMinCount, minSize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to find duplicates: %v\n", err)
		os.Exit(2)
	}

	// Format and display
	formatter := duplicate.NewFormatter()

	if duplicatesJSON {
		output, err := formatter.FormatJSON(sets)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to format JSON: %v\n", err)
			os.Exit(2)
		}
		fmt.Println(output)
	} else if duplicatesDetails {
		// Detailed view with file paths (original behavior)
		output := formatter.FormatTable(sets)
		fmt.Print(output)
	} else {
		// Summary view grouped by root folder (default)
		output := formatter.FormatSummary(sets)
		fmt.Print(output)
	}
}

// parseSize parses human-readable size strings like "10M", "512K", "1G"
// Returns size in bytes
func parseSize(sizeStr string) (int64, error) {
	if sizeStr == "" || sizeStr == "0" {
		return 0, nil
	}

	// Check for suffix
	lastChar := sizeStr[len(sizeStr)-1]
	var multiplier int64 = 1
	numStr := sizeStr

	if lastChar >= 'A' && lastChar <= 'Z' || lastChar >= 'a' && lastChar <= 'z' {
		numStr = sizeStr[:len(sizeStr)-1]
		switch lastChar {
		case 'K', 'k':
			multiplier = 1024
		case 'M', 'm':
			multiplier = 1024 * 1024
		case 'G', 'g':
			multiplier = 1024 * 1024 * 1024
		case 'T', 't':
			multiplier = 1024 * 1024 * 1024 * 1024
		default:
			return 0, fmt.Errorf("unknown size suffix '%c' (use K, M, G, or T)", lastChar)
		}
	}

	// Parse the numeric part
	var size int64
	_, err := fmt.Sscanf(numStr, "%d", &size)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %w", err)
	}

	return size * multiplier, nil
}
