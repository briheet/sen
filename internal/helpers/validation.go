package helpers

import (
	"errors"
	"os"
	"path/filepath"
)

var (
	ErrIsNotDir = errors.New(`
		the given source path is not a directory.
		Please give valid source directory to work with.
	`)
)

// This does validation for main source path
func ValidateSourcePath(sourcePath string) error {
	// Create a absolute Source path if not
	// Helps in future and reduces absolute path checks
	sourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return err
	}

	// Get source info
	info, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}

	// Check if it is Dir or not.
	// Golang's packages takes in Dir which has
	// access to entrypoint of the application.
	if !info.IsDir() {
		return ErrIsNotDir
	}
	return nil
}
