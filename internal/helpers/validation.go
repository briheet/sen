package helpers

import (
	"errors"
	"os"
)

var (
	ErrIsNotDir = errors.New(`
		the given source path is not a directory.
		Please give valid source directory to work with.
	`)
)

// This does validation for main source path
func ValidateSourcePath(sourcePath string) error {
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
