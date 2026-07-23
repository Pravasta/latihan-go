package note

import (
	"encoding/json"
	"fmt"
	"os"
)

type Storage interface {
	Load() ([]Note, error)
	Save([]Note) error
}

type storage struct {
	path string
}

// Load implements [Storage].
func (s *storage) Load() ([]Note, error) {
	file, err := os.Open(s.path)
	if err != nil {
		fmt.Printf("[Storage Load] Failed to open file: %v\n", err)
		return nil, err
	}
	defer file.Close()

	var notes []Note
	if err := json.NewDecoder(file).Decode(&notes); err != nil {
		fmt.Printf("[Storage Load] Failed to decode notes: %v\n", err)
		return nil, err
	}
	return notes, nil
}

// Save implements [Storage].
func (s *storage) Save(notes []Note) error {
	data, err := json.MarshalIndent(notes, "", "  ")
	if err != nil {
		fmt.Printf("[Storage Save] Failed to marshal notes: %v\n", err)
		return err
	}

	err = os.WriteFile(s.path, data, 0644)
	if err != nil {
		fmt.Printf("[Storage Save] Failed to write to file: %v\n", err)
		return err
	}

	return nil
}

func NewStorage(path string) Storage {
	return &storage{path: path}
}
