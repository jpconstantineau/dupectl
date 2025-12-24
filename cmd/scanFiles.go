package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jpconstantineau/dupectl/pkg/datastore"
	"github.com/jpconstantineau/dupectl/pkg/entities"
	"github.com/jpconstantineau/dupectl/pkg/scanner"
	"github.com/spf13/cobra"
)

var scanFilesVerbose bool

// scanFilesCmd represents the scanFiles command
var scanFilesCmd = &cobra.Command{
	Use:   "files <root-path>",
	Short: "Hash files in previously scanned folders",
	Long: `Hash files in folders that have already been scanned.

This is useful after running 'scan folders' to defer file hashing, or to re-hash
files after they have been modified.

Example:
  dupectl scan files /data/archive
  dupectl scan files /backup --verbose`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rootPath := args[0]

		// Convert to absolute path
		absPath, err := filepath.Abs(rootPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}

		// Verify path exists
		info, err := os.Stat(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("path does not exist: %s", absPath)
			}
			return fmt.Errorf("failed to access path: %w", err)
		}

		if !info.IsDir() {
			return fmt.Errorf("path is not a directory: %s", absPath)
		}

		// Ensure root folder exists in database
		rootFolderID, err := datastore.EnsureRootFolder(absPath, info.Name())
		if err != nil {
			return fmt.Errorf("failed to ensure root folder: %w", err)
		}

		// Create scanner
		s, err := scanner.NewFileSystemScanner(scanFilesVerbose)
		if err != nil {
			return fmt.Errorf("failed to create scanner: %w", err)
		}

		// Setup signal handler
		ctx := scanner.SetupSignalHandler()

		// Perform scan
		fmt.Printf("Scanning files: %s\n", absPath)
		err = s.Scan(ctx, absPath, rootFolderID, entities.ScanModeFiles)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		return nil
	},
}

func init() {
	scanCmd.AddCommand(scanFilesCmd)
	scanFilesCmd.Flags().BoolVarP(&scanFilesVerbose, "verbose", "v", false, "Enable verbose logging")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scanFilesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// scanFilesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
