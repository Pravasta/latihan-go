package main

import (
	"fmt"
	"net/http"
	"taskflow-api/internal/auth"
)

var secretKey = "mysecretkey"

func main() {
	storage := auth.NewStorage("data/users.json")
	jwt := auth.NewJWTService(secretKey)
	service := auth.NewService(storage, jwt)
	authMiddleware := auth.NewAuthMiddleware(jwt)

	handler := auth.NewHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", handler.Register)
	mux.HandleFunc("POST /login", handler.Login)
	mux.Handle(
		"GET /me",
		authMiddleware.Authenticate(
			http.HandlerFunc(handler.Me),
		),
	)

	fmt.Println("[Main] Server is Running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
