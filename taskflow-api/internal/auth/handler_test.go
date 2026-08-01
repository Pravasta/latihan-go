package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"taskflow-api/internal/common"
)

func TestHandler_Register(t *testing.T) {
	t.Run("invalid body returns 400", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("not-json"))
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("maps each service error to its status code", func(t *testing.T) {
		cases := []struct {
			name       string
			err        error
			wantStatus int
		}{
			{"invalid name", ErrInvalidName, http.StatusBadRequest},
			{"invalid email", ErrInvalidEmail, http.StatusBadRequest},
			{"invalid password", ErrInvalidPassword, http.StatusBadRequest},
			{"email already exists", ErrEmailAlreadyExists, http.StatusConflict},
			{"unexpected error", errors.New("disk exploded"), http.StatusInternalServerError},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				h := NewHandler(&fakeService{
					createUserFn: func(name, email, password string) (*User, error) {
						return nil, tc.err
					},
				})
				body := `{"name":"Alice","email":"alice@example.com","password":"Passw0rd!"}`
				req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
				rec := httptest.NewRecorder()

				h.Register(rec, req)

				if rec.Code != tc.wantStatus {
					t.Fatalf("got status %d, want %d", rec.Code, tc.wantStatus)
				}
			})
		}
	})

	t.Run("success returns 201 with the created user", func(t *testing.T) {
		want := &User{ID: "u1", Name: "Alice", Email: "alice@example.com"}
		h := NewHandler(&fakeService{
			createUserFn: func(name, email, password string) (*User, error) {
				return want, nil
			},
		})
		body := `{"name":"Alice","email":"alice@example.com","password":"Passw0rd!"}`
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusCreated)
		}
	})
}

func TestHandler_Login(t *testing.T) {
	t.Run("invalid body returns 400", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("not-json"))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("maps each service error to its status code", func(t *testing.T) {
		cases := []struct {
			name       string
			err        error
			wantStatus int
		}{
			{"invalid email", ErrInvalidEmail, http.StatusBadRequest},
			{"invalid credentials", ErrInvalidCredentials, http.StatusUnauthorized},
			{"user not found", ErrUserNotFound, http.StatusNotFound},
			{"unexpected error", errors.New("disk exploded"), http.StatusInternalServerError},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				h := NewHandler(&fakeService{
					authenticateFn: func(email, password string) (string, error) {
						return "", tc.err
					},
				})
				body := `{"email":"alice@example.com","password":"Passw0rd!"}`
				req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
				rec := httptest.NewRecorder()

				h.Login(rec, req)

				if rec.Code != tc.wantStatus {
					t.Fatalf("got status %d, want %d", rec.Code, tc.wantStatus)
				}
			})
		}
	})

	t.Run("success returns 200 with a token", func(t *testing.T) {
		h := NewHandler(&fakeService{
			authenticateFn: func(email, password string) (string, error) {
				return "signed.jwt.token", nil
			},
		})
		body := `{"email":"alice@example.com","password":"Passw0rd!"}`
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
		}

		var body2 loginUserResponse
		if err := json.NewDecoder(rec.Body).Decode(&body2); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if body2.Token != "signed.jwt.token" {
			t.Fatalf("got token %q, want %q", body2.Token, "signed.jwt.token")
		}
	})
}

func TestHandler_Me(t *testing.T) {
	t.Run("missing owner id returns 401", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		rec := httptest.NewRecorder()

		h.Me(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("not found maps to 404", func(t *testing.T) {
		h := NewHandler(&fakeService{
			meFn: func(userID string) (*User, error) {
				return nil, ErrUserNotFound
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req = req.WithContext(common.SetUserID(req.Context(), "u1"))
		rec := httptest.NewRecorder()

		h.Me(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("unexpected error returns 500", func(t *testing.T) {
		h := NewHandler(&fakeService{
			meFn: func(userID string) (*User, error) {
				return nil, errors.New("disk exploded")
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req = req.WithContext(common.SetUserID(req.Context(), "u1"))
		rec := httptest.NewRecorder()

		h.Me(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success returns 200", func(t *testing.T) {
		want := &User{ID: "u1", Name: "Alice", Email: "alice@example.com"}
		h := NewHandler(&fakeService{
			meFn: func(userID string) (*User, error) {
				return want, nil
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req = req.WithContext(common.SetUserID(req.Context(), "u1"))
		rec := httptest.NewRecorder()

		h.Me(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
		}
	})
}
