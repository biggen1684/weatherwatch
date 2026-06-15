package weather

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Struct to hold config.toml fields
type Config struct {
	Zone      string   `toml:"zone"`
	Area      string   `toml:"area"`
	UserAgent string   `toml:"user_agent"`
	Events    []string `toml:"events"`
}

func PreRunSetup() (string, Config, error) {
	key, err := getPushoverKey()
	if err != nil {
		return "", Config{}, err
	}

	cfg, err := loadConfig("config.toml")
	if err != nil {
		return "", Config{}, err
	}

	err = validateConfig(cfg)
	if err != nil {
		return "", Config{}, err
	}

	return key, cfg, nil
}

// Load Pushover API key from environment variable
func getPushoverKey() (string, error) {
	key := os.Getenv("PUSHOVER_API_KEY")
	if key == "" {
		return "", fmt.Errorf("PUSHOVER_API_KEY environment variable not set")
	}
	return key, nil
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

// Load User Agent needed from environment variable for NOAA api access
func getUserAgent() (string, error) {
	userAgent := os.Getenv("WEATHERWATCH_USER_AGENT")
	if userAgent == "" {
		return "", fmt.Errorf("WEATHERWATCH_USER_AGENT environment variable not set")
	}
	return userAgent, nil
}
