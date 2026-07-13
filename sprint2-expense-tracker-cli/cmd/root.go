package cmd

import (
	"errors"
	"fmt"
	"os"
	"sprint2-expense-tracker-cli/internal/expense"
	"strconv"
)

func Execute() error {
	storage := expense.NewStorage("data/expenses.json")
	service := expense.NewExpenseService(storage)

	args := os.Args

	if len(args) < 2 {
		return nil
	}

	switch args[1] {
	case "add":
		return handleAddExpense(service, args[2:])
	case "list":
		return handleListExpenses(service)
	case "summary":
		return handleSummaryExpenses(service)
	default:
		return fmt.Errorf("unknown command: %s", args[1])
	}
}

func handleAddExpense(service expense.ExpenseService, args []string) error {
	if len(args) < 3 {
		return errors.New("usage: add <amount> <category> <note>")
	}

	amount, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}

	category := expense.Category(args[1])
	note := args[2]

	exp := expense.Expense{
		Amount:   amount,
		Category: category,
		Note:     note,
	}

	if err := service.AddExpense(exp); err != nil {
		return err
	}

	return nil
}

func handleListExpenses(service expense.ExpenseService) error {
	// Parse --category=food
	var category *expense.Category
	var sortOrder *string

	for _, arg := range os.Args[2:] {
		if len(arg) > 10 && arg[:10] == "--category" {
			cat := expense.Category(arg[11:])
			category = &cat
		} else if len(arg) > 6 && arg[:6] == "--sort" {
			sort := arg[7:]
			sortOrder = &sort
		}
	}

	expenses, err := service.GetExpenses(category, sortOrder)
	if err != nil {
		return err
	}

	for _, exp := range expenses {
		fmt.Printf("%d | %d | %s | %s\n", exp.ID, exp.Amount, exp.Category, exp.Note)
	}

	return nil
}

func handleSummaryExpenses(service expense.ExpenseService) error {
	total, summary, err := service.ExpenseSummary()

	if err != nil {
		return err
	}

	fmt.Printf("Total: %d\n\n", total)

	// Print total summary
	for category, amount := range summary {
		fmt.Printf("%s: %d\n", category, amount)
		total += amount
	}

	return nil
}
