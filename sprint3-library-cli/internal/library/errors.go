package library

import "errors"

var (
	ErrTitleEmpty          = errors.New("title cannot be empty")
	ErrAuthorEmpty         = errors.New("author cannot be empty")
	ErrBorrowerEmpty       = errors.New("borrower cannot be empty")
	ErrBookNotAvailable    = errors.New("book is not available for borrowing")
	ErrBookNotFound        = errors.New("book not found")
	ErrBookAlreadyBorrowed = errors.New("book is already borrowed")
	ErrLoanNotFound        = errors.New("loan not found")
	ErrLoanAlreadyReturned = errors.New("loan is already returned")
)
