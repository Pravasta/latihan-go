package task

import (
	"encoding/json"
	"net/http"
	"taskflow-api/internal/common"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
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

	ownerID, ok := common.GetUserID(r.Context())
	if !ok {
		common.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	task, err := h.service.Create(ownerID, projectIdStr, req.Title, req.Description)
	if err != nil {
		switch err {
		case ErrInvalidOwnerID, ErrInvalidProjectID, ErrInvalidTaskTitle:
			defaultRes.Message = err.Error()
			common.WriteError(w, http.StatusBadRequest, defaultRes.Message)
		case ErrProjectNotFound:
			defaultRes.Message = err.Error()
			common.WriteError(w, http.StatusNotFound, defaultRes.Message)
		default:
			defaultRes.Message = "Internal server error"
			common.WriteError(w, http.StatusInternalServerError, defaultRes.Message)
		}
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

	ownerID, ok := common.GetUserID(r.Context())
	if !ok {
		common.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err := h.service.Delete(ownerID, projectID, taskID)
	if err != nil {
		switch err {
		case ErrInvalidOwnerID, ErrInvalidProjectID, ErrInvalidTaskID:
			defaultRes.Message = err.Error()
			common.WriteError(w, http.StatusBadRequest, defaultRes.Message)
		case ErrProjectNotFound, ErrTaskNotFound:
			defaultRes.Message = err.Error()
			common.WriteError(w, http.StatusNotFound, defaultRes.Message)
		default:
			defaultRes.Message = "Internal server error"
			common.WriteError(w, http.StatusInternalServerError, defaultRes.Message)
		}
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

	ownerID, ok := common.GetUserID(r.Context())
	if !ok {
		common.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	task, err := h.service.GetByID(ownerID, projectID, taskID)
	if err != nil {
		switch err {
		case ErrInvalidOwnerID, ErrInvalidProjectID, ErrInvalidTaskID:
			defaultRes.Message = err.Error()
			common.WriteError(w, http.StatusBadRequest, defaultRes.Message)
		case ErrProjectNotFound, ErrTaskNotFound:
			defaultRes.Message = err.Error()
			common.WriteError(w, http.StatusNotFound, defaultRes.Message)
		default:
			defaultRes.Message = "Internal server error"
			common.WriteError(w, http.StatusInternalServerError, defaultRes.Message)
		}
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
	)

	ownerID, ok := common.GetUserID(r.Context())
	if !ok {
		common.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tasks, err := h.service.List(ownerID, projectID)
	if err != nil {
		switch err {
		case ErrInvalidOwnerID, ErrInvalidProjectID:
			defaultRes.Message = err.Error()
			common.WriteError(w, http.StatusBadRequest, defaultRes.Message)
		case ErrProjectNotFound:
			defaultRes.Message = err.Error()
			common.WriteError(w, http.StatusNotFound, defaultRes.Message)
		default:
			defaultRes.Message = "Internal server error"
			common.WriteError(w, http.StatusInternalServerError, defaultRes.Message)
		}
		return
	}

	defaultRes.Message = "Tasks retrieved successfully"
	defaultRes.Data = tasks
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

	ownerID, ok := common.GetUserID(r.Context())
	if !ok {
		common.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	task, err := h.service.Update(ownerID, projectID, taskID, req.Title, req.Description)
	if err != nil {
		switch err {
		case ErrInvalidOwnerID, ErrInvalidProjectID, ErrInvalidTaskID:
			defaultRes.Message = err.Error()
			common.WriteError(w, http.StatusBadRequest, defaultRes.Message)
		case ErrProjectNotFound, ErrTaskNotFound:
			defaultRes.Message = err.Error()
			common.WriteError(w, http.StatusNotFound, defaultRes.Message)
		default:
			defaultRes.Message = "Internal server error"
			common.WriteError(w, http.StatusInternalServerError, defaultRes.Message)
		}
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

	ownerID, ok := common.GetUserID(r.Context())
	if !ok {
		common.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	task, err := h.service.UpdateStatus(ownerID, projectID, taskID, req.Status)
	if err != nil {
		switch err {
		case ErrInvalidOwnerID, ErrInvalidProjectID, ErrInvalidTaskID:
			defaultRes.Message = err.Error()
			common.WriteError(w, http.StatusBadRequest, defaultRes.Message)
		case ErrProjectNotFound, ErrTaskNotFound, ErrInvalidTaskStatus, ErrInvalidTaskStatusTransition:
			defaultRes.Message = err.Error()
			common.WriteError(w, http.StatusNotFound, defaultRes.Message)
		default:
			defaultRes.Message = "Internal server error"
			common.WriteError(w, http.StatusInternalServerError, defaultRes.Message)
		}
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
