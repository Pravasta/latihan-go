package project

import (
	"encoding/json"
	"fmt"
	"os"
)

type Storage interface {
	Load() ([]Project, error)
	Save(projects []Project) error
}

type storage struct {
	path string
}

// Load implements Storage.
func (s *storage) Load() ([]Project, error) {
	file, err := os.Open(s.path)
	if err != nil {
		fmt.Printf("[Storage] Failed to open file: %v\n", err)
		return nil, err
	}
	defer file.Close()

	var projects []Project
	if err := json.NewDecoder(file).Decode(&projects); err != nil {
		fmt.Printf("[Storage] Failed to decode projects: %v\n", err)
		return nil, err
	}
	return projects, nil
}

// Save implements Storage.
func (s *storage) Save(projects []Project) error {
	projectData, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		fmt.Printf("[Storage] Failed to marshal projects: %v\n", err)
		return err
	}

	err = os.WriteFile(s.path, projectData, 0644)
	if err != nil {
		fmt.Printf("[Storage] Failed to write to file: %v\n", err)
		return err
	}

	return nil
}

func NewStorage(path string) Storage {
	return &storage{path: path}
}
