package components

import (
	"encoding/json"

	"github.com/kkstas/tener/internal/model/expense"
)

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
