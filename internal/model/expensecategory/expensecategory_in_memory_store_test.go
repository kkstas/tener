package expensecategory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kkstas/tjener/internal/model/expensecategory"
)

func TestInMemoryCreate(t *testing.T) {
	ctx := context.Background()

	store := expensecategory.InMemoryStore{}

	categories, _ := store.Query(ctx)

	err := store.Create(ctx, expensecategory.Category{})
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

func TestInMemoryDelete(t *testing.T) {
	t.Run("deletes existing categories", func(t *testing.T) {
		ctx := context.Background()
		store := expensecategory.InMemoryStore{}
		name := "some name"

		_ = store.Create(ctx, expensecategory.Category{Name: name})

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
		store := expensecategory.InMemoryStore{}
		nonExistingName := "asdf"

		err := store.Delete(ctx, nonExistingName)
		if err == nil {
			t.Fatal("expected an error but didn't get one")
		}

		var notFoundErr *expensecategory.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &expensecategory.NotFoundError{SK: nonExistingName})
		}
	})
}
