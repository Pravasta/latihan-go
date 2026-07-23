# Sprint 04 — Notes REST API with net/http

## Sprint Information

- Sprint: 04
- Title: Notes REST API
- Difficulty: Intermediate
- Estimated Duration: 1–2 Days
- Primary Focus: HTTP Fundamentals

---

# Background

Tim membutuhkan REST API sederhana untuk mengelola personal notes.

Pada sprint sebelumnya aplikasi menerima input melalui CLI. Pada sprint ini, aplikasi harus menerima request melalui HTTP dan mengembalikan response dalam format JSON.

Aplikasi belum menggunakan database.

Data tetap disimpan menggunakan file JSON agar fokus utama sprint tetap pada HTTP fundamentals.

---

# Learning Objectives

Sprint ini berfokus pada:

- `net/http`
- HTTP server
- HTTP methods
- RESTful routing
- Request body
- JSON encoding and decoding
- HTTP status codes
- Path parameters
- Query parameters
- Handler layer
- Service layer
- Storage abstraction
- HTTP testing menggunakan `httptest`

---

# Project Setup

Buat project baru:

```bash
mkdir notes-api
cd notes-api

go mod init github.com/<username>/notes-api

mkdir -p cmd/api
mkdir -p internal/note
mkdir -p data
mkdir -p docs

echo "[]" > data/notes.json
```

---

# Required Project Structure

```text
notes-api/
├── cmd/
│   └── api/
│       └── main.go
│
├── internal/
│   └── note/
│       ├── model.go
│       ├── service.go
│       ├── storage.go
│       ├── handler.go
│       ├── errors.go
│       ├── service_test.go
│       ├── storage_test.go
│       └── handler_test.go
│
├── data/
│   └── notes.json
│
├── docs/
│   ├── sprint-04-task.md
│   └── sprint-04-review.md
│
└── go.mod
```

---

# Restrictions

Sprint ini hanya boleh menggunakan Go standard library.

Dilarang menggunakan:

- Gin
- Echo
- Fiber
- Chi
- Gorilla Mux
- GORM
- Database

Gunakan:

```go
net/http
encoding/json
```

---

# Domain Model

```go
type Note struct {
    ID        int       `json:"id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

---

# API Endpoints

## API-001 — Create Note

```http
POST /notes
```

Request body:

```json
{
  "title": "Belajar Go",
  "content": "Belajar net/http"
}
```

Expected response:

```http
201 Created
```

```json
{
  "data": {
    "id": 1,
    "title": "Belajar Go",
    "content": "Belajar net/http",
    "created_at": "...",
    "updated_at": "..."
  }
}
```

---

## API-002 — List Notes

```http
GET /notes
```

Expected:

```http
200 OK
```

```json
{
  "data": [
    {
      "id": 1,
      "title": "Belajar Go",
      "content": "Belajar net/http",
      "created_at": "...",
      "updated_at": "..."
    }
  ]
}
```

---

## API-003 — Get Note Detail

```http
GET /notes/{id}
```

Example:

```http
GET /notes/1
```

Expected:

```http
200 OK
```

```json
{
  "data": {
    "id": 1,
    "title": "Belajar Go",
    "content": "Belajar net/http"
  }
}
```

---

## API-004 — Update Note

```http
PUT /notes/{id}
```

Request:

```json
{
  "title": "Belajar Golang",
  "content": "Belajar standard library net/http"
}
```

Expected:

```http
200 OK
```

`CreatedAt` tidak boleh berubah.

`UpdatedAt` harus diperbarui.

---

## API-005 — Delete Note

```http
DELETE /notes/{id}
```

Expected:

```http
204 No Content
```

Response body harus kosong.

---

# Bonus Endpoint — Search Notes

```http
GET /notes?search=golang
```

Search harus:

- Case-insensitive
- Mencari pada title
- Mencari pada content

Jika `search` kosong atau tidak diberikan, tampilkan semua notes.

---

# Validation Rules

## VR-001 — Empty Title

Response:

```http
400 Bad Request
```

```json
{
  "error": "note title cannot be empty"
}
```

---

## VR-002 — Empty Content

Response:

```http
400 Bad Request
```

```json
{
  "error": "note content cannot be empty"
}
```

---

## VR-003 — Invalid JSON

Request:

```json
{
  "title":
}
```

Response:

```http
400 Bad Request
```

```json
{
  "error": "invalid request body"
}
```

---

## VR-004 — Invalid ID

Example:

```http
GET /notes/abc
```

Response:

```http
400 Bad Request
```

```json
{
  "error": "invalid note id"
}
```

---

## VR-005 — Note Not Found

Example:

```http
GET /notes/999
```

Response:

```http
404 Not Found
```

```json
{
  "error": "note not found"
}
```

---

## VR-006 — Method Not Allowed

Example:

```http
PATCH /notes/1
```

Response:

```http
405 Method Not Allowed
```

---

# Architecture Requirements

Flow:

```text
HTTP Request
    ↓
Handler
    ↓
Service
    ↓
Storage Interface
    ↓
JSON File
```

Responsibilities:

## Handler

Handler bertanggung jawab untuk:

- Membaca HTTP request
- Membaca path parameter
- Membaca query parameter
- Decode JSON
- Memanggil service
- Mapping domain error ke HTTP status code
- Encode JSON response

Handler tidak boleh:

- Generate Note ID
- Mengubah `CreatedAt`
- Mengubah `UpdatedAt`
- Melakukan business logic

---

## Service

Service bertanggung jawab untuk:

- Validation
- Generate ID
- Create note
- Update note
- Delete note
- Search note
- Business rules

---

## Storage

Storage hanya bertanggung jawab untuk:

- Load notes
- Save notes

Interface:

```go
type Storage interface {
    Load() ([]Note, error)
    Save([]Note) error
}
```

---

# Required Service API

Nama method boleh berbeda, tetapi behavior harus setara dengan:

```go
type Service struct {
    storage Storage
}

func NewService(storage Storage) *Service

func (s *Service) Create(title, content string) (Note, error)

func (s *Service) List(search string) ([]Note, error)

func (s *Service) GetByID(id int) (Note, error)

func (s *Service) Update(id int, title, content string) (Note, error)

func (s *Service) Delete(id int) error
```

---

# Routing Requirements

Gunakan `http.ServeMux`.

Routing boleh menggunakan kemampuan routing standard library dari versi Go yang digunakan project.

Contoh konsep:

```go
mux := http.NewServeMux()
```

Handler harus menerima dependency melalui constructor.

Contoh:

```go
type Handler struct {
    service *Service
}

func NewHandler(service *Service) *Handler
```

Dilarang membuat storage langsung di handler.

---

# Response Format

Success response:

```json
{
  "data": {}
}
```

Error response:

```json
{
  "error": "error message"
}
```

Gunakan helper jika diperlukan agar response konsisten.

---

# HTTP Status Code Requirements

| Scenario           | Status |
| ------------------ | ------ |
| Create success     | 201    |
| Read success       | 200    |
| Update success     | 200    |
| Delete success     | 204    |
| Validation error   | 400    |
| Invalid JSON       | 400    |
| Invalid ID         | 400    |
| Not found          | 404    |
| Method not allowed | 405    |
| Internal error     | 500    |

---

# Error Handling Requirements

Gunakan sentinel errors untuk domain error.

Contoh:

```go
var (
    ErrTitleEmpty   = errors.New("note title cannot be empty")
    ErrContentEmpty = errors.New("note content cannot be empty")
    ErrNoteNotFound = errors.New("note not found")
)
```

Gunakan:

```go
errors.Is(...)
```

ketika melakukan mapping error.

Internal error tidak boleh membocorkan detail implementasi ke client.

Contoh yang dilarang:

```json
{
  "error": "open data/notes.json: permission denied"
}
```

Client cukup menerima:

```json
{
  "error": "internal server error"
}
```

---

# Testing Requirements

## Service Tests

Minimal:

- `TestCreateNote`
- `TestCreateNoteValidation`
- `TestListNotes`
- `TestSearchNotes`
- `TestGetNoteByID`
- `TestGetNoteNotFound`
- `TestUpdateNote`
- `TestUpdateNoteNotFound`
- `TestDeleteNote`
- `TestDeleteNoteNotFound`

Validation test wajib menggunakan table-driven test.

---

## Storage Tests

Minimal:

- `TestStorageSaveAndLoad`
- `TestStorageInvalidJSON`

Gunakan:

```go
t.TempDir()
```

Dilarang menggunakan file production untuk testing.

---

## Handler Tests

Gunakan:

```go
net/http/httptest
```

Minimal:

- `TestCreateNoteHandler`
- `TestCreateNoteInvalidJSON`
- `TestGetNoteHandler`
- `TestGetNoteInvalidID`
- `TestGetNoteNotFound`
- `TestUpdateNoteHandler`
- `TestDeleteNoteHandler`

Contoh konsep:

```go
req := httptest.NewRequest(
    http.MethodPost,
    "/notes",
    strings.NewReader(`{
        "title": "Belajar Go",
        "content": "Belajar HTTP"
    }`),
)

rec := httptest.NewRecorder()

handler.ServeHTTP(rec, req)
```

Test minimal harus memverifikasi:

- HTTP status code
- Response body
- Content-Type jika response memiliki JSON body

---

# Important Scenarios

## Scenario 1 — Create and Retrieve

Create:

```http
POST /notes
```

Kemudian:

```http
GET /notes/1
```

Note harus dapat ditemukan.

---

## Scenario 2 — Update Timestamp

Create note.

Simpan nilai:

```text
CreatedAt
UpdatedAt
```

Update note.

Expected:

```text
CreatedAt tetap sama
UpdatedAt berubah
```

---

## Scenario 3 — Delete

Delete note.

Kemudian:

```http
GET /notes/{id}
```

harus menghasilkan:

```http
404 Not Found
```

---

# Acceptance Criteria

## API

- [ ] POST `/notes` works
- [ ] GET `/notes` works
- [ ] GET `/notes/{id}` works
- [ ] PUT `/notes/{id}` works
- [ ] DELETE `/notes/{id}` works

## HTTP

- [ ] Correct HTTP methods
- [ ] Correct HTTP status codes
- [ ] JSON request decoding
- [ ] JSON response encoding
- [ ] Invalid JSON handled
- [ ] Invalid ID handled
- [ ] Method not allowed handled

## Architecture

- [ ] Handler layer exists
- [ ] Service layer exists
- [ ] Storage abstraction exists
- [ ] Dependency injection used
- [ ] No business logic in handler
- [ ] No persistence logic in service

## Testing

- [ ] Service tests implemented
- [ ] Storage tests implemented
- [ ] Handler tests implemented with `httptest`
- [ ] All tests pass
- [ ] Minimum coverage: 80%

## Quality

- [ ] No panic
- [ ] No ignored errors
- [ ] `go vet ./...` passes
- [ ] Server shuts down cleanly when possible
- [ ] Internal errors are not exposed to clients

---

# Definition of Done

Sprint selesai jika:

```bash
go test ./...
go test ./... -cover
go vet ./...
```

berhasil dan acceptance criteria terpenuhi.

Manual API testing juga harus dilakukan menggunakan salah satu:

- curl
- Postman
- Bruno
- Insomnia

---

# Commands

Run server:

```bash
go run ./cmd/api
```

Example request:

```bash
curl -X POST http://localhost:8080/notes \
  -H "Content-Type: application/json" \
  -d '{"title":"Belajar Go","content":"Belajar net/http"}'
```

List notes:

```bash
curl http://localhost:8080/notes
```
