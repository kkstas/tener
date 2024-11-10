package components

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kkstas/tener/internal/model/expense"
)

func parseDate(date, layout string) string {
	parsedDate, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return date
	}
	readableDate := parsedDate.Format(layout)
	return readableDate
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func extractCategories(expenses []expense.Expense) []string {
	uniqueCategories := make(map[string]bool)

	for _, exp := range expenses {
		if exp.Category != "" {
			uniqueCategories[exp.Category] = true
		}
	}

	foundCategories := make([]string, 0, len(uniqueCategories))

	for key := range uniqueCategories {
		foundCategories = append(foundCategories, key)
	}

	return foundCategories
}

func getTotalAmount(expenses []expense.Expense) string {
	var sum float64
	for _, val := range expenses {
		sum += val.Amount
	}
	return fmt.Sprintf("%.2f", sum)
}
