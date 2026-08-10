package project

import (
	"context"
	"database/sql"
)

// TaskDeleter is the minimal capability Service needs from the task
// package to cascade-delete a project's tasks inside the same
// transaction as the project deletion. project cannot import task (task
// already imports project, so the reverse would be an import cycle), so
// it depends on this narrow interface instead — task's postgres
// repository satisfies it structurally, with no import needed here.
type TaskDeleter interface {
	DeleteAllByProjectTx(ctx context.Context, tx *sql.Tx, projectID string) error
}

type Repository interface {
	Create(ctx context.Context, project *Project) (*Project, error)
	ListByOwner(ctx context.Context, ownerID string) ([]Project, error)
	GetByID(ctx context.Context, ownerID, projectID string) (*Project, error)
	Update(ctx context.Context, project *Project) (*Project, error)
	WithTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error
	DeleteTx(ctx context.Context, tx *sql.Tx, ownerID, projectID string) error
}
