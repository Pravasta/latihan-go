package auth

import (
	"fmt"
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
	if common.IsValidEmail(email) == false {
		fmt.Println("[Service Authenticate] Invalid email format")
		return "", ErrInvalidEmail
	}

	password = strings.TrimSpace(password)
	if common.IsValidPassword(password) == false {
		fmt.Println("[Service Authenticate] Invalid password format")
		return "", ErrInvalidPassword
	}

	users, err := s.storage.Load()
	if err != nil {
		fmt.Printf("[Service Authenticate] Failed to load users: %v\n", err)
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
		fmt.Println("[Service Authenticate] User not found")
		return "", ErrUserNotFound
	}

	if !CheckPasswordHash(password, foundUser.PasswordHash) {
		fmt.Println("[Service Authenticate] Invalid password")
		return "", ErrInvalidPassword
	}

	token, err := s.jwt.Generate(foundUser.ID)
	if err != nil {
		fmt.Printf("[Service Authenticate] Failed to generate token: %v\n", err)
		return "", err
	}

	return token, nil
}

// CreateUser implements Service.
func (s *service) CreateUser(name string, email string, password string) (*User, error) {

	name = strings.TrimSpace(name)
	if name == "" {
		fmt.Println("[Service CreateUser] Name cannot be empty")
		return nil, ErrInvalidName
	}

	email = strings.TrimSpace(email)
	if common.IsValidEmail(email) == false {
		fmt.Println("[Service CreateUser] Invalid email format")
		return nil, ErrInvalidEmail
	}

	password = strings.TrimSpace(password)
	if common.IsValidPassword(password) == false {
		fmt.Println("[Service CreateUser] Invalid password format")
		return nil, ErrInvalidPassword
	}

	users, err := s.storage.Load()
	if err != nil {
		fmt.Printf("[Service CreateUser] Failed to load users: %v\n", err)
		return nil, err
	}

	for _, u := range users {
		if u.Email == email {
			fmt.Println("[Service CreateUser] Email already exists")
			return nil, ErrEmailAlreadyExists
		}
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		fmt.Printf("[Service CreateUser] Failed to hash password: %v\n", err)
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
		fmt.Printf("[Service CreateUser] Failed to save users: %v\n", err)
		return nil, err
	}

	return user, nil
}

// Me implements Service.
func (s *service) Me(userID string) (*User, error) {
	if userID == "" {
		fmt.Println("[Service Me] User ID cannot be empty")
		return nil, ErrUserNotFound
	}

	users, err := s.storage.Load()
	if err != nil {
		fmt.Printf("[Service Me] Failed to load users: %v\n", err)
		return nil, err
	}

	for _, u := range users {
		if u.ID == userID {
			return &u, nil
		}
	}

	fmt.Println("[Service Me] User not found")
	return nil, ErrUserNotFound
}

func NewService(storage Storage, jwt *JWTService) Service {
	return &service{
		storage: storage,
		jwt:     jwt,
	}
}
