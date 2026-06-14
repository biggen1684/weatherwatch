package weather

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Office    string   `toml:"office"`
	Area      string   `toml:"area"`
	UserAgent string   `toml:"user_agent"`
	Events    []string `toml:"events"`
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
