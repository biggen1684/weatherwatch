package weather

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Sends a Pushover notification for a matching alert
func SendPushover(client *http.Client, pushoverURL string, apiKey string, userKey string, alert FlatAlert) error {

	// Build POST form data
	data := url.Values{}
	data.Set("token", apiKey)
	data.Set("user", userKey)
	data.Set("title", alert.Event)
	data.Set("message", alert.Headline)

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
