package model_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kkstas/tjener/internal/model"
)

func TestCreateInMemoryExpenseCategory(t *testing.T) {
	ctx := context.Background()

	store := model.ExpenseCategoryInMemoryStore{}

	categories, _ := store.Query(ctx)

	err := store.Create(ctx, model.ExpenseCategory{})
	if err != nil {
		t.Fatalf("failed creating expense category, %v", err)
	}
	newCategories, err := store.Query(ctx)

	if err != nil {
		t.Fatalf("failed querying for expense categories after creating one, %v", err)
	}
	if (len(newCategories) - 1) != len(categories) {
		t.Errorf("expected one new category added. got %d", len(newCategories)-len(categories))
	}
}

func TestDeleteInMemoryExpenseCategory(t *testing.T) {
	t.Run("deletes existing categories", func(t *testing.T) {
		ctx := context.Background()
		store := model.ExpenseCategoryInMemoryStore{}
		name := "some name"

		_ = store.Create(ctx, model.ExpenseCategory{Name: name})

		categories, _ := store.Query(ctx)
		if len(categories) != 1 {
			t.Fatalf("expected one category saved in the store, got %#v", categories)
		}

		err := store.Delete(ctx, name)
		if err != nil {
			t.Fatalf("didn't expect an error while deleting category but got one: %v", err)
		}

		categories, _ = store.Query(ctx)
		if len(categories) != 0 {
			t.Errorf("expected 0 categories after deleting, got %d", len(categories))
		}
	})

	t.Run("returns proper error when category for deletion does not exist", func(t *testing.T) {
		ctx := context.Background()
		store := model.ExpenseCategoryInMemoryStore{}
		nonExistingName := "asdf"

		err := store.Delete(ctx, nonExistingName)
		if err == nil {
			t.Fatal("expected an error but didn't get one")
		}

		var notFoundErr *model.ExpenseCategoryNotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &model.ExpenseCategoryNotFoundError{CreatedAt: nonExistingName})
		}
	})
}
