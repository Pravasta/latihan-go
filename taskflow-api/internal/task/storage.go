package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type Storage interface {
	Load() ([]Task, error)
	Save(tasks []Task) error
}

type storage struct {
	path string
}

// Load implements Storage.
func (s *storage) Load() ([]Task, error) {
	file, err := os.Open(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return []Task{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("open task storage: %w", err)
	}
	defer file.Close()

	var tasks []Task
	if err := json.NewDecoder(file).Decode(&tasks); err != nil {
		if errors.Is(err, io.EOF) {
			return []Task{}, nil
		}
		return nil, fmt.Errorf("decode task storage: %w", err)
	}
	return tasks, nil
}

// Save implements Storage.
func (s *storage) Save(tasks []Task) error {
	taskData, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal tasks: %w", err)
	}

	if err := os.WriteFile(s.path, taskData, 0644); err != nil {
		return fmt.Errorf("write task storage: %w", err)
	}

	return nil
}

func NewStorage(path string) Storage {
	return &storage{path: path}
}
