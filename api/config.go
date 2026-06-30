package weather

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// Location holds the configuration for a single NWS forecast zone to monitor
type Location struct {
	Name   string `toml:"name"`
	Area   string `toml:"area"`
	Zone   string `toml:"zone"`
	County string `toml:"county"`
}

// Config holds all configuration settings loaded from config.toml
type Config struct {
	Locations []Location `toml:"locations"`
	Events    []string   `toml:"events"`
}

func PreRunSetup() (string, string, Config, error) {

	// getPushoverKey() is located in env.go
	apiKey, err := getPushoverAPIKey()
	if err != nil {
		return "", "", Config{}, err
	}

	// getPushoverKey() is located in env.go
	userKey, err := getPushoverUserKey()
	if err != nil {
		return "", "", Config{}, err
	}

	cfg, err := loadConfig("config.toml")
	if err != nil {
		return "", "", Config{}, err
	}

	err = validateConfig(cfg)
	if err != nil {
		return "", "", Config{}, err
	}

	return apiKey, userKey, cfg, nil
}

// Check if config is there and unmarshal if it is
func loadConfig(path string) (Config, error) {
	var cfg Config
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("could not load config.toml: %v", err)
	}
	return cfg, nil
}

// validateConfig checks that all required fields are present
func validateConfig(cfg Config) error {
	if len(cfg.Locations) == 0 {
		return fmt.Errorf("no locations defined in config.toml — add at least one [[locations]] block")
	}
	for i, loc := range cfg.Locations {
		if loc.Name == "" {
			return fmt.Errorf("location %d is missing a name", i+1)
		}
		if loc.Area == "" {
			return fmt.Errorf("location %d (%s) is missing area", i+1, loc.Name)
		}
		if loc.Zone == "" {
			return fmt.Errorf("location %d (%s) is missing zone", i+1, loc.Name)
		}
		if loc.County == "" {
			return fmt.Errorf("location %d (%s) is missing county", i+1, loc.Name)
		}
	}
	if len(cfg.Events) == 0 {
		return fmt.Errorf("events is missing from config.toml — add at least one event type")
	}
	return nil
}
