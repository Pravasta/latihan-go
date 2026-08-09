package project

import (
	"context"
	"fmt"
)

// Compile-time checks: if a fake drifts from the interface it stands in
// for (wrong method name, wrong signature), this fails the build instead
// of silently compiling as an unrelated type with no test ever catching it.
var (
	_ Repository = (*fakeRepository)(nil)
	_ Service    = (*fakeService)(nil)
)

// fakeRepository is a test double for Repository. It keeps projects in
// memory and lets a test force any operation to fail, so service_test.go
// can exercise error paths without touching a real database.
type fakeRepository struct {
	projects  []Project
	createErr error
	listErr   error
	getErr    error
	updateErr error
	deleteErr error
	nextID    int
}

func (f *fakeRepository) Create(ctx context.Context, p *Project) (*Project, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.nextID++
	created := *p
	created.ID = fmt.Sprintf("p%d", f.nextID)
	f.projects = append(f.projects, created)
	return &created, nil
}

func (f *fakeRepository) ListByOwner(ctx context.Context, ownerID string) ([]Project, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	var result []Project
	for _, p := range f.projects {
		if p.OwnerID == ownerID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (f *fakeRepository) GetByID(ctx context.Context, ownerID, projectID string) (*Project, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	for _, p := range f.projects {
		if p.ID == projectID && p.OwnerID == ownerID {
			return &p, nil
		}
	}
	return nil, ErrProjectNotFound
}

func (f *fakeRepository) Update(ctx context.Context, p *Project) (*Project, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	for i, existing := range f.projects {
		if existing.ID == p.ID && existing.OwnerID == p.OwnerID {
			updated := existing
			updated.Name = p.Name
			updated.Description = p.Description
			updated.UpdatedAt = p.UpdatedAt
			f.projects[i] = updated
			return &updated, nil
		}
	}
	return nil, ErrProjectNotFound
}

func (f *fakeRepository) Delete(ctx context.Context, ownerID, projectID string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	for i, p := range f.projects {
		if p.ID == projectID && p.OwnerID == ownerID {
			f.projects = append(f.projects[:i], f.projects[i+1:]...)
			return nil
		}
	}
	return ErrProjectNotFound
}

// fakeService is a test double for Service, used by handler_test.go. Each
// method is a func field so a test can plug in only the behavior it needs
// without maintaining in-memory state.
type fakeService struct {
	createFn func(ctx context.Context, ownerID, name, description string) (*Project, error)
	listFn   func(ctx context.Context, ownerID string) ([]Project, error)
	getFn    func(ctx context.Context, ownerID, projectID string) (*Project, error)
	updateFn func(ctx context.Context, ownerID, projectID, name, description string) (*Project, error)
	deleteFn func(ctx context.Context, ownerID, projectID string) error
}

func (f *fakeService) Create(ctx context.Context, ownerID, name, description string) (*Project, error) {
	return f.createFn(ctx, ownerID, name, description)
}

func (f *fakeService) ListByOwner(ctx context.Context, ownerID string) ([]Project, error) {
	return f.listFn(ctx, ownerID)
}

func (f *fakeService) GetByID(ctx context.Context, ownerID, projectID string) (*Project, error) {
	return f.getFn(ctx, ownerID, projectID)
}

func (f *fakeService) Update(ctx context.Context, ownerID, projectID, name, description string) (*Project, error) {
	return f.updateFn(ctx, ownerID, projectID, name, description)
}

func (f *fakeService) Delete(ctx context.Context, ownerID, projectID string) error {
	return f.deleteFn(ctx, ownerID, projectID)
}
