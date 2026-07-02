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
events = ["Tornado Warning", "Heat Advisory"]

[[locations]]
name = "Home"
area = "FL"
zone = "FLZ112"
county = "FLC005"
`
		path := filepath.Join(t.TempDir(), "config.toml")
		err := os.WriteFile(path, []byte(content), 0644)
		assert.NoError(t, err)

		cfg, err := loadConfig(path)
		assert.NoError(t, err)
		assert.Len(t, cfg.Locations, 1)
		assert.Equal(t, "Home", cfg.Locations[0].Name)
		assert.Equal(t, "FL", cfg.Locations[0].Area)
		assert.Equal(t, "FLZ112", cfg.Locations[0].Zone)
		assert.Equal(t, "FLC005", cfg.Locations[0].County)
		assert.Equal(t, []string{"Tornado Warning", "Heat Advisory"}, cfg.Events)
	})

	t.Run("multiple locations", func(t *testing.T) {
		content := `
events = ["Tornado Warning"]

[[locations]]
name = "Home"
area = "FL"
zone = "FLZ112"
county = "FLC005"

[[locations]]
name = "Vacation"
area = "FL"
zone = "FLZ108"
county = "FLC131"
`
		path := filepath.Join(t.TempDir(), "config.toml")
		err := os.WriteFile(path, []byte(content), 0644)
		assert.NoError(t, err)

		cfg, err := loadConfig(path)
		assert.NoError(t, err)
		assert.Len(t, cfg.Locations, 2)
		assert.Equal(t, "Home", cfg.Locations[0].Name)
		assert.Equal(t, "Vacation", cfg.Locations[1].Name)
	})

	t.Run("file does not exist", func(t *testing.T) {
		_, err := loadConfig("nonexistent.toml")
		assert.Error(t, err)
	})

	t.Run("malformed toml", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "bad.toml")
		err := os.WriteFile(path, []byte("locations = ["), 0644)
		assert.NoError(t, err)

		_, err = loadConfig(path)
		assert.Error(t, err)
	})
}

func TestValidateConfig(t *testing.T) {
	valid := Config{
		Locations: []Location{
			{Name: "Home", Area: "FL", Zone: "FLZ112", County: "FLC005"},
		},
		Events: []string{"Tornado Warning"},
	}

	t.Run("valid config", func(t *testing.T) {
		err := validateConfig(valid)
		assert.NoError(t, err)
	})

	t.Run("no locations", func(t *testing.T) {
		cfg := valid
		cfg.Locations = nil
		err := validateConfig(cfg)
		assert.Error(t, err)
	})

	t.Run("missing location name", func(t *testing.T) {
		cfg := Config{
			Locations: []Location{
				{Name: "", Area: "FL", Zone: "FLZ112", County: "FLC005"},
			},
			Events: []string{"Tornado Warning"},
		}
		err := validateConfig(cfg)
		assert.Error(t, err)
	})

	t.Run("missing area", func(t *testing.T) {
		cfg := Config{
			Locations: []Location{
				{Name: "Home", Area: "", Zone: "FLZ112", County: "FLC005"},
			},
			Events: []string{"Tornado Warning"},
		}
		err := validateConfig(cfg)
		assert.Error(t, err)
	})

	t.Run("missing zone", func(t *testing.T) {
		cfg := Config{
			Locations: []Location{
				{Name: "Home", Area: "FL", Zone: "", County: "FLC005"},
			},
			Events: []string{"Tornado Warning"},
		}
		err := validateConfig(cfg)
		assert.Error(t, err)
	})

	t.Run("missing county", func(t *testing.T) {
		cfg := Config{
			Locations: []Location{
				{Name: "Home", Area: "FL", Zone: "FLZ112", County: ""},
			},
			Events: []string{"Tornado Warning"},
		}
		err := validateConfig(cfg)
		assert.Error(t, err)
	})

	t.Run("missing events", func(t *testing.T) {
		cfg := valid
		cfg.Events = nil
		err := validateConfig(cfg)
		assert.Error(t, err)
	})

	t.Run("valid config with multiple locations", func(t *testing.T) {
		cfg := Config{
			Locations: []Location{
				{Name: "Home", Area: "FL", Zone: "FLZ112", County: "FLC005"},
				{Name: "Vacation", Area: "FL", Zone: "FLZ108", County: "FLC131"},
			},
			Events: []string{"Tornado Warning"},
		}
		err := validateConfig(cfg)
		assert.NoError(t, err)
	})
	t.Run("lat without lon returns error", func(t *testing.T) {
		cfg := Config{
			Locations: []Location{
				{Name: "Home", Area: "FL", Zone: "FLZ112", County: "FLC005", Lat: 30.19},
			},
			Events: []string{"Tornado Warning"},
		}
		err := validateConfig(cfg)
		assert.Error(t, err)
	})

	t.Run("lon without lat returns error", func(t *testing.T) {
		cfg := Config{
			Locations: []Location{
				{Name: "Home", Area: "FL", Zone: "FLZ112", County: "FLC005", Lon: -85.81},
			},
			Events: []string{"Tornado Warning"},
		}
		err := validateConfig(cfg)
		assert.Error(t, err)
	})

	t.Run("both lat and lon configured is valid", func(t *testing.T) {
		cfg := Config{
			Locations: []Location{
				{Name: "Home", Area: "FL", Zone: "FLZ112", County: "FLC005", Lat: 30.19, Lon: -85.81},
			},
			Events: []string{"Tornado Warning"},
		}
		err := validateConfig(cfg)
		assert.NoError(t, err)
	})

	t.Run("neither lat nor lon configured is valid", func(t *testing.T) {
		cfg := Config{
			Locations: []Location{
				{Name: "Home", Area: "FL", Zone: "FLZ112", County: "FLC005"},
			},
			Events: []string{"Tornado Warning"},
		}
		err := validateConfig(cfg)
		assert.NoError(t, err)
	})
}
