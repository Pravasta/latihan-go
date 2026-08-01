package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
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
	ErrInvalidName:        http.StatusBadRequest,
	ErrInvalidEmail:       http.StatusBadRequest,
	ErrInvalidPassword:    http.StatusBadRequest,
	ErrInvalidCredentials: http.StatusUnauthorized,
	ErrEmailAlreadyExists: http.StatusConflict,
	ErrUserNotFound:       http.StatusNotFound,
}

func writeServiceError(w http.ResponseWriter, err error) {
	if status, ok := errStatusCode[err]; ok {
		common.WriteError(w, status, err.Error())
		return
	}
	slog.Error("unhandled auth service error", "error", err)
	common.WriteError(w, http.StatusInternalServerError, "Internal server error")
}

type createUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (h *Handler) Register(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req createUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, ErrInvalidRequestBody.Error())
		return
	}

	user, err := h.service.CreateUser(req.Name, req.Email, req.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	userResp := userResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	common.WriteJSON(
		w,
		http.StatusCreated,
		map[string]any{
			"data": userResp,
		},
	)
}

type loginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginUserResponse struct {
	Token string `json:"token"`
}

func (h *Handler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req loginUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, ErrInvalidRequestBody.Error())
		return
	}

	token, err := h.service.Authenticate(req.Email, req.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	resp := loginUserResponse{
		Token: token,
	}

	common.WriteJSON(
		w,
		http.StatusOK,
		resp,
	)
}

func (h *Handler) Me(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := common.GetUserID(r.Context())
	if !ok {
		common.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.service.Me(userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	userResp := userResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	common.WriteJSON(
		w,
		http.StatusOK,
		map[string]any{
			"data": userResp,
		},
	)
}
