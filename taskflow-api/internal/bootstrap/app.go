package bootstrap

import (
	"net/http"
	"taskflow-api/internal/auth"
	"taskflow-api/internal/config"
	"taskflow-api/internal/project"
	"taskflow-api/internal/repository/postgres"
	"taskflow-api/internal/task"
)

type App struct {
	Config     *config.Config
	DB         *postgres.DB
	JWTService *auth.JWTService
	Server     *http.Server
}

func NewApp() (*App, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, err
	}

	db, err := postgres.ConnectDB(cfg)
	if err != nil {
		return nil, err
	}

	jwtService := auth.NewJWTService(cfg)

	// [NOW - UNUSED]
	// storage := auth.NewStorage("data/users.json")
	// projectStorage := project.NewStorage("data/projects.json")
	// taskStorage := task.NewStorage("data/tasks.json")

	// REPOSITORIES
	userRepository := postgres.NewAuthRepository(db)
	projectRepository := postgres.NewProjectRepository(db)
	taskRepository := postgres.NewTaskRepository(db)

	// SERVICES
	authService := auth.NewService(userRepository, jwtService)
	projectService := project.NewService(projectRepository, taskRepository)
	taskService := task.NewService(taskRepository, projectRepository)
	authMiddleware := auth.NewAuthMiddleware(jwtService)

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

	addr := ":" + cfg.Server.Port

	return &App{
		Config:     cfg,
		DB:         db,
		JWTService: jwtService,
		Server:     &http.Server{Addr: addr, Handler: mux},
	}, nil
}
