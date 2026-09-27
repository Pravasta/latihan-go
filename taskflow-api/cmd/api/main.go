package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"taskflow-api/internal/bootstrap"
	"time"
)

func main() {
	app, err := bootstrap.NewApp()
	if err != nil {
		log.Fatalf("Failed to bootstrap application: %v", err)
	}

	log.Printf("Starting server on %s", app.Server.Addr)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	defer stop()

	go func() {
		if err := app.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to run application: %v", err)
		}
	}()

	<-ctx.Done()

	log.Println("Shutting down gracefully, press Ctrl+C again to force")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)

	defer cancel()

	if err := app.Server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	if err := app.Close(); err != nil {
		log.Fatalf("Failed to close application: %v", err)
	}

	log.Println("Server exiting")
}

// To Run
// go run cmd/api/main.go
