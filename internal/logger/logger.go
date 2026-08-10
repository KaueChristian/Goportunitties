// Package logger builds the application's slog handler.
//
// The project used to ship a hand-rolled logger wrapping the standard log
// package. log/slog covers the same ground with levels and structured
// attributes, which is what makes logs searchable once the app runs somewhere
// other than a terminal.
package logger

import (
	"log/slog"
	"os"
)

// New returns the root logger and installs it as slog's default.
//
// Development gets human-readable text at debug level; production gets JSON at
// info level, which is what log aggregators expect to ingest.
func New(env string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}

	var handler slog.Handler = slog.NewTextHandler(os.Stdout, opts)
	if env == "production" {
		opts.Level = slog.LevelInfo
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	log := slog.New(handler)
	slog.SetDefault(log)
	return log
}
