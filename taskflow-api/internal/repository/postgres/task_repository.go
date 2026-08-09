package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	tsx "taskflow-api/internal/task"
)

type TaskRepository struct {
	db *DB
}

func NewTaskRepository(db *DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *tsx.Task) (*tsx.Task, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO tasks (
			project_id, title, description, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, project_id, title, description, status, created_at, updated_at
	`, task.ProjectID, task.Title, task.Description, task.Status, task.CreatedAt, task.UpdatedAt)

	var createdTask tsx.Task
	if err := row.Scan(&createdTask.ID, &createdTask.ProjectID, &createdTask.Title, &createdTask.Description, &createdTask.Status, &createdTask.CreatedAt, &createdTask.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	return &createdTask, nil
}

func (r *TaskRepository) List(ctx context.Context, ownerID, projectID string, query tsx.TaskQuery) (*tsx.TaskListResult, error) {
	whereQuery := `
		FROM tasks t
		JOIN projects p ON t.project_id = p.id
		WHERE p.owner_id = $1 AND t.project_id = $2
	`

	args := []any{ownerID, projectID}

	if query.Status != "" {
		placeholder := "$" + strconv.Itoa(len(args)+1)

		whereQuery += " AND t.status = " + placeholder
		args = append(args, query.Status)
	}

	if query.Search != "" {
		placeholder := "$" + strconv.Itoa(len(args)+1)
		whereQuery += `
			AND (
				LOWER(t.title) LIKE ` + placeholder + `
				OR 
				LOWER(t.description) LIKE ` + placeholder + `
			)
		`

		args = append(args, "%"+strings.ToLower(query.Search)+"%")
	}

	countQuery := `
		SELECT COUNT(*)
	` + whereQuery

	filteredArgs := append([]any{}, args...)

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, filteredArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count tasks: %w", err)
	}

	selectQuery := `
		SELECT t.id, t.project_id, t.title, t.description, t.status, t.created_at, t.updated_at
	` + whereQuery

	if query.Sort != "" {
		switch query.Sort {
		case "created_at":
			selectQuery += " ORDER BY t.created_at"
		case "updated_at":
			selectQuery += " ORDER BY t.updated_at"
		case "title":
			selectQuery += " ORDER BY t.title"
		default:
			return nil, fmt.Errorf("invalid sort field: %s", query.Sort)
		}
	} else {
		selectQuery += " ORDER BY t.created_at"
	}

	if query.Order != "" {
		switch query.Order {
		case tsx.OrderAsc:
			selectQuery += " ASC"
		case tsx.OrderDesc:
			selectQuery += " DESC"
		default:
			return nil, fmt.Errorf("invalid order: %s", query.Order)
		}
	} else {
		selectQuery += " DESC"
	}

	args = append(args, query.Limit)
	limitPlaceholder := "$" + strconv.Itoa(len(args))
	args = append(args, (query.Page-1)*query.Limit)
	offsetPlaceholder := "$" + strconv.Itoa(len(args))

	selectQuery += " LIMIT " + limitPlaceholder + " OFFSET " + offsetPlaceholder

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []tsx.Task
	for rows.Next() {
		var task tsx.Task
		if err := rows.Scan(&task.ID, &task.ProjectID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over task rows: %w", err)
	}

	totalPages := (total + query.Limit - 1) / query.Limit

	paginationMeta := tsx.PaginationMeta{
		Page:       query.Page,
		Limit:      query.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	return &tsx.TaskListResult{
		Tasks:          tasks,
		PaginationMeta: paginationMeta,
	}, nil
}

func (r *TaskRepository) GetByID(ctx context.Context, ownerID, projectID, taskID string) (*tsx.Task, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, project_id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE id = $1 AND project_id = $2
	`, taskID, projectID)

	var task tsx.Task
	if err := row.Scan(&task.ID, &task.ProjectID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, tsx.ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to get task by ID: %w", err)
	}
	return &task, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *tsx.Task) (*tsx.Task, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, updated_at = $4
		WHERE id = $5 AND project_id = $6
		RETURNING id, project_id, title, description, status, created_at, updated_at
	`, task.Title, task.Description, task.Status, task.UpdatedAt, task.ID, task.ProjectID)

	var updatedTask tsx.Task
	if err := row.Scan(&updatedTask.ID, &updatedTask.ProjectID, &updatedTask.Title, &updatedTask.Description, &updatedTask.Status, &updatedTask.CreatedAt, &updatedTask.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, tsx.ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to update task: %w", err)
	}
	return &updatedTask, nil
}

func (r *TaskRepository) Delete(ctx context.Context, ownerID, projectID, taskID string) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM tasks
		WHERE id = $1 AND project_id = $2
	`, taskID, projectID)

	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return tsx.ErrTaskNotFound
	}

	return nil
}
