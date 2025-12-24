package pathutil

import (
	"path/filepath"
	"runtime"
	"strings"
)

// ToAbsolute converts a relative path to absolute path
func ToAbsolute(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(absPath), nil
}

// NormalizePath normalizes a path for consistent storage
// Uses platform-specific separators
func NormalizePath(path string) string {
	return filepath.Clean(path)
}

// NormalizePathForStorage normalizes a path for database storage
// On Windows, converts drive letters to uppercase for consistency
func NormalizePathForStorage(path string) string {
	path = filepath.Clean(path)

	// On Windows, ensure drive letter is uppercase
	if runtime.GOOS == "windows" && len(path) >= 2 && path[1] == ':' {
		return strings.ToUpper(path[:1]) + path[1:]
	}

	return path
}

// IsSubpath checks if child is a subpath of parent
func IsSubpath(parent, child string) bool {
	parent = filepath.Clean(parent)
	child = filepath.Clean(child)

	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}

	// If relative path starts with "..", child is not under parent
	return !strings.HasPrefix(rel, "..") && rel != "."
}

// Join joins path elements using platform-specific separator
func Join(elements ...string) string {
	return filepath.Join(elements...)
}
