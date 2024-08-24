package model

import (
	"context"
	"slices"
)

type ExpenseCategoryInMemoryStore struct {
	categories []ExpenseCategory
}

func (e *ExpenseCategoryInMemoryStore) Create(ctx context.Context, categoryFC ExpenseCategory) error {
	e.categories = append(e.categories, categoryFC)
	return nil
}

func (e *ExpenseCategoryInMemoryStore) Delete(ctx context.Context, createdAt string) error {
	var deleted bool

	e.categories = slices.DeleteFunc(e.categories, func(category ExpenseCategory) bool {
		deleted = true
		return category.Name == createdAt
	})

	if !deleted {
		return &ExpenseCategoryNotFoundError{CreatedAt: createdAt}
	}

	return nil
}

func (e *ExpenseCategoryInMemoryStore) Query(ctx context.Context) ([]ExpenseCategory, error) {
	return e.categories, nil
}
