package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

// SHA256Hasher implements the Hasher interface using SHA-256
type SHA256Hasher struct{}

// HashFile calculates the SHA-256 hash of a file
func (h *SHA256Hasher) HashFile(reader io.Reader) (string, error) {
	hasher := sha256.New()

	// Use 64KB buffer for efficient reading
	buf := make([]byte, 64*1024)

	_, err := io.CopyBuffer(hasher, reader, buf)
	if err != nil {
		return "", fmt.Errorf("failed to read file for hashing: %w", err)
	}

	// Return hex-encoded hash
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// Algorithm returns the algorithm name
func (h *SHA256Hasher) Algorithm() string {
	return "sha256"
}
