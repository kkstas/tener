package components

import (
	"encoding/json"

	"github.com/kkstas/tener/internal/model/expense"
	"github.com/kkstas/tener/internal/model/expensecategory"
)

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func getUniqueCategoryNames(categoriesFromExpenses []string, categories []expensecategory.Category) []string {
	m := make(map[string]bool)

	for _, c := range categoriesFromExpenses {
		m[c] = true
	}

	for _, c := range categories {
		m[c.Name] = true
	}

	result := []string{}

	for k := range m {
		result = append(result, k)
	}

	return result
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
