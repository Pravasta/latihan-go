package library

import (
	"testing"
	"time"
)

// DRIVEN TEST
/*
   	- TestAddBook
	- TestAddBookValidation
	- TestBorrowBook
	- TestBorrowUnavailableBook
	- TestBorrowBookNotFound
	- TestReturnBook
	- TestReturnLoanNotFound
	- TestReturnAlreadyReturnedLoan
	- TestListActiveLoans
*/

type fakeBookStorage struct {
	books []Book
}

type fakeLoanStorage struct {
	loans []Loan
}

func (f *fakeBookStorage) Load() ([]Book, error) {
	return f.books, nil
}

func (f *fakeBookStorage) Save(books []Book) error {
	f.books = books
	return nil
}

func (f *fakeLoanStorage) Load() ([]Loan, error) {
	return f.loans, nil
}

func (f *fakeLoanStorage) Save(loans []Loan) error {
	f.loans = loans
	return nil
}

// [TestAddBook] =>
func TestAddBook(t *testing.T) {
	fakeBookStorage := &fakeBookStorage{}
	fakeLoanStorage := &fakeLoanStorage{}

	service := NewService(fakeBookStorage, fakeLoanStorage)

	book := Book{
		Title:     "Menu Masakan Padang",
		Author:    "John Doe",
		Available: true,
	}

	err := service.AddBook(book.Title, book.Author)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(fakeBookStorage.books) != 1 {
		t.Fatalf("expected 1 book, got %d", len(fakeBookStorage.books))
	}

	got := fakeBookStorage.books[0]
	if got.Title != book.Title || got.Author != book.Author || got.Available != book.Available {
		t.Fatalf("expected %v, got %v", book, got)
	}
}

func TestAddBookValidation(t *testing.T) {
	fakeBookStorage := &fakeBookStorage{}
	fakeLoanStorage := &fakeLoanStorage{}

	service := NewService(fakeBookStorage, fakeLoanStorage)

	// Test empty title
	err := service.AddBook("", "John Doe")
	if err == nil || err != ErrTitleEmpty {
		t.Fatalf("expected error %v, got %v", ErrTitleEmpty, err)
	}

	// Test empty author
	err = service.AddBook("Menu Masakan Padang", "")
	if err == nil || err != ErrAuthorEmpty {
		t.Fatalf("expected error %v, got %v", ErrAuthorEmpty, err)
	}
}

func TestBorrowBook(t *testing.T) {
	fakeBookStorage := &fakeBookStorage{
		books: []Book{
			{ID: 1, Title: "Menu Masakan Padang", Author: "John Doe", Available: true},
		},
	}
	fakeLoanStorage := &fakeLoanStorage{}

	service := NewService(fakeBookStorage, fakeLoanStorage)

	err := service.BorrowBook(1, "Alice")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(fakeLoanStorage.loans) != 1 {
		t.Fatalf("expected 1 loan, got %d", len(fakeLoanStorage.loans))
	}

	got := fakeLoanStorage.loans[0]
	if got.BookID != 1 || got.Borrower != "Alice" || got.ReturnedAt != nil {
		t.Fatalf("expected loan with BookID=1, Borrower=Alice and ReturnedAt=nil, got %v", got)
	}
}

func TestBorrowUnavailableBook(t *testing.T) {
	fakeBookStorage := &fakeBookStorage{
		books: []Book{
			{ID: 1, Title: "Menu Masakan Padang", Author: "John Doe", Available: false},
		},
	}
	fakeLoanStorage := &fakeLoanStorage{}

	service := NewService(fakeBookStorage, fakeLoanStorage)

	err := service.BorrowBook(1, "Alice")
	if err == nil || err != ErrBookAlreadyBorrowed {
		t.Fatalf("expected error %v, got %v", ErrBookAlreadyBorrowed, err)
	}
}

func TestBorrowBookNotFound(t *testing.T) {
	fakeBookStorage := &fakeBookStorage{
		books: []Book{
			{ID: 1, Title: "Menu Masakan Padang", Author: "John Doe", Available: true},
		},
	}
	fakeLoanStorage := &fakeLoanStorage{}

	service := NewService(fakeBookStorage, fakeLoanStorage)

	err := service.BorrowBook(2, "Alice") // Book ID 2 does not exist
	if err == nil || err != ErrBookNotFound {
		t.Fatalf("expected error %v, got %v", ErrBookNotFound, err)
	}
}

func TestReturnBook(t *testing.T) {
	fakeBookStorage := &fakeBookStorage{
		books: []Book{
			{ID: 1, Title: "Menu Masakan Padang", Author: "John Doe", Available: false},
		},
	}
	fakeLoanStorage := &fakeLoanStorage{
		loans: []Loan{
			{ID: 1, BookID: 1, Borrower: "Alice", ReturnedAt: nil},
		},
	}

	service := NewService(fakeBookStorage, fakeLoanStorage)

	err := service.ReturnBook(1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(fakeLoanStorage.loans) != 1 {
		t.Fatalf("expected 1 loan, got %d", len(fakeLoanStorage.loans))
	}

	got := fakeLoanStorage.loans[0]
	if got.ID != 1 || got.BookID != 1 || got.Borrower != "Alice" || got.ReturnedAt == nil {
		t.Fatalf("expected loan with ID=1, BookID=1, Borrower=Alice and ReturnedAt!=nil, got %v", got)
	}
}

func TestReturnLoanNotFound(t *testing.T) {
	fakeBookStorage := &fakeBookStorage{
		books: []Book{
			{ID: 1, Title: "Menu Masakan Padang", Author: "John Doe", Available: false},
		},
	}
	fakeLoanStorage := &fakeLoanStorage{
		loans: []Loan{
			{ID: 1, BookID: 1, Borrower: "Alice", ReturnedAt: nil},
		},
	}

	service := NewService(fakeBookStorage, fakeLoanStorage)

	err := service.ReturnBook(2) // Loan ID 2 does not exist
	if err == nil || err != ErrLoanNotFound {
		t.Fatalf("expected error %v, got %v", ErrLoanNotFound, err)
	}
}

func TestReturnAlreadyReturnedLoan(t *testing.T) {
	fiveMinsAgo := time.Now().Add(-5 * time.Minute)
	returnedAt := fiveMinsAgo
	fakeBookStorage := &fakeBookStorage{
		books: []Book{
			{ID: 1, Title: "Menu Masakan Padang", Author: "John Doe", Available: true},
		},
	}
	fakeLoanStorage := &fakeLoanStorage{
		loans: []Loan{
			{ID: 1, BookID: 1, Borrower: "Alice", ReturnedAt: &returnedAt},
		},
	}

	service := NewService(fakeBookStorage, fakeLoanStorage)

	err := service.ReturnBook(1) // Loan ID 1 is already returned
	if err == nil {
		t.Fatalf("expected error %v, got nil", ErrLoanAlreadyReturned)
	}

	if err != ErrLoanAlreadyReturned {
		t.Fatalf("expected error %v, got %v", ErrLoanAlreadyReturned, err)
	}
}

func TestListActiveLoans(t *testing.T) {
	fakeBookStorage := &fakeBookStorage{
		books: []Book{
			{ID: 1, Title: "Menu Masakan Padang", Author: "John Doe", Available: false},
			{ID: 2, Title: "Menu Masakan Jawa", Author: "Jane Doe", Available: true},
		},
	}
	fakeLoanStorage := &fakeLoanStorage{
		loans: []Loan{
			{ID: 1, BookID: 1, Borrower: "Alice", ReturnedAt: nil},
			{ID: 2, BookID: 2, Borrower: "Bob", ReturnedAt: nil},
		},
	}

	service := NewService(fakeBookStorage, fakeLoanStorage)

	activeLoans, err := service.ListActiveLoans()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(activeLoans) != 2 {
		t.Fatalf("expected 2 active loans, got %d", len(activeLoans))
	}

	loan1 := activeLoans[0]
	if loan1.ID != 1 || loan1.BookID != 1 || loan1.Borrower != "Alice" || loan1.ReturnedAt != nil {
		t.Fatalf("expected loan with ID=1, BookID=1, Borrower=Alice and ReturnedAt=nil, got %v", loan1)
	}

	loan2 := activeLoans[1]
	if loan2.ID != 2 || loan2.BookID != 2 || loan2.Borrower != "Bob" || loan2.ReturnedAt != nil {
		t.Fatalf("expected loan with ID=2, BookID=2, Borrower=Bob and ReturnedAt=nil, got %v", loan2)
	}
}
