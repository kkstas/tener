package components

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/kkstas/tjener/internal/model"
)

func TestParseExpenses(t *testing.T) {
	expenses := []model.Expense{
		{CreatedAt: "2024-08-18T14:36:51.351629414+02:00", Name: "zero", Category: "food", Amount: 1, Currency: "PLN"},
		{CreatedAt: "2024-08-17T15:26:14.136402263+02:00", Name: "one", Category: "aa", Amount: 124, Currency: "PLN"},
		{CreatedAt: "2024-08-18T15:26:14.136402263+02:00", Name: "two", Category: "asdf", Amount: 14, Currency: "PLN"},
		{CreatedAt: "2024-08-16T15:26:14.136402263+02:00", Name: "three", Category: "dasdfas", Amount: 82, Currency: "PLN"},
	}

	t.Run("returns expenseDays in date descending order", func(t *testing.T) {
		expenseDays := parseExpensesIntoExpenseDays(expenses)
		got := make([]string, 0, len(expenseDays))
		for _, expenseDay := range expenseDays {
			got = append(got, expenseDay.date)
		}

		want := make([]string, len(expenseDays))
		copy(want, got)

		sort.Sort(sort.Reverse(sort.StringSlice(want)))

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got = %#v| want = %#v", got, want)
		}
	})

	expenseDays := parseExpensesIntoExpenseDays(expenses)

	cases := []struct {
		date         string
		expenseCount int
	}{
		{"2024-08-16", 1},
		{"2024-08-17", 1},
		{"2024-08-18", 2},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("should contain %d expense(s) for date %s", c.expenseCount, c.date), func(t *testing.T) {
			var foundExpenses []model.Expense
			var found bool
			for _, day := range expenseDays {
				if day.date == c.date {
					foundExpenses = day.expenses
					found = true
				}
			}
			if !found {
				t.Fatalf("not found expenseDay struct for date %s", c.date)
			}

			if len(foundExpenses) != c.expenseCount {
				t.Errorf("expected expense amount %d, got %d for date %s", c.expenseCount, len(foundExpenses), c.date)
			}
		})
	}
}
