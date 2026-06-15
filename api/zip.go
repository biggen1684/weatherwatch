package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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
