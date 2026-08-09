package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log"
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
			log.Printf("[AuthRepository-ERROR] Email already exists: %v", err)
			return nil, auth.ErrEmailAlreadyExists
		}
		log.Printf("[AuthRepository-ERROR] Failed to create user: %v", err)
		return nil, err
	}
	log.Printf("[AuthRepository-INFO] User created successfully: %v", createdUser)
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
			log.Printf("[AuthRepository-ERROR] User not found: %v", err)
			return nil, auth.ErrUserNotFound
		}
		log.Printf("[AuthRepository-ERROR] Failed to find user by ID: %v", err)
		return nil, err
	}
	log.Printf("[AuthRepository-INFO] User found by ID: %v", user)
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
			log.Printf("[AuthRepository-ERROR] User not found: %v", err)
			return nil, auth.ErrUserNotFound
		}
		log.Printf("[AuthRepository-ERROR] Failed to find user by email: %v", err)
		return nil, err
	}
	log.Printf("[AuthRepository-INFO] User found by email: %v", user)
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
