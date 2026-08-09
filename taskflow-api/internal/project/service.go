package project

import (
	"context"
	"time"
)

type Service interface {
	Create(ctx context.Context, ownerID, name, description string) (*Project, error)
	ListByOwner(ctx context.Context, ownerID string) ([]Project, error)
	GetByID(ctx context.Context, ownerID, projectID string) (*Project, error)
	Update(
		ctx context.Context,
		ownerID,
		projectID,
		name,
		description string,
	) (*Project, error)
	Delete(ctx context.Context, ownerID, projectID string) error
}

type service struct {
	repo Repository
}

// Create implements Service.
func (s *service) Create(ctx context.Context, ownerID string, name string, description string) (*Project, error) {
	if ownerID == "" {
		return nil, ErrInvalidOwnerID
	}

	if name == "" {
		return nil, ErrInvalidProjectName
	}

	if description == "" {
		return nil, ErrInvalidProjectDescription
	}

	timeNow := time.Now()

	project := &Project{
		OwnerID:     ownerID,
		Name:        name,
		Description: description,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}

	result, err := s.repo.Create(ctx, project)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Delete implements Service.
func (s *service) Delete(ctx context.Context, ownerID string, projectID string) error {
	if ownerID == "" {
		return ErrInvalidOwnerID
	}

	if projectID == "" {
		return ErrInvalidProjectID
	}

	err := s.repo.Delete(ctx, ownerID, projectID)
	if err != nil {
		return err
	}

	return nil
}

// GetByID implements Service.
func (s *service) GetByID(ctx context.Context, ownerID string, projectID string) (*Project, error) {
	if ownerID == "" {
		return nil, ErrInvalidOwnerID
	}

	if projectID == "" {
		return nil, ErrInvalidProjectID
	}

	project, err := s.repo.GetByID(ctx, ownerID, projectID)
	if err != nil {
		return nil, err
	}

	return project, nil
}

// ListByOwner implements Service.
func (s *service) ListByOwner(ctx context.Context, ownerID string) ([]Project, error) {
	if ownerID == "" {
		return nil, ErrInvalidOwnerID
	}

	projects, err := s.repo.ListByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

// Update implements Service.
func (s *service) Update(ctx context.Context, ownerID string, projectID string, name string, description string) (*Project, error) {
	if ownerID == "" {
		return nil, ErrInvalidOwnerID
	}

	if projectID == "" {
		return nil, ErrInvalidProjectID
	}

	if name == "" {
		return nil, ErrInvalidProjectName
	}

	if description == "" {
		return nil, ErrInvalidProjectDescription
	}

	project := &Project{
		ID:          projectID,
		OwnerID:     ownerID,
		Name:        name,
		Description: description,
		UpdatedAt:   time.Now(),
	}

	result, err := s.repo.Update(ctx, project)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}
