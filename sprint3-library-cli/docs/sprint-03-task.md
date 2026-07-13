# Sprint 03 — Library Management CLI

## Sprint Information

- Sprint: 03
- Title: Library Management CLI
- Difficulty: Beginner → Intermediate
- Estimated Duration: 1–2 Days

---

# Background

Sebuah perpustakaan kecil membutuhkan aplikasi CLI untuk mengelola buku dan proses peminjaman.

Aplikasi masih menggunakan JSON sebagai persistence dan belum menggunakan database.

Berbeda dengan sprint sebelumnya, aplikasi memiliki lebih dari satu domain entity dan business rule yang saling berhubungan.

---

# Learning Objectives

Sprint ini berfokus pada:

- Multiple entities
- Relationship antar data
- Business rules
- Domain validation
- Multiple storage files
- Error propagation
- Table-driven tests
- Separation of concerns

---

# Project Setup

Buat project baru:

```bash
mkdir library-cli
cd library-cli

go mod init github.com/<username>/library-cli
```

Buat file data:

```bash
mkdir -p data docs
echo "[]" > data/books.json
echo "[]" > data/loans.json
```

---

# Required Project Structure

```text
library-cli/
├── cmd/
│   └── root.go
│
├── internal/
│   └── library/
│       ├── model.go
│       ├── service.go
│       ├── storage.go
│       ├── errors.go
│       └── service_test.go
│
├── data/
│   ├── books.json
│   └── loans.json
│
├── docs/
│   ├── sprint-03-task.md
│   └── sprint-03-review.md
│
├── main.go
└── go.mod
```

---

# Domain Models

## Book

```go
type Book struct {
    ID        int    `json:"id"`
    Title     string `json:"title"`
    Author    string `json:"author"`
    Available bool   `json:"available"`
}
```

## Loan

```go
type Loan struct {
    ID         int        `json:"id"`
    BookID     int        `json:"book_id"`
    Borrower   string     `json:"borrower"`
    BorrowedAt time.Time  `json:"borrowed_at"`
    ReturnedAt *time.Time `json:"returned_at,omitempty"`
}
```

`ReturnedAt == nil` berarti buku masih dipinjam.

---

# Functional Requirements

## FR-001 — Add Book

Command:

```bash
go run . book add "Clean Code" "Robert C. Martin"
```

Expected output:

```text
Book successfully added
```

Buku baru harus memiliki:

```text
Available = true
```

---

## FR-002 — List Books

Command:

```bash
go run . book list
```

Example output:

```text
1 | Clean Code | Robert C. Martin | AVAILABLE
2 | The Pragmatic Programmer | Andrew Hunt | BORROWED
```

---

## FR-003 — Borrow Book

Command:

```bash
go run . borrow 1 "Pravasta"
```

Expected output:

```text
Book successfully borrowed
```

Ketika buku dipinjam:

1. Pastikan buku dengan ID tersebut tersedia.
2. Buat Loan baru.
3. Set `Book.Available = false`.
4. Simpan perubahan books.
5. Simpan loan baru.

---

## FR-004 — Return Book

Command:

```bash
go run . return 1
```

Parameter `1` adalah Loan ID.

Expected output:

```text
Book successfully returned
```

Ketika buku dikembalikan:

1. Cari Loan berdasarkan Loan ID.
2. Pastikan loan belum pernah dikembalikan.
3. Isi `ReturnedAt` dengan waktu saat ini.
4. Set buku terkait menjadi `Available = true`.
5. Simpan perubahan.

---

## FR-005 — List Active Loans

Command:

```bash
go run . loan list
```

Hanya tampilkan loan yang:

```text
ReturnedAt == nil
```

Example:

```text
Loan #1 | Book ID: 2 | Borrower: Pravasta
```

---

# Validation Rules

## VR-001 — Empty Book Title

Title tidak boleh kosong.

Expected error:

```text
book title cannot be empty
```

---

## VR-002 — Empty Author

Author tidak boleh kosong.

Expected error:

```text
book author cannot be empty
```

---

## VR-003 — Empty Borrower

Borrower tidak boleh kosong.

Expected error:

```text
borrower cannot be empty
```

---

## VR-004 — Book Not Found

Jika Book ID tidak ditemukan:

```text
book not found
```

---

## VR-005 — Book Already Borrowed

Buku yang:

```text
Available == false
```

tidak boleh dipinjam kembali.

Expected:

```text
book is not available
```

---

## VR-006 — Loan Not Found

Jika Loan ID tidak ditemukan:

```text
loan not found
```

---

## VR-007 — Loan Already Returned

Loan dengan:

```text
ReturnedAt != nil
```

tidak boleh dikembalikan dua kali.

Expected:

```text
loan already returned
```

---

# Storage Requirements

Gunakan dua storage abstraction.

```go
type BookStorage interface {
    Load() ([]Book, error)
    Save([]Book) error
}
```

```go
type LoanStorage interface {
    Load() ([]Loan, error)
    Save([]Loan) error
}
```

Implementasi JSON storage harus menyimpan data ke:

```text
data/books.json
data/loans.json
```

Service menerima kedua dependency melalui constructor.

Contoh konsep:

```go
type Service struct {
    bookStorage BookStorage
    loanStorage LoanStorage
}
```

Dilarang membuat concrete storage langsung di dalam service.

---

# Business Rules

## BR-001

Satu buku hanya dapat memiliki satu active loan.

---

## BR-002

Book hanya dianggap tersedia jika:

```text
Available == true
```

---

## BR-003

Return menggunakan Loan ID, bukan Book ID.

---

## BR-004

Loan history tidak boleh dihapus ketika buku dikembalikan.

Yang berubah hanya:

```text
ReturnedAt
```

---

## BR-005

ID tidak boleh menggunakan:

```go
len(items) + 1
```

Cari ID terbesar, lalu tambahkan satu.

---

# Technical Requirements

- Gunakan standard library Go.
- Tidak menggunakan database.
- Tidak menggunakan Cobra.
- Tidak menggunakan global mutable variable.
- Tidak menggunakan `panic`.
- Business logic harus berada di service.
- Storage hanya bertanggung jawab terhadap persistence.
- Error harus dikembalikan kepada caller.
- Gunakan custom sentinel errors jika sesuai.

---

# Testing Requirements

Minimal wajib memiliki:

- `TestAddBook`
- `TestAddBookValidation`
- `TestBorrowBook`
- `TestBorrowUnavailableBook`
- `TestBorrowBookNotFound`
- `TestReturnBook`
- `TestReturnLoanNotFound`
- `TestReturnAlreadyReturnedLoan`
- `TestListActiveLoans`

Validation test wajib menggunakan table-driven test.

---

# Important Test Scenario

Initial state:

```text
Book:
ID        = 1
Title     = Clean Code
Available = true
```

Action:

```text
Borrow Book ID 1
```

Expected state:

```text
Book.Available = false

Loan:
BookID     = 1
Borrower   = Pravasta
ReturnedAt = nil
```

Kemudian:

```text
Return Loan ID 1
```

Expected final state:

```text
Book.Available = true
Loan.ReturnedAt != nil
```

---

# Acceptance Criteria

## Features

- [ ] Add book works
- [ ] List books works
- [ ] Borrow book works
- [ ] Return book works
- [ ] List active loans works

## Business Rules

- [ ] Borrowed book cannot be borrowed again
- [ ] Returning a book updates both Book and Loan
- [ ] Returned loan cannot be returned twice
- [ ] Loan history remains after return
- [ ] IDs remain unique after deletion or data gaps

## Architecture

- [ ] `BookStorage` interface exists
- [ ] `LoanStorage` interface exists
- [ ] Both storages use dependency injection
- [ ] Business logic is located in service layer
- [ ] JSON storage only handles persistence
- [ ] CLI layer only handles input/output and command routing

## Testing

- [ ] Minimum required tests implemented
- [ ] Table-driven validation tests implemented
- [ ] All tests pass
- [ ] Minimum test coverage: 80%

## Quality

- [ ] No panic
- [ ] No ignored errors
- [ ] Idiomatic Go naming
- [ ] No unnecessary code duplication
- [ ] `go vet ./...` passes

---

# Commands Summary

```bash
go run . book add "Clean Code" "Robert C. Martin"

go run . book list

go run . borrow 1 "Pravasta"

go run . loan list

go run . return 1
```

---

# Definition of Done

Sprint dinyatakan selesai jika:

```bash
go test ./...
go test ./... -cover
go vet ./...
```

berhasil dan seluruh Acceptance Criteria terpenuhi.
