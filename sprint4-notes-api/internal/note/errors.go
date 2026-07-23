package note

import "errors"

var (
	ErrTitleEmpty       = errors.New("note title cannot be empty")
	ErrContentEmpty     = errors.New("note content cannot be empty")
	ErrInvalidJSON      = errors.New("invalid request body")
	ErrInvalidID        = errors.New("invalid note ID")
	ErrNoteNotFound     = errors.New("note not found")
	ErrMethodNotAllowed = errors.New("method not allowed")
)
