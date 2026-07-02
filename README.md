# weatherwatch

A command-line daemon that polls the National Weather Service (NWS) API for active severe weather alerts in configurable NWS forecast zones sending push notifications via [Pushover](https://pushover.net) and structured JSON to stdout when a configured alert type is issued.

> ⚠️ **Disclaimer:** weatherwatch is a personal learning project intended for educational and recreational use only. It should not be relied upon as a primary source of severe weather alerts or for any life-safety decisions. Always monitor official sources such as the [National Weather Service](https://www.weather.gov), local emergency management agencies, and NOAA Weather Radio for authoritative, real-time severe weather information. The developer makes no guarantees regarding the accuracy, timeliness, or completeness of alerts delivered by this application.

## Features

- Polls `api.weather.gov` for active alerts at 60 second intervals
- Outputs the full matched alert as JSON to stdout for every new notification. Suitable for piping into other tools
- Monitors one or more locations simultaneously, each with their own NWS zone/county code and lat/lon
- Optionally filters alerts by NWS polygon geometry when `lat`/`lon` are configured, falling back to zone/county matching for alerts without polygon data
- Filters alerts by user configurable event type (e.g. Tornado Warning, Flash Flood Warning) — shared across all locations
- Sends push notifications to your device via Pushover when a new matching alert is found
- Avoids duplicate notifications using an in-memory cache — re-notifies only if NWS extends an active alert's expiry time
- Looks up your NWS zone/county code and lat/lon from a zip code (no need to know them ahead of time)
- Lists all valid NWS alert event types so you know what events to put in your config
- Structured logging to stderr, designed to run as a systemd service with automatic log capture via journald
- Sends a Pushover notification on startup and shutdown so you always know when the daemon is running or has stopped

### NWS Alerting Primer

The NWS alert system works at the zone and county level. When NWS issues an alert, they assign it to one or more forecast zones or county codes covering the affected area.  These are called UGC (Universal Geographic Code). Each code represents a specific forecast zone or county, formatted as a two-letter state abbreviation followed by a letter indicating the type (`Z` for forecast zone, `C` for county) and a three-digit number. For example, an alert for zip code 90210 would include these UGC codes:

- `CAZ368` — California (`CA`), forecast zone (`Z`), zone number `368` (Santa Monica Mountains)
- `CAC037` — California (`CA`), county (`C`), county FIPS code `037` (Los Angeles County)

> Note that NWS alerts don't consistently use one type of code.  Some alerts use forecast zone codes, others use county codes, and some include both. This is why weatherwatch requires both a zone and county code in your config to check for either as a match.

These geographic areas can be large — a watch may cover dozens of counties across multiple states, while a warning typically covers one or two counties.  Warnings are generally issued with storm polygon information indicating the size of the storm but the actual storm polygon may only affect a small portion of that area. For example, a Severe Thunderstorm Warning for a storm in the northern part of a county will still list the entire county in its UGC list, meaning anyone monitoring that county code would receive the alert even if the storm is 50 miles from their location.  However, the polygon data included with the same alert will give a precise position of the storm within the county/zone that is affected.  

weatherwatch with polygon filtering bridges the gap between getting alerts for the entire affected zone and only getting alerts when your configured `lat`/`lon` coordinates fall within the alert polygon.  In the above example, if you had configured `lat` and `lon` for your location, weatherwatch would only have notified you if your coordinates fell inside the storm polygon and not just because the storm was somewhere in your county.  This cuts down on notifications for storm or events that are not near your location.  

Not all alerts include polygon geometry — broad condition alerts like Heat Advisories, Tornado Watches, and Extreme Heat Warnings cover entire zones by nature and never include a polygon. For these alerts, weatherwatch always falls back to zone/county matching regardless of whether `lat`/`lon` are configured.

### Alert Filtering Logic

weatherwatch filters alerts using the following logic depending on whether the NWS alert includes polygon geometry and whether `lat`/`lon` are configured in the `config.toml` file.  Again, NWS does not always include polygon geometry with their alerts — storm-based warnings typically do, while watches and broad condition advisories typically do not.

| Alert has geometry | Location has `lat`/`lon` | Result |
|---|---|---|
| No | No | Notify if zone/county matches |
| No | Yes | Notify if zone/county matches |
| Yes | No | Notify if zone/county matches |
| Yes | Yes | Notify only if coordinates fall within polygon bounding box |

Zone/county matching is always the first filter — if your zone or county code isn't in the alert's list, the alert is skipped regardless of geometry. Polygon filtering is an additional precision filter applied on top of zone/county matching when both geometry and coordinates are available.  Configuring `lat` and `lon` is recommended for more precise alerting. Without `lat` and `lon` coordinates, weatherwatch notifies for any alert covering your entire zone or county — which can be large geographic areas. With coordinates, alerts are filtered to only those whose polygon actually covers your specific location, reducing notifications for storms or events that are in your county but far from you.

## Requirements

- Linux (binaries are built and tested for Linux only)
- A [Pushover](https://pushover.net) account (see [Pushover setup](#pushover-setup) below)

## Quick Setup

For anyone who just wants to get running fast — full details for each step are further down.

**1. Get the binary**

Download the latest Linux binary from the [Releases](https://github.com/biggen1684/weatherwatch/releases) page, or build from source:

```bash
git clone https://github.com/biggen1684/weatherwatch.git
cd weatherwatch
go build -o weatherwatch .
```

**2. Set your environment variables**

```bash
export PUSHOVER_API_KEY="your_app_token_here"
export PUSHOVER_USER_KEY="your_user_key_here"
export WEATHERWATCH_USER_AGENT="weatherwatch (you@example.com)"
```

**3. Configure your zone, county, and events in the config.toml file (lat/lon are optional)**

```bash
cp config.example.toml config.toml
./weatherwatch -zip <your_zip_code> # run once per location you want to monitor
./weatherwatch -listevents # see all valid event types
nano config.toml
```

**4. Run it**

```bash
./weatherwatch
```


## In-depth Installation and Configuration
 
### Pre-built binary (Linux)
 
Download the latest Linux binary from the [Releases](https://github.com/biggen1684/weatherwatch/releases) page.
 
> **Note:** weatherwatch is currently only tested and distributed for Linux. It relies on environment variables for configuration secrets and Windows handles these differently (`setx` or the System Properties GUI rather than `export`). Windows support hasn't been tested — building from source on Windows probably works fine, but isn't guaranteed.

### Build from source
 
Requires Go 1.21 or later (uses `log/slog`).
 
```bash
git clone https://github.com/biggen1684/weatherwatch.git
cd weatherwatch
go build -o weatherwatch .
```

### Pushover Setup

1. Create a Pushover account at [pushover.net](https://pushover.net)
2. Your **User Key** is shown on your dashboard after logging in.
3. Create an **Application** (also from the dashboard) to get an **API Token** — this becomes `PUSHOVER_API_KEY`.
4. Add these as Environment Variables (instructions below)
5. Run with the `-test` flag to verify your Pushover keys are configured correctly.

### Environment Variables

weatherwatch requires three environment variables to be set. These are not stored in `config.toml` since these are secrets.

| Variable | Purpose |
|---|---|
| `PUSHOVER_API_KEY` | Your Pushover Application Token |
| `PUSHOVER_USER_KEY` | Your Pushover User Key |
| `WEATHERWATCH_USER_AGENT` | A contact string sent to the NWS API, e.g. `weatherwatch (you@example.com)` |

>**Note:** NWS requires a `User-Agent` header identifying who's making requests in case they need to reach you about unusual traffic. Technically, you can put just about anything here including a fake email.  However, they will eventually require an API key per their own [authentication documentation](https://www.weather.gov/documentation/services-web-api) which can be included in this field in the future.

The location you store these three environment variables will depend on your chosen method of running the program:  

**1. Running directly (shell, screen, nohup):** Set these in your shell profile (`~/.bashrc`, `~/.profile`) so they persist across reboots, or export them before running for a one-off:

```bash
export PUSHOVER_API_KEY="your_app_token_here"
export PUSHOVER_USER_KEY="your_user_key_here"
export WEATHERWATCH_USER_AGENT="weatherwatch (you@example.com)"
```

**2. Running via systemd:**  See [Running Long-Term](#running-long-term) below.

### Configuration

Copy the example config and fill in your values:

```bash
cp config.example.toml config.toml
```

`config.toml` fields:
```toml
# Event types to notify on — shared across all locations
# Run with -listevents to see all valid event type strings
# Note: weatherwatch is designed for land-based alerts only
# NWS marine area codes are not supported
events = [
    "Tornado Warning",
    "Severe Thunderstorm Warning",
    "Flash Flood Warning",
    "Hurricane Warning"
]

# Add one [[locations]] block per area you want to monitor
# You must add at least one [[locations]] block
# Run with -zip <zipcode> to look up your zone, county, and lat/lon
# name must be unique per location block
[[locations]]
name = "Home"
area = "CA"
zone = "CAZ368"
county = "CAC037"
lat = 34.0901   # optional — enables polygon-based filtering for more precise alerts
lon = -118.4065  # optional — enables polygon-based filtering for more precise alerts

# Add additional locations as needed
[[locations]]
name = "Vacation Home"
area = "AL"
zone = "ALZ043"
county = "ALC051"
# lat and lon omitted — falls back to zone/county matching only
```

| Field | Description |
|---|---|
| `events` | List of NWS event type strings to notify on — shared across all locations |
| `name` | A label for this location — appears in Pushover notification titles e.g. `[Home] Tornado Warning` |
| `area` | Two-letter state abbreviation (e.g. `FL`, `AL`, `GA`) |
| `zone` | NWS forecast zone code — run `-zip` to find yours |
| `county` | NWS county code — run `-zip` to find yours |
| `lat` | *(optional)* Latitude of your location — run `-zip` to find yours. Enables polygon-based filtering when combined with `lon`.  Should give more precision for alerting. |
| `lon` | *(optional)* Longitude of your location — run `-zip` to find yours. Enables polygon-based filtering when combined with `lat`.  Should give more precision for alerting. |

## Finding Your Zone

If you don't know your NWS zone/county code or lat/lon, run weatherwatch with the `-zip` flag and your zip code:

```bash
./weatherwatch -zip 90210
```

Run `-zip` once for each location you want to monitor — each location requires its own zone and county code while lat/lon is optional.

This feature looks up the latitude/longitude for that zip, queries the NWS API, and prints your zone/county codes along with your lat/lon. You have to add both the zone and county codes to `zone` and `county` fields in `config.toml`. Lat/lon is optional but should give more precision for alerts that are closer to your location.

`-zip` only supports zip codes for the 50 US states. US territories such as Puerto Rico, Guam, and the US Virgin Islands are not supported. Users in these areas would need to look up their NWS zone and county codes manually from the [NWS website](https://www.weather.gov) and enter them directly into `config.toml`.

> **Note:** zip-code-to-coordinate lookups use the geographic centroid of the zip code's boundary. For zip codes covering narrow areas like barrier islands, this can occasionally resolve to a marine zone instead of land. weatherwatch will warn you if this happens — try another nearby zip code if this happens.

## Listing Valid Events

To see every alert event type the NWS API recognizes:

```bash
./weatherwatch -listevents
```

Copy whichever event names are relevant to you into the `events` array in `config.toml`. Event names must match exactly (including capitalization). Each event must be inside quotes and comma separated.

## Usage

### Running manually

```bash
./weatherwatch
```

Runs continuously, polling NWS every 60 seconds and sending Pushover notifications for new matching alerts. Use Ctrl-C to stop.

### Flags

| Flag | Description |
|---|---|
| `-zip <zipcode>` | Look up your NWS zone/county codes and lat/lon from a zip code then exit |
| `-listevents` | Print all valid NWS alert event type strings, then exit |
| `-print` | Fetch alerts, print any matching your config, then exit (no notifications sent) |
| `-debug` | Print raw API responses for troubleshooting |
| `-test` | Send a test Pushover notification to verify your API and User keys are configured correctly, then exit |

`-zip`, `-listevents`, `-print`, and `-test` are one-shot utility commands — none of them start the long-running daemon loop.

### Running Long-Term

weatherwatch is designed to run continuously. Since it's a single process holding state in memory (which alerts have already been notified about), it needs to keep running rather than being re-invoked periodically via cron or manually.

A few options in increasing order of robustness:

A. **`screen` or `tmux`** — good for quick/manual use on a machine you're logged into directly. Needs `screen` installed:

```bash
screen -S weatherwatch
./weatherwatch
# Ctrl+A then d to detach — output continues to accumulate in the screen buffer
```

B. **`nohup`** — survives terminal closure, doesn't survive a reboot:

```bash
nohup ./weatherwatch &
# nohup redirects stdout to nohup.out by default
```

C. **`systemd` (recommended for long-term/unattended use)** — survives reboots and restarts automatically on crash. A service file is included as `weatherwatch.service`.

1. Edit the paths and username in `weatherwatch.service` to match your deployment

2. Create a `.env` file alongside the binary with your three environment variables:

```bash
   nano /home/user/weatherwatch/.env
```

   Paste in your values:

```bash
   PUSHOVER_API_KEY=your_app_token_here
   PUSHOVER_USER_KEY=your_user_key_here
   WEATHERWATCH_USER_AGENT=weatherwatch (you@example.com)
```

   Restrict its permissions since it contains secrets:

```bash
   chmod 600 .env
```

   The `weatherwatch.service` service file references this `.env` file with `EnvironmentFile=`.

3. Install and start the service:

```bash
   sudo cp weatherwatch.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable --now weatherwatch
```

4. Check status and logs:

```bash
   sudo systemctl status weatherwatch
   journalctl -u weatherwatch -f
```

## Logging

weatherwatch writes structured logs to **stderr** using Go's `log/slog`. Matched alert data is written to **stdout** as JSON (see [JSON Output](#json-output) below). Keeping these on separate streams means you can pipe or redirect the alert data independently from the logs. This is separate from the one-shot utility flags (`-zip`, `-listevents`, `-print`), which print directly to the console since they're interactive commands.

When run directly or under `screen`/`nohup`, both streams appear interleaved in the terminal by default. When run under systemd, journald captures both stdout and stderr automatically — view combined logs with `journalctl -u weatherwatch -f`. journald also handles log rotation and retention on its own, so no manual log management is needed.

Logged events:

- Startup/configuration failures
- Location configuration summary on startup (name, zone, county, and precision mode — polygon or zone/county)
- Failed connections to the NWS API
- Failed Pushover notifications
- Successfully sent notifications (event type, location, headline, VTEC dedup key, event expiration)
- Re-notifications when NWS extends an active alert's expiry (logged with updated `seen_until` value)

## JSON Output

For every new (non-duplicate) matching alert, weatherwatch writes the full alert object as a single line of JSON to stdout, alongside sending the Pushover notification.  When available, the NWS alert polygon geometry is also included in the JSON output. This feature makes it easy to pipe weatherwatch's output into other tools:

```bash
./weatherwatch | jq '.headline'
```

Since logs go to stderr and alert data goes to stdout, you can capture them separately:

```bash
./weatherwatch > alerts.jsonl 2> weatherwatch.log
```

Each JSON line contains the full decoded NWS alert — event type, severity, headline, description, affected zones, VTEC parameters, and timing fields (`onset`, `expires`, `ends`).

## Project structure

```
weatherwatch/
├── main.go                # Entry point, flag parsing, daemon loop
├── config.example.toml    # Template config — copy to config.toml
├── weatherwatch.service   # systemd service file
├── LICENSE                # MIT
└── api/                   # weather package
    ├── config.go           # Config struct, loading, validation
    ├── env.go               # Environment variable helpers
    ├── zip.go               # Zip code validation and geocoding
    ├── zone.go              # Zip-to-NWS-zone lookup
    ├── alerts.go            # NWS alert fetching, filtering, seen-alert tracking
    ├── pushover.go          # Pushover notification sending
```

## License

MIT
