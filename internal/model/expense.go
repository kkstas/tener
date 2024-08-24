package model

import (
	"github.com/kkstas/tjener/pkg/validator"
)

var ValidCurrencies = []string{"PLN", "EUR", "GBP", "USD", "CZK", "CHF", "NOK", "SEK", "DKK", "HUF", "CAD", "AUD", "JPY", "CNY", "TRY"}

const (
	ExpenseNameMinLength     = 2
	ExpenseNameMaxLength     = 50
	ExpenseCategoryMinLength = 2
	ExpenseCategoryMaxLength = 50
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

func NewExpenseFC(name, category string, amount float64, currency string) (Expense, error) {
	return validateExpense(Expense{
		PK:        expensePK,
		CreatedAt: generateCurrentTimestamp(),
		Name:      name,
		Category:  category,
		Amount:    amount,
		Currency:  currency,
	})
}

func NewExpenseFU(name, createdAt, category string, amount float64, currency string) (Expense, error) {
	return validateExpense(Expense{
		PK:        expensePK,
		CreatedAt: createdAt,
		Name:      name,
		Category:  category,
		Amount:    amount,
		Currency:  currency,
	})
}

func validateExpense(expense Expense) (Expense, error) {
	expense.Check(validator.StringLengthBetween("name", expense.Name, ExpenseNameMinLength, ExpenseNameMaxLength))
	expense.Check(validator.StringLengthBetween("category", expense.Category, ExpenseCategoryMinLength, ExpenseCategoryMaxLength))
	expense.Check(validator.OneOf("currency", expense.Currency, ValidCurrencies))
	expense.Check(validator.IsValidAmountPrecision("amount", expense.Amount))
	expense.Check(validator.IsNonZero("amount", expense.Amount))

	if err := expense.Validate(); err != nil {
		return Expense{}, err
	}

	return expense, nil
}
