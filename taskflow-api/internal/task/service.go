package task

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"taskflow-api/internal/project"
	"time"

	"github.com/google/uuid"
)

// Dependency Interface for ProjectService
type ProjectService interface {
	GetByID(ownerID, projectID string) (*project.Project, error)
}

type Service interface {
	Create(
		ownerID,
		projectID,
		title,
		description string,
	) (*Task, error)
	List(
		ownerID,
		projectID string,
		query TaskQuery,
	) (*TaskListResult, error)
	GetByID(
		ownerID,
		projectID,
		taskID string,
	) (*Task, error)
	Update(
		ownerID,
		projectID,
		taskID,
		title,
		description string,
	) (*Task, error)
	UpdateStatus(
		ownerID,
		projectID,
		taskID string,
		status TaskStatus,
	) (*Task, error)
	Delete(
		ownerID,
		projectID,
		taskID string,
	) error
}

type service struct {
	storage        Storage
	projectService ProjectService
}

// Create implements Service.
func (s *service) Create(ownerID string, projectID string, title string, description string) (*Task, error) {
	if ownerID == "" {
		return nil, ErrInvalidOwnerID
	}

	fmt.Printf("[Service Create] Creating task for ownerID: %s, projectID: %s, title: %s\n", ownerID, projectID, title)

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
	projectData, err := s.projectService.GetByID(ownerID, projectID)
	if err != nil {
		if errors.Is(err, project.ErrProjectNotFound) {
			fmt.Printf("[Service Create] Project with ID %s not found for owner %s\n", projectID, ownerID)
			return nil, ErrProjectNotFound
		}

		fmt.Printf("[Service Create] Failed to get project: %v\n", err)
		return nil, err
	}

	if projectData == nil {
		fmt.Printf("[Service Create] Project with ID %s not found for owner %s\n", projectID, ownerID)
		return nil, ErrProjectNotFound
	}

	tasks, err := s.storage.Load()
	if err != nil {
		fmt.Printf("[Service Create] Failed to load tasks: %v\n", err)
		return nil, err
	}

	timeNow := time.Now()

	task := &Task{
		ID:          uuid.NewString(),
		ProjectID:   projectID,
		Title:       title,
		Description: description,
		Status:      TaskStatusTodo,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}

	tasks = append(tasks, *task)

	err = s.storage.Save(tasks)
	if err != nil {
		fmt.Printf("[Service Create] Failed to save tasks: %v\n", err)
		return nil, err
	}

	return task, nil
}

// Delete implements Service.
func (s *service) Delete(ownerID string, projectID string, taskID string) error {
	if ownerID == "" {
		return ErrInvalidOwnerID
	}
	if projectID == "" {
		return ErrInvalidProjectID
	}
	if taskID == "" {
		return ErrInvalidTaskID
	}

	projectData, err := s.projectService.GetByID(ownerID, projectID)
	if err != nil {
		if errors.Is(err, project.ErrProjectNotFound) {
			fmt.Printf("[Service Delete] Project with ID %s not found for owner %s\n", projectID, ownerID)
			return ErrProjectNotFound
		}
		fmt.Printf("[Service Delete] Failed to get project: %v\n", err)
		return err
	}

	if projectData == nil {
		fmt.Printf("[Service Delete] Project with ID %s not found for owner %s\n", projectID, ownerID)
		return ErrProjectNotFound
	}

	tasks, err := s.storage.Load()
	if err != nil {
		fmt.Printf("[Service Delete] Failed to load tasks: %v\n", err)
		return err
	}

	for i, t := range tasks {
		if t.ID == taskID && t.ProjectID == projectID {
			tasks = append(tasks[:i], tasks[i+1:]...)
			err = s.storage.Save(tasks)
			if err != nil {
				fmt.Printf("[Service Delete] Failed to save tasks: %v\n", err)
				return err
			}
			return nil
		}
	}

	fmt.Printf("[Service Delete] Task with ID %s not found for project %s\n", taskID, projectID)
	return ErrTaskNotFound
}

// GetByID implements Service.
func (s *service) GetByID(ownerID string, projectID string, taskID string) (*Task, error) {
	if ownerID == "" {
		return nil, ErrInvalidOwnerID
	}
	if projectID == "" {
		return nil, ErrInvalidProjectID
	}
	if taskID == "" {
		return nil, ErrInvalidTaskID
	}

	projectData, err := s.projectService.GetByID(ownerID, projectID)
	if err != nil {
		if errors.Is(err, project.ErrProjectNotFound) {
			fmt.Printf("[Service GetByID] Project with ID %s not found for owner %s\n", projectID, ownerID)
			return nil, ErrProjectNotFound
		}
		fmt.Printf("[Service GetByID] Failed to get project: %v\n", err)
		return nil, err
	}

	if projectData == nil {
		fmt.Printf("[Service GetByID] Project with ID %s not found for owner %s\n", projectID, ownerID)
		return nil, ErrProjectNotFound
	}

	tasks, err := s.storage.Load()
	if err != nil {
		fmt.Printf("[Service GetByID] Failed to load tasks: %v\n", err)
		return nil, err
	}

	for _, t := range tasks {
		if t.ID == taskID && t.ProjectID == projectID {
			return &t, nil
		}
	}

	fmt.Printf("[Service GetByID] Task with ID %s not found for project %s\n", taskID, projectID)
	return nil, ErrTaskNotFound
}

// List implements Service.
func (s *service) List(ownerID string, projectID string, query TaskQuery) (*TaskListResult, error) {
	if ownerID == "" {
		return nil, ErrInvalidOwnerID
	}
	if projectID == "" {
		return nil, ErrInvalidProjectID
	}

	projectData, err := s.projectService.GetByID(ownerID, projectID)
	if err != nil {
		if errors.Is(err, project.ErrProjectNotFound) {
			fmt.Printf("[Service List] Project with ID %s not found for owner %s\n", projectID, ownerID)
			return nil, ErrProjectNotFound
		}
		fmt.Printf("[Service List] Failed to get project: %v\n", err)
		return nil, err
	}

	if projectData == nil {
		fmt.Printf("[Service List] Project with ID %s not found for owner %s\n", projectID, ownerID)
		return nil, ErrProjectNotFound
	}

	filtered := make([]Task, 0)
	var search string
	var status TaskStatus
	page := query.Page
	if page < 1 {
		return nil, ErrInvalidPageNumber
	}
	limit := query.Limit
	if limit < 1 || limit > 100 {
		return nil, ErrInvalidLimitNumber
	}

	sortBy := query.Sort
	if sortBy == "" {
		sortBy = "created_at"
	}

	order := query.Order
	if order == "" {
		order = OrderDesc
	}

	if query.Status != "" {
		if !isValidStatus(query.Status) {
			fmt.Printf("[Service List] Invalid status filter: %s\n", query.Status)
			return nil, ErrInvalidTaskStatus
		}

		status = query.Status
	}

	if query.Search != "" {
		search = strings.ToLower(query.Search)
	}

	if query.Sort != "" {
		if !isValidSort(query.Sort) {
			fmt.Printf("[Service List] Invalid sort field: %s\n", query.Sort)
			return nil, ErrInvalidSortField
		}
	}

	if query.Order != "" {
		if !isValidOrder(query.Order) {
			fmt.Printf("[Service List] Invalid order: %s\n", query.Order)
			return nil, ErrInvalidOrder
		}
	}

	tasks, err := s.storage.Load()
	if err != nil {
		fmt.Printf("[Service List] Failed to load tasks: %v\n", err)
		return nil, err
	}

	for _, t := range tasks {
		if t.ProjectID == projectID {
			filtered = append(filtered, t)
		}
	}

	// Apply status filter
	if status != "" {
		var statusFiltered []Task
		for _, t := range filtered {
			if t.Status == status {
				statusFiltered = append(statusFiltered, t)
			}
		}
		filtered = statusFiltered
	}

	// Apply search filter
	if search != "" {
		var searchFiltered []Task
		for _, t := range filtered {
			if strings.Contains(strings.ToLower(t.Title), search) || strings.Contains(strings.ToLower(t.Description), search) {
				searchFiltered = append(searchFiltered, t)
			}
		}
		filtered = searchFiltered
	}

	// Apply sorting
	switch sortBy {
	case "title":
		if order == OrderAsc {
			sort.Slice(filtered, func(i, j int) bool {
				return filtered[i].Title < filtered[j].Title
			})
		} else {
			sort.Slice(filtered, func(i, j int) bool {
				return filtered[i].Title > filtered[j].Title
			})
		}
	case "created_at":
		if order == OrderAsc {
			sort.Slice(filtered, func(i, j int) bool {
				return filtered[i].CreatedAt.Before(filtered[j].CreatedAt)
			})
		} else {
			sort.Slice(filtered, func(i, j int) bool {
				return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
			})
		}
	case "updated_at":
		if order == OrderAsc {
			sort.Slice(filtered, func(i, j int) bool {
				return filtered[i].UpdatedAt.Before(filtered[j].UpdatedAt)
			})
		} else {
			sort.Slice(filtered, func(i, j int) bool {
				return filtered[i].UpdatedAt.After(filtered[j].UpdatedAt)
			})
		}
	default:
		fmt.Printf("[Service List] Invalid sort field: %s\n", sortBy)
		return nil, ErrInvalidSortField
	}

	totalCount := len(filtered)

	// Apply pagination
	start := (page - 1) * limit
	end := start + limit

	totalPages := (totalCount + limit - 1) / limit

	if start >= len(filtered) {
		return &TaskListResult{
			Tasks: []Task{},
			PaginationMeta: PaginationMeta{
				Page:       page,
				Limit:      limit,
				Total:      totalCount,
				TotalPages: totalPages,
			},
		}, nil
	}

	if end > len(filtered) {
		end = len(filtered)
	}

	filtered = filtered[start:end]

	return &TaskListResult{
		Tasks: filtered,
		PaginationMeta: PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      totalCount,
			TotalPages: totalPages,
		},
	}, nil
}

// Update implements Service.
func (s *service) Update(ownerID string, projectID string, taskID string, title string, description string) (*Task, error) {
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

	projectData, err := s.projectService.GetByID(ownerID, projectID)
	if err != nil {
		if errors.Is(err, project.ErrProjectNotFound) {
			fmt.Printf("[Service Update] Project with ID %s not found for owner %s\n", projectID, ownerID)
			return nil, ErrProjectNotFound
		}
		fmt.Printf("[Service Update] Failed to get project: %v\n", err)
		return nil, err
	}

	if projectData == nil {
		fmt.Printf("[Service Update] Project with ID %s not found for owner %s\n", projectID, ownerID)
		return nil, ErrProjectNotFound
	}

	tasks, err := s.storage.Load()
	if err != nil {
		fmt.Printf("[Service Update] Failed to load tasks: %v\n", err)
		return nil, err
	}

	for i, t := range tasks {
		if t.ID == taskID && t.ProjectID == projectID {
			t.Title = title
			t.Description = description
			t.UpdatedAt = time.Now()
			tasks[i] = t

			err = s.storage.Save(tasks)
			if err != nil {
				fmt.Printf("[Service Update] Failed to save tasks: %v\n", err)
				return nil, err
			}

			return &t, nil
		}
	}

	fmt.Printf("[Service Update] Task with ID %s not found for project %s\n", taskID, projectID)
	return nil, ErrTaskNotFound
}

// UpdateStatus implements Service.
func (s *service) UpdateStatus(ownerID string, projectID string, taskID string, status TaskStatus) (*Task, error) {
	if ownerID == "" {
		return nil, ErrInvalidOwnerID
	}
	if projectID == "" {
		return nil, ErrInvalidProjectID
	}
	if taskID == "" {
		return nil, ErrInvalidTaskID
	}

	projectData, err := s.projectService.GetByID(ownerID, projectID)
	if err != nil {
		if errors.Is(err, project.ErrProjectNotFound) {
			fmt.Printf("[Service UpdateStatus] Project with ID %s not found for owner %s\n", projectID, ownerID)
			return nil, ErrProjectNotFound
		}
		fmt.Printf("[Service UpdateStatus] Failed to get project: %v\n", err)
		return nil, err
	}

	if projectData == nil {
		fmt.Printf("[Service UpdateStatus] Project with ID %s not found for owner %s\n", projectID, ownerID)
		return nil, ErrProjectNotFound
	}

	tasks, err := s.storage.Load()
	if err != nil {
		fmt.Printf("[Service UpdateStatus] Failed to load tasks: %v\n", err)
		return nil, err
	}

	if !isValidStatus(status) {
		fmt.Printf("[Service UpdateStatus] Invalid status: %s\n", status)
		return nil, ErrInvalidTaskStatus
	}

	for i, t := range tasks {
		if t.ID == taskID && t.ProjectID == projectID {
			if !isValidTransition(t.Status, status) {
				fmt.Printf("[Service UpdateStatus] Invalid status transition from %s to %s\n", t.Status, status)
				return nil, ErrInvalidTaskStatusTransition
			}

			t.Status = status
			t.UpdatedAt = time.Now()
			tasks[i] = t

			err = s.storage.Save(tasks)
			if err != nil {
				fmt.Printf("[Service UpdateStatus] Failed to save tasks: %v\n", err)
				return nil, err
			}

			return &t, nil
		}
	}

	fmt.Printf("[Service UpdateStatus] Task with ID %s not found for project %s\n", taskID, projectID)
	return nil, ErrTaskNotFound
}

func NewService(
	storage Storage,
	projectService ProjectService,
) Service {
	return &service{
		storage:        storage,
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
