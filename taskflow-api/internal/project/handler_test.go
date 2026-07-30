package project

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"taskflow-api/internal/common"
)

func withOwner(r *http.Request, ownerID string) *http.Request {
	return r.WithContext(common.SetUserID(r.Context(), ownerID))
}

func TestHandler_CreateProject(t *testing.T) {
	t.Run("missing owner id returns 401", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(`{"name":"n","description":"d"}`))
		rec := httptest.NewRecorder()

		h.CreateProject(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("invalid body returns 400", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		req := withOwner(httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(`not-json`)), "owner-1")
		rec := httptest.NewRecorder()

		h.CreateProject(rec, req)

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
			{"invalid name", ErrInvalidProjectName, http.StatusBadRequest},
			{"invalid description", ErrInvalidProjectDescription, http.StatusBadRequest},
			{"invalid owner", ErrInvalidOwnerID, http.StatusBadRequest},
			{"unexpected error", errors.New("disk exploded"), http.StatusInternalServerError},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				h := NewHandler(&fakeService{
					createFn: func(ownerID, name, description string) (*Project, error) {
						return nil, tc.err
					},
				})
				req := withOwner(httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(`{"name":"n","description":"d"}`)), "owner-1")
				rec := httptest.NewRecorder()

				h.CreateProject(rec, req)

				if rec.Code != tc.wantStatus {
					t.Fatalf("got status %d, want %d", rec.Code, tc.wantStatus)
				}
			})
		}
	})

	t.Run("success returns 201 with the created project", func(t *testing.T) {
		want := &Project{ID: "p1", OwnerID: "owner-1", Name: "n", Description: "d"}
		h := NewHandler(&fakeService{
			createFn: func(ownerID, name, description string) (*Project, error) {
				return want, nil
			},
		})
		req := withOwner(httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(`{"name":"n","description":"d"}`)), "owner-1")
		rec := httptest.NewRecorder()

		h.CreateProject(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusCreated)
		}

		var got Project
		if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if got.ID != want.ID {
			t.Fatalf("got id %q, want %q", got.ID, want.ID)
		}
	})
}

func TestHandler_GetProject(t *testing.T) {
	t.Run("missing path id returns 400", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		req := withOwner(httptest.NewRequest(http.MethodGet, "/projects/", nil), "owner-1")
		rec := httptest.NewRecorder()

		h.GetProject(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing owner id returns 401", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		req := httptest.NewRequest(http.MethodGet, "/projects/p1", nil)
		req.SetPathValue("id", "p1")
		rec := httptest.NewRecorder()

		h.GetProject(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("not found maps to 404", func(t *testing.T) {
		h := NewHandler(&fakeService{
			getFn: func(ownerID, projectID string) (*Project, error) {
				return nil, ErrProjectNotFound
			},
		})
		req := withOwner(httptest.NewRequest(http.MethodGet, "/projects/p1", nil), "owner-1")
		req.SetPathValue("id", "p1")
		rec := httptest.NewRecorder()

		h.GetProject(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("unexpected error returns 500", func(t *testing.T) {
		h := NewHandler(&fakeService{
			getFn: func(ownerID, projectID string) (*Project, error) {
				return nil, errors.New("disk exploded")
			},
		})
		req := withOwner(httptest.NewRequest(http.MethodGet, "/projects/p1", nil), "owner-1")
		req.SetPathValue("id", "p1")
		rec := httptest.NewRecorder()

		h.GetProject(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success returns 200", func(t *testing.T) {
		want := &Project{ID: "p1", OwnerID: "owner-1"}
		h := NewHandler(&fakeService{
			getFn: func(ownerID, projectID string) (*Project, error) {
				return want, nil
			},
		})
		req := withOwner(httptest.NewRequest(http.MethodGet, "/projects/p1", nil), "owner-1")
		req.SetPathValue("id", "p1")
		rec := httptest.NewRecorder()

		h.GetProject(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
		}
	})
}

func TestHandler_ListProjects(t *testing.T) {
	t.Run("missing owner id returns 401", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		req := httptest.NewRequest(http.MethodGet, "/projects", nil)
		rec := httptest.NewRecorder()

		h.ListProjects(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("unexpected service error returns 500", func(t *testing.T) {
		h := NewHandler(&fakeService{
			listFn: func(ownerID string) ([]Project, error) {
				return nil, errors.New("disk exploded")
			},
		})
		req := withOwner(httptest.NewRequest(http.MethodGet, "/projects", nil), "owner-1")
		rec := httptest.NewRecorder()

		h.ListProjects(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusInternalServerError)
		}
	})

	t.Run("success returns 200 with data wrapper", func(t *testing.T) {
		want := []Project{{ID: "p1", OwnerID: "owner-1"}, {ID: "p2", OwnerID: "owner-1"}}
		h := NewHandler(&fakeService{
			listFn: func(ownerID string) ([]Project, error) {
				return want, nil
			},
		})
		req := withOwner(httptest.NewRequest(http.MethodGet, "/projects", nil), "owner-1")
		rec := httptest.NewRecorder()

		h.ListProjects(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
		}

		var body struct {
			Data []Project `json:"data"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(body.Data) != 2 {
			t.Fatalf("got %d projects, want 2", len(body.Data))
		}
	})
}

func TestHandler_DeleteProject(t *testing.T) {
	t.Run("missing path id returns 400", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		req := withOwner(httptest.NewRequest(http.MethodDelete, "/projects/", nil), "owner-1")
		rec := httptest.NewRecorder()

		h.DeleteProject(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing owner id returns 401", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		req := httptest.NewRequest(http.MethodDelete, "/projects/p1", nil)
		req.SetPathValue("id", "p1")
		rec := httptest.NewRecorder()

		h.DeleteProject(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("not found maps to 404", func(t *testing.T) {
		h := NewHandler(&fakeService{
			deleteFn: func(ownerID, projectID string) error {
				return ErrProjectNotFound
			},
		})
		req := withOwner(httptest.NewRequest(http.MethodDelete, "/projects/p1", nil), "owner-1")
		req.SetPathValue("id", "p1")
		rec := httptest.NewRecorder()

		h.DeleteProject(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("success returns 200", func(t *testing.T) {
		h := NewHandler(&fakeService{
			deleteFn: func(ownerID, projectID string) error {
				return nil
			},
		})
		req := withOwner(httptest.NewRequest(http.MethodDelete, "/projects/p1", nil), "owner-1")
		req.SetPathValue("id", "p1")
		rec := httptest.NewRecorder()

		h.DeleteProject(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
		}
	})
}

func TestHandler_UpdateProject(t *testing.T) {
	t.Run("invalid body returns 400", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		req := withOwner(httptest.NewRequest(http.MethodPut, "/projects/p1", bytes.NewBufferString(`not-json`)), "owner-1")
		req.SetPathValue("id", "p1")
		rec := httptest.NewRecorder()

		h.UpdateProject(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing path id returns 400", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		req := withOwner(httptest.NewRequest(http.MethodPut, "/projects/", bytes.NewBufferString(`{"name":"n","description":"d"}`)), "owner-1")
		rec := httptest.NewRecorder()

		h.UpdateProject(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing owner id returns 401", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		req := httptest.NewRequest(http.MethodPut, "/projects/p1", bytes.NewBufferString(`{"name":"n","description":"d"}`))
		req.SetPathValue("id", "p1")
		rec := httptest.NewRecorder()

		h.UpdateProject(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("validation error maps to 400", func(t *testing.T) {
		h := NewHandler(&fakeService{
			updateFn: func(ownerID, projectID, name, description string) (*Project, error) {
				return nil, ErrInvalidProjectName
			},
		})
		req := withOwner(httptest.NewRequest(http.MethodPut, "/projects/p1", bytes.NewBufferString(`{"name":"","description":"d"}`)), "owner-1")
		req.SetPathValue("id", "p1")
		rec := httptest.NewRecorder()

		h.UpdateProject(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found maps to 404", func(t *testing.T) {
		h := NewHandler(&fakeService{
			updateFn: func(ownerID, projectID, name, description string) (*Project, error) {
				return nil, ErrProjectNotFound
			},
		})
		req := withOwner(httptest.NewRequest(http.MethodPut, "/projects/p1", bytes.NewBufferString(`{"name":"n","description":"d"}`)), "owner-1")
		req.SetPathValue("id", "p1")
		rec := httptest.NewRecorder()

		h.UpdateProject(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("success returns 200 with updated fields", func(t *testing.T) {
		want := &Project{ID: "p1", Name: "New", Description: "New desc"}
		h := NewHandler(&fakeService{
			updateFn: func(ownerID, projectID, name, description string) (*Project, error) {
				return want, nil
			},
		})
		req := withOwner(httptest.NewRequest(http.MethodPut, "/projects/p1", bytes.NewBufferString(`{"name":"New","description":"New desc"}`)), "owner-1")
		req.SetPathValue("id", "p1")
		rec := httptest.NewRecorder()

		h.UpdateProject(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
		}

		var got updateProjectResponse
		if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if got.Name != "New" || got.Description != "New desc" {
			t.Fatalf("got %+v, want name/description updated", got)
		}
	})
}
