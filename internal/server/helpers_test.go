package server

import (
	"slices"
	"testing"

	"github.com/kkstas/tjener/internal/model/expense"
	"github.com/kkstas/tjener/internal/model/expensecategory"
)

func TestExtractUserIDs(t *testing.T) {
	id1 := "one"
	id2 := "two"

	t.Run("finds ids from expenses and expense categories", func(t *testing.T) {
		expenses := []expense.Expense{{CreatedBy: id1}, {CreatedBy: id1}}
		expenseCategories := []expensecategory.Category{{CreatedBy: id2}, {CreatedBy: id2}}

		ids := extractUserIDs(expenses, expenseCategories)

		got := len(ids)
		want := 2
		if got != want {
			t.Errorf("expected ids slice with length of %d, got %d", want, got)
		}
		if !slices.Contains(ids, id1) {
			t.Errorf("expected %s in %v", id1, ids)
		}
		if !slices.Contains(ids, id2) {
			t.Errorf("expected %s in %v", id2, ids)
		}
	})

	t.Run("finds ids only from expenses if expense categories are empty", func(t *testing.T) {
		expenses := []expense.Expense{{CreatedBy: id1}, {CreatedBy: id1}}
		expenseCategories := []expensecategory.Category{}

		ids := extractUserIDs(expenses, expenseCategories)

		got := len(ids)
		want := 1
		if got != want {
			t.Errorf("expected ids slice with length of %d, got %d", want, got)
		}
		if !slices.Contains(ids, id1) {
			t.Errorf("expected %s in %v", id1, ids)
		}
	})

	t.Run("finds ids only from expense categories if expenses are empty", func(t *testing.T) {
		expenses := []expense.Expense{}
		expenseCategories := []expensecategory.Category{{CreatedBy: id1}, {CreatedBy: id1}}

		ids := extractUserIDs(expenses, expenseCategories)

		got := len(ids)
		want := 1
		if got != want {
			t.Errorf("expected ids slice with length of %d, got %d", want, got)
		}
		if !slices.Contains(ids, id1) {
			t.Errorf("expected %s in %v", id1, ids)
		}
	})

	t.Run("return empty slice if nothing is found", func(t *testing.T) {
		ids := extractUserIDs([]expense.Expense{}, []expensecategory.Category{})

		got := len(ids)
		want := 0
		if got != want {
			t.Errorf("expected ids slice with length of %d, got %d", want, got)
		}
	})
}
