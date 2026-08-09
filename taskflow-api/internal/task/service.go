package task

import (
	"context"
	"errors"
	"strings"
	"taskflow-api/internal/project"
	"time"

	"github.com/google/uuid"
)

// Dependency Interface for ProjectService
type ProjectService interface {
	GetByID(ctx context.Context, ownerID, projectID string) (*project.Project, error)
}

type Service interface {
	Create(
		ctx context.Context,
		ownerID,
		projectID,
		title,
		description string,
	) (*Task, error)
	List(
		ctx context.Context,
		ownerID,
		projectID string,
		query TaskQuery,
	) (*TaskListResult, error)
	GetByID(
		ctx context.Context,
		ownerID,
		projectID,
		taskID string,
	) (*Task, error)
	Update(
		ctx context.Context,
		ownerID,
		projectID,
		taskID,
		title,
		description string,
	) (*Task, error)
	UpdateStatus(
		ctx context.Context,
		ownerID,
		projectID,
		taskID string,
		status TaskStatus,
	) (*Task, error)
	Delete(
		ctx context.Context,
		ownerID,
		projectID,
		taskID string,
	) error
}

type service struct {
	repo           Repository
	projectService ProjectService
}

// Create implements Service.
func (s *service) Create(ctx context.Context, ownerID string, projectID string, title string, description string) (*Task, error) {
	if ownerID == "" {
		return nil, ErrInvalidOwnerID
	}

	newProjectID := strings.TrimSpace(projectID)
	if newProjectID == "" {
		return nil, ErrInvalidProjectID
	}

	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if title == "" {
		return nil, ErrInvalidTaskTitle
	}

	// Check if the project exists and belongs to the owner
	if err := ensureProjectExists(ctx, s.projectService, ownerID, newProjectID); err != nil {
		return nil, err
	}

	timeNow := time.Now()

	newTask := &Task{
		ID:          uuid.NewString(),
		ProjectID:   newProjectID,
		Title:       title,
		Description: description,
		Status:      TaskStatusTodo,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}

	task, err := s.repo.Create(ctx, newTask)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// Delete implements Service.
func (s *service) Delete(ctx context.Context, ownerID string, projectID string, taskID string) error {
	if ownerID == "" {
		return ErrInvalidOwnerID
	}
	if projectID == "" {
		return ErrInvalidProjectID
	}
	if taskID == "" {
		return ErrInvalidTaskID
	}

	if err := ensureProjectExists(ctx, s.projectService, ownerID, projectID); err != nil {
		return err
	}

	err := s.repo.Delete(ctx, ownerID, projectID, taskID)
	if err != nil {
		return err
	}

	return nil
}

// GetByID implements Service.
func (s *service) GetByID(ctx context.Context, ownerID string, projectID string, taskID string) (*Task, error) {
	if ownerID == "" {
		return nil, ErrInvalidOwnerID
	}
	if projectID == "" {
		return nil, ErrInvalidProjectID
	}
	if taskID == "" {
		return nil, ErrInvalidTaskID
	}

	if err := ensureProjectExists(ctx, s.projectService, ownerID, projectID); err != nil {
		return nil, err
	}

	task, err := s.repo.GetByID(ctx, ownerID, projectID, taskID)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// List implements Service.
func (s *service) List(ctx context.Context, ownerID string, projectID string, query TaskQuery) (*TaskListResult, error) {
	if ownerID == "" {
		return nil, ErrInvalidOwnerID
	}
	if projectID == "" {
		return nil, ErrInvalidProjectID
	}

	if err := ensureProjectExists(ctx, s.projectService, ownerID, projectID); err != nil {
		return nil, err
	}

	if query.Page < 1 {
		return nil, ErrInvalidPageNumber
	}

	if query.Limit < 1 || query.Limit > 100 {
		return nil, ErrInvalidLimitNumber
	}

	if query.Sort == "" {
		query.Sort = "created_at"
	}

	if query.Order == "" {
		query.Order = OrderDesc
	}

	if query.Status != "" {
		if !isValidStatus(query.Status) {
			return nil, ErrInvalidTaskStatus
		}

	}

	if query.Sort != "" {
		if !isValidSort(query.Sort) {
			return nil, ErrInvalidSortField
		}
	}

	if query.Order != "" {
		if !isValidOrder(query.Order) {
			return nil, ErrInvalidOrder
		}
	}

	tasks, err := s.repo.List(ctx, ownerID, projectID, query)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// Update implements Service.
func (s *service) Update(ctx context.Context, ownerID string, projectID string, taskID string, title string, description string) (*Task, error) {
	if ownerID == "" {
		return nil, ErrInvalidOwnerID
	}
	if projectID == "" {
		return nil, ErrInvalidProjectID
	}
	if taskID == "" {
		return nil, ErrInvalidTaskID
	}
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if title == "" {
		return nil, ErrInvalidTaskTitle
	}

	if err := ensureProjectExists(ctx, s.projectService, ownerID, projectID); err != nil {
		return nil, err
	}

	// Update overwrites every column the repository's SQL touches, so the
	// current status must be carried forward explicitly — otherwise editing
	// title/description would silently reset status to empty.
	existing, err := s.repo.GetByID(ctx, ownerID, projectID, taskID)
	if err != nil {
		return nil, err
	}

	task, err := s.repo.Update(ctx, &Task{
		ID:          taskID,
		ProjectID:   projectID,
		Title:       title,
		Description: description,
		Status:      existing.Status,
		UpdatedAt:   time.Now(),
	})

	if err != nil {
		return nil, err
	}

	return task, nil

}

// UpdateStatus implements Service.
func (s *service) UpdateStatus(ctx context.Context, ownerID string, projectID string, taskID string, status TaskStatus) (*Task, error) {
	if ownerID == "" {
		return nil, ErrInvalidOwnerID
	}
	if projectID == "" {
		return nil, ErrInvalidProjectID
	}
	if taskID == "" {
		return nil, ErrInvalidTaskID
	}

	if err := ensureProjectExists(ctx, s.projectService, ownerID, projectID); err != nil {
		return nil, err
	}

	task, err := s.repo.GetByID(ctx, ownerID, projectID, taskID)
	if err != nil {
		return nil, err
	}

	if !isValidStatus(status) {
		return nil, ErrInvalidTaskStatus
	}

	if !isValidTransition(task.Status, status) {
		return nil, ErrInvalidTaskStatusTransition
	}

	// Same reasoning as Update: carry forward title/description so this
	// status-only change doesn't blank them out in the repository's SET.
	newTask, err := s.repo.Update(ctx, &Task{
		ID:          taskID,
		ProjectID:   projectID,
		Title:       task.Title,
		Description: task.Description,
		Status:      status,
		UpdatedAt:   time.Now(),
	})

	if err != nil {
		return nil, err
	}

	return newTask, nil
}

func NewService(
	repo Repository,
	projectService ProjectService,
) Service {
	return &service{
		repo:           repo,
		projectService: projectService,
	}
}

func isValidStatus(status TaskStatus) bool {
	switch status {
	case TaskStatusTodo, TaskStatusInProgress, TaskStatusDone:
		return true
	default:
		return false
	}
}

func isValidTransition(currentStatus, newStatus TaskStatus) bool {
	switch currentStatus {
	case TaskStatusTodo:
		return newStatus == TaskStatusInProgress || newStatus == TaskStatusDone
	case TaskStatusInProgress:
		return newStatus == TaskStatusDone
	case TaskStatusDone:
		return false
	default:
		return false
	}
}

func isValidSort(sort string) bool {
	switch sort {
	case "title", "created_at", "updated_at":
		return true
	default:
		return false
	}
}

func isValidOrder(order Order) bool {
	switch order {
	case OrderAsc, OrderDesc:
		return true
	default:
		return false
	}
}

// ensureProjectExists translates a project-lookup failure into task's own
// ErrProjectNotFound. It fully owns that translation, so callers just need
// to check err != nil — they should not re-inspect the error afterwards.
func ensureProjectExists(ctx context.Context, projectService ProjectService, ownerID, projectID string) error {
	projectData, err := projectService.GetByID(ctx, ownerID, projectID)
	if err != nil {
		if errors.Is(err, project.ErrProjectNotFound) {
			return ErrProjectNotFound
		}

		return err
	}

	if projectData == nil {
		return ErrProjectNotFound
	}

	return nil
}
