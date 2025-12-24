/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/jpconstantineau/dupectl/pkg/datastore"
	"github.com/jpconstantineau/dupectl/pkg/formatter"
	"github.com/spf13/cobra"
)

var (
	getDuplicatesJSON     bool
	getDuplicatesMinCount int
	getDuplicatesRoot     string
	getDuplicatesMinSize  int64
	getDuplicatesSort     string
)

// getDuplicatesCmd represents the getDuplicates command
var getDuplicatesCmd = &cobra.Command{
	Use:   "duplicates",
	Short: "Query and display duplicate files",
	Long: `Query duplicate files identified during scans.

Display duplicate file sets with optional filtering by count, size, or root folder.
Output can be formatted as human-readable table (default) or JSON.

Example:
  dupectl get duplicates
  dupectl get duplicates --min-count 5
  dupectl get duplicates --min-size 1048576 --json
  dupectl get duplicates --sort count`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Build filter
		filter := datastore.DuplicateFilter{
			MinCount:  getDuplicatesMinCount,
			RootPath:  getDuplicatesRoot,
			MinSize:   getDuplicatesMinSize,
			SortField: getDuplicatesSort,
		}

		// Query duplicates
		sets, err := datastore.GetDuplicateFiles(filter)
		if err != nil {
			return fmt.Errorf("failed to query duplicates: %w", err)
		}

		// Get statistics
		totalSets, totalFiles, wastedBytes, recoverableBytes, err := datastore.GetDuplicateStats()
		if err != nil {
			return fmt.Errorf("failed to get duplicate stats: %w", err)
		}

		// Format output
		if getDuplicatesJSON {
			output, err := formatter.FormatDuplicatesJSON(sets, totalSets, totalFiles, wastedBytes, recoverableBytes)
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(output)
		} else {
			output := formatter.FormatDuplicatesTable(sets, totalSets, totalFiles, wastedBytes, recoverableBytes)
			fmt.Print(output)
		}

		return nil
	},
}

func init() {
	getCmd.AddCommand(getDuplicatesCmd)

	getDuplicatesCmd.Flags().BoolVar(&getDuplicatesJSON, "json", false, "Output in JSON format")
	getDuplicatesCmd.Flags().IntVar(&getDuplicatesMinCount, "min-count", 2, "Minimum number of duplicates")
	getDuplicatesCmd.Flags().StringVar(&getDuplicatesRoot, "root", "", "Filter by root folder path")
	getDuplicatesCmd.Flags().Int64Var(&getDuplicatesMinSize, "min-size", 0, "Minimum file size in bytes")
	getDuplicatesCmd.Flags().StringVar(&getDuplicatesSort, "sort", "size", "Sort by: size, count, or path")
}
