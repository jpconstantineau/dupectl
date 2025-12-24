package datastore

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jpconstantineau/dupectl/pkg/entities"
)

// DuplicateFilter contains filtering options for duplicate queries
type DuplicateFilter struct {
	MinCount  int
	RootPath  string
	MinSize   int64
	SortField string
}

// GetDuplicateFiles retrieves duplicate file sets with optional filtering
func GetDuplicateFiles(filter DuplicateFilter) ([]entities.DuplicateFileSet, error) {
	db, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Build query with filters
	query := `
		WITH duplicate_groups AS (
			SELECT size, hash_value, hash_algorithm, COUNT(*) as file_count
			FROM files
			WHERE hash_value IS NOT NULL
	`

	args := []interface{}{}

	// Add min size filter
	if filter.MinSize > 0 {
		query += " AND size >= ?"
		args = append(args, filter.MinSize)
	}

	query += `
			GROUP BY size, hash_value, hash_algorithm
			HAVING file_count >= ?
		)
		SELECT dg.size, dg.hash_value, dg.hash_algorithm, dg.file_count,
		       f.id, f.root_folder_id, f.folder_id, f.path, f.name, f.size, f.mtime, 
		       f.hash_value, f.hash_algorithm, f.error_status, f.scanned_at
		FROM duplicate_groups dg
		INNER JOIN files f ON f.size = dg.size AND f.hash_value = dg.hash_value
	`

	args = append(args, filter.MinCount)

	// Add root path filter if specified
	if filter.RootPath != "" {
		query += " WHERE f.path LIKE ?"
		args = append(args, filter.RootPath+"%")
	}

	// Add sorting
	switch filter.SortField {
	case "size":
		query += " ORDER BY dg.size DESC, dg.hash_value"
	case "count":
		query += " ORDER BY dg.file_count DESC, dg.size DESC"
	case "path":
		query += " ORDER BY f.path"
	default:
		query += " ORDER BY dg.size DESC, dg.hash_value"
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query duplicates: %w", err)
	}
	defer rows.Close()

	// Group files by duplicate set
	duplicateSets := make(map[string]*entities.DuplicateFileSet)

	for rows.Next() {
		var size, fileCount int64
		var hashValue, hashAlgorithm string
		var file entities.File
		var mtime, scannedAt *int64

		err = rows.Scan(
			&size, &hashValue, &hashAlgorithm, &fileCount,
			&file.ID, &file.RootFolderID, &file.FolderID, &file.Path, &file.Name,
			&file.Size, &mtime, &file.HashValue, &file.HashAlgorithm,
			&file.ErrorStatus, &scannedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan duplicate row: %w", err)
		}

		// Convert timestamps
		if mtime != nil {
			t := time.Unix(*mtime, 0)
			file.Mtime = t
		}
		if scannedAt != nil {
			t := time.Unix(*scannedAt, 0)
			file.ScannedAt = &t
		}

		// Group by hash
		key := hashValue + hashAlgorithm
		if _, exists := duplicateSets[key]; !exists {
			duplicateSets[key] = &entities.DuplicateFileSet{
				Size:          size,
				HashValue:     hashValue,
				HashAlgorithm: hashAlgorithm,
				FileCount:     int(fileCount),
				Files:         []entities.File{},
			}
		}

		duplicateSets[key].Files = append(duplicateSets[key].Files, file)
	}

	// Convert map to slice
	result := make([]entities.DuplicateFileSet, 0, len(duplicateSets))
	for _, set := range duplicateSets {
		result = append(result, *set)
	}

	return result, nil
}

// GetDuplicateStats returns summary statistics about duplicates
func GetDuplicateStats() (totalSets int, totalFiles int, wastedBytes int64, recoverableBytes int64, err error) {
	db, err := GetDB()
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	query := `
		WITH duplicate_groups AS (
			SELECT size, hash_value, COUNT(*) as file_count
			FROM files
			WHERE hash_value IS NOT NULL
			GROUP BY size, hash_value
			HAVING file_count >= 2
		)
		SELECT 
			COUNT(*) as total_sets,
			SUM(file_count) as total_files,
			SUM(size * file_count) as total_wasted,
			SUM(size * (file_count - 1)) as recoverable
		FROM duplicate_groups
	`

	err = db.QueryRow(query).Scan(&totalSets, &totalFiles, &wastedBytes, &recoverableBytes)
	if err == sql.ErrNoRows {
		return 0, 0, 0, 0, nil
	}
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to query duplicate stats: %w", err)
	}

	return totalSets, totalFiles, wastedBytes, recoverableBytes, nil
}
