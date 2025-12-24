package formatter

import (
	"fmt"
	"strings"

	"github.com/jpconstantineau/dupectl/pkg/entities"
)

// FormatDuplicatesTable formats duplicate file sets as a human-readable table
func FormatDuplicatesTable(sets []entities.DuplicateFileSet, totalSets, totalFiles int, wastedBytes, recoverableBytes int64) string {
	var builder strings.Builder

	builder.WriteString("Duplicate Files Report\n")
	builder.WriteString("======================\n\n")

	if len(sets) == 0 {
		builder.WriteString("No duplicate files found. All files are unique.\n")
		return builder.String()
	}

	for i, set := range sets {
		builder.WriteString(fmt.Sprintf("Duplicate Set #%d (Size: %s, Hash: %s, Algorithm: %s)\n",
			i+1, formatBytes(set.Size), truncateHash(set.HashValue), set.HashAlgorithm))
		builder.WriteString(fmt.Sprintf("  File Count: %d\n", set.FileCount))

		for j, file := range set.Files {
			prefix := "├─"
			if j == len(set.Files)-1 {
				prefix = "└─"
			}
			builder.WriteString(fmt.Sprintf("  %s %s\n", prefix, file.Path))
		}

		builder.WriteString("\n")
	}

	builder.WriteString("────────────────────────────────────────────────────────────────\n")
	builder.WriteString("Summary:\n")
	builder.WriteString(fmt.Sprintf("  Total Duplicate Sets: %d\n", totalSets))
	builder.WriteString(fmt.Sprintf("  Total Duplicate Files: %d\n", totalFiles))
	builder.WriteString(fmt.Sprintf("  Total Wasted Space: %s (across all copies)\n", formatBytes(wastedBytes)))
	builder.WriteString(fmt.Sprintf("  Storage Recoverable: %s (if keeping 1 copy per set)\n", formatBytes(recoverableBytes)))

	return builder.String()
}

// formatBytes formats bytes in human-readable form
func formatBytes(bytes int64) string {
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

// truncateHash returns first 16 characters of hash for display
func truncateHash(hash string) string {
	if len(hash) > 16 {
		return hash[:16] + "..."
	}
	return hash
}
