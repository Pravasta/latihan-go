package project

import "errors"

var (
	ErrInvalidProjectName        = errors.New("project name is required")
	ErrProjectNotFound           = errors.New("project not found")
	ErrInvalidOwnerID            = errors.New("owner ID is required")
	ErrInvalidProjectDescription = errors.New("project description is required")
	ErrInvalidProjectID          = errors.New("project ID is required")
)
