package expense

import (
	"context"
	"fmt"
	"slices"

	"github.com/kkstas/tjener/internal/helpers"
)

type InMemoryStore struct {
	expenses []Expense
}

func (e *InMemoryStore) Create(ctx context.Context, expenseFC Expense, userID, vaultID string) (Expense, error) {
	expenseFC.CreatedBy = userID
	e.expenses = append(e.expenses, expenseFC)
	return expenseFC, nil
}

func (e *InMemoryStore) Delete(ctx context.Context, SK, vaultID string) error {
	var deleted bool

	e.expenses = slices.DeleteFunc(e.expenses, func(expense Expense) bool {
		deleted = true
		return expense.SK == SK
	})

	if !deleted {
		return &NotFoundError{SK: SK}
	}

	return nil
}

func (e *InMemoryStore) Update(ctx context.Context, expenseFU Expense, vaultID string) error {
	var found bool

	for i, el := range e.expenses {
		if el.SK == expenseFU.SK {
			found = true
			e.expenses[i] = expenseFU
		}
	}

	if !found {
		return &NotFoundError{SK: expenseFU.SK}
	}
	return nil
}

func (e *InMemoryStore) FindOne(ctx context.Context, SK, vaultID string) (Expense, error) {
	for _, el := range e.expenses {
		if el.SK == SK {
			return el, nil
		}
	}
	return Expense{}, &NotFoundError{SK: SK}
}

func (es *InMemoryStore) GetMonthlySums(ctx context.Context, monthsAgo int, vaultID string) ([]MonthlySum, error) {
	expenses, err := es.Query(ctx, helpers.MonthsAgo(monthsAgo), helpers.DaysAgo(0), []string{}, vaultID)
	if err != nil {
		return nil, err
	}

	m := make(map[string]MonthlySum)

	for _, val := range expenses {
		sum, found := m[val.Date[:7]+val.Category]
		if !found {
			m[val.Date[:7]+val.Category] = MonthlySum{
				SK:       val.Date[:7] + val.Category,
				Category: val.Category,
				Sum:      val.Amount,
			}
			continue
		}
		sum.Sum += val.Amount
		m[val.Date[:7]+val.Category] = sum
	}

	var results []MonthlySum

	for _, v := range m {
		results = append(results, v)
	}

	return results, err
}

// Retrieves expenses between the given `from` and `to` YYYY-MM-DD dates (inclusive).
func (e *InMemoryStore) Query(ctx context.Context, from, to string, categories []string, vaultID string) ([]Expense, error) {
	daysDiff, err := helpers.DaysBetween(from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get number of days between 'from' and 'to' date: %w", err)
	}
	if daysDiff < minQueryRangeDaysDiff || daysDiff > maxQueryRangeDaysDiff {
		return nil, fmt.Errorf(
			"invalid difference between 'from' and 'to' date; got=%d, max=%d, min=%d",
			daysDiff,
			minQueryRangeDaysDiff,
			maxQueryRangeDaysDiff,
		)
	}

	var expenses []Expense

	for _, expense := range e.expenses {
		daysAfterFrom, err := helpers.DaysBetween(from, expense.Date)
		if err != nil {
			return nil, fmt.Errorf("failed to get number of days between 'from' and 'expense.Date' for expense: %+v: %w", expense, err)
		}
		daysBeforeTo, err := helpers.DaysBetween(expense.Date, to)
		if err != nil {
			return nil, fmt.Errorf("failed to get number of days between 'expense.Date' and 'to for expense: %+v: %w", expense, err)
		}

		hasQueriedCategory := true

		if len(categories) > 0 {
			if !slices.Contains(categories, expense.Category) {
				hasQueriedCategory = false
			}
		}

		if daysAfterFrom >= 0 && daysBeforeTo >= 0 && hasQueriedCategory {
			expenses = append(expenses, expense)
		}
	}

	return expenses, nil
}
