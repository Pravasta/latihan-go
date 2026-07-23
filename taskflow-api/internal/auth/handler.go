package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"taskflow-api/internal/common"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
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
		switch err {
		case ErrInvalidName:
			common.WriteError(w, http.StatusBadRequest, err.Error())
		case ErrInvalidEmail:
			common.WriteError(w, http.StatusBadRequest, err.Error())
		case ErrInvalidPassword:
			common.WriteError(w, http.StatusBadRequest, err.Error())
		case ErrEmailAlreadyExists:
			common.WriteError(w, http.StatusConflict, err.Error())
		default:
			common.WriteError(w, http.StatusInternalServerError, "Internal server error")
		}
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
		switch err {
		case ErrInvalidEmail:
			common.WriteError(w, http.StatusBadRequest, err.Error())
		case ErrInvalidPassword:
			common.WriteError(w, http.StatusUnauthorized, err.Error())
		case ErrUserNotFound:
			common.WriteError(w, http.StatusNotFound, err.Error())
		default:
			common.WriteError(w, http.StatusInternalServerError, "Internal server error")
		}
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
	// read from jwt middleware context
	userID, ok := common.GetUserID(r.Context())
	fmt.Println("[Handler] User ID from context:", userID)
	if !ok {
		fmt.Println("[Handler] User ID is missing")
		common.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.service.Me(userID)
	if err != nil {
		switch err {
		case ErrUserNotFound:
			common.WriteError(w, http.StatusNotFound, err.Error())
		default:
			common.WriteError(w, http.StatusInternalServerError, "Internal server error")
		}
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
