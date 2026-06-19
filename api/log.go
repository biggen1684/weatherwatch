package weather

import (
	"log/slog"
	"os"
)

// SetupLogger configures structured logging to stdout, which systemd/journald
// captures automatically when running as a service
func SetupLogger() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}
