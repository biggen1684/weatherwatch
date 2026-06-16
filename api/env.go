package weather

import (
	"fmt"
	"os"
)

// Shared "utility" file". Both of the functions below are called at a couple different places
// inside the program so it makes sense to group them together here to help with code readability

// Load User Agent needed from environment variable for NOAA api access
func getUserAgent() (string, error) {
	userAgent := os.Getenv("WEATHERWATCH_USER_AGENT")
	if userAgent == "" {
		return "", fmt.Errorf("WEATHERWATCH_USER_AGENT environment variable not set")
	}
	return userAgent, nil
}

// Load Pushover API key from environment variable
func getPushoverAPIKey() (string, error) {
	apiKey := os.Getenv("PUSHOVER_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("PUSHOVER_API_KEY environment variable not set")
	}
	return apiKey, nil
}

// Load Pushover USER key from environment variable
func getPushoverUserKey() (string, error) {
	userKey := os.Getenv("PUSHOVER_USER_KEY")
	if userKey == "" {
		return "", fmt.Errorf("PUSHOVER_USER_KEY environment variable not set")
	}
	return userKey, nil
}
