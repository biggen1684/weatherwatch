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
	debug := flag.Bool("debug", false, "Print raw API responses for troubleshooting")
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

	key, cfg, err := weather.PreRunSetup()
	if err != nil {
		fmt.Printf("Error: %v.\n", err)
		return
	}
	fmt.Println(key)
	fmt.Println(cfg)

	// debug := flag.Bool("debug", false, "print raw API response (use -debug to enable)")
	// flag.Parse()

	// alerts, err := weather.ConnectNOAA(client, alertsURL, *debug)
	// if err != nil {
	// 	fmt.Printf("Error: %s.\n", err)
	// 	return
	// }
	// fmt.Println(alerts)
}
