package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	weather "github.com/biggen1684/weatherwatch/api"
)

const alertsURL = "https://api.weather.gov/alerts/"
const zipURL = "https://api.zippopotam.us/us/"
const pointsURL = "https://api.weather.gov/points/"
const pushoverURL = "https://api.pushover.net/1/messages.json"

func main() {

	// Setup logger for output of logs
	logFile, err := weather.SetupLogger("weatherwatch.log")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer logFile.Close()

	client := &http.Client{Timeout: 30 * time.Second}

	zip := flag.String("zip", "", "Zip code to look up your NWS zone/county codes (e.g. -zip 32547)")
	listevents := flag.Bool("listevents", false, "List all valid NWS alert event types")
	debug := flag.Bool("debug", false, "Print raw API responses for troubleshooting")
	print := flag.Bool("print", false, "Print alerts matching your configured zone and events, then exit")
	flag.Parse()

	// Convert zip to lat/long if zip flag is sent in - End program after
	if *zip != "" {
		zone, county, err := weather.LookupZone(client, zipURL, pointsURL, *zip, *debug)
		if err != nil {
			fmt.Printf("Error: %v.\n", err)
			return
		}
		fmt.Printf("Your NWS Zone is: %s\n", zone)
		fmt.Printf("Your NWS County is: %s\n", county)
		fmt.Println("Add both to their respective fields in config.toml")
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
		slog.Error("pre-run setup failed", "error", err)
		return
	}

	// Declare empty variable here pre-loop to set and compare later
	seen := weather.SeenAlerts{}

	for {
		// Connect to NWS API to retrieve alerts of type and zone as defined in our config
		alerts, err := weather.ConnectNOAA(client, alertsURL, cfg, *debug)
		if err != nil {
			slog.Error("connect to NOAA failed", "error", err)
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
			key := weather.VtecKey(p)
			_, alreadySeen := seen[key]
			if alreadySeen {
				continue
			}

			err := weather.SendPushover(client, pushoverURL, apiKey, userKey, p)
			if err != nil {
				slog.Error("pushover notification failed", "error", err)
				continue
			}

			// Use the event end time if available, as it represents when the weather event actually ends.
			// p.Expires is when the NWS message version expires (often only a few hours),
			// while p.Ends is when the actual weather event ends (could be days away).
			// Falling back to p.Expires if p.Ends is zero (not set) or earlier than p.Expires.
			seenExpiry := p.Expires
			if !p.Ends.IsZero() && p.Ends.After(p.Expires) {
				seenExpiry = p.Ends
			}
			seen[key] = seenExpiry

			// Log either the p.Ends or p.Expires time in addition to other information
			slog.Info("alert sent", "event", p.Event, "headline", p.Headline, "dedup_key", key, "seen_until", seenExpiry)
		}

		time.Sleep(60 * time.Second)
	}
}
