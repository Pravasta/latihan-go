package expense

import (
	"encoding/json"
	"fmt"
	"os"
)

type Storage interface {
	Load() ([]Expense, error)
	Save(expenses []Expense) error
}

type storage struct {
	filePath string
}

// Load implements Storage.
func (s *storage) Load() ([]Expense, error) {
	// Open the file for reading
	file, err := os.Open(s.filePath)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		return nil, err
	}
	defer file.Close()

	var expenses []Expense
	if err := json.NewDecoder(file).Decode(&expenses); err != nil {
		fmt.Printf("Failed to decode expenses: %v\n", err)
		return nil, err
	}

	return expenses, nil
}

// Save implements Storage.
func (s *storage) Save(expenses []Expense) error {
	// Fetch the existing expenses from the file
	file, err := os.Open(s.filePath)

	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		return err
	}

	var existingExpenses []Expense
	if err := json.NewDecoder(file).Decode(&existingExpenses); err != nil {
		fmt.Printf("Failed to decode existing expenses: %v\n", err)
		return err
	}

	file.Close()

	// update ID
	for i := range expenses {
		expenses[i].ID = len(existingExpenses) + i + 1
	}

	// Append the new expenses to the existing expenses
	existingExpenses = append(existingExpenses, expenses...)

	// Write the updated expenses back to the file
	file, err = os.Create(s.filePath)
	if err != nil {
		fmt.Printf("Failed to create file: %v\n", err)
		return err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(existingExpenses); err != nil {
		fmt.Printf("Failed to encode expenses: %v\n", err)
		return err
	}

	return nil
}

func NewStorage(filePath string) Storage {
	return &storage{filePath: filePath}
}
