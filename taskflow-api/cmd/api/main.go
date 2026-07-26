package main

import (
	"fmt"
	"net/http"
	"taskflow-api/internal/auth"
	"taskflow-api/internal/project"
	"taskflow-api/internal/task"
)

var secretKey = "mysecretkey"

func main() {
	storage := auth.NewStorage("data/users.json")
	projectStorage := project.NewStorage("data/projects.json")
	taskStorage := task.NewStorage("data/tasks.json")
	jwt := auth.NewJWTService(secretKey)
	service := auth.NewService(storage, jwt)
	projectService := project.NewService(projectStorage)
	taskService := task.NewService(taskStorage, projectService)
	authMiddleware := auth.NewAuthMiddleware(jwt)

	handler := auth.NewHandler(service)

	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("POST /register", handler.Register)
	mux.HandleFunc("POST /login", handler.Login)
	mux.Handle(
		"GET /me",
		authMiddleware.Authenticate(
			http.HandlerFunc(handler.Me),
		),
	)

	// Project routes
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

	// Task routes
	taskHandler := task.NewHandler(taskService)
	mux.Handle(
		"POST /projects/{projectID}/tasks",
		authMiddleware.Authenticate(
			http.HandlerFunc(taskHandler.CreateTask),
		),
	)
	mux.Handle(
		"DELETE /projects/{projectID}/tasks/{taskID}",
		authMiddleware.Authenticate(
			http.HandlerFunc(taskHandler.DeleteTask),
		),
	)
	mux.Handle(
		"GET /projects/{projectID}/tasks/{taskID}",
		authMiddleware.Authenticate(
			http.HandlerFunc(taskHandler.GetByID),
		),
	)
	mux.Handle(
		"GET /projects/{projectID}/tasks",
		authMiddleware.Authenticate(
			http.HandlerFunc(taskHandler.ListTasks),
		),
	)
	mux.Handle(
		"PUT /projects/{projectID}/tasks/{taskID}",
		authMiddleware.Authenticate(
			http.HandlerFunc(taskHandler.UpdateTask),
		),
	)
	mux.Handle(
		"PATCH /projects/{projectID}/tasks/{taskID}/status",
		authMiddleware.Authenticate(
			http.HandlerFunc(taskHandler.UpdateTaskStatus),
		),
	)

	fmt.Println("[Main] Server is Running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
