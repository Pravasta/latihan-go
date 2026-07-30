package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	if errors.Is(err, os.ErrNotExist) {
		return []Project{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open project storage: %w", err)
	}
	defer file.Close()

	var projects []Project
	if err := json.NewDecoder(file).Decode(&projects); err != nil {
		if errors.Is(err, io.EOF) {
			return []Project{}, nil
		}
		return nil, fmt.Errorf("decode project storage: %w", err)
	}
	return projects, nil
}

// Save implements Storage.
func (s *storage) Save(projects []Project) error {
	projectData, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal projects: %w", err)
	}

	if err := os.WriteFile(s.path, projectData, 0644); err != nil {
		return fmt.Errorf("write project storage: %w", err)
	}

	return nil
}

func NewStorage(path string) Storage {
	return &storage{path: path}
}
