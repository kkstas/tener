package expensecategory

import (
	"context"
	"slices"
)

type InMemoryStore struct {
	categories []Category
}

func (e *InMemoryStore) Create(ctx context.Context, categoryFC Category) error {
	e.categories = append(e.categories, categoryFC)
	return nil
}

func (e *InMemoryStore) Delete(ctx context.Context, SK string) error {
	var deleted bool

	e.categories = slices.DeleteFunc(e.categories, func(category Category) bool {
		deleted = true
		return category.Name == SK
	})

	if !deleted {
		return &NotFoundError{SK: SK}
	}

	return nil
}

func (e *InMemoryStore) Query(ctx context.Context) ([]Category, error) {
	return e.categories, nil
}
