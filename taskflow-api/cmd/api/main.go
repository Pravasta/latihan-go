package main

import (
	"fmt"
	"log"
	"net/http"
	"taskflow-api/internal/auth"
	"taskflow-api/internal/project"
	"taskflow-api/internal/repository/postgres"
	"taskflow-api/internal/task"

	"github.com/joho/godotenv"
)

var secretKey = "mysecretkey"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("[Main] No .env file found, relying on environment variables")
	}

	// postgres db
	db, err := postgres.NewDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// [NOW - UNUSED]
	// storage := auth.NewStorage("data/users.json")
	// projectStorage := project.NewStorage("data/projects.json")
	// taskStorage := task.NewStorage("data/tasks.json")

	// REPOSITORIES
	userRepository := postgres.NewAuthRepository(db)
	projectRepository := postgres.NewProjectRepository(db)
	taskRepository := postgres.NewTaskRepository(db)

	// SERVICES
	jwt := auth.NewJWTService(secretKey)
	authService := auth.NewService(userRepository, jwt)
	projectService := project.NewService(projectRepository, taskRepository)
	taskService := task.NewService(taskRepository, projectRepository)
	authMiddleware := auth.NewAuthMiddleware(jwt)

	handler := auth.NewHandler(authService)

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

// To Run
// go run cmd/api/main.go
