package main

import (
	"fmt"
	"net/http"
	"sprint4-notes-api/internal/note"
)

func main() {
	storage := note.NewStorage("data/notes.json")
	service := note.NewService(storage)

	handler := note.NewHandler(service)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /notes", handler.CreateNote)
	mux.HandleFunc("GET /notes", handler.ListNotes)
	mux.HandleFunc("GET /notes/{id}", handler.GetNoteByID)
	mux.HandleFunc("PUT /notes/{id}", handler.UpdateNote)
	mux.HandleFunc("DELETE /notes/{id}", handler.DeleteNote)

	fmt.Println("[Main] Server is Running on http://localhost:8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
