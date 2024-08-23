package model

import (
	"github.com/kkstas/tjener/pkg/validator"
)

var ValidCurrencies = []string{"PLN", "USD", "EUR", "GBP", "CHF", "NOK", "SEK", "DKK", "HUF", "CZK", "CAD", "AUD", "JPY", "CNY", "TRY"}

const (
	expenseNameMinLength     = 2
	expenseNameMaxLength     = 50
	expenseCategoryMinLength = 2
	expenseCategoryMaxLength = 50
)

type Expense struct {
	PK                  string  `dynamodbav:"PK"`
	CreatedAt           string  `dynamodbav:"SK"`
	Name                string  `dynamodbav:"name"`
	Category            string  `dynamodbav:"category"`
	Amount              float64 `dynamodbav:"amount"`
	Currency            string  `dynamodbav:"currency"`
	validator.Validator `dynamodbav:"-"`
}

func NewExpense(name, category string, amount float64, currency string) (Expense, error) {
	return newExpenseInternal(expensePK, generateCurrentTimestamp(), name, category, amount, currency)
}

func newExpenseInternal(PK, createdAt, name, category string, amount float64, currency string) (Expense, error) {
	expense := Expense{}
	expense.Check(validator.StringLengthBetween("name", name, expenseNameMinLength, expenseNameMaxLength))
	expense.Check(validator.StringLengthBetween("category", category, expenseCategoryMinLength, expenseCategoryMaxLength))
	expense.Check(validator.OneOf("currency", currency, ValidCurrencies))
	expense.Check(validator.IsValidAmountPrecision("amount", amount))
	expense.Check(validator.IsNonZero("amount", amount))

	if err := expense.Validate(); err != nil {
		return Expense{}, err
	}

	return Expense{
		PK:        PK,
		CreatedAt: createdAt,
		Name:      name,
		Category:  category,
		Amount:    amount,
		Currency:  currency,
	}, nil
}
