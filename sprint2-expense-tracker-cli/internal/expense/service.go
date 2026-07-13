package expense

import "errors"

type ExpenseService interface {
	AddExpense(expense Expense) error
	GetExpenses(
		category *Category,
		sortOrder *string,
	) ([]Expense, error)
	ExpenseSummary() (int, map[Category]int, error)
}

type expenseService struct {
	storage Storage
}

// ExpenseSummary implements ExpenseService.
func (e *expenseService) ExpenseSummary() (int, map[Category]int, error) {
	summaryByCategory := make(map[Category]int)

	// Load all expenses from storage
	expenses, err := e.storage.Load()
	if err != nil {
		return 0, nil, errors.New(ErrInternal)
	}

	for _, exp := range expenses {
		summaryByCategory[exp.Category] += exp.Amount
	}

	total := 0
	for _, amount := range summaryByCategory {
		total += amount
	}

	return total, summaryByCategory, nil
}

// AddExpense implements ExpenseService.
func (e *expenseService) AddExpense(expense Expense) error {

	if expense.Amount <= 0 {
		return errors.New(ErrGreaterThanZero)
	}

	if expense.Category.IsValid() == false {
		return errors.New(ErrInvalidCategory)
	}

	if expense.Note == "" {
		return errors.New(ErrEmptyNote)
	}

	err := e.storage.Save([]Expense{expense})
	if err != nil {
		return errors.New(ErrInternal)
	}

	return nil
}

// GetExpenses implements ExpenseService.
func (e *expenseService) GetExpenses(
	category *Category,
	sortOrder *string,
) ([]Expense, error) {
	// Load all expenses from storage
	expenses, err := e.storage.Load()
	if err != nil {
		return nil, errors.New(ErrInternal)
	}

	// Filter by Category if provided
	if category != nil {
		filtered := []Expense{}
		for _, exp := range expenses {
			if exp.Category == *category {
				filtered = append(filtered, exp)
			}
		}
		expenses = filtered
	}

	// Sort by Amount if sortOrder is provided
	if sortOrder != nil {
		// Example usage:
		// go run . list --sort=amount => asc
		// go run . list --sort=-amount => desc
		// using slice

		switch *sortOrder {
		case "amount":
			// Sort ascending
			for i := 0; i < len(expenses)-1; i++ {
				for j := 0; j < len(expenses)-i-1; j++ {
					if expenses[j].Amount > expenses[j+1].Amount {
						expenses[j], expenses[j+1] = expenses[j+1], expenses[j]
					}
				}
			}
		case "-amount":
			// Sort descending
			for i := 0; i < len(expenses)-1; i++ {
				for j := 0; j < len(expenses)-i-1; j++ {
					if expenses[j].Amount < expenses[j+1].Amount {
						expenses[j], expenses[j+1] = expenses[j+1], expenses[j]
					}
				}
			}
		}
	}

	return expenses, nil
}

func NewExpenseService(storage Storage) ExpenseService {
	return &expenseService{storage: storage}
}
