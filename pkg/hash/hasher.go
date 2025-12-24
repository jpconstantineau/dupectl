package hash

import (
	"fmt"
	"io"
)

// Hasher is the interface for file hashing algorithms
type Hasher interface {
	// HashFile calculates the hash of a file from its reader
	HashFile(reader io.Reader) (string, error)
	// Algorithm returns the algorithm name
	Algorithm() string
}

// NewHasher creates a new hasher based on the algorithm name
func NewHasher(algorithm string) (Hasher, error) {
	switch algorithm {
	case "sha256":
		return &SHA256Hasher{}, nil
	case "sha512":
		return &SHA512Hasher{}, nil
	case "sha3-256":
		return &SHA3Hasher{}, nil
	default:
		return nil, fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}
}
