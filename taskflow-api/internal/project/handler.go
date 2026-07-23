package project

import (
	"encoding/json"
	"fmt"
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

type createProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *Handler) CreateProject(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req createProjectRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, auth.ErrInvalidRequestBody.Error())
		return
	}

	ownerID, ok := common.GetUserID(r.Context())
	if !ok {
		fmt.Println("[Handler CreateProject] Failed to get user ID from context")
		common.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	project, err := h.service.Create(ownerID, req.Name, req.Description)
	if err != nil {
		switch err {
		case ErrInvalidProjectName:
			common.WriteError(w, http.StatusBadRequest, err.Error())
		case ErrInvalidProjectDescription:
			common.WriteError(w, http.StatusBadRequest, err.Error())
		case ErrInvalidOwnerID:
			common.WriteError(w, http.StatusBadRequest, err.Error())
		default:
			common.WriteError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	common.WriteJSON(w, http.StatusCreated, project)
}

func (h *Handler) ListProjects(
	w http.ResponseWriter,
	r *http.Request,
) {
	ownerID, ok := common.GetUserID(r.Context())
	if !ok {
		fmt.Println("[Handler ListProjects] Failed to get user ID from context")
		common.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projects, err := h.service.ListByOwner(ownerID)
	if err != nil {
		switch err {
		case ErrInvalidOwnerID:
			common.WriteError(w, http.StatusBadRequest, err.Error())
		default:
			common.WriteError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]any{
		"data": projects,
	})
}

func (h *Handler) GetProject(
	w http.ResponseWriter,
	r *http.Request,
) {
	idStr := r.PathValue("id")

	if idStr == "" {
		common.WriteError(w, http.StatusBadRequest, "Missing project ID")
		return
	}

	ownerID, ok := common.GetUserID(r.Context())
	if !ok {
		fmt.Println("[Handler GetProject] Failed to get user ID from context")
		common.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	project, err := h.service.GetByID(ownerID, idStr)
	if err != nil {
		switch err {
		case ErrProjectNotFound:
			common.WriteError(w, http.StatusNotFound, err.Error())
		case ErrInvalidOwnerID:
			common.WriteError(w, http.StatusBadRequest, err.Error())
		default:
			common.WriteError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	common.WriteJSON(w, http.StatusOK, project)
}

func (h *Handler) DeleteProject(
	w http.ResponseWriter,
	r *http.Request,
) {
	idStr := r.PathValue("id")

	if idStr == "" {
		common.WriteError(w, http.StatusBadRequest, "Missing project ID")
		return
	}

	ownerID, ok := common.GetUserID(r.Context())

	if !ok {
		fmt.Println("[Handler DeleteProject] Failed to get user ID from context")
		common.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err := h.service.Delete(ownerID, idStr)
	if err != nil {
		switch err {
		case ErrProjectNotFound:
			common.WriteError(w, http.StatusNotFound, err.Error())
		case ErrInvalidOwnerID:
			common.WriteError(w, http.StatusBadRequest, err.Error())
		case ErrInvalidProjectID:
			common.WriteError(w, http.StatusBadRequest, err.Error())
		default:
			common.WriteError(w, http.StatusInternalServerError, "Internal server error")
		}
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

func (h *Handler) UpdateProject(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req updateProjectRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, auth.ErrInvalidRequestBody.Error())
		return
	}

	idStr := r.PathValue("id")

	if idStr == "" {
		common.WriteError(w, http.StatusBadRequest, "Missing project ID")
		return
	}

	ownerID, ok := common.GetUserID(r.Context())
	if !ok {
		fmt.Println("[Handler UpdateProject] Failed to get user ID from context")
		common.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	project, err := h.service.Update(ownerID, idStr, req.Name, req.Description)
	if err != nil {
		switch err {
		case ErrProjectNotFound:
			common.WriteError(w, http.StatusNotFound, err.Error())
		case ErrInvalidOwnerID:
			common.WriteError(w, http.StatusBadRequest, err.Error())
		case ErrInvalidProjectID:
			common.WriteError(w, http.StatusBadRequest, err.Error())
		case ErrInvalidProjectName:
			common.WriteError(w, http.StatusBadRequest, err.Error())
		case ErrInvalidProjectDescription:
			common.WriteError(w, http.StatusBadRequest, err.Error())
		default:
			common.WriteError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	resp := updateProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
	}

	common.WriteJSON(w, http.StatusOK, resp)
}
