package main

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/term"
)

/*
Logging represents the configuration for logging behavior.

It includes the log level, format, and output destination.
*/
type Logging struct {
	Level  string `yaml:"level,omitempty" default:"debug"`   // Log level (e.g., "debug", "info")
	Format string `yaml:"format,omitempty" default:"text"`   // Log format ("text" or "json")
	Output string `yaml:"output,omitempty" default:"stdout"` // Log output destination ("stdout", "stderr", or file path)
}

/*
setup configures the logging based on the Logging configuration.

It sets the log level, determines the output destination,
enables or disables color output, and sets the log format.
*/
func (l *Logging) setup() {
	// Parse and set the log level
	if logLevel, err := zerolog.ParseLevel(l.Level); err != nil || l.Level == "" {
		log.Warn().Str("input", l.Level).Msg("Log level is unknown or empty, falling back to 'info' level")
		log.Logger = log.Level(zerolog.InfoLevel)
	} else {
		log.Logger = log.Level(logLevel)
	}

	// Determine the log output destination
	var writer io.Writer
	switch l.Output {
	case "stdout", "out", "1":
		writer = os.Stdout
	case "stderr", "err", "2":
		writer = os.Stderr
	default:
		file, err := os.OpenFile(l.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatal().Err(err).Str("output", l.Output).Msg("Failed to open log file")
		}
		writer = file
	}

	// Determine if color output is enabled
	var useColors bool
	if f, ok := writer.(*os.File); ok {
		// Check if the output is a terminal
		useColors = term.IsTerminal(int(f.Fd())) // #nosec G115
	} else {
		useColors = false
	}

	// Set the log format based on the configuration
	switch l.Format {
	case "text":
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: time.RFC3339,
			NoColor:    !useColors,
		})
	case "json":
		fallthrough
	default:
		log.Logger = log.Output(writer)
	}
}
