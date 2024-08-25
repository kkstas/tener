package model

import (
	"context"
	"slices"
)

type ExpenseInMemoryStore struct {
	expenses []Expense
}

func (e *ExpenseInMemoryStore) Create(ctx context.Context, expenseFC Expense) (Expense, error) {
	e.expenses = append(e.expenses, expenseFC)
	return expenseFC, nil
}

func (e *ExpenseInMemoryStore) Delete(ctx context.Context, createdAt string) error {
	var deleted bool

	e.expenses = slices.DeleteFunc(e.expenses, func(expense Expense) bool {
		deleted = true
		return expense.CreatedAt == createdAt
	})

	if !deleted {
		return &ExpenseNotFoundError{CreatedAt: createdAt}
	}

	return nil
}

func (e *ExpenseInMemoryStore) Update(ctx context.Context, expenseFU Expense) (Expense, error) {
	var found bool

	for i, el := range e.expenses {
		if el.CreatedAt == expenseFU.CreatedAt {
			found = true
			e.expenses[i] = expenseFU
		}
	}

	if !found {
		return Expense{}, &ExpenseNotFoundError{CreatedAt: expenseFU.CreatedAt}
	}
	return expenseFU, nil
}

func (e *ExpenseInMemoryStore) FindOne(ctx context.Context, createdAt string) (Expense, error) {
	for _, el := range e.expenses {
		if el.CreatedAt == createdAt {
			return el, nil
		}
	}
	return Expense{}, &ExpenseNotFoundError{CreatedAt: createdAt}
}

func (e *ExpenseInMemoryStore) Query(ctx context.Context) ([]Expense, error) {
	return e.expenses, nil
}
