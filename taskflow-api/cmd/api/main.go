package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"taskflow-api/internal/bootstrap"
	"time"
)

func main() {
	app, err := bootstrap.NewApp()
	if err != nil {
		slog.Error("Failed to initialize application", "error", err)
		os.Exit(1)
	}

	slog.Info("Application initialized successfully")

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	defer stop()

	go func() {
		if err := app.Run(); err != nil && err != http.ErrServerClosed {
			app.Logger.Error("Failed to run application", "error", err)
		}
	}()

	<-ctx.Done()

	app.Logger.Info("Shutting down gracefully, press Ctrl+C again to force")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)

	defer cancel()

	if err := app.Server.Shutdown(shutdownCtx); err != nil {
		app.Logger.Error("Server forced to shutdown", "error", err)
	}

	if err := app.Close(); err != nil {
		app.Logger.Error("Failed to close application", "error", err)
	}

	app.Logger.Info("Server exiting")
}

// To Run
// go run cmd/api/main.go
