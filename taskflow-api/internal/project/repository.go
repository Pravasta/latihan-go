package project

import "context"

type Repository interface {
	Create(ctx context.Context, project *Project) (*Project, error)
	ListByOwner(ctx context.Context, ownerID string) ([]Project, error)
	GetByID(ctx context.Context, ownerID, projectID string) (*Project, error)
	Update(ctx context.Context, project *Project) (*Project, error)
	Delete(ctx context.Context, ownerID, projectID string) error
}
