package main

import (
	"fmt"
	"net/http"
	"taskflow-api/internal/auth"
	"taskflow-api/internal/project"
)

var secretKey = "mysecretkey"

func main() {
	storage := auth.NewStorage("data/users.json")
	projectStorage := project.NewStorage("data/projects.json")
	jwt := auth.NewJWTService(secretKey)
	service := auth.NewService(storage, jwt)
	projectService := project.NewService(projectStorage)
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

	projectHandler := project.NewHandler(projectService)
	mux.Handle(
		"POST /projects",
		authMiddleware.Authenticate(
			http.HandlerFunc(projectHandler.CreateProject),
		),
	)
	mux.Handle(
		"GET /projects",
		authMiddleware.Authenticate(
			http.HandlerFunc(projectHandler.ListProjects),
		),
	)
	mux.Handle(
		"GET /projects/{id}",
		authMiddleware.Authenticate(
			http.HandlerFunc(projectHandler.GetProject),
		),
	)
	mux.Handle(
		"PUT /projects/{id}",
		authMiddleware.Authenticate(
			http.HandlerFunc(projectHandler.UpdateProject),
		),
	)
	mux.Handle(
		"DELETE /projects/{id}",
		authMiddleware.Authenticate(
			http.HandlerFunc(projectHandler.DeleteProject),
		),
	)

	fmt.Println("[Main] Server is Running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
