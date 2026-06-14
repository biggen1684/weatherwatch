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

	zip := flag.String("zip", "", "Zip code to look up your NWS office (e.g. -zip 32547)")
	flag.Parse()

	if *zip != "" {
		office, err := weather.LookupOffice(client, zipURL, pointsURL, *zip)
		if err != nil {
			fmt.Printf("Error: %v.\n", err)
			return
		}
		fmt.Println("Your NWS office:", office)
		fmt.Println("Add this to the 'office' field in config.toml")
		return
	}

	key, cfg, err := weather.PreRunSetup()
	if err != nil {
		fmt.Printf("Error: %v.\n", err)
		return
	}
	fmt.Println(key)
	fmt.Println(cfg)

	// // Check if Pushover environment variable is set
	// _, err := weather.GetPushoverKey()
	// if err != nil {
	// 	fmt.Printf("Error: %v.\n", err)
	// 	return
	// }

	// // Load config file
	// cfg, err := weather.LoadConfig("config.toml")
	// if err != nil {
	// 	fmt.Printf("Error: %v.\n", err)
	// 	return
	// }

	// // Validate config file
	// err = weather.ValidateConfig(cfg)
	// if err != nil {
	// 	fmt.Printf("Error: %v.\n", err)
	// 	return
	// }

	// debug := flag.Bool("debug", false, "print raw API response (use -debug to enable)")
	// flag.Parse()

	// alerts, err := weather.ConnectNOAA(client, alertsURL, *debug)
	// if err != nil {
	// 	fmt.Printf("Error: %s.\n", err)
	// 	return
	// }
	// fmt.Println(alerts)
}
