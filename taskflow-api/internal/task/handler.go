package task

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"taskflow-api/internal/common"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// errStatusCode maps a known service error to the HTTP status it should
// produce. Any error not listed here is treated as unexpected and becomes a
// 500, logged with the real cause instead of exposed to the client.
var errStatusCode = map[error]int{
	ErrInvalidOwnerID:              http.StatusBadRequest,
	ErrInvalidProjectID:            http.StatusBadRequest,
	ErrInvalidTaskID:               http.StatusBadRequest,
	ErrInvalidTaskTitle:            http.StatusBadRequest,
	ErrInvalidTaskStatus:           http.StatusBadRequest,
	ErrInvalidTaskStatusTransition: http.StatusBadRequest,
	ErrInvalidSortField:            http.StatusBadRequest,
	ErrInvalidOrder:                http.StatusBadRequest,
	ErrInvalidPageNumber:           http.StatusBadRequest,
	ErrInvalidLimitNumber:          http.StatusBadRequest,
	ErrProjectNotFound:             http.StatusNotFound,
	ErrTaskNotFound:                http.StatusNotFound,
}

func writeServiceError(w http.ResponseWriter, err error) {
	if status, ok := errStatusCode[err]; ok {
		common.WriteError(w, status, err.Error())
		return
	}
	slog.Error("unhandled task service error", "error", err)
	common.WriteError(w, http.StatusInternalServerError, "Internal server error")
}

func requireOwnerID(w http.ResponseWriter, r *http.Request) (string, bool) {
	ownerID, ok := common.GetUserID(r.Context())
	if !ok {
		common.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return "", false
	}
	return ownerID, true
}

func (h *Handler) CreateTask(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req CreateTaskRequest
	var res CreateTaskResponse
	var defaultRes DefaultResponse
	projectIdStr := r.PathValue("projectID")

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ownerID, ok := requireOwnerID(w, r)
	if !ok {
		return
	}

	task, err := h.service.Create(ownerID, projectIdStr, req.Title, req.Description)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	res = CreateTaskResponse{
		ID:          task.ID,
		ProjectID:   task.ProjectID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}

	defaultRes.Message = "Task created successfully"
	defaultRes.Data = res

	common.WriteJSON(w, http.StatusCreated, defaultRes)
}

func (h *Handler) DeleteTask(
	w http.ResponseWriter,
	r *http.Request,
) {
	var defaultRes DefaultResponse
	var projectID = r.PathValue("projectID")
	var taskID = r.PathValue("taskID")

	ownerID, ok := requireOwnerID(w, r)
	if !ok {
		return
	}

	if err := h.service.Delete(ownerID, projectID, taskID); err != nil {
		writeServiceError(w, err)
		return
	}

	defaultRes.Message = "Task deleted successfully"
	common.WriteJSON(w, http.StatusOK, defaultRes)
}

func (h *Handler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	var (
		defaultRes DefaultResponse
		taskID     = r.PathValue("taskID")
		projectID  = r.PathValue("projectID")
	)

	ownerID, ok := requireOwnerID(w, r)
	if !ok {
		return
	}

	task, err := h.service.GetByID(ownerID, projectID, taskID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	defaultRes.Message = "Task retrieved successfully"
	defaultRes.Data = task
	common.WriteJSON(w, http.StatusOK, defaultRes)
}

func (h *Handler) ListTasks(
	w http.ResponseWriter,
	r *http.Request,
) {
	var (
		defaultRes DefaultResponse
		projectID  = r.PathValue("projectID")
		limitStr   = r.URL.Query().Get("limit")
		pageStr    = r.URL.Query().Get("page")
		taskQuery  = TaskQuery{
			Limit: 10, // Default limit
			Page:  1,  // Default page
		}
		search = r.URL.Query().Get("search")
		sort   = r.URL.Query().Get("sort")
		order  = r.URL.Query().Get("order")
		status = r.URL.Query().Get("status")
	)

	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			common.WriteError(w, http.StatusBadRequest, "Invalid limit number")
			return
		}
		taskQuery.Limit = limit
	}

	if pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			common.WriteError(w, http.StatusBadRequest, "Invalid page number")
			return
		}
		taskQuery.Page = page
	}

	if status != "" {
		taskQuery.Status = TaskStatus(status)
	}

	if search != "" {
		taskQuery.Search = search
	}

	if sort != "" {
		taskQuery.Sort = sort
	}

	if order != "" {
		taskQuery.Order = Order(order)
	}

	ownerID, ok := requireOwnerID(w, r)
	if !ok {
		return
	}

	tasks, err := h.service.List(ownerID, projectID, taskQuery)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	defaultRes.Message = "Tasks retrieved successfully"
	defaultRes.Data = tasks.Tasks
	defaultRes.Meta = tasks.PaginationMeta
	common.WriteJSON(w, http.StatusOK, defaultRes)
}

func (h *Handler) UpdateTask(
	w http.ResponseWriter,
	r *http.Request,
) {
	var (
		req        UpdateTaskRequest
		res        UpdateTaskResponse
		defaultRes DefaultResponse
		projectID  = r.PathValue("projectID")
		taskID     = r.PathValue("taskID")
	)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ownerID, ok := requireOwnerID(w, r)
	if !ok {
		return
	}

	task, err := h.service.Update(ownerID, projectID, taskID, req.Title, req.Description)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	res = UpdateTaskResponse{
		ID:          task.ID,
		ProjectID:   task.ProjectID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}

	defaultRes.Message = "Task updated successfully"
	defaultRes.Data = res

	common.WriteJSON(w, http.StatusOK, defaultRes)
}

func (h *Handler) UpdateTaskStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	var (
		req        UpdateTaskStatusRequest
		res        UpdateTaskResponse
		defaultRes DefaultResponse
		projectID  = r.PathValue("projectID")
		taskID     = r.PathValue("taskID")
	)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ownerID, ok := requireOwnerID(w, r)
	if !ok {
		return
	}

	task, err := h.service.UpdateStatus(ownerID, projectID, taskID, req.Status)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	res = UpdateTaskResponse{
		ID:          task.ID,
		ProjectID:   task.ProjectID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}

	defaultRes.Message = "Task status updated successfully"
	defaultRes.Data = res

	common.WriteJSON(w, http.StatusOK, defaultRes)
}
