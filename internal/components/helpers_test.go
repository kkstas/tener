package components

import (
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/kkstas/tjener/internal/model/expense"
)

func TestParseDate(t *testing.T) {
	t.Run("returns parsed date if provided date is valid YYYY-MM-DD", func(t *testing.T) {
		got := parseDate("2024-09-07", time.RFC3339)
		want := "2024-09-07T00:00:00Z"
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})

	t.Run("returns input date if provided date is not valid YYYY-MM-DD", func(t *testing.T) {
		input := "2024-0907"
		got := parseDate(input, time.RFC3339)
		if got != input {
			t.Errorf("got %s, want %s", got, input)
		}
	})
}

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
