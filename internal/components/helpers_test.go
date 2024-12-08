package components

import (
	"reflect"
	"slices"
	"testing"

	"github.com/kkstas/tener/internal/model/expense"
)

func TestExtractCategories(t *testing.T) {

	t.Run("returns unique categories", func(t *testing.T) {
		categories := []expense.Expense{
			{Category: "cat1"},
			{Category: "cat1"},
			{Category: "cat2"},
		}

		got := extractCategories(categories)
		want := []string{"cat1", "cat2"}

		slices.Sort(got)
		slices.Sort(want)

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %#v, want %#v", got, want)
		}
	})

	t.Run("doesn't return empty string categories", func(t *testing.T) {
		categories := []expense.Expense{
			{Category: "cat1"},
			{Category: ""},
			{Category: "cat2"},
		}

		got := extractCategories(categories)
		want := []string{"cat1", "cat2"}

		slices.Sort(got)
		slices.Sort(want)

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %#v, want %#v", got, want)
		}
	})

	t.Run("returns empty string slice if no expenses were provided", func(t *testing.T) {
		categories := []expense.Expense{}

		got := extractCategories(categories)
		want := []string{}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %#v, want %#v", got, want)
		}
	})
}
