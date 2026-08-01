package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"taskflow-api/internal/common"
)

func TestAuthMiddleware_Authenticate(t *testing.T) {
	jwtSvc := NewJWTService("test-secret")
	mw := NewAuthMiddleware(jwtSvc)

	newNextHandler := func(called *bool, gotUserID *string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*called = true
			userID, _ := common.GetUserID(r.Context())
			*gotUserID = userID
			w.WriteHeader(http.StatusOK)
		})
	}

	t.Run("missing Authorization header returns 401", func(t *testing.T) {
		var called bool
		var gotUserID string
		handler := mw.Authenticate(newNextHandler(&called, &gotUserID))

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
		}
		if called {
			t.Fatal("next handler should not have been called")
		}
	})

	t.Run("header without Bearer prefix returns 401", func(t *testing.T) {
		var called bool
		var gotUserID string
		handler := mw.Authenticate(newNextHandler(&called, &gotUserID))

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.Header.Set("Authorization", "Basic abc123")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
		}
		if called {
			t.Fatal("next handler should not have been called")
		}
	})

	t.Run("empty token after Bearer prefix returns 401", func(t *testing.T) {
		var called bool
		var gotUserID string
		handler := mw.Authenticate(newNextHandler(&called, &gotUserID))

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.Header.Set("Authorization", "Bearer ")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
		}
		if called {
			t.Fatal("next handler should not have been called")
		}
	})

	t.Run("invalid token returns 401", func(t *testing.T) {
		var called bool
		var gotUserID string
		handler := mw.Authenticate(newNextHandler(&called, &gotUserID))

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.Header.Set("Authorization", "Bearer not-a-real-token")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
		}
		if called {
			t.Fatal("next handler should not have been called")
		}
	})

	t.Run("valid token sets the user id in context and calls next", func(t *testing.T) {
		var called bool
		var gotUserID string
		handler := mw.Authenticate(newNextHandler(&called, &gotUserID))

		token, err := jwtSvc.Generate("user-1")
		if err != nil {
			t.Fatalf("failed to prepare test fixture: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
		}
		if !called {
			t.Fatal("expected next handler to be called")
		}
		if gotUserID != "user-1" {
			t.Fatalf("got user id %q, want %q", gotUserID, "user-1")
		}
	})
}
