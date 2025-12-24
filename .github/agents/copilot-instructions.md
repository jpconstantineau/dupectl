# dupectl Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-12-24

## Active Technologies

- Go 1.21+ (leverages goroutines for concurrent scanning, standard library crypto/hash for hashing) + `spf13/cobra` (CLI framework, already present), `spf13/viper` (configuration, already present), `modernc.org/sqlite` (embedded database with WAL mode, already present), `crypto/sha256`, `crypto/sha512`, `golang.org/x/crypto/sha3` (hash algorithms from stdlib + golang.org/x) (001-duplicate-scan-system-01)

## Project Structure

```text
src/
tests/
```

## Commands

# Add commands for Go 1.21+ (leverages goroutines for concurrent scanning, standard library crypto/hash for hashing)

## Code Style

Go 1.21+ (leverages goroutines for concurrent scanning, standard library crypto/hash for hashing): Follow standard conventions

## Recent Changes

- 001-duplicate-scan-system-01: Added Go 1.21+ (leverages goroutines for concurrent scanning, standard library crypto/hash for hashing) + `spf13/cobra` (CLI framework, already present), `spf13/viper` (configuration, already present), `modernc.org/sqlite` (embedded database with WAL mode, already present), `crypto/sha256`, `crypto/sha512`, `golang.org/x/crypto/sha3` (hash algorithms from stdlib + golang.org/x)

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
