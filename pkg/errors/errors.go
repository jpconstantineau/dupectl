package errors

import "fmt"

// ScanError represents a scan-specific error with context
type ScanError struct {
	Path      string
	Operation string
	Err       error
}

func (e *ScanError) Error() string {
	return fmt.Sprintf("%s failed for %s: %v", e.Operation, e.Path, e.Err)
}

func (e *ScanError) Unwrap() error {
	return e.Err
}

// NewScanError creates a new scan error
func NewScanError(path, operation string, err error) *ScanError {
	return &ScanError{
		Path:      path,
		Operation: operation,
		Err:       err,
	}
}

// PermissionError represents a permission denied error
type PermissionError struct {
	Path string
	Err  error
}

func (e *PermissionError) Error() string {
	return fmt.Sprintf("Permission denied accessing %s: %v. Try running with elevated privileges or check file permissions.", e.Path, e.Err)
}

func (e *PermissionError) Unwrap() error {
	return e.Err
}

// NewPermissionError creates a permission error with actionable message
func NewPermissionError(path string, err error) *PermissionError {
	return &PermissionError{
		Path: path,
		Err:  err,
	}
}

// DatabaseError represents a database operation error
type DatabaseError struct {
	Operation string
	Err       error
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("Database %s failed: %v. Check database file permissions and disk space.", e.Operation, e.Err)
}

func (e *DatabaseError) Unwrap() error {
	return e.Err
}

// NewDatabaseError creates a database error
func NewDatabaseError(operation string, err error) *DatabaseError {
	return &DatabaseError{
		Operation: operation,
		Err:       err,
	}
}
