package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jpconstantineau/dupectl/pkg/datastore"
	"github.com/jpconstantineau/dupectl/pkg/entities"
	"github.com/jpconstantineau/dupectl/pkg/scanner"
	"github.com/spf13/cobra"
)

var (
	scanAllVerbose bool
	scanAllRescan  bool
)

// scanAllCmd represents the scanAll command
var scanAllCmd = &cobra.Command{
	Use:   "all <root-path>",
	Short: "Scan all files and folders completely",
	Long: `Perform complete scan of folder structure and file hashing for duplicate detection.

This command will:
  1. Recursively traverse all folders under the root path
  2. Calculate cryptographic hash for each file
  3. Store results in database for later analysis
  4. Support resumption from checkpoint if interrupted

Example:
  dupectl scan all /home/user/documents
  dupectl scan all ./my-data --verbose
  dupectl scan all /backup --rescan`,
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

		// Check for --rescan flag
		if scanAllRescan {
			err = datastore.DeleteScanState(rootFolderID)
			if err != nil {
				return fmt.Errorf("failed to delete existing scan state: %w", err)
			}
			if scanAllVerbose {
				fmt.Println("Deleted existing scan state, starting fresh scan")
			}
		}

		// Create scanner
		s, err := scanner.NewFileSystemScanner(scanAllVerbose)
		if err != nil {
			return fmt.Errorf("failed to create scanner: %w", err)
		}

		// Setup signal handler for graceful shutdown
		ctx := scanner.SetupSignalHandler()

		// Perform scan
		fmt.Printf("Scanning: %s\n", absPath)
		err = s.Scan(ctx, absPath, rootFolderID, entities.ScanModeAll)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		return nil
	},
}

func init() {
	scanCmd.AddCommand(scanAllCmd)

	scanAllCmd.Flags().BoolVarP(&scanAllVerbose, "verbose", "v", false, "Enable verbose logging")
	scanAllCmd.Flags().BoolVar(&scanAllRescan, "rescan", false, "Force restart scan from beginning")
}

// promptForConfirmation prompts the user for yes/no confirmation
func promptForConfirmation(message string) bool {
	fmt.Printf("%s (y/n): ", message)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
