package components

import (
	"slices"
	"time"

	"github.com/kkstas/tjener/internal/model"
)

func parseTime(timestamp string) string {
	t, err := time.Parse("2006-01-02", timestamp)
	if err != nil {
		return err.Error()
	}
	return t.Local().Format("Monday, 02 Jan")
}

type expenseDay struct {
	date     string
	expenses []model.Expense
}

func parseExpensesIntoExpenseDays(expenses []model.Expense) []expenseDay {
	days := make(map[string][]model.Expense)

	for _, expense := range expenses {
		days[expense.CreatedAt[:10]] = append(days[expense.CreatedAt[:10]], expense)
	}

	keys := make([]string, 0, len(days))

	for k := range days {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	expenseDays := make([]expenseDay, 0, len(keys))

	for i := len(keys) - 1; i >= 0; i-- {
		expenseDays = append(expenseDays, expenseDay{date: keys[i], expenses: days[keys[i]]})
	}
	return expenseDays
}
