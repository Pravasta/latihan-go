package task

import "errors"

var (
	ErrInvalidTaskTitle            = errors.New("task title cannot be empty")
	ErrTaskNotFound                = errors.New("task not found")
	ErrInvalidTaskStatus           = errors.New("invalid task status")
	ErrInvalidTaskStatusTransition = errors.New("invalid task status transition")
	ErrInvalidOwnerID              = errors.New("invalid owner ID")
	ErrInvalidProjectID            = errors.New("invalid project ID")
	ErrInvalidTaskID               = errors.New("invalid task ID")
	ErrProjectNotFound             = errors.New("project not found")
)
