package library

import (
	"encoding/json"
	"fmt"
	"os"
)

type BookStorage interface {
	Load() ([]Book, error)
	Save(books []Book) error
}

type LoanStorage interface {
	Load() ([]Loan, error)
	Save(loans []Loan) error
}

type bookStorage struct {
	path string
}

type loanStorage struct {
	path string
}

// Load implements [LoanStorage].
func (l *loanStorage) Load() ([]Loan, error) {
	file, err := os.Open(l.path)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		return nil, err
	}

	defer file.Close()

	var loans []Loan
	if err := json.NewDecoder(file).Decode(&loans); err != nil {
		fmt.Printf("Failed to decode loans: %v\n", err)
		return nil, err
	}
	return loans, nil
}

// Save implements [LoanStorage].
func (l *loanStorage) Save(loans []Loan) error {
	data, err := json.MarshalIndent(loans, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal loans: %v\n", err)
		return err
	}

	err = os.WriteFile(l.path, data, 0644)
	if err != nil {
		fmt.Printf("Failed to write to file: %v\n", err)
		return err
	}

	return nil
}

// Load implements [BookStorage].
func (b *bookStorage) Load() ([]Book, error) {
	file, err := os.Open(b.path)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		return nil, err
	}

	defer file.Close()

	var books []Book
	if err := json.NewDecoder(file).Decode(&books); err != nil {
		fmt.Printf("Failed to decode books: %v\n", err)
		return nil, err
	}
	return books, nil
}

// Save implements [BookStorage].
func (b *bookStorage) Save(books []Book) error {
	data, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal books: %v\n", err)
		return err
	}

	err = os.WriteFile(b.path, data, 0644)
	if err != nil {
		fmt.Printf("Failed to write to file: %v\n", err)
		return err
	}

	return nil
}

func NewBookStorage(path string) BookStorage {
	return &bookStorage{path: path}
}

func NewLoanStorage(path string) LoanStorage {
	return &loanStorage{path: path}
}
