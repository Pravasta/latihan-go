package project

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"taskflow-api/internal/auth"
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
	ErrInvalidProjectName:        http.StatusBadRequest,
	ErrInvalidProjectDescription: http.StatusBadRequest,
	ErrInvalidOwnerID:            http.StatusBadRequest,
	ErrInvalidProjectID:          http.StatusBadRequest,
	ErrProjectNotFound:           http.StatusNotFound,
}

func writeServiceError(w http.ResponseWriter, err error) {
	if status, ok := errStatusCode[err]; ok {
		common.WriteError(w, status, err.Error())
		return
	}
	slog.Error("unhandled project service error", "error", err)
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

func requirePathID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := r.PathValue("id")
	if id == "" {
		common.WriteError(w, http.StatusBadRequest, "Missing project ID")
		return "", false
	}
	return id, true
}

type createProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, auth.ErrInvalidRequestBody.Error())
		return
	}

	ownerID, ok := requireOwnerID(w, r)
	if !ok {
		return
	}

	project, err := h.service.Create(ownerID, req.Name, req.Description)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	common.WriteJSON(w, http.StatusCreated, project)
}

func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := requireOwnerID(w, r)
	if !ok {
		return
	}

	projects, err := h.service.ListByOwner(ownerID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]any{
		"data": projects,
	})
}

func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	idStr, ok := requirePathID(w, r)
	if !ok {
		return
	}

	ownerID, ok := requireOwnerID(w, r)
	if !ok {
		return
	}

	project, err := h.service.GetByID(ownerID, idStr)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	common.WriteJSON(w, http.StatusOK, project)
}

func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	idStr, ok := requirePathID(w, r)
	if !ok {
		return
	}

	ownerID, ok := requireOwnerID(w, r)
	if !ok {
		return
	}

	if err := h.service.Delete(ownerID, idStr); err != nil {
		writeServiceError(w, err)
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]string{"message": "Project deleted successfully"})
}

type updateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type updateProjectResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	var req updateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, auth.ErrInvalidRequestBody.Error())
		return
	}

	idStr, ok := requirePathID(w, r)
	if !ok {
		return
	}

	ownerID, ok := requireOwnerID(w, r)
	if !ok {
		return
	}

	project, err := h.service.Update(ownerID, idStr, req.Name, req.Description)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	resp := updateProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
	}

	common.WriteJSON(w, http.StatusOK, resp)
}
