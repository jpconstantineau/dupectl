package hash

import (
	"encoding/hex"
	"fmt"
	"io"

	"golang.org/x/crypto/sha3"
)

// SHA3Hasher implements the Hasher interface using SHA3-256
type SHA3Hasher struct{}

// HashFile calculates the SHA3-256 hash of a file
func (h *SHA3Hasher) HashFile(reader io.Reader) (string, error) {
	hasher := sha3.New256()

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
func (h *SHA3Hasher) Algorithm() string {
	return "sha3-256"
}
