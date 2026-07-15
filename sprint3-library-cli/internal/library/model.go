package library

import "time"

type Book struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Available bool   `json:"available"`
}

type Loan struct {
	ID         int        `json:"id"`
	BookID     int        `json:"book_id"`
	Borrower   string     `json:"borrower"`
	BorrowedAt time.Time  `json:"borrowed_at"`
	ReturnedAt *time.Time `json:"returned_at,omitempty"`
}
