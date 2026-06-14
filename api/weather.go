package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

// Config objec to hold config.toml fields
type Config struct {
	Office    string   `toml:"office"`
	Area      string   `toml:"area"`
	UserAgent string   `toml:"user_agent"`
	Events    []string `toml:"events"`
}

// Zip and Place struct to hold retuned zip to lat long fields
type ZipResponse struct {
	Places []Place `json:"places"`
}

type Place struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

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

func validateConfig(cfg Config) error {
	if cfg.Area == "" {
		return fmt.Errorf("area is missing from config.toml")
	}
	if len(cfg.Events) == 0 {
		return fmt.Errorf("events are missing from config.toml")
	}
	if cfg.Office == "" {
		return fmt.Errorf("NWS office is missing — run with -zip to find your NWS office")
	}
	if cfg.UserAgent == "" {
		return fmt.Errorf("useragent field is empty")
	}

	return nil
}

func LookupOffice(client *http.Client, zipURL string, pointsURL string, zip string) (string, error) {
	err := validateZip(zip)
	if err != nil {
		return "", err
	}

	lat, lon, err := zipToLatLon(client, zipURL, zip)
	if err != nil {
		return "", err
	}
	fmt.Println(lat)
	fmt.Print(lon)

	// office, err := latLonToOffice(client, lat, lon)
	// if err != nil {
	// 	return "", err
	// }
	return "", nil
}

// Check to see if user entered zip flag is 5 digits
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

func zipToLatLon(client *http.Client, zipURL string, zip string) (string, string, error) {
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
	// Places[0] is empty of lat/long
	if len(zipResp.Places) == 0 {
		return "", "", fmt.Errorf("no location data returned for zip %q", zip)
	}

	return zipResp.Places[0].Latitude, zipResp.Places[0].Longitude, nil
}

func latLonToOffice(client *http.Client, lat string, long string) {
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
