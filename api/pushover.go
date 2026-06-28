package weather

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func SendPushover(client *http.Client, pushoverURL string, apiKey string, userKey string, alert AlertProperties) error {

	// Build POST form data
	data := url.Values{}
	data.Set("token", apiKey)
	data.Set("user", userKey)
	data.Set("title", alert.Headline)
	// Pushover has a 1024 char limit. Make sure we truncate ourselves if we go over
	// because Pushover will reject it if so
	message := alert.Description + "\n\nArea: " + alert.AreaDesc
	if len(message) > 1024 {
		message = message[:1021] + "..."
	}
	data.Set("message", message)

	// Build request with context and form body
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodPost, pushoverURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to build request: %v", err)
	}

	// Build header and send data
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("network error: %v", err)
	}
	defer res.Body.Close()

	// Read Body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading body: %v", err)
	}

	// Return any error messages the API sends
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("pushover API error %d: %s", res.StatusCode, string(body))
	}

	return nil
}

// Wrapper function that creates a fake alert when the -test flag is used
func SendPushoverTest(client *http.Client, pushoverURL string, apiKey string, userKey string) error {
	testAlert := AlertProperties{
		Headline:    "weatherwatch test notification",
		Description: "If you received this, your Pushover API key and User key are configured correctly.",
		AreaDesc:    "weatherwatch",
	}
	err := SendPushover(client, pushoverURL, apiKey, userKey, testAlert)
	return err
}

// Wrapper functions that sends an alert that weatherwatch is shutting down
func SendPushoverShutdown(client *http.Client, pushoverURL string, apiKey string, userKey string) {
	testAlert := AlertProperties{
		Headline:    "Shutting down weatherwatch",
		Description: "Weatherwatch has exited. Please restart if you wish to get alerts again.",
		AreaDesc:    "weatherwatch",
	}
	_ = SendPushover(client, pushoverURL, apiKey, userKey, testAlert)

}
