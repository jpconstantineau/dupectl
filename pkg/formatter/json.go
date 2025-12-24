package formatter

import (
	"encoding/json"

	"github.com/jpconstantineau/dupectl/pkg/entities"
)

// DuplicatesJSON represents the JSON output structure
type DuplicatesJSON struct {
	DuplicateSets []DuplicateSetJSON `json:"duplicate_sets"`
	Summary       SummaryJSON        `json:"summary"`
}

// DuplicateSetJSON represents a duplicate set in JSON format
type DuplicateSetJSON struct {
	Size          int64      `json:"size"`
	HashValue     string     `json:"hash_value"`
	HashAlgorithm string     `json:"hash_algorithm"`
	FileCount     int        `json:"file_count"`
	Files         []FileJSON `json:"files"`
}

// FileJSON represents a file in JSON format
type FileJSON struct {
	ID        int    `json:"id"`
	Path      string `json:"path"`
	Name      string `json:"name"`
	Mtime     string `json:"mtime"`
	ScannedAt string `json:"scanned_at,omitempty"`
}

// SummaryJSON represents summary statistics in JSON format
type SummaryJSON struct {
	TotalSets        int   `json:"total_sets"`
	TotalFiles       int   `json:"total_files"`
	TotalWastedBytes int64 `json:"total_wasted_bytes"`
	RecoverableBytes int64 `json:"recoverable_bytes"`
}

// FormatDuplicatesJSON formats duplicate file sets as JSON
func FormatDuplicatesJSON(sets []entities.DuplicateFileSet, totalSets, totalFiles int, wastedBytes, recoverableBytes int64) (string, error) {
	output := DuplicatesJSON{
		DuplicateSets: make([]DuplicateSetJSON, 0, len(sets)),
		Summary: SummaryJSON{
			TotalSets:        totalSets,
			TotalFiles:       totalFiles,
			TotalWastedBytes: wastedBytes,
			RecoverableBytes: recoverableBytes,
		},
	}

	for _, set := range sets {
		jsonSet := DuplicateSetJSON{
			Size:          set.Size,
			HashValue:     set.HashValue,
			HashAlgorithm: set.HashAlgorithm,
			FileCount:     set.FileCount,
			Files:         make([]FileJSON, 0, len(set.Files)),
		}

		for _, file := range set.Files {
			jsonFile := FileJSON{
				ID:    file.ID,
				Path:  file.Path,
				Name:  file.Name,
				Mtime: file.Mtime.Format("2006-01-02T15:04:05Z"),
			}

			if file.ScannedAt != nil {
				jsonFile.ScannedAt = file.ScannedAt.Format("2006-01-02T15:04:05Z")
			}

			jsonSet.Files = append(jsonSet.Files, jsonFile)
		}

		output.DuplicateSets = append(output.DuplicateSets, jsonSet)
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}
