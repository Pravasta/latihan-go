package cmd

import (
	"fmt"
	"os"
	"sprint3-library-cli/internal/library"
	"strconv"
)

func Execute() error {
	bookStorage := library.NewBookStorage("data/books.json")
	loanStorage := library.NewLoanStorage("data/loans.json")

	service := library.NewService(bookStorage, loanStorage)

	args := os.Args

	if len(args) < 2 {
		return nil
	}

	switch args[1] {
	case "book":
		return handleBookCommand(service, args[2:])
	case "loan":
		return handleLoanCommand(service, args[2:])
	case "borrow":
		return handleBorrowBook(service, args[2:])
	case "return":
		return handleReturnBook(service, args[2:])
	default:
		return fmt.Errorf("unknown command: %s", args[1])
	}
}

func handleBookCommand(service *library.Service, args []string) error {
	if len(args) < 1 {
		return nil
	}

	switch args[0] {
	case "add":
		return handleAddBook(service, args[1:])
	case "list":
		return handleListBooks(service)
	default:
		return fmt.Errorf("unknown book command: %s", args[0])
	}
}

func handleLoanCommand(service *library.Service, args []string) error {
	if len(args) < 1 {
		return nil
	}

	switch args[0] {
	case "list":
		return handleLoanBook(service)
	default:
		return fmt.Errorf("unknown loan command: %s", args[0])
	}
}

func handleAddBook(service *library.Service, args []string) error {
	fmt.Println("Adding a new book...")
	if len(args) < 2 {
		fmt.Println("Usage: add <title> <author>")
		fmt.Println("Example: add 'The Great Gatsby' 'F. Scott Fitzgerald'")
		return nil
	}

	title := args[0]
	author := args[1]
	fmt.Printf("Adding book: %s by %s\n", title, author)

	err := service.AddBook(title, author)
	if err != nil {
		fmt.Printf("Error adding book: %v\n", err)
		return err
	}

	fmt.Println("Book added successfully.")
	return nil
}

func handleListBooks(service *library.Service) error {
	books, err := service.ListBooks()
	if err != nil {
		return err
	}

	fmt.Println("List of books:")
	for _, book := range books {
		fmt.Printf("ID: %d, Title: %s, Author: %s, Available: %t\n", book.ID, book.Title, book.Author, book.Available)
	}

	return nil
}

func handleLoanBook(service *library.Service) error {
	// loanList
	loans, err := service.ListActiveLoans()
	if err != nil {
		return err
	}

	for _, loan := range loans {
		fmt.Printf("Book ID: %d, Borrower: %s\n", loan.BookID, loan.Borrower)
	}

	return nil
}

func handleReturnBook(service *library.Service, args []string) error {
	if len(args) < 1 {
		return nil
	}

	bookID := args[0]
	fmt.Printf("Returning book ID: %s\n", bookID)

	bookIDInt, err := strconv.Atoi(bookID)
	if err != nil {
		return fmt.Errorf("invalid book ID: %v", err)
	}

	err = service.ReturnBook(bookIDInt)
	if err != nil {
		return err
	}

	fmt.Println("Book returned successfully.")
	return nil
}

func handleBorrowBook(service *library.Service, args []string) error {
	if len(args) < 2 {
		return nil
	}

	bookID := args[0]
	borrower := args[1]

	bookIDInt, err := strconv.Atoi(bookID)
	if err != nil {
		return fmt.Errorf("invalid book ID: %v", err)
	}

	err = service.BorrowBook(bookIDInt, borrower)
	if err != nil {
		return err
	}

	fmt.Println("Book borrowed successfully.")
	return nil
}
