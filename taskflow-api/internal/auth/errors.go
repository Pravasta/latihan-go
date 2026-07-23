package auth

import "errors"

var (
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidPassword    = errors.New("invalid password format")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUserNotFound       = errors.New("user not found")
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters long")
	ErrInvalidName        = errors.New("name cannot be empty")
	ErrInvalidRequestBody = errors.New("invalid request body")
)
