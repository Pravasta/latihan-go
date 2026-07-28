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
	ErrInvalidSortField            = errors.New("invalid sort field")
	ErrInvalidOrder                = errors.New("invalid order")
	ErrInvalidPageNumber           = errors.New("invalid page number")
	ErrInvalidLimitNumber          = errors.New("invalid limit number")
)
