package auth

import (
	"context"
	"strings"
	"taskflow-api/internal/common"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	CreateUser(ctx context.Context, name, email, password string) (*User, error)
	Authenticate(ctx context.Context, email, password string) (string, error)
	Me(ctx context.Context, userID string) (*User, error)
}

type service struct {
	repo Repository
	jwt  *JWTService
}

// Authenticate implements Service.
func (s *service) Authenticate(ctx context.Context, email string, password string) (string, error) {
	email = strings.TrimSpace(email)
	if !common.IsValidEmail(email) {
		return "", ErrInvalidEmail
	}

	password = strings.TrimSpace(password)
	if !common.IsValidPassword(password) {
		return "", ErrInvalidPassword
	}

	user, err := s.repo.FindByEmail(ctx, email)

	if err != nil {
		return "", err
	}

	if !CheckPasswordHash(password, user.PasswordHash) {
		return "", ErrInvalidCredentials
	}

	token, err := s.jwt.Generate(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

// CreateUser implements Service.
func (s *service) CreateUser(ctx context.Context, name string, email string, password string) (*User, error) {

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidName
	}

	email = strings.TrimSpace(email)
	if !common.IsValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	password = strings.TrimSpace(password)
	if !common.IsValidPassword(password) {
		return nil, ErrInvalidPassword
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:           uuid.New().String(),
		Name:         name,
		Email:        email,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
	}

	user, err = s.repo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Me implements Service.
func (s *service) Me(ctx context.Context, userID string) (*User, error) {
	if userID == "" {
		return nil, ErrUserNotFound
	}

	user, err := s.repo.FindByID(ctx, userID)

	if err != nil {
		return nil, err
	}

	return user, nil

}

func NewService(repo Repository, jwt *JWTService) Service {
	return &service{
		repo: repo,
		jwt:  jwt,
	}
}
