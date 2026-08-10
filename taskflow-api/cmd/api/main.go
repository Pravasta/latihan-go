package main

import (
	"fmt"
	"log"
	"taskflow-api/internal/bootstrap"
)

func main() {
	app, err := bootstrap.NewApp()
	if err != nil {
		log.Fatalf("Failed to bootstrap application: %v", err)
	}
	defer app.DB.Close()

	fmt.Printf("[Main] Server is Running on http://localhost%s\n", app.Server.Addr)
	if err := app.Server.ListenAndServe(); err != nil {
		panic(err)
	}
}

// To Run
// go run cmd/api/main.go
