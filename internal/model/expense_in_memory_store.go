package model

import (
	"context"
	"fmt"
	"slices"

	"github.com/kkstas/tjener/internal/helpers"
)

type ExpenseInMemoryStore struct {
	expenses []Expense
}

func (e *ExpenseInMemoryStore) Create(ctx context.Context, expenseFC Expense) (Expense, error) {
	e.expenses = append(e.expenses, expenseFC)
	return expenseFC, nil
}

func (e *ExpenseInMemoryStore) Delete(ctx context.Context, SK string) error {
	var deleted bool

	e.expenses = slices.DeleteFunc(e.expenses, func(expense Expense) bool {
		deleted = true
		return expense.SK == SK
	})

	if !deleted {
		return &ExpenseNotFoundError{SK: SK}
	}

	return nil
}

func (e *ExpenseInMemoryStore) Update(ctx context.Context, expenseFU Expense) error {
	var found bool

	for i, el := range e.expenses {
		if el.SK == expenseFU.SK {
			found = true
			e.expenses[i] = expenseFU
		}
	}

	if !found {
		return &ExpenseNotFoundError{SK: expenseFU.SK}
	}
	return nil
}

func (e *ExpenseInMemoryStore) FindOne(ctx context.Context, SK string) (Expense, error) {
	for _, el := range e.expenses {
		if el.SK == SK {
			return el, nil
		}
	}
	return Expense{}, &ExpenseNotFoundError{SK: SK}
}

func (e *ExpenseInMemoryStore) Query(ctx context.Context) ([]Expense, error) {
	return e.expenses, nil
}

// Retrieves expenses between the given `from` and `to` YYYY-MM-DD dates (inclusive).
func (e *ExpenseInMemoryStore) QueryByDateRange(ctx context.Context, from, to string) ([]Expense, error) {
	daysDiff, err := helpers.DaysBetween(from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get number of days between 'from' and 'to' date: %w", err)
	}
	if daysDiff < minQueryRangeDaysDiff || daysDiff > maxQueryRangeDaysDiff {
		return nil, fmt.Errorf("invalid difference between 'from' and 'to' date; got=%d, max=%d, min=%d", daysDiff, minQueryRangeDaysDiff, maxQueryRangeDaysDiff)
	}

	var expenses []Expense

	for _, expense := range e.expenses {
		daysAfterFrom, err := helpers.DaysBetween(from, expense.Date)
		if err != nil {
			return nil, fmt.Errorf("failed to get number of days between 'from' and 'expense.Date' for expense: %+v: %w", expense, err)
		}
		daysBeforeTo, err := helpers.DaysBetween(expense.Date, to)
		if err != nil {
			return nil, fmt.Errorf("failed to get number of days between 'expense.Date' and 'to for expense: %+v: %w", expense, err)
		}
		if daysAfterFrom >= 0 && daysBeforeTo >= 0 {
			expenses = append(expenses, expense)
		}
	}

	return expenses, nil
}
