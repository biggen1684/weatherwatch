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
			slog.Error("pre-run setup failed", "error", err)
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

		// Remove alerts for entries that have expired
		seen = weather.PruneSeenAlerts(seen)

		// Pushover logic - skip anything already notified about
		for _, p := range matches {
			key := weather.VtecKey(p)
			_, alreadySeen := seen[key]
			if alreadySeen {
				continue
			}

			// Send alerts to Pushover
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

			// Output the full alert as JSON to stdout for downstream consumers
			data, err := json.Marshal(p)
			if err != nil {
				slog.Error("json marshal failed", "error", err)
				continue
			}
			fmt.Println(string(data))
		}

		time.Sleep(60 * time.Second)
	}
}
