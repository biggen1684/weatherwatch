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
