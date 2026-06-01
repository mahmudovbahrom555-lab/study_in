// Package logger предоставляет структурированный логгер на базе slog.
// В development использует читаемый текстовый формат, в production — JSON.
package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New создаёт новый логгер на основе уровня и формата.
//
// level: debug | info | warn | error
// format: text | json
func New(level, format string) *slog.Logger {
	var slogLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	var handler slog.Handler
	if strings.ToLower(format) == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
