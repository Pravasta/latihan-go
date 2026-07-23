package note

import (
	"fmt"
	"strings"
	"time"
)

type Service struct {
	storage Storage
}

func NewService(storage Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) Create(title, content string) (Note, error) {
	var note Note

	if title == "" {
		fmt.Println("[Service Create] Title cannot be empty")
		return note, ErrTitleEmpty
	}

	if content == "" {
		fmt.Println("[Service Create] Content cannot be empty")
		return note, ErrContentEmpty
	}

	notes, err := s.storage.Load()
	if err != nil {
		fmt.Printf("[Service Create] Failed to load notes: %v\n", err)
		return note, err
	}

	maxID := 0
	for _, n := range notes {
		if n.ID > maxID {
			maxID = n.ID
		}
	}

	now := time.Now()

	note = Note{
		ID:        maxID + 1,
		Title:     title,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	notes = append(notes, note)

	err = s.storage.Save(notes)
	if err != nil {
		fmt.Printf("[Service Create] Failed to save notes: %v\n", err)
		return note, err
	}

	return note, nil
}

func (s *Service) List(search string) ([]Note, error) {
	notes, err := s.storage.Load()
	if err != nil {
		fmt.Printf("[Service List] Failed to load notes: %v\n", err)
		return nil, err
	}

	if search == "" {
		return notes, nil
	}

	var filtered []Note
	search = strings.ToLower(search)

	for _, note := range notes {
		if strings.Contains(
			strings.ToLower(note.Title),
			search,
		) || strings.Contains(
			strings.ToLower(note.Content),
			search,
		) {
			filtered = append(filtered, note)
		}
	}

	if len(filtered) == 0 {
		fmt.Printf("[Service List] No notes found matching search: %s\n", search)
		return nil, ErrNoteNotFound
	}

	return filtered, nil
}

func (s *Service) Get(id int) (Note, error) {
	var note Note

	notes, err := s.storage.Load()
	if err != nil {
		fmt.Printf("[Service Get] Failed to load notes: %v\n", err)
		return note, err
	}

	for _, n := range notes {
		if n.ID == id {
			return n, nil
		}
	}

	fmt.Printf("[Service Get] Note with ID %d not found\n", id)
	return note, ErrNoteNotFound
}

func (s *Service) Update(id int, title, content string) (Note, error) {
	var updatedNote Note

	notes, err := s.storage.Load()
	if err != nil {
		fmt.Printf("[Service Update] Failed to load notes: %v\n", err)
		return updatedNote, err
	}

	for i, n := range notes {
		if n.ID == id {
			if title != "" {
				n.Title = title
			}
			if content != "" {
				n.Content = content
			}
			n.UpdatedAt = time.Now()
			notes[i] = n
			updatedNote = n

			err = s.storage.Save(notes)
			if err != nil {
				fmt.Printf("[Service Update] Failed to save notes: %v\n", err)
				return updatedNote, err
			}

			return updatedNote, nil
		}
	}

	fmt.Printf("[Service Update] Note with ID %d not found\n", id)
	return updatedNote, ErrNoteNotFound
}

func (s *Service) Delete(id int) error {
	notes, err := s.storage.Load()
	if err != nil {
		fmt.Printf("[Service Delete] Failed to load notes: %v\n", err)
		return err
	}

	for i, n := range notes {
		if n.ID == id {
			notes = append(notes[:i], notes[i+1:]...)

			err = s.storage.Save(notes)
			if err != nil {
				fmt.Printf("[Service Delete] Failed to save notes: %v\n", err)
				return err
			}

			return nil
		}
	}

	fmt.Printf("[Service Delete] Note with ID %d not found\n", id)
	return ErrNoteNotFound
}
