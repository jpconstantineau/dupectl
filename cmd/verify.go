package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var verifyRepair bool
var verifyJSON bool

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify all <root-folder-path>",
	Short: "Check database consistency and detect data integrity issues",
	Long: `Check database consistency and detect data integrity issues.

This operation validates foreign key integrity, timestamp validity, removed flag cascade,
statistics accuracy, and hash algorithm consistency. Use --repair to automatically fix safe issues.

Examples:
  dupectl verify all /path/to/root
  dupectl verify all /path/to/root --repair
  dupectl verify all /path/to/root --json`,
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
		db, err := openDatabase()
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

		// Run verification
		result := runVerification(db, rootID, absPath, verifyRepair)

		// Output results
		if verifyJSON {
			outputJSON(result)
		} else {
			outputTable(result)
		}

		// Exit with error code if issues found
		if result.IssuesFound > 0 && !verifyRepair {
			os.Exit(1)
		}
	},
}

type VerifyResult struct {
	RootPath      string        `json:"root_path"`
	ChecksRun     int           `json:"checks_run"`
	IssuesFound   int           `json:"issues_found"`
	IssuesFixed   int           `json:"issues_fixed"`
	Checks        []CheckResult `json:"checks"`
	ScanTimestamp string        `json:"scan_timestamp"`
}

type CheckResult struct {
	Name   string        `json:"name"`
	Status string        `json:"status"` // pass, warning, error
	Issues []IssueDetail `json:"issues"`
	Fixed  int           `json:"fixed,omitempty"`
}

type IssueDetail struct {
	Description string `json:"description"`
	Severity    string `json:"severity"` // warning, error
}

func runVerification(db *sql.DB, rootID int64, rootPath string, repair bool) VerifyResult {
	result := VerifyResult{
		RootPath:      rootPath,
		ChecksRun:     5,
		Checks:        make([]CheckResult, 0),
		ScanTimestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Check 1: Foreign key integrity
	check1 := checkForeignKeyIntegrity(db, rootID, repair)
	result.Checks = append(result.Checks, check1)
	result.IssuesFound += len(check1.Issues)
	result.IssuesFixed += check1.Fixed

	// Check 2: Timestamp validity
	check2 := checkTimestampValidity(db, rootID, repair)
	result.Checks = append(result.Checks, check2)
	result.IssuesFound += len(check2.Issues)
	result.IssuesFixed += check2.Fixed

	// Check 3: Removed flag cascade
	check3 := checkRemovedFlagCascade(db, rootID, repair)
	result.Checks = append(result.Checks, check3)
	result.IssuesFound += len(check3.Issues)
	result.IssuesFixed += check3.Fixed

	// Check 4: Statistics accuracy
	check4 := checkStatisticsAccuracy(db, rootID, repair)
	result.Checks = append(result.Checks, check4)
	result.IssuesFound += len(check4.Issues)
	result.IssuesFixed += check4.Fixed

	// Check 5: Hash algorithm consistency
	check5 := checkHashAlgorithmConsistency(db, rootID, repair)
	result.Checks = append(result.Checks, check5)
	result.IssuesFound += len(check5.Issues)
	result.IssuesFixed += check5.Fixed

	return result
}

func checkForeignKeyIntegrity(db *sql.DB, rootID int64, repair bool) CheckResult {
	check := CheckResult{
		Name:   "foreign_key_integrity",
		Status: "pass",
		Issues: make([]IssueDetail, 0),
	}

	// Check for orphaned files (folder_id references non-existent folder)
	var orphanedFiles int
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM files f 
		LEFT JOIN folders fo ON f.folder_id = fo.id 
		WHERE f.root_folder_id = ? AND fo.id IS NULL
	`, rootID).Scan(&orphanedFiles)

	if err == nil && orphanedFiles > 0 {
		check.Status = "error"
		check.Issues = append(check.Issues, IssueDetail{
			Description: fmt.Sprintf("%d files reference non-existent folders", orphanedFiles),
			Severity:    "error",
		})

		if repair {
			// Delete orphaned files
			result, err := db.Exec("DELETE FROM files WHERE root_folder_id = ? AND folder_id NOT IN (SELECT id FROM folders)", rootID)
			if err == nil {
				rows, _ := result.RowsAffected()
				check.Fixed = int(rows)
			}
		}
	}

	// Check for orphaned folders (parent_id references non-existent folder)
	var orphanedFolders int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM folders f 
		LEFT JOIN folders p ON f.parent_id = p.id 
		WHERE f.root_folder_id = ? AND f.parent_id IS NOT NULL AND p.id IS NULL
	`, rootID).Scan(&orphanedFolders)

	if err == nil && orphanedFolders > 0 {
		check.Status = "error"
		check.Issues = append(check.Issues, IssueDetail{
			Description: fmt.Sprintf("%d folders reference non-existent parent folders", orphanedFolders),
			Severity:    "error",
		})

		if repair {
			// Set parent_id to NULL for orphaned folders
			result, err := db.Exec(`
				UPDATE folders 
				SET parent_id = NULL 
				WHERE root_folder_id = ? 
				AND parent_id IS NOT NULL 
				AND parent_id NOT IN (SELECT id FROM folders)
			`, rootID)
			if err == nil {
				rows, _ := result.RowsAffected()
				check.Fixed += int(rows)
			}
		}
	}

	return check
}

func checkTimestampValidity(db *sql.DB, rootID int64, repair bool) CheckResult {
	check := CheckResult{
		Name:   "timestamp_validity",
		Status: "pass",
		Issues: make([]IssueDetail, 0),
	}

	now := time.Now()

	// Check for future timestamps in files
	var futureDates int
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM files 
		WHERE root_folder_id = ? AND (
			datetime(created_at) > datetime(?) OR 
			datetime(updated_at) > datetime(?) OR
			datetime(scanned_at) > datetime(?)
		)
	`, rootID, now, now, now).Scan(&futureDates)

	if err == nil && futureDates > 0 {
		check.Status = "warning"
		check.Issues = append(check.Issues, IssueDetail{
			Description: fmt.Sprintf("%d files have future timestamps", futureDates),
			Severity:    "warning",
		})

		if repair {
			// Reset future timestamps to current time
			result, err := db.Exec(`
				UPDATE files 
				SET created_at = CASE WHEN datetime(created_at) > datetime(?) THEN ? ELSE created_at END,
					updated_at = CASE WHEN datetime(updated_at) > datetime(?) THEN ? ELSE updated_at END,
					scanned_at = CASE WHEN datetime(scanned_at) > datetime(?) THEN ? ELSE scanned_at END
				WHERE root_folder_id = ?
			`, now, now, now, now, now, now, rootID)
			if err == nil {
				rows, _ := result.RowsAffected()
				check.Fixed = int(rows)
			}
		}
	}

	// Check for inconsistent timestamps (created_at > updated_at)
	var inconsistentTimestamps int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM files 
		WHERE root_folder_id = ? AND datetime(created_at) > datetime(updated_at)
	`, rootID).Scan(&inconsistentTimestamps)

	if err == nil && inconsistentTimestamps > 0 {
		check.Status = "warning"
		check.Issues = append(check.Issues, IssueDetail{
			Description: fmt.Sprintf("%d files have created_at > updated_at", inconsistentTimestamps),
			Severity:    "warning",
		})

		if repair {
			// Set updated_at = created_at for inconsistent records
			result, err := db.Exec(`
				UPDATE files 
				SET updated_at = created_at 
				WHERE root_folder_id = ? AND datetime(created_at) > datetime(updated_at)
			`, rootID)
			if err == nil {
				rows, _ := result.RowsAffected()
				check.Fixed += int(rows)
			}
		}
	}

	return check
}

func checkRemovedFlagCascade(db *sql.DB, rootID int64, repair bool) CheckResult {
	check := CheckResult{
		Name:   "removed_flag_cascade",
		Status: "pass",
		Issues: make([]IssueDetail, 0),
	}

	// Check for files not marked removed when parent folder is removed
	rows, err := db.Query(`
		SELECT f.path, fo.path
		FROM files f
		JOIN folders fo ON f.folder_id = fo.id
		WHERE f.root_folder_id = ? AND fo.removed = 1 AND f.removed = 0
	`, rootID)

	if err == nil {
		defer rows.Close()
		var filePath, folderPath string
		count := 0

		for rows.Next() {
			err := rows.Scan(&filePath, &folderPath)
			if err == nil {
				count++
				if count <= 3 {
					check.Issues = append(check.Issues, IssueDetail{
						Description: fmt.Sprintf("Folder %s removed but child file %s not marked removed", folderPath, filePath),
						Severity:    "error",
					})
				}
			}
		}

		if count > 0 {
			check.Status = "error"
			if count > 3 {
				check.Issues = append(check.Issues, IssueDetail{
					Description: fmt.Sprintf("...and %d more inconsistencies", count-3),
					Severity:    "error",
				})
			}

			if repair {
				// Mark files as removed when parent folder is removed
				result, err := db.Exec(`
					UPDATE files 
					SET removed = 1, removed_at = CURRENT_TIMESTAMP
					WHERE root_folder_id = ? 
					AND folder_id IN (SELECT id FROM folders WHERE removed = 1)
					AND removed = 0
				`, rootID)
				if err == nil {
					rows, _ := result.RowsAffected()
					check.Fixed = int(rows)
				}
			}
		}
	}

	return check
}

func checkStatisticsAccuracy(db *sql.DB, rootID int64, repair bool) CheckResult {
	check := CheckResult{
		Name:   "statistics_accuracy",
		Status: "pass",
		Issues: make([]IssueDetail, 0),
	}

	// Get current statistics from root folder
	var storedFolders, storedFiles int64
	err := db.QueryRow(`
		SELECT folder_count, file_count 
		FROM rootfolders 
		WHERE id = ?
	`, rootID).Scan(&storedFolders, &storedFiles)

	if err != nil {
		return check
	}

	// Count actual folders and files
	var actualFolders, actualFiles int64
	db.QueryRow("SELECT COUNT(*) FROM folders WHERE root_folder_id = ?", rootID).Scan(&actualFolders)
	db.QueryRow("SELECT COUNT(*) FROM files WHERE root_folder_id = ?", rootID).Scan(&actualFiles)

	if storedFolders != actualFolders || storedFiles != actualFiles {
		check.Status = "warning"
		check.Issues = append(check.Issues, IssueDetail{
			Description: fmt.Sprintf("Statistics mismatch - Expected: %d folders, %d files; Actual: %d folders, %d files",
				storedFolders, storedFiles, actualFolders, actualFiles),
			Severity: "warning",
		})

		if repair {
			// Update statistics
			_, err := db.Exec(`
				UPDATE rootfolders 
				SET folder_count = (SELECT COUNT(*) FROM folders WHERE root_folder_id = ?),
					file_count = (SELECT COUNT(*) FROM files WHERE root_folder_id = ?)
				WHERE id = ?
			`, rootID, rootID, rootID)
			if err == nil {
				check.Fixed = 1
			}
		}
	}

	return check
}

func checkHashAlgorithmConsistency(db *sql.DB, rootID int64, repair bool) CheckResult {
	check := CheckResult{
		Name:   "hash_algorithm_consistency",
		Status: "pass",
		Issues: make([]IssueDetail, 0),
	}

	// Check if all files have the same hash algorithm
	rows, err := db.Query(`
		SELECT DISTINCT hash_algorithm, COUNT(*) as count
		FROM files 
		WHERE root_folder_id = ? AND hash_algorithm IS NOT NULL
		GROUP BY hash_algorithm
	`, rootID)

	if err == nil {
		defer rows.Close()
		algorithms := make(map[string]int)

		for rows.Next() {
			var algo string
			var count int
			if rows.Scan(&algo, &count) == nil {
				algorithms[algo] = count
			}
		}

		if len(algorithms) > 1 {
			check.Status = "warning"
			desc := fmt.Sprintf("Multiple hash algorithms detected:")
			for algo, count := range algorithms {
				desc += fmt.Sprintf(" %s (%d files)", algo, count)
			}
			check.Issues = append(check.Issues, IssueDetail{
				Description: desc,
				Severity:    "warning",
			})
			// Note: Cannot auto-repair this - requires re-scanning
		}
	}

	return check
}

func outputTable(result VerifyResult) {
	if verifyRepair {
		fmt.Printf("Verifying database consistency for: %s\n", result.RootPath)
		fmt.Println("Running consistency checks...")
	} else {
		fmt.Printf("Verifying database consistency for: %s\n\n", result.RootPath)
		fmt.Println("Check Results:")
		fmt.Println("═════════════")
	}

	for _, check := range result.Checks {
		displayName := formatCheckName(check.Name)

		if check.Status == "pass" {
			if !verifyRepair {
				fmt.Printf("✓ %s: PASS\n", displayName)
			}
		} else if check.Status == "warning" {
			fmt.Printf("⚠ %s: WARNING", displayName)
			if len(check.Issues) > 0 {
				fmt.Printf(" - %d issue(s) found\n", len(check.Issues))
			} else {
				fmt.Println()
			}
			for _, issue := range check.Issues {
				fmt.Printf("  - %s\n", issue.Description)
			}
			if verifyRepair && check.Fixed > 0 {
				fmt.Printf("  Fixing: %d issue(s) fixed ✓\n", check.Fixed)
			}
		} else if check.Status == "error" {
			fmt.Printf("✗ %s: %d inconsistenc", displayName, len(check.Issues))
			if len(check.Issues) == 1 {
				fmt.Print("y found\n")
			} else {
				fmt.Print("ies found\n")
			}
			for _, issue := range check.Issues {
				fmt.Printf("  - %s\n", issue.Description)
			}
			if verifyRepair && check.Fixed > 0 {
				fmt.Printf("  Fixing: %d issue(s) fixed ✓\n", check.Fixed)
			}
		}
	}

	fmt.Println()
	if verifyRepair {
		fmt.Printf("Summary: %d checks run, %d issues found, %d issues fixed\n",
			result.ChecksRun, result.IssuesFound, result.IssuesFixed)
	} else {
		errorCount := 0
		warningCount := 0
		for _, check := range result.Checks {
			if check.Status == "error" {
				errorCount += len(check.Issues)
			} else if check.Status == "warning" {
				warningCount += len(check.Issues)
			}
		}

		fmt.Printf("Summary: %d checks run, %d issues found", result.ChecksRun, result.IssuesFound)
		if errorCount > 0 || warningCount > 0 {
			fmt.Printf(" (%d ERROR, %d WARNING)\n", errorCount, warningCount)
		} else {
			fmt.Println()
		}

		if result.IssuesFound > 0 {
			fmt.Println("Run with --repair to attempt automatic fixes.")
		}
	}
}

func outputJSON(result VerifyResult) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func formatCheckName(name string) string {
	switch name {
	case "foreign_key_integrity":
		return "Foreign key integrity"
	case "timestamp_validity":
		return "Timestamp validity"
	case "removed_flag_cascade":
		return "Removed flag cascade"
	case "statistics_accuracy":
		return "Statistics accuracy"
	case "hash_algorithm_consistency":
		return "Hash algorithm consistency"
	default:
		return name
	}
}

func openDatabase() (*sql.DB, error) {
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
	rootCmd.AddCommand(verifyCmd)

	verifyCmd.Flags().BoolVar(&verifyRepair, "repair", false, "Attempt automatic fixes for safe issues")
	verifyCmd.Flags().BoolVar(&verifyJSON, "json", false, "Output results in JSON format")
}
