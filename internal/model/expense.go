package model

import (
	"strings"
	"time"

	"github.com/kkstas/tjener/pkg/validator"
)

var ValidCurrencies = []string{"PLN", "EUR", "GBP", "USD", "CZK", "CHF", "NOK", "SEK", "DKK", "HUF", "CAD", "AUD", "JPY", "CNY", "TRY"}

const (
	expensePK             = "expense"
	minQueryRangeDaysDiff = 0
	maxQueryRangeDaysDiff = 365

	ExpenseNameMinLength     = 2
	ExpenseNameMaxLength     = 50
	ExpenseCategoryMinLength = 2
	ExpenseCategoryMaxLength = 50
)

type Expense struct {
	PK                  string  `dynamodbav:"PK"`
	SK                  string  `dynamodbav:"SK"`
	Name                string  `dynamodbav:"name"`
	Date                string  `dynamodbav:"date"`
	Category            string  `dynamodbav:"category"`
	Amount              float64 `dynamodbav:"amount"`
	Currency            string  `dynamodbav:"currency"`
	CreatedAt           string  `dynamodbav:"createdAt"`
	validator.Validator `dynamodbav:"-"`
}

func NewExpenseFC(name, date, category string, amount float64, currency string) (Expense, error) {
	currentTimestamp := generateCurrentTimestamp()
	return validateExpense(Expense{
		PK:        expensePK,
		SK:        buildSK(date, currentTimestamp),
		Name:      strings.TrimSpace(name),
		Date:      date,
		Category:  strings.TrimSpace(category),
		Amount:    amount,
		Currency:  strings.TrimSpace(currency),
		CreatedAt: currentTimestamp,
	})
}

func NewExpenseFU(name, SK, date, category string, amount float64, currency string) (Expense, error) {
	return validateExpense(Expense{
		PK:       expensePK,
		SK:       SK,
		Name:     strings.TrimSpace(name),
		Date:     date,
		Category: strings.TrimSpace(category),
		Amount:   amount,
		Currency: strings.TrimSpace(currency),
	})
}

func validateExpense(expense Expense) (Expense, error) {
	expense.Check(validator.StringLengthBetween("name", expense.Name, ExpenseNameMinLength, ExpenseNameMaxLength))
	expense.Check(validator.StringLengthBetween("category", expense.Category, ExpenseCategoryMinLength, ExpenseCategoryMaxLength))
	expense.Check(validator.OneOf("currency", expense.Currency, ValidCurrencies))
	expense.Check(validator.IsValidAmountPrecision("amount", expense.Amount))
	expense.Check(validator.IsNonZero("amount", expense.Amount))
	expense.Check(validator.IsValidTime("date", time.DateOnly, expense.Date))

	if err := expense.Validate(); err != nil {
		return Expense{}, err
	}

	return expense, nil
}
