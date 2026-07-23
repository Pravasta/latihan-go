package note

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type createNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (h *Handler) CreateNote(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req createNoteRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Println("[Handler CreateNote] Failed to decode request body:", err)
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{"error": ErrInvalidJSON.Error()},
		)
		return
	}

	note, err := h.service.Create(req.Title, req.Content)

	if err != nil {
		switch {
		case errors.Is(err, ErrTitleEmpty):
			writeJSON(
				w,
				http.StatusBadRequest,
				map[string]string{"error": err.Error()},
			)

		case errors.Is(err, ErrContentEmpty):
			writeJSON(
				w,
				http.StatusBadRequest,
				map[string]string{"error": err.Error()},
			)

		default:
			writeJSON(
				w,
				http.StatusInternalServerError,
				map[string]string{"error": err.Error()},
			)
		}

		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		map[string]any{
			"data": note,
		},
	)
}

func (h *Handler) ListNotes(
	w http.ResponseWriter,
	r *http.Request,
) {
	search := r.URL.Query().Get("search")

	notes, err := h.service.List(search)
	if err != nil {
		if errors.Is(err, ErrNoteNotFound) {
			writeJSON(
				w,
				http.StatusNotFound,
				map[string]string{"error": err.Error()},
			)
			return
		}

		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{"error": err.Error()},
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]any{
			"data": notes,
		},
	)
}

func (h *Handler) GetNoteByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	idStr := r.PathValue("id")

	// Convert idStr to an integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{"error": err.Error()},
		)
		return
	}

	note, err := h.service.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, ErrNoteNotFound):
			writeJSON(
				w,
				http.StatusNotFound,
				map[string]string{"error": err.Error()},
			)

		default:
			writeJSON(
				w,
				http.StatusInternalServerError,
				map[string]string{"error": err.Error()},
			)
		}

		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]any{
			"data": note,
		},
	)
}

func (h *Handler) UpdateNote(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req createNoteRequest
	var idStr = r.PathValue("id")

	// Convert idStr to an integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{"error": err.Error()},
		)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Println("[Handler UpdateNote] Failed to decode request body:", err)
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{"error": ErrInvalidJSON.Error()},
		)
		return
	}

	note, err := h.service.Update(id, req.Title, req.Content)

	if err != nil {
		switch {
		case errors.Is(err, ErrNoteNotFound):
			writeJSON(
				w,
				http.StatusNotFound,
				map[string]string{"error": err.Error()},
			)

		default:
			writeJSON(
				w,
				http.StatusInternalServerError,
				map[string]string{"error": err.Error()},
			)
		}

		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]any{
			"data": note,
		},
	)
}

func (h *Handler) DeleteNote(
	w http.ResponseWriter,
	r *http.Request,
) {
	idStr := r.PathValue("id")

	// Convert idStr to an integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{"error": err.Error()},
		)
		return
	}

	err = h.service.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, ErrNoteNotFound):
			writeJSON(
				w,
				http.StatusNotFound,
				map[string]string{"error": err.Error()},
			)

		default:
			writeJSON(
				w,
				http.StatusInternalServerError,
				map[string]string{"error": err.Error()},
			)
		}

		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]string{"message": "Note deleted successfully"},
	)
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}
