package logger

import (
	"log/slog"
	"os"
	"strings"
	"taskflow-api/internal/config"
)

func NewLogger(
	cfg *config.Config,
) *slog.Logger {
	var level slog.Level

	switch strings.ToLower(cfg.Logger.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler

	if cfg.Logger.Format == "json" || cfg.Env.Environment == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)

	slog.SetDefault(logger)

	return logger
}
