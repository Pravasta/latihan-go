package expense

import "testing"

// Driven Test
/*
   - TestAddExpense
   - TestInvvalidAmount
   - TestInvalidCategory
   - TestSummary
   - TestFilterCategory
*/

type fakeStorage struct {
	expenses []Expense
}

func (f *fakeStorage) Load() ([]Expense, error) {
	return f.expenses, nil
}

func (f *fakeStorage) Save(expenses []Expense) error {
	f.expenses = expenses
	return nil
}

func TestAddExpense(t *testing.T) {
	storage := &fakeStorage{}
	service := NewExpenseService(storage)

	expense := Expense{
		Amount:   100,
		Category: Food,
		Note:     "Lunch",
	}

	err := service.AddExpense(expense)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(storage.expenses) != 1 {
		t.Fatalf("expected 1 expense, got %d", len(storage.expenses))
	}

	got := storage.expenses[0]
	if got.Amount != expense.Amount || got.Category != expense.Category || got.Note != expense.Note {
		t.Fatalf("expected %v, got %v", expense, got)
	}

}

func TestInvalidAmount(t *testing.T) {
	storage := &fakeStorage{}
	service := NewExpenseService(storage)

	expense := Expense{
		Amount:   -50,
		Category: Food,
		Note:     "Dinner",
	}

	err := service.AddExpense(expense)

	if err == nil || err.Error() != ErrGreaterThanZero {
		t.Fatalf("expected error %v, got %v", ErrGreaterThanZero, err)
	}
}

func TestInvalidCategory(t *testing.T) {
	storage := &fakeStorage{}
	service := NewExpenseService(storage)

	expense := Expense{
		Amount:   50,
		Category: "invalid_category",
		Note:     "Dinner",
	}

	err := service.AddExpense(expense)

	if err == nil || err.Error() != ErrInvalidCategory {
		t.Fatalf("expected error %v, got %v", ErrInvalidCategory, err)
	}
}

func TestSummary(t *testing.T) {
	storage := &fakeStorage{
		expenses: []Expense{
			{Amount: 100, Category: Food, Note: "Lunch"},
			{Amount: 200, Category: Transport, Note: "Taxi"},
			{Amount: 50, Category: Food, Note: "Snack"},
		},
	}
	service := NewExpenseService(storage)

	total, summary, err := service.ExpenseSummary()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if total != 350 {
		t.Fatalf("expected total 350, got %d", total)
	}

	if summary[Food] != 150 {
		t.Fatalf("expected food total 150, got %d", summary[Food])
	}

	if summary[Transport] != 200 {
		t.Fatalf("expected transport total 200, got %d", summary[Transport])
	}
}

func TestFilterCategory(t *testing.T) {
	storage := &fakeStorage{
		expenses: []Expense{
			{Amount: 100, Category: Food, Note: "Lunch"},
			{Amount: 200, Category: Transport, Note: "Taxi"},
			{Amount: 50, Category: Food, Note: "Snack"},
		},
	}
	service := NewExpenseService(storage)

	category := Food
	expenses, err := service.GetExpenses(&category, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(expenses) != 2 {
		t.Fatalf("expected 2 expenses, got %d", len(expenses))
	}

	for _, exp := range expenses {
		if exp.Category != Food {
			t.Fatalf("expected category %v, got %v", Food, exp.Category)
		}
	}
}
