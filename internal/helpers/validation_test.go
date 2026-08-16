package helpers

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSourcePath(t *testing.T) {
	t.Parallel()

	t.Run("valid directory", func(t *testing.T) {
		sourcePath := t.TempDir()
		err := ValidateSourcePath(sourcePath)
		assert.NoError(t, err)
	})

	t.Run("invalid directory", func(t *testing.T) {
		sourcePath := t.TempDir()
		sourcePath = filepath.Join(sourcePath, "main.go")
		err := ValidateSourcePath(sourcePath)
		assert.NotNil(t, err)
	})
}
