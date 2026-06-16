package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"
)

type AlertResponse struct {
	Features []Feature `json:"features"`
	Title    string    `json:"title"`
	Updated  time.Time `json:"updated"`
}

type Feature struct {
	Properties AlertProperties `json:"properties"`
}

type AlertProperties struct {
	ID          string    `json:"id"`
	Event       string    `json:"event"`
	Severity    string    `json:"severity"`
	Urgency     string    `json:"urgency"`
	MessageType string    `json:"messageType"`
	Status      string    `json:"status"`
	Headline    string    `json:"headline"`
	AreaDesc    string    `json:"areaDesc"`
	Onset       time.Time `json:"onset"`
	Expires     time.Time `json:"expires"`
	Ends        time.Time `json:"ends"`
	SenderName  string    `json:"senderName"`
	Description string    `json:"description"`
	Instruction string    `json:"instruction"`
	Geocode     Geocode   `json:"geocode"`
}

type Geocode struct {
	UGC []string `json:"UGC"`
}

// Struct to hold alerts types from NWS
type AlertTypesResponse struct {
	EventTypes []string `json:"eventTypes"`
}

// Print all valid alerts types from NWS if -listevents flag is passed in
func ListEventTypes(client *http.Client, alertsURL string, debug bool) error {

	// getUserAgent() is located in env.go
	userAgent, err := getUserAgent()
	if err != nil {
		return err
	}

	url := alertsURL + "types"
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to build request: %s", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/ld+json")
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("network error: %s", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading body: %s", err)
	}

	if debug {
		fmt.Printf("\n--- Raw response from %s ---\n%s\n", url, string(body))
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("API error %d: %s", res.StatusCode, string(body))
	}

	var types AlertTypesResponse
	if err := json.Unmarshal(body, &types); err != nil {
		return fmt.Errorf("unmarshal failed: %s", err)
	}

	fmt.Print("The following are all valid Alert Event types for the NWS:\n\n")
	for _, v := range types.EventTypes {
		fmt.Println(v)
	}
	return nil
}

func ConnectNOAA(client *http.Client, alertsURL string, cfg Config, debug bool) (AlertResponse, error) {

	// Get UserAgent is located in env.go
	userAgent, err := getUserAgent()
	if err != nil {
		return AlertResponse{}, err
	}

	// Setup context, Get, and URL
	url := alertsURL + "active"
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodGet, url, nil)
	if err != nil {
		return AlertResponse{}, fmt.Errorf("failed to build request %s", err)
	}

	// Build query to send to API
	q := req.URL.Query()
	q.Add("area", cfg.Area)
	req.URL.RawQuery = q.Encode()

	// Build the headers
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/geo+json")
	res, err := client.Do(req)
	if err != nil {
		return AlertResponse{}, fmt.Errorf("network error %s", err)
	}
	defer res.Body.Close()

	// Read body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return AlertResponse{}, fmt.Errorf("reading body: %s", err)
	}

	// Return any error messages the API sends
	if res.StatusCode != http.StatusOK {
		return AlertResponse{}, fmt.Errorf("API error %d: %s", res.StatusCode, string(body))
	}

	//Print raw body if debug flag true
	if debug {
		fmt.Printf("\nStatus code: %d\n", res.StatusCode)
		fmt.Printf("\n--- Raw response from %s ---\n%s\n", url, string(body))
	}

	//Finally unmarshal into a slice containing the struct declared above
	var alerts AlertResponse
	err = json.Unmarshal(body, &alerts)
	if err != nil {
		return AlertResponse{}, fmt.Errorf("unmarshal failed %s", err)
	}
	return alerts, nil
}

// Returns alerts matching the configured zone and events for printing or Pushover
func FilterAlerts(alerts AlertResponse, cfg Config) []AlertProperties {
	var matches []AlertProperties
	for _, f := range alerts.Features {
		if !slices.Contains(f.Properties.Geocode.UGC, cfg.Zone) {
			continue
		}
		if !slices.Contains(cfg.Events, f.Properties.Event) {
			continue
		}
		matches = append(matches, f.Properties)
	}
	return matches
}

// Prints alerts to the screen if flag is sent - Useful for seeing alerts without firing off notifications
func PrintMatchingAlerts(matches []AlertProperties) {
	for _, v := range matches {
		fmt.Println("--- Matching Alert ---")
		fmt.Println("Event: ", v.Event)
		fmt.Println("Severity: ", v.Severity)
		fmt.Println("Headline: ", v.Headline)
		fmt.Println("Description:")
		fmt.Println(v.Description)
		fmt.Println("Area: ", v.AreaDesc)
		fmt.Println()
	}
}

// SeenAlerts tracks alert IDs already notified about, mapped to their expiration time
type SeenAlerts map[string]time.Time

// PruneSeenAlerts removes any entries whose alert has already expired
func PruneSeenAlerts(seen SeenAlerts) SeenAlerts {
	now := time.Now()      // Get the current time
	pruned := SeenAlerts{} // Create a new empty map of seen alerts object
	for id, expires := range seen {
		if expires.After(now) { // Only keep alerts that haven't expired yet
			pruned[id] = expires // Add ID that hasn't expired to the new pruned map
		}
	}
	return pruned
}
