package task

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

func TestHandler_CreateTask(t *testing.T) {
	newReq := func(body string) *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/projects/p1/tasks", bytes.NewBufferString(body))
		req.SetPathValue("projectID", "p1")
		return req
	}

	t.Run("invalid body returns 400", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		req := withOwner(newReq("not-json"), "owner-1")
		rec := httptest.NewRecorder()

		h.CreateTask(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing owner id returns 401", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		req := newReq(`{"title":"t","description":"d"}`)
		rec := httptest.NewRecorder()

		h.CreateTask(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("maps each service error to its status code", func(t *testing.T) {
		cases := []struct {
			name       string
			err        error
			wantStatus int
		}{
			{"invalid title", ErrInvalidTaskTitle, http.StatusBadRequest},
			{"project not found", ErrProjectNotFound, http.StatusNotFound},
			{"unexpected error", errors.New("disk exploded"), http.StatusInternalServerError},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				h := NewHandler(&fakeService{
					createFn: func(ownerID, projectID, title, description string) (*Task, error) {
						return nil, tc.err
					},
				})
				req := withOwner(newReq(`{"title":"t","description":"d"}`), "owner-1")
				rec := httptest.NewRecorder()

				h.CreateTask(rec, req)

				if rec.Code != tc.wantStatus {
					t.Fatalf("got status %d, want %d", rec.Code, tc.wantStatus)
				}
			})
		}
	})

	t.Run("success returns 201 with the created task", func(t *testing.T) {
		want := &Task{ID: "t1", ProjectID: "p1", Title: "t", Description: "d", Status: TaskStatusTodo}
		h := NewHandler(&fakeService{
			createFn: func(ownerID, projectID, title, description string) (*Task, error) {
				return want, nil
			},
		})
		req := withOwner(newReq(`{"title":"t","description":"d"}`), "owner-1")
		rec := httptest.NewRecorder()

		h.CreateTask(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusCreated)
		}

		var body DefaultResponse
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if body.Message != "Task created successfully" {
			t.Fatalf("got message %q, want %q", body.Message, "Task created successfully")
		}
	})
}

func TestHandler_DeleteTask(t *testing.T) {
	newReq := func() *http.Request {
		req := httptest.NewRequest(http.MethodDelete, "/projects/p1/tasks/t1", nil)
		req.SetPathValue("projectID", "p1")
		req.SetPathValue("taskID", "t1")
		return req
	}

	t.Run("missing owner id returns 401", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		rec := httptest.NewRecorder()

		h.DeleteTask(rec, newReq())

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("not found maps to 404", func(t *testing.T) {
		h := NewHandler(&fakeService{
			deleteFn: func(ownerID, projectID, taskID string) error {
				return ErrTaskNotFound
			},
		})
		req := withOwner(newReq(), "owner-1")
		rec := httptest.NewRecorder()

		h.DeleteTask(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("success returns 200", func(t *testing.T) {
		h := NewHandler(&fakeService{
			deleteFn: func(ownerID, projectID, taskID string) error {
				return nil
			},
		})
		req := withOwner(newReq(), "owner-1")
		rec := httptest.NewRecorder()

		h.DeleteTask(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
		}
	})
}

func TestHandler_GetByID(t *testing.T) {
	newReq := func() *http.Request {
		req := httptest.NewRequest(http.MethodGet, "/projects/p1/tasks/t1", nil)
		req.SetPathValue("projectID", "p1")
		req.SetPathValue("taskID", "t1")
		return req
	}

	t.Run("missing owner id returns 401", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		rec := httptest.NewRecorder()

		h.GetByID(rec, newReq())

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("not found maps to 404", func(t *testing.T) {
		h := NewHandler(&fakeService{
			getByIDFn: func(ownerID, projectID, taskID string) (*Task, error) {
				return nil, ErrTaskNotFound
			},
		})
		req := withOwner(newReq(), "owner-1")
		rec := httptest.NewRecorder()

		h.GetByID(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("success returns 200", func(t *testing.T) {
		want := &Task{ID: "t1", ProjectID: "p1"}
		h := NewHandler(&fakeService{
			getByIDFn: func(ownerID, projectID, taskID string) (*Task, error) {
				return want, nil
			},
		})
		req := withOwner(newReq(), "owner-1")
		rec := httptest.NewRecorder()

		h.GetByID(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
		}
	})
}

func TestHandler_ListTasks(t *testing.T) {
	newReq := func(rawQuery string) *http.Request {
		req := httptest.NewRequest(http.MethodGet, "/projects/p1/tasks?"+rawQuery, nil)
		req.SetPathValue("projectID", "p1")
		return req
	}

	t.Run("invalid limit query param returns 400", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		req := withOwner(newReq("limit=abc"), "owner-1")
		rec := httptest.NewRecorder()

		h.ListTasks(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid page query param returns 400", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		req := withOwner(newReq("page=0"), "owner-1")
		rec := httptest.NewRecorder()

		h.ListTasks(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing owner id returns 401", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		rec := httptest.NewRecorder()

		h.ListTasks(rec, newReq(""))

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("unexpected service error returns 500", func(t *testing.T) {
		h := NewHandler(&fakeService{
			listFn: func(ownerID, projectID string, query TaskQuery) (*TaskListResult, error) {
				return nil, errors.New("disk exploded")
			},
		})
		req := withOwner(newReq(""), "owner-1")
		rec := httptest.NewRecorder()

		h.ListTasks(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusInternalServerError)
		}
	})

	t.Run("valid query params are all forwarded to the service", func(t *testing.T) {
		var gotQuery TaskQuery
		h := NewHandler(&fakeService{
			listFn: func(ownerID, projectID string, query TaskQuery) (*TaskListResult, error) {
				gotQuery = query
				return &TaskListResult{Tasks: []Task{}}, nil
			},
		})
		req := withOwner(newReq("limit=5&page=2&status=done&search=foo&sort=title&order=asc"), "owner-1")
		rec := httptest.NewRecorder()

		h.ListTasks(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
		}
		if gotQuery.Limit != 5 || gotQuery.Page != 2 || gotQuery.Status != TaskStatusDone ||
			gotQuery.Search != "foo" || gotQuery.Sort != "title" || gotQuery.Order != OrderAsc {
			t.Fatalf("got query %+v, want all fields parsed from the URL", gotQuery)
		}
	})

	t.Run("success returns 200 with data and pagination meta", func(t *testing.T) {
		want := &TaskListResult{
			Tasks:          []Task{{ID: "t1"}, {ID: "t2"}},
			PaginationMeta: PaginationMeta{Page: 1, Limit: 10, Total: 2, TotalPages: 1},
		}
		h := NewHandler(&fakeService{
			listFn: func(ownerID, projectID string, query TaskQuery) (*TaskListResult, error) {
				return want, nil
			},
		})
		req := withOwner(newReq(""), "owner-1")
		rec := httptest.NewRecorder()

		h.ListTasks(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
		}

		var body DefaultResponse
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if body.Meta.Total != 2 {
			t.Fatalf("got total %d, want 2", body.Meta.Total)
		}
	})
}

func TestHandler_UpdateTask(t *testing.T) {
	newReq := func(body string) *http.Request {
		req := httptest.NewRequest(http.MethodPut, "/projects/p1/tasks/t1", bytes.NewBufferString(body))
		req.SetPathValue("projectID", "p1")
		req.SetPathValue("taskID", "t1")
		return req
	}

	t.Run("invalid body returns 400", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		req := withOwner(newReq("not-json"), "owner-1")
		rec := httptest.NewRecorder()

		h.UpdateTask(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing owner id returns 401", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		rec := httptest.NewRecorder()

		h.UpdateTask(rec, newReq(`{"title":"n","description":"d"}`))

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("validation error maps to 400", func(t *testing.T) {
		h := NewHandler(&fakeService{
			updateFn: func(ownerID, projectID, taskID, title, description string) (*Task, error) {
				return nil, ErrInvalidTaskTitle
			},
		})
		req := withOwner(newReq(`{"title":"","description":"d"}`), "owner-1")
		rec := httptest.NewRecorder()

		h.UpdateTask(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found maps to 404", func(t *testing.T) {
		h := NewHandler(&fakeService{
			updateFn: func(ownerID, projectID, taskID, title, description string) (*Task, error) {
				return nil, ErrTaskNotFound
			},
		})
		req := withOwner(newReq(`{"title":"n","description":"d"}`), "owner-1")
		rec := httptest.NewRecorder()

		h.UpdateTask(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("success returns 200 with updated fields", func(t *testing.T) {
		want := &Task{ID: "t1", ProjectID: "p1", Title: "New", Description: "New desc"}
		h := NewHandler(&fakeService{
			updateFn: func(ownerID, projectID, taskID, title, description string) (*Task, error) {
				return want, nil
			},
		})
		req := withOwner(newReq(`{"title":"New","description":"New desc"}`), "owner-1")
		rec := httptest.NewRecorder()

		h.UpdateTask(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
		}
	})
}

func TestHandler_UpdateTaskStatus(t *testing.T) {
	newReq := func(body string) *http.Request {
		req := httptest.NewRequest(http.MethodPatch, "/projects/p1/tasks/t1/status", bytes.NewBufferString(body))
		req.SetPathValue("projectID", "p1")
		req.SetPathValue("taskID", "t1")
		return req
	}

	t.Run("invalid body returns 400", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		req := withOwner(newReq("not-json"), "owner-1")
		rec := httptest.NewRecorder()

		h.UpdateTaskStatus(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing owner id returns 401", func(t *testing.T) {
		h := NewHandler(&fakeService{})
		rec := httptest.NewRecorder()

		h.UpdateTaskStatus(rec, newReq(`{"status":"done"}`))

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	// Regression test: these two used to be miswired to 404 in the old
	// switch-based handler (case ErrProjectNotFound, ErrTaskNotFound,
	// ErrInvalidTaskStatus, ErrInvalidTaskStatusTransition: ...404). They
	// are client input errors and must be 400.
	t.Run("invalid status value maps to 400, not 404", func(t *testing.T) {
		h := NewHandler(&fakeService{
			updateStatusFn: func(ownerID, projectID, taskID string, status TaskStatus) (*Task, error) {
				return nil, ErrInvalidTaskStatus
			},
		})
		req := withOwner(newReq(`{"status":"bogus"}`), "owner-1")
		rec := httptest.NewRecorder()

		h.UpdateTaskStatus(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid status transition maps to 400, not 404", func(t *testing.T) {
		h := NewHandler(&fakeService{
			updateStatusFn: func(ownerID, projectID, taskID string, status TaskStatus) (*Task, error) {
				return nil, ErrInvalidTaskStatusTransition
			},
		})
		req := withOwner(newReq(`{"status":"todo"}`), "owner-1")
		rec := httptest.NewRecorder()

		h.UpdateTaskStatus(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found maps to 404", func(t *testing.T) {
		h := NewHandler(&fakeService{
			updateStatusFn: func(ownerID, projectID, taskID string, status TaskStatus) (*Task, error) {
				return nil, ErrTaskNotFound
			},
		})
		req := withOwner(newReq(`{"status":"done"}`), "owner-1")
		rec := httptest.NewRecorder()

		h.UpdateTaskStatus(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("success returns 200 with updated status", func(t *testing.T) {
		want := &Task{ID: "t1", ProjectID: "p1", Status: TaskStatusDone}
		h := NewHandler(&fakeService{
			updateStatusFn: func(ownerID, projectID, taskID string, status TaskStatus) (*Task, error) {
				return want, nil
			},
		})
		req := withOwner(newReq(`{"status":"done"}`), "owner-1")
		rec := httptest.NewRecorder()

		h.UpdateTaskStatus(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
		}
	})
}
