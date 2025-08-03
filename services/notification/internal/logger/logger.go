package logger

import (
	"log/slog"
	"os"
)

// NewLogger initializes and returns a structured logger using slog.
// It outputs JSON-formatted logs to stdout, suitable for production.
func NewLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
	})
	logger := slog.New(handler)
	return logger
}
