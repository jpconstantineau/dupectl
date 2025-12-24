package duplicate

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Formatter formats duplicate results for display
type Formatter struct{}

// NewFormatter creates a result formatter
func NewFormatter() *Formatter {
	return &Formatter{}
}

// FormatTable formats duplicates as a table
func (f *Formatter) FormatTable(sets []*DuplicateSet) string {
	if len(sets) == 0 {
		return "No duplicates found.\n"
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Found %d duplicate sets:\n\n", len(sets)))

	for i, set := range sets {
		sb.WriteString(fmt.Sprintf("Set %d: %d files, %s each (hash: %s...)\n",
			i+1, len(set.Files), formatSize(set.Size), set.Hash[:16]))

		for _, file := range set.Files {
			sb.WriteString(fmt.Sprintf("  - %s\n", file.Path))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatSummary formats duplicates as a summary table grouped by root folder
func (f *Formatter) FormatSummary(sets []*DuplicateSet) string {
	if len(sets) == 0 {
		return "No duplicates found.\n"
	}

	// Group duplicates by root folder
	type RootSummary struct {
		Path           string
		DuplicateSets  int
		DuplicateFiles int
		TotalSize      int64
	}

	rootMap := make(map[string]*RootSummary)

	for _, set := range sets {
		for _, file := range set.Files {
			root := file.RootFolderPath
			if root == "" {
				root = "(unknown)"
			}

			if _, exists := rootMap[root]; !exists {
				rootMap[root] = &RootSummary{
					Path: root,
				}
			}

			rootMap[root].DuplicateFiles++
			rootMap[root].TotalSize += file.Size
		}
	}

	// Count sets per root (a set may span multiple roots)
	for _, set := range sets {
		rootsInSet := make(map[string]bool)
		for _, file := range set.Files {
			root := file.RootFolderPath
			if root == "" {
				root = "(unknown)"
			}
			rootsInSet[root] = true
		}
		for root := range rootsInSet {
			rootMap[root].DuplicateSets++
		}
	}

	var sb strings.Builder
	sb.WriteString("Duplicate Files Summary\n")
	sb.WriteString("═══════════════════════\n\n")
	sb.WriteString(fmt.Sprintf("%-50s  %-10s  %-10s  %s\n", "Root Folder", "Sets", "Files", "Total Size"))
	sb.WriteString(strings.Repeat("─", 120))
	sb.WriteString("\n")

	// Sort by root path for consistent display
	var roots []*RootSummary
	for _, summary := range rootMap {
		roots = append(roots, summary)
	}

	for _, summary := range roots {
		path := summary.Path
		if len(path) > 50 {
			path = "..." + path[len(path)-47:]
		}

		sb.WriteString(fmt.Sprintf("%-50s  %-10s  %-10s  %s\n",
			path,
			formatNumber(int64(summary.DuplicateSets)),
			formatNumber(int64(summary.DuplicateFiles)),
			formatSize(summary.TotalSize)))
	}

	sb.WriteString("\n")

	totalSets := len(sets)
	totalFiles := 0
	totalSize := int64(0)
	for _, set := range sets {
		totalFiles += len(set.Files)
		totalSize += set.Size * int64(len(set.Files))
	}

	sb.WriteString(fmt.Sprintf("Total: %d duplicate sets, %d files, %s\n",
		totalSets, totalFiles, formatSize(totalSize)))
	sb.WriteString("\nUse 'dupectl get duplicates --details' to see individual file paths.\n")

	return sb.String()
}

// FormatJSON formats duplicates as JSON
func (f *Formatter) FormatJSON(sets []*DuplicateSet) (string, error) {
	type JSONFile struct {
		Path string `json:"path"`
		Size int64  `json:"size"`
	}

	type JSONSet struct {
		Hash  string     `json:"hash"`
		Size  int64      `json:"size"`
		Count int        `json:"count"`
		Files []JSONFile `json:"files"`
	}

	result := make([]JSONSet, len(sets))
	for i, set := range sets {
		files := make([]JSONFile, len(set.Files))
		for j, file := range set.Files {
			files[j] = JSONFile{
				Path: file.Path,
				Size: file.Size,
			}
		}

		result[i] = JSONSet{
			Hash:  set.Hash,
			Size:  set.Size,
			Count: len(set.Files),
			Files: files,
		}
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// formatSize formats byte size in human-readable format
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatNumber formats an integer with thousand separators
func formatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	str := fmt.Sprintf("%d", n)
	result := ""
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result
}
