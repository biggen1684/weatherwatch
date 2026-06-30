package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	weather "github.com/biggen1684/weatherwatch/api"
)

const alertsURL = "https://api.weather.gov/alerts/"
const zipURL = "https://api.zippopotam.us/us/"
const pointsURL = "https://api.weather.gov/points/"
const pushoverURL = "https://api.pushover.net/1/messages.json"

func main() {

	// Setup logger for output of errors to stderr
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	// Setup http client
	client := &http.Client{Timeout: 30 * time.Second}

	// Configure runtime flags
	zip := flag.String("zip", "", "Zip code to look up your NWS zone/county codes (e.g. -zip 32547)")
	listevents := flag.Bool("listevents", false, "List all valid NWS alert event types")
	debug := flag.Bool("debug", false, "Print raw API responses for troubleshooting")
	print := flag.Bool("print", false, "Print alerts matching your configured zone and events then exit")
	test := flag.Bool("test", false, "Checks Pushover settings are valid.")
	flag.Parse()

	// Test pushover api and user keys for connectivity
	if *test {
		apiKey, userKey, _, err := weather.PreRunSetup()
		if err != nil {
			fmt.Printf("Error: %v.\n", err)
			return
		}
		err = weather.SendPushoverTest(client, pushoverURL, apiKey, userKey)
		if err != nil {
			fmt.Printf("Error sending test notification: %v.\n", err)
			return
		}
		fmt.Println("Test notification sent successfully — check your Pushover app.")
		return
	}

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

	// Handle -print flag before starting daemon
	if *print {
		for _, loc := range cfg.Locations {
			alerts, err := weather.ConnectNOAA(client, alertsURL, loc, *debug)
			if err != nil {
				fmt.Printf("Error: %v.\n", err)
				return
			}
			matches := weather.FilterAlerts(alerts, loc, cfg.Events)
			fmt.Printf("--- %s ---\n", loc.Name)
			weather.PrintMatchingAlerts(matches)
		}
		return
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Go routine to catch SIGTERM OR SIGINT and send Pushover notification to user to let them know program is exiting
	go func() {
		sig := <-sigChan
		slog.Info("shutting down weatherwatch", "signal", sig)
		err := weather.SendPushoverShutdown(client, pushoverURL, apiKey, userKey)
		if err != nil {
			slog.Error("pushover shutdown notification failed", "error", err)

		}
		os.Exit(0)
	}()

	// Send startup alert once to let user know weatherwatch is running
	err = weather.SendPushoverStartup(client, pushoverURL, apiKey, userKey)
	if err != nil {
		slog.Error("pushover startup notification failed", "error", err)
	} else {
		slog.Info("starting up weatherwatch")
	}

	// Brief pause to allow startup notification to deliver before first poll
	time.Sleep(2 * time.Second)

	// Declare empty variable here pre-loop to set and compare later
	seen := weather.SeenAlerts{}

	// Main daemon loop
	for {
		// Iterate over each configured location
		for _, loc := range cfg.Locations {
			// Connect to NWS API to retrieve alerts for this location
			alerts, err := weather.ConnectNOAA(client, alertsURL, loc, *debug)
			if err != nil {
				slog.Error("connect to NOAA failed", "error", err, "location", loc.Name)
				continue
			}

			// Filter alerts returned for either printing or Pushover
			matches := weather.FilterAlerts(alerts, loc, cfg.Events)

			// Pushover logic - notify on new alerts or on existing alerts whose end time has been extended
			for _, p := range matches {
				// Build a dedup key unique to this location and alert event
				key := loc.Name + "." + weather.VtecKey(p)

				// Determine the effective end time for this alert (favors p.Ends over p.Expires)
				incomingExpiry := weather.EffectiveExpiry(p)

				// Skip if already notified and the end time hasn't extended further
				if !weather.ShouldNotify(seen, key, incomingExpiry) {
					continue
				}

				// Send alerts to Pushover
				err := weather.SendPushover(client, pushoverURL, apiKey, userKey, loc.Name, p)
				if err != nil {
					slog.Error("pushover notification failed", "error", err, "location", loc.Name)
					continue
				}

				// Update seen map with the latest known end time for this alert
				seen[key] = incomingExpiry

				// Log successful Pushover notification
				slog.Info("alert sent", "event", p.Event, "location", loc.Name, "headline", p.Headline, "dedup_key", key, "seen_until", incomingExpiry)

				// Marshal the full alert to JSON and write it to stdout for downstream consumers
				data, err := json.Marshal(p)
				if err != nil {
					slog.Error("json marshal failed", "error", err)
					continue
				}
				fmt.Println(string(data))
			}
		}

		// Exit after printing all locations
		if *print {
			return
		}

		// Remove alerts for entries that have expired
		seen = weather.PruneSeenAlerts(seen)

		// Sleep for 60 seconds before looping again
		time.Sleep(60 * time.Second)
	}
}
