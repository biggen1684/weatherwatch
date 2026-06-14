package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/BurntSushi/toml"
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

func getUserAgent() (string, error) {
	userAgent := os.Getenv("WEATHERWATCH_USER_AGENT")
	if userAgent == "" {
		return "", fmt.Errorf("WEATHERWATCH_USER_AGENT environment variable not set")
	}
	return userAgent, nil
}

// Struct to hold config.toml fields
type Config struct {
	Zone      string   `toml:"zone"`
	Area      string   `toml:"area"`
	UserAgent string   `toml:"user_agent"`
	Events    []string `toml:"events"`
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

func validateConfig(cfg Config) error {
	if cfg.Area == "" {
		return fmt.Errorf("area is missing from config.toml")
	}
	if len(cfg.Events) == 0 {
		return fmt.Errorf("events are missing from config.toml")
	}
	if cfg.Zone == "" {
		return fmt.Errorf("NWS Zone is missing — run with -zip to find your NWS Zone")
	}

	return nil
}

// Runs four functions to retrieve NWS Zone if -zip flag is sent in
func LookupZone(client *http.Client, zipURL string, pointsURL string, zip string, debug bool) (string, error) {

	userAgent, err := getUserAgent()
	if err != nil {
		return "", err
	}

	err = validateZip(zip)
	if err != nil {
		return "", err
	}

	lat, lon, err := zipToLatLon(client, zipURL, zip, debug)
	if err != nil {
		return "", err
	}

	zone, err := latLonToZone(client, pointsURL, userAgent, lat, lon, debug)
	if err != nil {
		return "", err
	}
	return zone, nil
}

// Check to see if the entered zip flag is 5 digits
func validateZip(zip string) error {
	if len(zip) != 5 {
		return fmt.Errorf("%q is not valid, must be 5 digits in length", zip)
	}
	for _, c := range zip {
		if c < '0' || c > '9' {
			return fmt.Errorf("%q is not valid, must be 5 digits only", zip)
		}
	}
	return nil
}

// Struct to hold returned lat/long fields
type ZipResponse struct {
	Places []Place `json:"places"`
}

type Place struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

// Send to Zippopotam.us to retrieve Lat/Long
func zipToLatLon(client *http.Client, zipURL string, zip string, debug bool) (string, string, error) {
	//Setup context, Get, and URL
	url := zipURL + zip
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodGet, url, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to build request %s", err)
	}

	// Build headers
	req.Header.Set("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("network error: %v", err)
	}
	defer res.Body.Close()

	// Read body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", fmt.Errorf("reading body: %s", err)
	}

	if debug {
		fmt.Printf("\n--- Raw response from %s ---\n%s\n", url, string(body))
	}

	// Return any error messages the API sends
	if res.StatusCode == http.StatusNotFound {
		return "", "", fmt.Errorf("zip code %q not found", zip)
	}
	if res.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("zippopotam API error %d: %s", res.StatusCode, string(body))
	}

	//Finally unmarshal into a slice containing the struct declared above
	var zipResp ZipResponse
	err = json.Unmarshal(body, &zipResp)
	if err != nil {
		return "", "", fmt.Errorf("unmarshal failed %s", err)
	}

	// Guard against there ever being a sucessful 200 response but for whatever reason
	// zipResp.Places[0] is empty of lat/long
	if len(zipResp.Places) == 0 {
		return "", "", fmt.Errorf("no location data returned for zip %q", zip)
	}

	return zipResp.Places[0].Latitude, zipResp.Places[0].Longitude, nil
}

// Structs to hold returned NWS zone
type PointsResponse struct {
	Properties PointsProperties `json:"properties"`
}

type PointsProperties struct {
	ForecastZone string `json:"forecastZone"`
}

func latLonToZone(client *http.Client, pointsURL string, userAgent string, lat string, long string, debug bool) (string, error) {
	// Setup context, Get, and URL
	url := pointsURL + lat + "," + long
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to build request %s", err)
	}

	// Build the headers
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/geo+json")
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("network error %s", err)
	}
	defer res.Body.Close()

	// Read body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("reading body: %s", err)
	}

	if debug {
		fmt.Printf("\n--- Raw response from %s ---\n%s\n", url, string(body))
	}

	// Return any error messages the API sends
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error %d: %s", res.StatusCode, string(body))
	}

	// Finally unmarshal into a slice containing the struct declared above
	var response PointsResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("unmarshal failed %s", err)
	}
	// Returns the last segment of the address after the "/"
	zone := path.Base(response.Properties.ForecastZone)
	return zone, nil
}

// func ConnectNOAA(client *http.Client, alertsURL string, state string, debug bool) (AlertResponse, error) {
// 	// Setup context, Get, and URL
// 	url := alertsURL + "active"
// 	req, err := http.NewRequestWithContext(context.Background(),
// 		http.MethodGet, url, nil)
// 	if err != nil {
// 		return AlertResponse{}, fmt.Errorf("failed to build request %s", err)
// 	}

// 	// Build query to send to API
// 	q := req.URL.Query()
// 	q.Add("area", state)
// 	req.URL.RawQuery = q.Encode()

// 	// Build the headers
// 	req.Header.Set("User-Agent", "weatherwatch (joe@joemoe.net)")
// 	req.Header.Set("Accept", "application/geo+json")
// 	res, err := client.Do(req)
// 	if err != nil {
// 		return AlertResponse{}, fmt.Errorf("network error %s", err)
// 	}
// 	defer res.Body.Close()

// 	// Read body
// 	body, err := io.ReadAll(res.Body)
// 	if err != nil {
// 		return AlertResponse{}, fmt.Errorf("reading body: %s", err)
// 	}

// 	// Return any error messages the API sends
// 	if res.StatusCode != http.StatusOK {
// 		return AlertResponse{}, fmt.Errorf("API error %d: %s", res.StatusCode, string(body))
// 	}

// 	//Print raw body if debug flag true
// 	if debug == true {
// 		fmt.Printf("\nStatus code: %d\n", res.StatusCode)
// 		fmt.Println(string(body))
// 	}

// 	//Finally unmarshal into a slice containing the struct declared above
// 	var alerts AlertResponse
// 	err = json.Unmarshal(body, &alerts)
// 	if err != nil {
// 		return AlertResponse{}, fmt.Errorf("unmarshal failed %s", err)
// 	}
// 	return alerts, nil
//}
