package task

import "context"

type Repository interface {
	Create(ctx context.Context, task *Task) (*Task, error)
	List(ctx context.Context, ownerID, projectID string, query TaskQuery) (*TaskListResult, error)
	GetByID(ctx context.Context, ownerID, projectID, taskID string) (*Task, error)
	Update(ctx context.Context, task *Task) (*Task, error)
	Delete(ctx context.Context, ownerID, projectID, taskID string) error
}
