package weather

import (
	"fmt"
	"log/slog"
	"os"
)

// Create the .log file and append logs to it.  Open it for writing only
func SetupLogger() (*os.File, error) {
	logFile, err := os.OpenFile("weatherwatch.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("could not open log file: %v", err)
	}

	// Create handler - use default formatting
	logger := slog.New(slog.NewTextHandler(logFile, nil))

	// Default logger for entire program
	slog.SetDefault(logger)

	return logFile, nil
}
