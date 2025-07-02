package logger

import (
	"os"

	kitlog "github.com/go-kit/log"
)

type Config struct {
	Service string
	Version string
}

// New creates a new structured logger using go-kit/log
func New(config Config) kitlog.Logger {
	// Using logfmt format, human readable and easy to parse by log aggregators like datadog, ELK stack etc.
	logger := kitlog.NewLogfmtLogger(os.Stderr)
	// Add timestamp with UTC timezone
	logger = kitlog.With(logger, "ts", kitlog.DefaultTimestampUTC)
	// Add caller information, which is the file and line number of the code that called the logger
	logger = kitlog.With(logger, "caller", kitlog.DefaultCaller)
	// Add service and version information
	logger = kitlog.With(logger, "service", config.Service, "version", config.Version)
	return logger
}
