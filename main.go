package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	weather "github.com/biggen1684/weatherwatch/api"
)

const alertsURL = "https://api.weather.gov/alerts/"
const zipURL = "https://api.zippopotam.us/us/"
const pointsURL = "https://api.weather.gov/points/"

func main() {
	client := &http.Client{Timeout: 30 * time.Second}

	zip := flag.String("zip", "", "Zip code to look up your NWS Zone (e.g. -zip 32547)")
	listevents := flag.Bool("listevents", false, "List all valid NWS alert event types")
	debug := flag.Bool("debug", false, "Print raw API responses for troubleshooting")
	print := flag.Bool("print", false, "Print alerts matching your configured zone and events, then exit")
	flag.Parse()

	// Convert zip to lat/long if zip flag is sent in - End program after
	if *zip != "" {
		zone, err := weather.LookupZone(client, zipURL, pointsURL, *zip, *debug)
		if err != nil {
			fmt.Printf("Error: %v.\n", err)
			return
		}
		fmt.Printf("Your NWS Zone is %s:\n", zone)
		fmt.Println("Add this to the 'zone' field in config.toml")
		return
	}

	// List valid event types - End program after
	if *listevents {
		err := weather.ListEventTypes(client, alertsURL, *debug)
		if err != nil {
			fmt.Printf("Error: %v.\n", err)
			return
		}
		return
	}

	// Make sure we have environment variables set and the config.toml
	apiKey, userKey, cfg, err := weather.PreRunSetup()
	if err != nil {
		fmt.Printf("Error: %v.\n", err)
		return
	}

	// Declare empty variable here pre-loop to set and compare later
	seen := weather.SeenAlerts{}

	for {
		// Connect to NWS API to retrive alerts of type and zone as defined in our config
		alerts, err := weather.ConnectNOAA(client, alertsURL, cfg, *debug)
		if err != nil {
			fmt.Printf("Error: %s.\n", err)
			time.Sleep(60 * time.Second)
			continue
		}

		// Filter alerts returned for either printing or Pushover
		matches := weather.FilterAlerts(alerts, cfg)

		// Print alerts to console and then end program. Useful for seeing alerts without firing off notifications
		if *print {
			weather.PrintMatchingAlerts(matches)
			return
		}

		seen = weather.PruneSeenAlerts(seen)

		// Pushover logic - skip anything already notified about
		for _, p := range matches {
			if _, alreadySeen := seen[p.ID]; alreadySeen {
				continue
			}

			// Pushover call
			err := weather.SendPushover(client, apiKey, userKey, p)
			if err != nil {
				fmt.Printf("Error sending Pushover: %v.\n", err)
				continue
			}

			seen[p.ID] = p.Expires
		}

		time.Sleep(60 * time.Second)
	}
}
