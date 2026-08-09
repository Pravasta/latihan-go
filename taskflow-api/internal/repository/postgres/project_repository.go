package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"taskflow-api/internal/project"
)

type ProjectRepository struct {
	db *DB
}

func NewProjectRepository(db *DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, proj *project.Project) (*project.Project, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO projects (
			owner_id, name, description, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, owner_id, name, description, created_at, updated_at
	`, proj.OwnerID, proj.Name, proj.Description, proj.CreatedAt, proj.UpdatedAt)

	var createdProject project.Project
	if err := row.Scan(&createdProject.ID, &createdProject.OwnerID, &createdProject.Name, &createdProject.Description, &createdProject.CreatedAt, &createdProject.UpdatedAt); err != nil {
		return nil, err
	}
	return &createdProject, nil
}

func (r *ProjectRepository) ListByOwner(ctx context.Context, ownerID string) ([]project.Project, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, owner_id, name, description, created_at, updated_at
		FROM projects
		WHERE owner_id = $1
	`, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects by owner: %w", err)
	}
	defer rows.Close()

	var projects []project.Project
	for rows.Next() {
		var proj project.Project
		if err := rows.Scan(&proj.ID, &proj.OwnerID, &proj.Name, &proj.Description, &proj.CreatedAt, &proj.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, proj)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, ownerID, projectID string) (*project.Project, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, owner_id, name, description, created_at, updated_at
		FROM projects
		WHERE owner_id = $1 AND id = $2
	`, ownerID, projectID)

	var proj project.Project
	if err := row.Scan(&proj.ID, &proj.OwnerID, &proj.Name, &proj.Description, &proj.CreatedAt, &proj.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, project.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project by ID: %w", err)
	}
	return &proj, nil
}

func (r *ProjectRepository) Update(ctx context.Context, proj *project.Project) (*project.Project, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE projects
		SET name = $1, description = $2, updated_at = $3
		WHERE owner_id = $4 AND id = $5
		RETURNING id, owner_id, name, description, created_at, updated_at
	`, proj.Name, proj.Description, proj.UpdatedAt, proj.OwnerID, proj.ID)

	var updatedProject project.Project
	if err := row.Scan(&updatedProject.ID, &updatedProject.OwnerID, &updatedProject.Name, &updatedProject.Description, &updatedProject.CreatedAt, &updatedProject.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, project.ErrProjectNotFound
		}
		return nil, err
	}
	return &updatedProject, nil
}

func (r *ProjectRepository) Delete(ctx context.Context, ownerID, projectID string) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM projects
		WHERE owner_id = $1 AND id = $2
	`, ownerID, projectID)

	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if affectedRows == 0 {
		return project.ErrProjectNotFound
	}

	return nil
}
