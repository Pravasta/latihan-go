package task

import (
	"encoding/json"
	"fmt"
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
	if err != nil {
		fmt.Printf("[Storage] Failed to open file: %v\n", err)
		return nil, err
	}
	defer file.Close()

	var tasks []Task
	if err := json.NewDecoder(file).Decode(&tasks); err != nil {
		fmt.Printf("[Storage] Failed to decode tasks: %v\n", err)
		return nil, err
	}
	return tasks, nil
}

// Save implements Storage.
func (s *storage) Save(tasks []Task) error {
	taskData, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		fmt.Printf("[Storage] Failed to marshal tasks: %v\n", err)
		return err
	}

	err = os.WriteFile(s.path, taskData, 0644)
	if err != nil {
		fmt.Printf("[Storage] Failed to write to file: %v\n", err)
		return err
	}

	return nil
}

func NewStorage(path string) Storage {
	return &storage{path: path}
}
