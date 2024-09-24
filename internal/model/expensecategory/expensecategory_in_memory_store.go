package expensecategory

import (
	"context"
	"slices"
)

type InMemoryStore struct {
	categories []Category
}

func (e *InMemoryStore) Create(ctx context.Context, categoryFC Category, vaultID string) error {
	e.categories = append(e.categories, categoryFC)
	return nil
}

func (e *InMemoryStore) Delete(ctx context.Context, name, vaultID string) error {
	var deleted bool

	e.categories = slices.DeleteFunc(e.categories, func(category Category) bool {
		deleted = true
		return category.Name == name
	})

	if !deleted {
		return &NotFoundError{SK: name}
	}

	return nil
}

func (e *InMemoryStore) FindAll(ctx context.Context, vaultID string) ([]Category, error) {
	return e.categories, nil
}
