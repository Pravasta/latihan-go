package auth

import (
	"encoding/json"
	"fmt"
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
	if err != nil {
		fmt.Printf(("[Storage] Failed to open file: %v\n"), err)
		return nil, err
	}

	defer file.Close()

	var users []User
	if err := json.NewDecoder(file).Decode(&users); err != nil {
		fmt.Printf(("[Storage] Failed to decode users: %v\n"), err)
		return nil, err
	}
	return users, nil
}

// Save implements Storage.
func (s *storage) Save(users []User) error {
	user, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		fmt.Printf(("[Storage] Failed to marshal users: %v\n"), err)
		return err
	}

	err = os.WriteFile(s.path, user, 0644)
	if err != nil {
		fmt.Printf(("[Storage] Failed to write to file: %v\n"), err)
		return err
	}

	return nil
}

func NewStorage(path string) Storage {
	return &storage{path: path}
}
