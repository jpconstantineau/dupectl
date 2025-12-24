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

var scanFoldersVerbose bool

// scanFoldersCmd represents the scanFolders command
var scanFoldersCmd = &cobra.Command{
	Use:   "folders <root-path>",
	Short: "Scan folder structure only (no file hashing)",
	Long: `Perform quick folder structure mapping without file hashing.

This is useful for very large directory trees where you want to map the structure first, 
then hash files later during off-hours.

Example:
  dupectl scan folders /large/archive
  dupectl scan folders /backup --verbose`,
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
		s, err := scanner.NewFileSystemScanner(scanFoldersVerbose)
		if err != nil {
			return fmt.Errorf("failed to create scanner: %w", err)
		}

		// Setup signal handler
		ctx := scanner.SetupSignalHandler()

		// Perform scan
		fmt.Printf("Scanning folders: %s\n", absPath)
		err = s.Scan(ctx, absPath, rootFolderID, entities.ScanModeFolders)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		fmt.Println("\nNote: Run 'dupectl scan files <root-path>' to hash file contents")

		return nil
	},
}

func init() {
	scanCmd.AddCommand(scanFoldersCmd)
	scanFoldersCmd.Flags().BoolVarP(&scanFoldersVerbose, "verbose", "v", false, "Enable verbose logging")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scanFoldersCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// scanFoldersCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
