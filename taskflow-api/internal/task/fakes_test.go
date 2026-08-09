package task

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"taskflow-api/internal/project"
)

// Compile-time checks: if a fake drifts from the interface it stands in
// for (wrong method name, wrong signature), this fails the build instead
// of silently compiling as an unrelated type with no test ever catching it.
var (
	_ Repository     = (*fakeRepository)(nil)
	_ Service        = (*fakeService)(nil)
	_ ProjectService = (*fakeProjectService)(nil)
)

// fakeRepository is a test double for Repository. List mirrors the
// filter/search/sort/pagination semantics that TaskRepository.List pushes
// down to SQL, so service_test.go can exercise those paths without a real
// database.
type fakeRepository struct {
	tasks     []Task
	createErr error
	listErr   error
	getErr    error
	updateErr error
	deleteErr error
	nextID    int
}

func (f *fakeRepository) Create(ctx context.Context, task *Task) (*Task, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.nextID++
	created := *task
	created.ID = fmt.Sprintf("t%d", f.nextID)
	f.tasks = append(f.tasks, created)
	return &created, nil
}

func (f *fakeRepository) List(ctx context.Context, ownerID, projectID string, query TaskQuery) (*TaskListResult, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}

	var filtered []Task
	for _, t := range f.tasks {
		if t.ProjectID != projectID {
			continue
		}
		if query.Status != "" && t.Status != query.Status {
			continue
		}
		if query.Search != "" {
			search := strings.ToLower(query.Search)
			if !strings.Contains(strings.ToLower(t.Title), search) &&
				!strings.Contains(strings.ToLower(t.Description), search) {
				continue
			}
		}
		filtered = append(filtered, t)
	}

	less := func(i, j int) bool {
		switch query.Sort {
		case "title":
			return filtered[i].Title < filtered[j].Title
		case "updated_at":
			return filtered[i].UpdatedAt.Before(filtered[j].UpdatedAt)
		default:
			return filtered[i].CreatedAt.Before(filtered[j].CreatedAt)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if query.Order == OrderDesc {
			return less(j, i)
		}
		return less(i, j)
	})

	total := len(filtered)
	start := (query.Page - 1) * query.Limit
	end := start + query.Limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	totalPages := (total + query.Limit - 1) / query.Limit

	return &TaskListResult{
		Tasks: filtered[start:end],
		PaginationMeta: PaginationMeta{
			Page:       query.Page,
			Limit:      query.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func (f *fakeRepository) GetByID(ctx context.Context, ownerID, projectID, taskID string) (*Task, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	for _, t := range f.tasks {
		if t.ID == taskID && t.ProjectID == projectID {
			return &t, nil
		}
	}
	return nil, ErrTaskNotFound
}

func (f *fakeRepository) Update(ctx context.Context, task *Task) (*Task, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	for i, existing := range f.tasks {
		if existing.ID == task.ID && existing.ProjectID == task.ProjectID {
			updated := existing
			updated.Title = task.Title
			updated.Description = task.Description
			updated.Status = task.Status
			updated.UpdatedAt = task.UpdatedAt
			f.tasks[i] = updated
			return &updated, nil
		}
	}
	return nil, ErrTaskNotFound
}

func (f *fakeRepository) Delete(ctx context.Context, ownerID, projectID, taskID string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	for i, t := range f.tasks {
		if t.ID == taskID && t.ProjectID == projectID {
			f.tasks = append(f.tasks[:i], f.tasks[i+1:]...)
			return nil
		}
	}
	return ErrTaskNotFound
}

type fakeService struct {
	createFn       func(ctx context.Context, ownerID, projectID, title, description string) (*Task, error)
	listFn         func(ctx context.Context, ownerID, projectID string, query TaskQuery) (*TaskListResult, error)
	getByIDFn      func(ctx context.Context, ownerID, projectID, taskID string) (*Task, error)
	updateFn       func(ctx context.Context, ownerID, projectID, taskID, title, description string) (*Task, error)
	updateStatusFn func(ctx context.Context, ownerID, projectID, taskID string, status TaskStatus) (*Task, error)
	deleteFn       func(ctx context.Context, ownerID, projectID, taskID string) error
}

func (f *fakeService) Create(ctx context.Context, ownerID, projectID, title, description string) (*Task, error) {
	return f.createFn(ctx, ownerID, projectID, title, description)
}

func (f *fakeService) List(ctx context.Context, ownerID, projectID string, query TaskQuery) (*TaskListResult, error) {
	return f.listFn(ctx, ownerID, projectID, query)
}

func (f *fakeService) GetByID(ctx context.Context, ownerID, projectID, taskID string) (*Task, error) {
	return f.getByIDFn(ctx, ownerID, projectID, taskID)
}

func (f *fakeService) Update(ctx context.Context, ownerID, projectID, taskID, title, description string) (*Task, error) {
	return f.updateFn(ctx, ownerID, projectID, taskID, title, description)
}

func (f *fakeService) UpdateStatus(ctx context.Context, ownerID, projectID, taskID string, status TaskStatus) (*Task, error) {
	return f.updateStatusFn(ctx, ownerID, projectID, taskID, status)
}

func (f *fakeService) Delete(ctx context.Context, ownerID, projectID, taskID string) error {
	return f.deleteFn(ctx, ownerID, projectID, taskID)
}

type fakeProjectService struct {
	getByIDFn func(ctx context.Context, ownerID, projectID string) (*project.Project, error)
}

func (f *fakeProjectService) GetByID(ctx context.Context, ownerID, projectID string) (*project.Project, error) {
	return f.getByIDFn(ctx, ownerID, projectID)
}
