# weatherwatch

A command-line daemon that polls the National Weather Service (NWS) API for active severe weather alerts in a specific zone and sends push notifications via [Pushover](https://pushover.net) when a configured alert type is issued.

> ⚠️ **Disclaimer:** weatherwatch is a personal learning project intended for educational and recreational use only. It should not be relied upon as a primary source of severe weather alerts or for any life-safety decisions. Always monitor official sources such as the [National Weather Service](https://www.weather.gov), local emergency management agencies, and NOAA Weather Radio for authoritative, real-time severe weather information. The developer makes no guarantees regarding the accuracy, timeliness, or completeness of alerts delivered by this application.

## Features

- Polls `api.weather.gov` for active alerts at 60 second intervals
- Filters alerts by user configurable NWS zone/county code and event type (e.g. Tornado Warning, Flash Flood Warning)
- Sends push notifications to your phone via Pushover when a new matching alert is found
- Avoids duplicate notifications for alerts already seen, using an in-memory cache with automatic expiration
- Looks up your NWS zone/county code from a zip code (no need to know them ahead of time)
- Lists all valid NWS alert event types so you know what events to put in your config
- Structured logging to a file for unattended/daemon operation

## Requirements
 
- Linux (binaries are built and tested for Linux only)
- A [Pushover](https://pushover.net) account (free tier supports 10,000 messages/month)
- A Pushover Application Token and User Key (see [Pushover setup](#pushover-setup) below)


## Installation
 
### Pre-built binary (Linux)
 
Download the latest Linux binary from the [Releases](https://github.com/biggen1684/weatherwatch/releases) page.
 
> **Note:** weatherwatch is currently only tested and distributed for Linux. It relies on environment variables for configuration secrets and Windows handles these differently (`setx` or the System Properties GUI rather than `export`). Windows support hasn't been tested — building from source on Windows may work, but isn't guaranteed.

### Build from source
 
Requires Go 1.21 or later (uses `log/slog`).
 
```bash
git clone https://github.com/biggen1684/weatherwatch.git
cd weatherwatch
go build -o weatherwatch .
```

## Environment Variables

weatherwatch requires three environment variables to be set. These are not stored in `config.toml` since they're either secrets or identifiers tied to your specific deployment.

| Variable | Purpose |
|---|---|
| `PUSHOVER_API_KEY` | Your Pushover Application Token |
| `PUSHOVER_USER_KEY` | Your Pushover User Key |
| `WEATHERWATCH_USER_AGENT` | A contact string sent to the NWS API, e.g. `weatherwatch (you@example.com)` |

NWS requires a `User-Agent` header identifying who's making requests in case they need to reach you about unusual traffic. Use your own email — don't reuse a placeholder.

Set them in your shell profile (`~/.bashrc`, `~/.profile`) so that they are persistant and will survive across reboots or export them before running for a one off:

```bash
export PUSHOVER_API_KEY="your_app_token_here"
export PUSHOVER_USER_KEY="your_user_key_here"
export WEATHERWATCH_USER_AGENT="weatherwatch (you@example.com)"
```

### Pushover Setup

1. Create a free account at [pushover.net](https://pushover.net)
2. Your **User Key** is shown on your dashboard after logging in
3. Create an **Application** (also from the dashboard) to get an **API Token** — this becomes `PUSHOVER_API_KEY`

## Configuration

Copy the example config and fill in your values:

```bash
cp config.example.toml config.toml
```

`config.toml` fields:

```toml
# Two-letter state abbreviation (or NWS marine area code)
area = "CA"

# Your NWS forecast zone code — see "Finding Your Zone" below
zone = "CAZ368"

# Your NWS forecast county code — see "Finding Your Zone" below
county = "CAC037"

# Event types to notify on — see "Listing Valid Events" below
events = [
    "Tornado Warning",
    "Severe Thunderstorm Warning",
    "Flash Flood Warning",
    "Hurricane Warning"
]
```

### Finding Your Zone

If you don't know your NWS zone code, run weatherwatch with the `-zip` flag and your zip code:

```bash
./weatherwatch -zip 90210
```

This looks up the latitude/longitude for that zip, queries the NWS API, and prints your zone/county codes. You have to add both the zone and county codes to `zone` and `county` fields in `config.toml`.

> **Note:** zip-code-to-coordinate lookups use the geographic centroid of the zip code's boundary. For zip codes covering narrow areas like barrier islands, this can occasionally resolve to a marine zone instead of land. weatherwatch will warn you if this happens — try another nearby zip code if this happens.

### Listing Valid Events

To see every alert event type the NWS API recognizes:

```bash
./weatherwatch -listevents
```

Copy whichever event names are relevant to you into the `events` array in `config.toml`. Event names must match exactly (including capitalization).

## Usage

### Normal operation (daemon mode)

```bash
./weatherwatch
```

Runs continuously, polling NWS every 60 seconds, sending Pushover notifications for new matching alerts. Designed to run in the background — see [Running Long-Term](#running-long-term) below.

### Flags

| Flag | Description |
|---|---|
| `-zip <zipcode>` | Look up your NWS zone/county codes from a zip code, then exit |
| `-listevents` | Print all valid NWS alert event type strings, then exit |
| `-print` | Fetch alerts, print any matching your config, then exit (no notifications sent) |
| `-debug` | Print raw API responses for troubleshooting |

`-zip`, `-listevents`, and `-print` are one-shot utility commands — none of them start the long-running daemon loop.

## Running Long-Term

weatherwatch is designed to run continuously. Since it's a single process holding state in memory (which alerts have already been notified about), it needs to keep running rather than being re-invoked periodically via cron.

A few options, in increasing order of robustness:

**`screen` or `tmux`** — good for quick/manual use on a machine you're logged into directly. Need `screen` installed of course:
```bash
screen -S weatherwatch
./weatherwatch
# Ctrl+A then d to detach
```

**`nohup`** — survives terminal closure, doesn't survive a reboot:
```bash
nohup ./weatherwatch &
```

**`systemd`** (recommended for long-term/unattended use) — survives reboots and restarts automatically on crash. *(Service file coming soon.)*

## Logging

weatherwatch writes structured logs to `weatherwatch.log` in its working directory, using Go's `log/slog`. This is separate from the one-shot utility flags (`-zip`, `-listevents`, `-print`), which print directly to the console since they're interactive commands.

Logged events:
- Startup/configuration failures
- Failed connections to the NWS API
- Failed Pushover notifications
- Successfully sent notifications (event type, headline, alert ID)
- No log rotation currently. You must manually delete/prune. (Coming soon!)

### Project structure

```
weatherwatch/
├── main.go              # Entry point, flag parsing, daemon loop
├── config.example.toml  # Template config — copy to config.toml
└── api/                 # weather package
    ├── config.go         # Config struct, loading, validation
    ├── env.go             # Environment variable helpers
    ├── zip.go             # Zip code validation and geocoding
    ├── zone.go            # Zip-to-NWS-zone lookup
    ├── alerts.go          # NWS alert fetching, filtering, seen-alert tracking
    ├── pushover.go        # Pushover notification sending
    └── log.go             # Logger setup
```

## License

MIT
