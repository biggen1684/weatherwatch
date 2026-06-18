package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
)

// Runs four functions to retrieve NWS Zone if -zip flag is sent in
func LookupZone(client *http.Client, zipURL string, pointsURL string, zip string, debug bool) (string, string, error) {

	// getUserAgent() is located in env.go
	userAgent, err := getUserAgent()
	if err != nil {
		return "", "", err
	}

	err = validateZip(zip)
	if err != nil {
		return "", "", err
	}

	lat, lon, err := zipToLatLon(client, zipURL, zip, debug)
	if err != nil {
		return "", "", err
	}

	zone, county, err := latLonToZone(client, pointsURL, userAgent, lat, lon, debug)
	if err != nil {
		return "", "", err
	}
	return zone, county, nil
}

// Structs to hold returned NWS zone
type PointsResponse struct {
	Properties PointsProperties `json:"properties"`
}

type PointsProperties struct {
	ForecastZone string `json:"forecastZone"`
	County       string `json:"county"`
	Type         string `json:"type"`
}

// Retrieve NWS zone from NWS API via lat/long that was returned earlier
func latLonToZone(client *http.Client, pointsURL string, userAgent string, lat string, long string, debug bool) (string, string, error) {
	// Setup context, Get, and URL
	url := pointsURL + lat + "," + long
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodGet, url, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to build request %s", err)
	}

	// Build the headers
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/geo+json")
	res, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("network error %s", err)
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
	if res.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("API error %d: %s", res.StatusCode, string(body))
	}

	// Finally unmarshal into a slice containing the struct declared above
	var response PointsResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", "", fmt.Errorf("unmarshal failed %s", err)
	}
	// Returns the last segment of the address after the "/". We need both zone and county returned
	zone := path.Base(response.Properties.ForecastZone)
	county := path.Base(response.Properties.County)

	// If zip corresponds to a marine zone, notifier the user
	if response.Properties.Type == "marine" {
		return zone, county, fmt.Errorf("this zip resolved to a marine zone (%s) which you probably don't want (but could use).\nThe geocoded coordinate likely fell over water so try a nearby zip instead", zone)
	}
	return zone, county, nil
}
