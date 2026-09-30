package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"taskflow-api/internal/auth"

	"github.com/lib/pq"
)

type AuthRepository struct {
	db *DB
}

func NewAuthRepository(db *DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) Create(ctx context.Context, user *auth.User) (*auth.User, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO users (name, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, password_hash, created_at
	`, user.Name, user.Email, user.PasswordHash, user.CreatedAt)

	var createdUser auth.User
	err := row.Scan(&createdUser.ID, &createdUser.Name, &createdUser.Email, &createdUser.PasswordHash, &createdUser.CreatedAt)
	if err != nil {
		if isUniqueViolationError(err) {
			slog.Error("Email already exists", "error", err)
			return nil, auth.ErrEmailAlreadyExists
		}
		slog.Error("Failed to create user", "error", err)
		return nil, err
	}
	slog.Info("User created successfully", "user_id", createdUser.ID, "email", createdUser.Email)
	return &createdUser, nil
}

func (r *AuthRepository) FindByID(ctx context.Context, id string) (*auth.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, password_hash, created_at
		FROM users
		WHERE id = $1
	`, id)

	var user auth.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Warn("User not found by ID", "user_id", id)
			return nil, auth.ErrUserNotFound
		}
		slog.Error("Failed to find user by ID", "user_id", id, "error", err)
		return nil, err
	}
	slog.Info("User found by ID", "user_id", user.ID, "email", user.Email)
	return &user, nil
}

func (r *AuthRepository) FindByEmail(ctx context.Context, email string) (*auth.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`, email)

	var user auth.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Warn("User not found by email", "email", email)
			return nil, auth.ErrUserNotFound
		}
		slog.Error("Failed to find user by email", "email", email, "error", err)
		return nil, err
	}
	slog.Info("User found by email", "user_id", user.ID, "email", user.Email)
	return &user, nil
}

func isUniqueViolationError(err error) bool {
	// Check if the error is a unique violation error (PostgreSQL error code 23505)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if pqErr.Code == "23505" {
			return true
		}
	}
	return false
}
