package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type Storage interface {
	Load() ([]User, error)
	Save(users []User) error
}

type storage struct {
	path string
}

// Load implements Storage.
func (s *storage) Load() ([]User, error) {
	file, err := os.Open(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return []User{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open auth storage: %w", err)
	}
	defer file.Close()

	var users []User
	if err := json.NewDecoder(file).Decode(&users); err != nil {
		if errors.Is(err, io.EOF) {
			return []User{}, nil
		}
		return nil, fmt.Errorf("decode auth storage: %w", err)
	}
	return users, nil
}

// Save implements Storage.
func (s *storage) Save(users []User) error {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal users: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return fmt.Errorf("write auth storage: %w", err)
	}

	return nil
}

func NewStorage(path string) Storage {
	return &storage{path: path}
}
