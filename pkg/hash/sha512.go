package hash

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
)

// SHA512Hasher implements the Hasher interface using SHA-512
type SHA512Hasher struct{}

// HashFile calculates the SHA-512 hash of a file
func (h *SHA512Hasher) HashFile(reader io.Reader) (string, error) {
	hasher := sha512.New()

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
func (h *SHA512Hasher) Algorithm() string {
	return "sha512"
}
