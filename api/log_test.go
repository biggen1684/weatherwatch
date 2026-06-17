package weather

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupLogger(t *testing.T) {
	t.Run("creates log file successfully", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "test.log")

		logFile, err := SetupLogger(path)
		assert.NoError(t, err)
		assert.NotNil(t, logFile)
		defer logFile.Close()

		_, err = os.Stat(path)
		assert.NoError(t, err)
	})

	t.Run("invalid path returns error", func(t *testing.T) {
		_, err := SetupLogger("/nonexistent-dir-xyz/test.log")
		assert.Error(t, err)
	})
}
