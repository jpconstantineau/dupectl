package duplicate

import (
	"database/sql"
	"fmt"

	"github.com/jpconstantineau/dupectl/pkg/datastore"
	"github.com/jpconstantineau/dupectl/pkg/logger"
)

// DuplicateSet represents a group of duplicate files
type DuplicateSet struct {
	Hash  string
	Size  int64
	Files []*datastore.File
}

// Detector finds duplicate files
type Detector struct {
	db *sql.DB
}

// NewDetector creates a duplicate detector
func NewDetector(db *sql.DB) *Detector {
	return &Detector{db: db}
}

// FindDuplicateFiles finds all duplicate file sets
func (d *Detector) FindDuplicateFiles(minCount int, minSize int64) ([]*DuplicateSet, error) {
	// Query for hashes that appear more than once
	query := `
	SELECT hash_value, size, COUNT(*) as count
	FROM files
	WHERE hash_value IS NOT NULL 
	  AND removed = 0 
	  AND error_status IS NULL
	  AND size > 0
	  AND size >= ?
	GROUP BY hash_value, size
	HAVING count >= ?
	ORDER BY size DESC, hash_value
	`

	rows, err := d.db.Query(query, minSize, minCount)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var duplicateSets []*DuplicateSet
	for rows.Next() {
		var hash string
		var size int64
		var count int

		if err := rows.Scan(&hash, &size, &count); err != nil {
			return nil, err
		}

		// Get all files with this hash
		files, err := datastore.GetFilesByHash(d.db, hash, size)
		if err != nil {
			logger.Warn("Failed to get files for hash %s: %v", hash, err)
			continue
		}

		duplicateSets = append(duplicateSets, &DuplicateSet{
			Hash:  hash,
			Size:  size,
			Files: files,
		})
	}

	return duplicateSets, rows.Err()
}

// CountDuplicates returns total duplicate count statistics
func (d *Detector) CountDuplicates() (sets, files int, err error) {
	// Count duplicate sets
	err = d.db.QueryRow(`
		SELECT COUNT(*) 
		FROM (
			SELECT hash_value
			FROM files
			WHERE hash_value IS NOT NULL AND removed = 0 AND error_status IS NULL AND size > 0
			GROUP BY hash_value, size
			HAVING COUNT(*) >= 2
		)
	`).Scan(&sets)
	if err != nil {
		return 0, 0, err
	}

	// Count duplicate files
	err = d.db.QueryRow(`
		SELECT COUNT(*)
		FROM files f
		WHERE hash_value IS NOT NULL 
		  AND removed = 0 
		  AND error_status IS NULL
		  AND size > 0
		  AND EXISTS (
			SELECT 1 FROM files f2
			WHERE f2.hash_value = f.hash_value 
			  AND f2.size = f.size
			  AND f2.id != f.id
			  AND f2.removed = 0
			  AND f2.error_status IS NULL
			  AND f2.size > 0
		  )
	`).Scan(&files)

	return sets, files, err
}
