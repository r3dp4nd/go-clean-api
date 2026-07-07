package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/r3dp4nd/go-clean-api/internal/config"
)

func New(cfg config.LogConfig) *slog.Logger {
	level := parseLevel(cfg.Level)

	options := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler

	switch strings.ToLower(strings.TrimSpace(cfg.Format)) {
	case "text":
		handler = slog.NewTextHandler(os.Stdout, options)
	default:
		handler = slog.NewJSONHandler(os.Stdout, options)
	}

	return slog.New(handler)
}

func parseLevel(value string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
