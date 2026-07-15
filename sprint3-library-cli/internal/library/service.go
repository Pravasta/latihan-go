package library

import (
	"fmt"
	"time"
)

type Service struct {
	bookStorage BookStorage
	loanStorage LoanStorage
}

func NewService(bookStorage BookStorage, loanStorage LoanStorage) *Service {
	return &Service{
		bookStorage: bookStorage,
		loanStorage: loanStorage,
	}
}

func (s *Service) AddBook(title, author string) error {
	books, err := s.bookStorage.Load()
	if err != nil {
		fmt.Printf("[AddBook Error] Failed to load books: %v", err)
		return err
	}

	if title == "" {
		fmt.Printf("[AddBook Error] Title cannot be empty: %v", ErrTitleEmpty)
		return ErrTitleEmpty
	}

	if author == "" {
		fmt.Printf("[AddBook Error] Author cannot be empty: %v", ErrAuthorEmpty)
		return ErrAuthorEmpty
	}

	maxID := 0
	for _, book := range books {
		if book.ID > maxID {
			maxID = book.ID
		}
	}

	newBook := Book{
		ID:        maxID + 1,
		Title:     title,
		Author:    author,
		Available: true,
	}

	books = append(books, newBook)

	err = s.bookStorage.Save(books)
	if err != nil {
		fmt.Printf("[AddBook Error] Failed to save book: %v", err)
		return err
	}

	return nil
}

func (s *Service) ListBooks() ([]Book, error) {
	books, err := s.bookStorage.Load()
	if err != nil {
		fmt.Printf("[ListBooks Error] Failed to load books: %v", err)
		return nil, err
	}

	return books, nil
}

func (s *Service) BorrowBook(id int, borrower string) error {
	loans, err := s.loanStorage.Load()
	if err != nil {
		fmt.Printf("[BorrowBook Error] Failed to load loans: %v", err)
		return err
	}

	if borrower == "" {
		fmt.Printf("[BorrowBook Error] Borrower cannot be empty: %v", ErrBorrowerEmpty)
		return ErrBorrowerEmpty
	}

	books, err := s.bookStorage.Load()
	if err != nil {
		fmt.Printf("[BorrowBook Error] Failed to load books: %v", err)
		return err
	}

	var book *Book
	for i := range books {
		if books[i].ID == id {
			book = &books[i]
			break
		}
	}
	if book == nil {
		fmt.Printf("[BorrowBook Error] Book not found: %v", ErrBookNotFound)
		return ErrBookNotFound
	}

	if !book.Available {
		fmt.Printf("[BorrowBook Error] Book is already borrowed: %v", ErrBookAlreadyBorrowed)
		return ErrBookAlreadyBorrowed
	}

	book.Available = false

	err = s.bookStorage.Save(books)
	if err != nil {
		fmt.Printf("[BorrowBook Error] Failed to save book: %v", err)
		return err
	}

	var maxLoanID int
	for _, loan := range loans {
		if loan.ID > maxLoanID {
			maxLoanID = loan.ID
		}
	}

	var newLoan Loan
	newLoan.ID = maxLoanID + 1
	newLoan.BookID = book.ID
	newLoan.Borrower = borrower
	newLoan.BorrowedAt = time.Now()
	newLoan.ReturnedAt = nil

	loans = append(loans, newLoan)

	err = s.loanStorage.Save(loans)
	if err != nil {
		fmt.Printf("[BorrowBook Error] Failed to save loan: %v", err)
		return err
	}

	return nil
}

func (s *Service) ReturnBook(id int) error {
	loans, err := s.loanStorage.Load()
	if err != nil {
		fmt.Printf("[ReturnBook Error] Failed to load loans: %v", err)
		return err
	}

	var loan *Loan
	for i := range loans {
		if loans[i].ID == id {
			loan = &loans[i]
			break
		}
	}

	if loan == nil {
		fmt.Printf("[ReturnBook Error] Loan not found: %v", ErrLoanNotFound)
		return ErrLoanNotFound
	}

	if loan.ReturnedAt != nil {
		fmt.Printf("[ReturnBook Error] Loan is already returned: %v", ErrLoanAlreadyReturned)
		return ErrLoanAlreadyReturned
	}

	books, err := s.bookStorage.Load()
	if err != nil {
		fmt.Printf("[ReturnBook Error] Failed to load books: %v", err)
		return err
	}

	var book *Book
	for i := range books {
		if books[i].ID == loan.BookID {
			book = &books[i]
			break
		}
	}

	if book == nil {
		fmt.Printf("[ReturnBook Error] Book not found: %v", ErrBookNotFound)
		return ErrBookNotFound
	}

	book.Available = true

	err = s.bookStorage.Save(books)
	if err != nil {
		fmt.Printf("[ReturnBook Error] Failed to save book: %v", err)
		return err
	}

	now := time.Now()
	loan.ReturnedAt = &now

	err = s.loanStorage.Save(loans)
	if err != nil {
		fmt.Printf("[ReturnBook Error] Failed to save loan: %v", err)
		return err
	}

	return nil
}

func (s *Service) ListActiveLoans() ([]Loan, error) {
	loans, err := s.loanStorage.Load()
	if err != nil {
		fmt.Printf("[ListActiveLoans Error] Failed to load loans: %v", err)
		return nil, err
	}

	activeLoans := []Loan{}
	for _, loan := range loans {
		if loan.ReturnedAt == nil {
			activeLoans = append(activeLoans, loan)
		}
	}

	if len(activeLoans) == 0 {
		fmt.Println("No active loans found.")
		return nil, nil
	}

	return activeLoans, nil
}
