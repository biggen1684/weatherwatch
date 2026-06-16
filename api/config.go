package weather

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// Struct to hold config.toml fields
type Config struct {
	Zone      string   `toml:"zone"`
	Area      string   `toml:"area"`
	UserAgent string   `toml:"user_agent"`
	Events    []string `toml:"events"`
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

// Validate fields are filled in - Doesn't validate they are accurate!
func validateConfig(cfg Config) error {
	if cfg.Area == "" {
		return fmt.Errorf("area is missing from config.toml - must be a two letter state code")
	}
	if len(cfg.Events) == 0 {
		return fmt.Errorf("events are missing from config.toml - run with -listevents to find all valid events available")
	}
	if cfg.Zone == "" {
		return fmt.Errorf("NWS Zone is missing from config.toml — run with -zip to find your NWS Zone")
	}

	return nil
}
