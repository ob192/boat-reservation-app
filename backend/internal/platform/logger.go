package platform

import (
	"log/slog"
	"os"
)

// NewLogger returns a JSON slog.Logger writing to stdout.
// Level can be tuned later via env if needed.
func NewLogger(debug bool) *slog.Logger {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(handler)
}
