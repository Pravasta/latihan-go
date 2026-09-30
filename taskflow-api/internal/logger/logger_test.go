package logger

import (
	"context"
	"log/slog"
	"taskflow-api/internal/config"
	"testing"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name          string
		cfg           *config.Config
		expectedLevel slog.Level
	}{
		{
			name: "Debug Level Development",
			cfg: &config.Config{
				Env: config.EnvConfig{Environment: "development"},
				Logger: config.LoggerConfig{
					Level:  "debug",
					Format: "text",
				},
			},
			expectedLevel: slog.LevelDebug,
		},
		{
			name: "Info Level Production JSON",
			cfg: &config.Config{
				Env: config.EnvConfig{Environment: "production"},
				Logger: config.LoggerConfig{
					Level:  "info",
					Format: "json",
				},
			},
			expectedLevel: slog.LevelInfo,
		},
		{
			name: "Warn Level",
			cfg: &config.Config{
				Env: config.EnvConfig{Environment: "development"},
				Logger: config.LoggerConfig{
					Level:  "warn",
					Format: "text",
				},
			},
			expectedLevel: slog.LevelWarn,
		},
		{
			name: "Error Level",
			cfg: &config.Config{
				Env: config.EnvConfig{Environment: "development"},
				Logger: config.LoggerConfig{
					Level:  "error",
					Format: "text",
				},
			},
			expectedLevel: slog.LevelError,
		},
		{
			name: "Default Level when unknown string given",
			cfg: &config.Config{
				Env: config.EnvConfig{Environment: "development"},
				Logger: config.LoggerConfig{
					Level:  "unknown",
					Format: "text",
				},
			},
			expectedLevel: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLogger(tt.cfg)

			if l == nil {
				t.Fatalf("expected logger to be non-nil")
			}

			// Verify if the handler level is enabled as expected
			ctx := context.Background()
			if !l.Handler().Enabled(ctx, tt.expectedLevel) {
				t.Errorf("expected level %v to be enabled", tt.expectedLevel)
			}

			// If level is info/warn/error, debug should be disabled
			if tt.expectedLevel > slog.LevelDebug {
				if l.Handler().Enabled(ctx, slog.LevelDebug) {
					t.Errorf("expected debug level to be disabled when level is %v", tt.expectedLevel)
				}
			}

			// Verify slog.Default() is set
			if slog.Default() != l {
				t.Errorf("expected slog.Default() to be set to the created logger")
			}
		})
	}
}
