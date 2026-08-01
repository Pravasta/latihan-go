package auth

import (
	"strings"
	"taskflow-api/internal/common"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	CreateUser(name, email, password string) (*User, error)
	Authenticate(email, password string) (string, error)
	Me(userID string) (*User, error)
}

type service struct {
	storage Storage
	jwt     *JWTService
}

// Authenticate implements Service.
func (s *service) Authenticate(email string, password string) (string, error) {
	email = strings.TrimSpace(email)
	if !common.IsValidEmail(email) {
		return "", ErrInvalidEmail
	}

	password = strings.TrimSpace(password)
	if !common.IsValidPassword(password) {
		return "", ErrInvalidPassword
	}

	users, err := s.storage.Load()
	if err != nil {
		return "", err
	}

	var foundUser *User
	for _, u := range users {
		if u.Email == email {
			foundUser = &u
			break
		}
	}

	if foundUser == nil {
		return "", ErrUserNotFound
	}

	if !CheckPasswordHash(password, foundUser.PasswordHash) {
		return "", ErrInvalidCredentials
	}

	token, err := s.jwt.Generate(foundUser.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

// CreateUser implements Service.
func (s *service) CreateUser(name string, email string, password string) (*User, error) {

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

	users, err := s.storage.Load()
	if err != nil {
		return nil, err
	}

	for _, u := range users {
		if u.Email == email {
			return nil, ErrEmailAlreadyExists
		}
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:           uuid.NewString(),
		Name:         name,
		Email:        email,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
	}

	users = append(users, *user)

	if err := s.storage.Save(users); err != nil {
		return nil, err
	}

	return user, nil
}

// Me implements Service.
func (s *service) Me(userID string) (*User, error) {
	if userID == "" {
		return nil, ErrUserNotFound
	}

	users, err := s.storage.Load()
	if err != nil {
		return nil, err
	}

	for _, u := range users {
		if u.ID == userID {
			return &u, nil
		}
	}

	return nil, ErrUserNotFound
}

func NewService(storage Storage, jwt *JWTService) Service {
	return &service{
		storage: storage,
		jwt:     jwt,
	}
}
