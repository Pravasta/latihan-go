package task

import (
	"errors"
	"taskflow-api/internal/project"
	"testing"
	"time"
)

func TestService_Create(t *testing.T) {
	t.Run("validation", func(t *testing.T) {
		cases := []struct {
			name        string
			ownerID     string
			projectID   string
			title       string
			description string
			wantErr     error
		}{
			{"missing owner id", "", "project-1", "Title", "Desc", ErrInvalidOwnerID},
			{"missing project id", "owner-1", "", "Title", "Desc", ErrInvalidProjectID},
			{"missing title", "owner-1", "project-1", "", "Desc", ErrInvalidTaskTitle},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				svc := NewService(&fakeStorage{}, &fakeProjectService{})

				_, err := svc.Create(tc.ownerID, tc.projectID, tc.title, tc.description)
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("got error %v, want %v", err, tc.wantErr)
				}
			})
		}
	})

	t.Run("success persists a new task", func(t *testing.T) {
		storage := &fakeStorage{}
		projectSvc := &fakeProjectService{
			getByIDFn: func(ownerID, projectID string) (*project.Project, error) {
				return &project.Project{ID: projectID}, nil
			},
		}

		svc := NewService(storage, projectSvc)

		got, err := svc.Create("owner-1", "project-1", "My Task", "desc")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID == "" {
			t.Fatal("expected a generated ID, got empty string")
		}
		if got.Title != "My Task" {
			t.Fatalf("got title %q, want %q", got.Title, "My Task")
		}
		if got.Description != "desc" {
			t.Fatalf("got description %q, want %q", got.Description, "desc")
		}

		if len(storage.tasks) != 1 {
			t.Fatalf("expected 1 task persisted, got %d", len(storage.tasks))
		}
	})

	t.Run("load failure is returned to the caller", func(t *testing.T) {
		svc := NewService(
			&fakeStorage{loadErr: errors.New("disk error")},
			&fakeProjectService{
				getByIDFn: func(ownerID, projectID string) (*project.Project, error) {
					return &project.Project{ID: projectID}, nil
				},
			})

		if _, err := svc.Create("owner-1", "project-1", "Title", "Desc"); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("save failure is returned to the caller", func(t *testing.T) {
		svc := NewService(
			&fakeStorage{saveErr: errors.New("disk error")},
			&fakeProjectService{
				getByIDFn: func(ownerID, projectID string) (*project.Project, error) {
					return &project.Project{ID: projectID}, nil
				},
			})

		if _, err := svc.Create("owner-1", "project-1", "Title", "Desc"); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("project not found", func(t *testing.T) {
		svc := NewService(
			&fakeStorage{},
			&fakeProjectService{
				getByIDFn: func(ownerID, projectID string) (*project.Project, error) {
					return nil, project.ErrProjectNotFound
				},
			})

		if _, err := svc.Create("owner-1", "project-1", "Title", "Desc"); !errors.Is(err, ErrProjectNotFound) {
			t.Fatalf("expected ErrProjectNotFound, got %v", err)
		}
	})
}

func TestEnsureProjectExists(t *testing.T) {
	t.Run("project found", func(t *testing.T) {
		err := ensureProjectExists(validProject(), "owner-1", "project-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("project lookup fails for a reason other than not-found", func(t *testing.T) {
		ps := &fakeProjectService{
			getByIDFn: func(ownerID, projectID string) (*project.Project, error) {
				return nil, errors.New("project storage unavailable")
			},
		}

		err := ensureProjectExists(ps, "owner-1", "project-1")
		if err == nil || errors.Is(err, ErrProjectNotFound) {
			t.Fatalf("got %v, want the underlying error passed through unchanged", err)
		}
	})

	t.Run("project lookup returns no error and no project", func(t *testing.T) {
		ps := &fakeProjectService{
			getByIDFn: func(ownerID, projectID string) (*project.Project, error) {
				return nil, nil
			},
		}

		err := ensureProjectExists(ps, "owner-1", "project-1")
		if !errors.Is(err, ErrProjectNotFound) {
			t.Fatalf("got %v, want %v", err, ErrProjectNotFound)
		}
	})
}

// validProject is a fakeProjectService that always confirms the project
// exists, so tests below can focus on task-level behavior instead of
// re-proving the project-lookup path already covered by TestService_Create.
func validProject() *fakeProjectService {
	return &fakeProjectService{
		getByIDFn: func(ownerID, projectID string) (*project.Project, error) {
			return &project.Project{ID: projectID}, nil
		},
	}
}

func TestService_Delete(t *testing.T) {
	existing := Task{ID: "t1", ProjectID: "project-1"}

	t.Run("validation", func(t *testing.T) {
		cases := []struct {
			name      string
			ownerID   string
			projectID string
			taskID    string
			wantErr   error
		}{
			{"missing owner id", "", "project-1", "t1", ErrInvalidOwnerID},
			{"missing project id", "owner-1", "", "t1", ErrInvalidProjectID},
			{"missing task id", "owner-1", "project-1", "", ErrInvalidTaskID},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				svc := NewService(&fakeStorage{tasks: []Task{existing}}, validProject())

				err := svc.Delete(tc.ownerID, tc.projectID, tc.taskID)
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("got error %v, want %v", err, tc.wantErr)
				}
			})
		}
	})

	t.Run("project not found", func(t *testing.T) {
		svc := NewService(&fakeStorage{}, &fakeProjectService{
			getByIDFn: func(ownerID, projectID string) (*project.Project, error) {
				return nil, project.ErrProjectNotFound
			},
		})

		if err := svc.Delete("owner-1", "project-1", "t1"); !errors.Is(err, ErrProjectNotFound) {
			t.Fatalf("expected ErrProjectNotFound, got %v", err)
		}
	})

	t.Run("task not found for a different project", func(t *testing.T) {
		svc := NewService(&fakeStorage{tasks: []Task{existing}}, validProject())

		if err := svc.Delete("owner-1", "project-2", "t1"); !errors.Is(err, ErrTaskNotFound) {
			t.Fatalf("expected ErrTaskNotFound, got %v", err)
		}
	})

	t.Run("load failure is returned to the caller", func(t *testing.T) {
		svc := NewService(&fakeStorage{loadErr: errors.New("disk error")}, validProject())

		if err := svc.Delete("owner-1", "project-1", "t1"); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("save failure is returned to the caller", func(t *testing.T) {
		svc := NewService(&fakeStorage{tasks: []Task{existing}, saveErr: errors.New("disk full")}, validProject())

		if err := svc.Delete("owner-1", "project-1", "t1"); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("deletes an existing task", func(t *testing.T) {
		storage := &fakeStorage{tasks: []Task{existing}}
		svc := NewService(storage, validProject())

		if err := svc.Delete("owner-1", "project-1", "t1"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(storage.tasks) != 0 {
			t.Fatalf("expected task to be removed, got %d remaining", len(storage.tasks))
		}
	})
}

func TestService_GetByID(t *testing.T) {
	existing := Task{ID: "t1", ProjectID: "project-1", Title: "Existing"}

	t.Run("validation", func(t *testing.T) {
		cases := []struct {
			name      string
			ownerID   string
			projectID string
			taskID    string
			wantErr   error
		}{
			{"missing owner id", "", "project-1", "t1", ErrInvalidOwnerID},
			{"missing project id", "owner-1", "", "t1", ErrInvalidProjectID},
			{"missing task id", "owner-1", "project-1", "", ErrInvalidTaskID},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				svc := NewService(&fakeStorage{tasks: []Task{existing}}, validProject())

				_, err := svc.GetByID(tc.ownerID, tc.projectID, tc.taskID)
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("got error %v, want %v", err, tc.wantErr)
				}
			})
		}
	})

	t.Run("found for the correct project", func(t *testing.T) {
		svc := NewService(&fakeStorage{tasks: []Task{existing}}, validProject())

		got, err := svc.GetByID("owner-1", "project-1", "t1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != "t1" {
			t.Fatalf("got id %q, want %q", got.ID, "t1")
		}
	})

	t.Run("not found for a different project", func(t *testing.T) {
		svc := NewService(&fakeStorage{tasks: []Task{existing}}, validProject())

		_, err := svc.GetByID("owner-1", "project-2", "t1")
		if !errors.Is(err, ErrTaskNotFound) {
			t.Fatalf("got error %v, want %v", err, ErrTaskNotFound)
		}
	})

	t.Run("project not found", func(t *testing.T) {
		svc := NewService(&fakeStorage{}, &fakeProjectService{
			getByIDFn: func(ownerID, projectID string) (*project.Project, error) {
				return nil, project.ErrProjectNotFound
			},
		})

		if _, err := svc.GetByID("owner-1", "project-1", "t1"); !errors.Is(err, ErrProjectNotFound) {
			t.Fatalf("expected ErrProjectNotFound, got %v", err)
		}
	})

	t.Run("load failure is returned to the caller", func(t *testing.T) {
		svc := NewService(&fakeStorage{loadErr: errors.New("disk error")}, validProject())

		if _, err := svc.GetByID("owner-1", "project-1", "t1"); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestService_List(t *testing.T) {
	tasks := []Task{
		{ID: "t1", ProjectID: "project-1", Title: "Alpha", Status: TaskStatusTodo, CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: "t2", ProjectID: "project-1", Title: "Beta", Status: TaskStatusDone, CreatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
		{ID: "t3", ProjectID: "project-2", Title: "Gamma", Status: TaskStatusTodo, CreatedAt: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)},
	}

	t.Run("validation", func(t *testing.T) {
		cases := []struct {
			name      string
			ownerID   string
			projectID string
			query     TaskQuery
			wantErr   error
		}{
			{"missing owner id", "", "project-1", TaskQuery{Page: 1, Limit: 10}, ErrInvalidOwnerID},
			{"missing project id", "owner-1", "", TaskQuery{Page: 1, Limit: 10}, ErrInvalidProjectID},
			{"page below 1", "owner-1", "project-1", TaskQuery{Page: 0, Limit: 10}, ErrInvalidPageNumber},
			{"limit below 1", "owner-1", "project-1", TaskQuery{Page: 1, Limit: 0}, ErrInvalidLimitNumber},
			{"limit above 100", "owner-1", "project-1", TaskQuery{Page: 1, Limit: 101}, ErrInvalidLimitNumber},
			{"invalid status", "owner-1", "project-1", TaskQuery{Page: 1, Limit: 10, Status: "bogus"}, ErrInvalidTaskStatus},
			{"invalid sort", "owner-1", "project-1", TaskQuery{Page: 1, Limit: 10, Sort: "bogus"}, ErrInvalidSortField},
			{"invalid order", "owner-1", "project-1", TaskQuery{Page: 1, Limit: 10, Order: "bogus"}, ErrInvalidOrder},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				svc := NewService(&fakeStorage{tasks: tasks}, validProject())

				_, err := svc.List(tc.ownerID, tc.projectID, tc.query)
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("got error %v, want %v", err, tc.wantErr)
				}
			})
		}
	})

	t.Run("filters by project", func(t *testing.T) {
		svc := NewService(&fakeStorage{tasks: tasks}, validProject())

		got, err := svc.List("owner-1", "project-1", TaskQuery{Page: 1, Limit: 10})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got.Tasks) != 2 {
			t.Fatalf("got %d tasks, want 2", len(got.Tasks))
		}
	})

	t.Run("filters by status", func(t *testing.T) {
		svc := NewService(&fakeStorage{tasks: tasks}, validProject())

		got, err := svc.List("owner-1", "project-1", TaskQuery{Page: 1, Limit: 10, Status: TaskStatusDone})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got.Tasks) != 1 || got.Tasks[0].ID != "t2" {
			t.Fatalf("got %+v, want only t2", got.Tasks)
		}
	})

	t.Run("filters by search across title and description", func(t *testing.T) {
		svc := NewService(&fakeStorage{tasks: tasks}, validProject())

		got, err := svc.List("owner-1", "project-1", TaskQuery{Page: 1, Limit: 10, Search: "alpha"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got.Tasks) != 1 || got.Tasks[0].ID != "t1" {
			t.Fatalf("got %+v, want only t1", got.Tasks)
		}
	})

	t.Run("sorts by title ascending", func(t *testing.T) {
		svc := NewService(&fakeStorage{tasks: tasks}, validProject())

		got, err := svc.List("owner-1", "project-1", TaskQuery{Page: 1, Limit: 10, Sort: "title", Order: OrderAsc})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got.Tasks) != 2 || got.Tasks[0].Title != "Alpha" || got.Tasks[1].Title != "Beta" {
			t.Fatalf("got %+v, want Alpha then Beta", got.Tasks)
		}
	})

	t.Run("sorts by title descending", func(t *testing.T) {
		svc := NewService(&fakeStorage{tasks: tasks}, validProject())

		got, err := svc.List("owner-1", "project-1", TaskQuery{Page: 1, Limit: 10, Sort: "title", Order: OrderDesc})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got.Tasks) != 2 || got.Tasks[0].Title != "Beta" || got.Tasks[1].Title != "Alpha" {
			t.Fatalf("got %+v, want Beta then Alpha", got.Tasks)
		}
	})

	t.Run("sorts by created_at ascending", func(t *testing.T) {
		svc := NewService(&fakeStorage{tasks: tasks}, validProject())

		got, err := svc.List("owner-1", "project-1", TaskQuery{Page: 1, Limit: 10, Sort: "created_at", Order: OrderAsc})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got.Tasks) != 2 || got.Tasks[0].ID != "t1" || got.Tasks[1].ID != "t2" {
			t.Fatalf("got %+v, want t1 then t2", got.Tasks)
		}
	})

	t.Run("sorts by updated_at ascending and descending", func(t *testing.T) {
		withUpdated := []Task{
			{ID: "t1", ProjectID: "project-1", Title: "Alpha", UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
			{ID: "t2", ProjectID: "project-1", Title: "Beta", UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
		}

		ascSvc := NewService(&fakeStorage{tasks: withUpdated}, validProject())
		asc, err := ascSvc.List("owner-1", "project-1", TaskQuery{Page: 1, Limit: 10, Sort: "updated_at", Order: OrderAsc})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(asc.Tasks) != 2 || asc.Tasks[0].ID != "t1" || asc.Tasks[1].ID != "t2" {
			t.Fatalf("got %+v, want t1 then t2 ascending", asc.Tasks)
		}

		descSvc := NewService(&fakeStorage{tasks: withUpdated}, validProject())
		desc, err := descSvc.List("owner-1", "project-1", TaskQuery{Page: 1, Limit: 10, Sort: "updated_at", Order: OrderDesc})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(desc.Tasks) != 2 || desc.Tasks[0].ID != "t2" || desc.Tasks[1].ID != "t1" {
			t.Fatalf("got %+v, want t2 then t1 descending", desc.Tasks)
		}
	})

	t.Run("paginates results", func(t *testing.T) {
		svc := NewService(&fakeStorage{tasks: tasks}, validProject())

		got, err := svc.List("owner-1", "project-1", TaskQuery{Page: 1, Limit: 1, Sort: "title", Order: OrderAsc})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got.Tasks) != 1 || got.Tasks[0].Title != "Alpha" {
			t.Fatalf("got %+v, want only Alpha on page 1", got.Tasks)
		}
		if got.PaginationMeta.TotalPages != 2 {
			t.Fatalf("got %d total pages, want 2", got.PaginationMeta.TotalPages)
		}
	})

	t.Run("page beyond available results returns an empty slice", func(t *testing.T) {
		svc := NewService(&fakeStorage{tasks: tasks}, validProject())

		got, err := svc.List("owner-1", "project-1", TaskQuery{Page: 99, Limit: 10})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got.Tasks) != 0 {
			t.Fatalf("got %d tasks, want 0", len(got.Tasks))
		}
	})

	t.Run("project not found", func(t *testing.T) {
		svc := NewService(&fakeStorage{}, &fakeProjectService{
			getByIDFn: func(ownerID, projectID string) (*project.Project, error) {
				return nil, project.ErrProjectNotFound
			},
		})

		if _, err := svc.List("owner-1", "project-1", TaskQuery{Page: 1, Limit: 10}); !errors.Is(err, ErrProjectNotFound) {
			t.Fatalf("expected ErrProjectNotFound, got %v", err)
		}
	})

	t.Run("load failure is returned to the caller", func(t *testing.T) {
		svc := NewService(&fakeStorage{loadErr: errors.New("disk error")}, validProject())

		if _, err := svc.List("owner-1", "project-1", TaskQuery{Page: 1, Limit: 10}); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestService_Update(t *testing.T) {
	existing := Task{ID: "t1", ProjectID: "project-1", Title: "Old", Description: "Old desc"}

	t.Run("validation", func(t *testing.T) {
		cases := []struct {
			name        string
			ownerID     string
			projectID   string
			taskID      string
			title       string
			description string
			wantErr     error
		}{
			{"missing owner id", "", "project-1", "t1", "New", "New desc", ErrInvalidOwnerID},
			{"missing project id", "owner-1", "", "t1", "New", "New desc", ErrInvalidProjectID},
			{"missing task id", "owner-1", "project-1", "", "New", "New desc", ErrInvalidTaskID},
			{"missing title", "owner-1", "project-1", "t1", "", "New desc", ErrInvalidTaskTitle},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				svc := NewService(&fakeStorage{tasks: []Task{existing}}, validProject())

				_, err := svc.Update(tc.ownerID, tc.projectID, tc.taskID, tc.title, tc.description)
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("got error %v, want %v", err, tc.wantErr)
				}
			})
		}
	})

	t.Run("updates title and description", func(t *testing.T) {
		storage := &fakeStorage{tasks: []Task{existing}}
		svc := NewService(storage, validProject())

		got, err := svc.Update("owner-1", "project-1", "t1", "New", "New desc")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Title != "New" || got.Description != "New desc" {
			t.Fatalf("got %+v, want updated title/description", got)
		}
		if storage.tasks[0].Title != "New" {
			t.Fatalf("update was not persisted: %+v", storage.tasks[0])
		}
	})

	t.Run("not found for a different project", func(t *testing.T) {
		svc := NewService(&fakeStorage{tasks: []Task{existing}}, validProject())

		_, err := svc.Update("owner-1", "project-2", "t1", "New", "New desc")
		if !errors.Is(err, ErrTaskNotFound) {
			t.Fatalf("got error %v, want %v", err, ErrTaskNotFound)
		}
	})

	t.Run("project not found", func(t *testing.T) {
		svc := NewService(&fakeStorage{}, &fakeProjectService{
			getByIDFn: func(ownerID, projectID string) (*project.Project, error) {
				return nil, project.ErrProjectNotFound
			},
		})

		if _, err := svc.Update("owner-1", "project-1", "t1", "New", "New desc"); !errors.Is(err, ErrProjectNotFound) {
			t.Fatalf("expected ErrProjectNotFound, got %v", err)
		}
	})

	t.Run("load failure is returned to the caller", func(t *testing.T) {
		svc := NewService(&fakeStorage{loadErr: errors.New("disk error")}, validProject())

		if _, err := svc.Update("owner-1", "project-1", "t1", "New", "New desc"); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("save failure is returned to the caller", func(t *testing.T) {
		svc := NewService(&fakeStorage{tasks: []Task{existing}, saveErr: errors.New("disk full")}, validProject())

		if _, err := svc.Update("owner-1", "project-1", "t1", "New", "New desc"); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestService_UpdateStatus(t *testing.T) {
	existing := Task{ID: "t1", ProjectID: "project-1", Status: TaskStatusTodo}

	t.Run("validation", func(t *testing.T) {
		cases := []struct {
			name      string
			ownerID   string
			projectID string
			taskID    string
			wantErr   error
		}{
			{"missing owner id", "", "project-1", "t1", ErrInvalidOwnerID},
			{"missing project id", "owner-1", "", "t1", ErrInvalidProjectID},
			{"missing task id", "owner-1", "project-1", "", ErrInvalidTaskID},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				svc := NewService(&fakeStorage{tasks: []Task{existing}}, validProject())

				_, err := svc.UpdateStatus(tc.ownerID, tc.projectID, tc.taskID, TaskStatusDone)
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("got error %v, want %v", err, tc.wantErr)
				}
			})
		}
	})

	t.Run("invalid status value", func(t *testing.T) {
		svc := NewService(&fakeStorage{tasks: []Task{existing}}, validProject())

		_, err := svc.UpdateStatus("owner-1", "project-1", "t1", TaskStatus("bogus"))
		if !errors.Is(err, ErrInvalidTaskStatus) {
			t.Fatalf("got error %v, want %v", err, ErrInvalidTaskStatus)
		}
	})

	t.Run("invalid transition is rejected", func(t *testing.T) {
		done := Task{ID: "t1", ProjectID: "project-1", Status: TaskStatusDone}
		svc := NewService(&fakeStorage{tasks: []Task{done}}, validProject())

		_, err := svc.UpdateStatus("owner-1", "project-1", "t1", TaskStatusTodo)
		if !errors.Is(err, ErrInvalidTaskStatusTransition) {
			t.Fatalf("got error %v, want %v", err, ErrInvalidTaskStatusTransition)
		}
	})

	t.Run("valid transition is persisted", func(t *testing.T) {
		storage := &fakeStorage{tasks: []Task{existing}}
		svc := NewService(storage, validProject())

		got, err := svc.UpdateStatus("owner-1", "project-1", "t1", TaskStatusInProgress)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Status != TaskStatusInProgress {
			t.Fatalf("got status %q, want %q", got.Status, TaskStatusInProgress)
		}
		if storage.tasks[0].Status != TaskStatusInProgress {
			t.Fatalf("update was not persisted: %+v", storage.tasks[0])
		}
	})

	t.Run("in_progress to done is a valid transition", func(t *testing.T) {
		inProgress := Task{ID: "t1", ProjectID: "project-1", Status: TaskStatusInProgress}
		storage := &fakeStorage{tasks: []Task{inProgress}}
		svc := NewService(storage, validProject())

		got, err := svc.UpdateStatus("owner-1", "project-1", "t1", TaskStatusDone)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Status != TaskStatusDone {
			t.Fatalf("got status %q, want %q", got.Status, TaskStatusDone)
		}
	})

	t.Run("not found for a different project", func(t *testing.T) {
		svc := NewService(&fakeStorage{tasks: []Task{existing}}, validProject())

		_, err := svc.UpdateStatus("owner-1", "project-2", "t1", TaskStatusDone)
		if !errors.Is(err, ErrTaskNotFound) {
			t.Fatalf("got error %v, want %v", err, ErrTaskNotFound)
		}
	})

	t.Run("project not found", func(t *testing.T) {
		svc := NewService(&fakeStorage{}, &fakeProjectService{
			getByIDFn: func(ownerID, projectID string) (*project.Project, error) {
				return nil, project.ErrProjectNotFound
			},
		})

		if _, err := svc.UpdateStatus("owner-1", "project-1", "t1", TaskStatusDone); !errors.Is(err, ErrProjectNotFound) {
			t.Fatalf("expected ErrProjectNotFound, got %v", err)
		}
	})

	t.Run("load failure is returned to the caller", func(t *testing.T) {
		svc := NewService(&fakeStorage{loadErr: errors.New("disk error")}, validProject())

		if _, err := svc.UpdateStatus("owner-1", "project-1", "t1", TaskStatusDone); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("save failure is returned to the caller", func(t *testing.T) {
		svc := NewService(&fakeStorage{tasks: []Task{existing}, saveErr: errors.New("disk full")}, validProject())

		if _, err := svc.UpdateStatus("owner-1", "project-1", "t1", TaskStatusDone); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
