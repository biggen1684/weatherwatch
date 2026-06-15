package weather

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	t.Run("valid toml file", func(t *testing.T) {
		content := `
zone = "FLZ112"
area = "FL"
events = ["Tornado Warning", "Heat Advisory"]
`
		path := filepath.Join(t.TempDir(), "config.toml")
		err := os.WriteFile(path, []byte(content), 0644)
		assert.NoError(t, err)

		cfg, err := loadConfig(path)
		assert.NoError(t, err)
		assert.Equal(t, "FLZ112", cfg.Zone)
		assert.Equal(t, "FL", cfg.Area)
		assert.Equal(t, []string{"Tornado Warning", "Heat Advisory"}, cfg.Events)
	})

	t.Run("file does not exist", func(t *testing.T) {
		_, err := loadConfig("nonexistent.toml")
		assert.Error(t, err)
	})

	t.Run("malformed toml", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "bad.toml")
		err := os.WriteFile(path, []byte("zone = ["), 0644)
		assert.NoError(t, err)

		_, err = loadConfig(path)
		assert.Error(t, err)
	})
}

func TestValidateConfig(t *testing.T) {
	valid := Config{
		Zone:   "FLZ112",
		Area:   "FL",
		Events: []string{"Tornado Warning"},
	}

	t.Run("valid config", func(t *testing.T) {
		err := validateConfig(valid)
		assert.NoError(t, err)
	})

	t.Run("missing area", func(t *testing.T) {
		cfg := valid
		cfg.Area = ""
		err := validateConfig(cfg)
		assert.Error(t, err)
	})

	t.Run("missing events", func(t *testing.T) {
		cfg := valid
		cfg.Events = nil
		err := validateConfig(cfg)
		assert.Error(t, err)
	})

	t.Run("missing zone", func(t *testing.T) {
		cfg := valid
		cfg.Zone = ""
		err := validateConfig(cfg)
		assert.Error(t, err)
	})
}
