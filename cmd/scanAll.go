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
	"github.com/jpconstantineau/dupectl/pkg/checkpoint"
	"github.com/jpconstantineau/dupectl/pkg/datastore"
	"github.com/jpconstantineau/dupectl/pkg/duplicate"
	"github.com/jpconstantineau/dupectl/pkg/logger"
	"github.com/jpconstantineau/dupectl/pkg/pathutil"
	"github.com/jpconstantineau/dupectl/pkg/scanner"
	"github.com/spf13/cobra"

	_ "modernc.org/sqlite"
)

var (
	scanAllProgress bool
	scanAllRestart  bool
)

// scanAllCmd represents the scanAll command
var scanAllCmd = &cobra.Command{
	Use:   "all <root-folder-path>",
	Short: "Scan all files and folders recursively",
	Long: `Scan all files and folders under the specified root folder.
Calculates cryptographic hashes for duplicate detection.

Supports checkpoint/resume: if interrupted, the scan will automatically resume
from where it left off. Use --restart to start fresh.

Examples:
  dupectl scan all /home/user/documents --progress
  dupectl scan all "C:\Users\user\Documents" --restart
  dupectl scan all ../relative/path`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runScanAll(args[0])
	},
}

func init() {
	scanCmd.AddCommand(scanAllCmd)

	scanAllCmd.Flags().BoolVar(&scanAllProgress, "progress", false, "Display real-time progress")
	scanAllCmd.Flags().BoolVar(&scanAllRestart, "restart", false, "Restart scan from beginning")
}

func runScanAll(rootFolderPath string) {
	// Convert to absolute path
	absPath, err := pathutil.ToAbsolute(rootFolderPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid path: %v\n", err)
		os.Exit(1)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Root folder does not exist: %s\n", absPath)
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to load configuration: %v\n", err)
		os.Exit(2)
	}

	// Initialize database
	db, err := sql.Open("sqlite", cfg.DatabasePath+"?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to open database: %v\n", err)
		os.Exit(2)
	}
	defer db.Close()

	// Run migrations
	if err := datastore.RunMigrations(db); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to run migrations: %v\n", err)
		os.Exit(2)
	}

	// Check if root folder is registered
	rootFolder, err := getRootFolderByPath(db, absPath)
	if err != nil {
		// Root not registered, prompt user
		fmt.Printf("Root folder not registered. Register now? (y/n): ")
		var response string
		fmt.Scanln(&response)

		if response != "y" && response != "Y" {
			fmt.Println("Scan cancelled.")
			os.Exit(0)
		}

		// Register root folder
		rootFolder, err = registerRootFolder(db, absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to register root folder: %v\n", err)
			os.Exit(2)
		}
		fmt.Printf("Root folder registered with ID: %d\n", rootFolder.ID)
	}

	// Setup signal handling for graceful shutdown
	ctx, cancel := checkpoint.SetupSignalHandler(func() {
		logger.Info("Saving checkpoint before exit...")
	})
	defer cancel()

	// Create scanner
	scannerCfg := &scanner.Config{
		RootFolderID:     int64(rootFolder.ID),
		RootPath:         absPath,
		ScanMode:         "all",
		HashAlgorithm:    cfg.HashAlgorithm,
		WorkerCount:      cfg.WorkerCount,
		ShowProgress:     scanAllProgress,
		ProgressInterval: time.Duration(cfg.ProgressInterval) * time.Second,
		TraverseLinks:    false, // Default to not following symlinks
	}

	s, err := scanner.NewScanner(db, scannerCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create scanner: %v\n", err)
		os.Exit(2)
	}

	// Start scan
	fmt.Printf("Scanning root folder: %s\n", absPath)
	if err := s.Scan(ctx, scanAllRestart); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Scan failed: %v\n", err)
		os.Exit(2)
	}

	// Display summary
	folders, files, duration := s.GetSummary()
	fmt.Printf("\nScan completed in %s\n", duration)
	fmt.Printf("Folders scanned: %d\n", folders)
	fmt.Printf("Files scanned: %d\n", files)

	// Count duplicates
	detector := duplicate.NewDetector(db)
	dupSets, dupFiles, err := detector.CountDuplicates()
	if err != nil {
		logger.Warn("Failed to count duplicates: %v", err)
		dupSets, dupFiles = 0, 0
	}
	fmt.Printf("Duplicates found: %d files in %d sets\n", dupFiles, dupSets)
}

// Temporary helpers - these should be moved to proper datastore functions

type RootFolder struct {
	ID   int
	Path string
}

func getRootFolderByPath(db *sql.DB, path string) (*RootFolder, error) {
	var rf RootFolder
	err := db.QueryRow("SELECT id, path FROM root_folders WHERE path = ?", path).Scan(&rf.ID, &rf.Path)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("root folder not registered")
	}
	if err != nil {
		return nil, err
	}
	return &rf, nil
}

func registerRootFolder(db *sql.DB, path string) (*RootFolder, error) {
	// Simple registration - in real implementation this would check for agent, host, etc.
	result, err := db.Exec("INSERT INTO root_folders (path, traverse_links) VALUES (?, 0)", path)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &RootFolder{ID: int(id), Path: path}, nil
}
