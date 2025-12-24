package hash

import (
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/sha3"
)

// Hasher calculates cryptographic hashes for files
type Hasher interface {
	Hash(ctx context.Context, filePath string) (string, error)
	Algorithm() string
}

// SHA256Hasher implements SHA-256 hashing
type SHA256Hasher struct{}

func (h *SHA256Hasher) Hash(ctx context.Context, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	buffer := make([]byte, 64*1024) // 64KB buffer

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		n, err := file.Read(buffer)
		if n > 0 {
			hasher.Write(buffer[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (h *SHA256Hasher) Algorithm() string {
	return "sha256"
}

// SHA512Hasher implements SHA-512 hashing (default)
type SHA512Hasher struct{}

func (h *SHA512Hasher) Hash(ctx context.Context, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha512.New()
	buffer := make([]byte, 64*1024) // 64KB buffer

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		n, err := file.Read(buffer)
		if n > 0 {
			hasher.Write(buffer[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (h *SHA512Hasher) Algorithm() string {
	return "sha512"
}

// SHA3256Hasher implements SHA3-256 hashing
type SHA3256Hasher struct{}

func (h *SHA3256Hasher) Hash(ctx context.Context, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha3.New256()
	buffer := make([]byte, 64*1024) // 64KB buffer

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		n, err := file.Read(buffer)
		if n > 0 {
			hasher.Write(buffer[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (h *SHA3256Hasher) Algorithm() string {
	return "sha3-256"
}

// NewHasher creates a hasher based on algorithm name
func NewHasher(algorithm string) (Hasher, error) {
	switch algorithm {
	case "sha256":
		return &SHA256Hasher{}, nil
	case "sha512":
		return &SHA512Hasher{}, nil
	case "sha3-256":
		return &SHA3256Hasher{}, nil
	default:
		return nil, fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}
}
