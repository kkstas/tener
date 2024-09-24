package expense

import (
	"strings"
	"time"

	"github.com/kkstas/tjener/internal/helpers"
	"github.com/kkstas/tjener/internal/model/expensecategory"
	"github.com/kkstas/tjener/pkg/validator"
)

var ValidCurrencies = []string{"PLN", "EUR", "GBP", "USD", "CZK", "CHF", "NOK", "SEK", "DKK", "HUF", "CAD", "AUD", "JPY", "CNY", "TRY"}

const (
	pkPrefix              = "expense"
	minQueryRangeDaysDiff = 0
	maxQueryRangeDaysDiff = 365

	NameMinLength = 2
	NameMaxLength = 50
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

func New(name, date, category string, amount float64, currency string) (Expense, error) {
	currentTimestamp := helpers.GenerateCurrentTimestamp()
	return validate(Expense{
		SK:        buildSK(date, currentTimestamp),
		Name:      strings.TrimSpace(name),
		Date:      date,
		Category:  strings.TrimSpace(category),
		Amount:    amount,
		Currency:  strings.TrimSpace(currency),
		CreatedAt: currentTimestamp,
	})
}

func NewFU(sk, name, date, category string, amount float64, currency string) (Expense, error) {
	return validate(Expense{
		SK:       sk,
		Name:     strings.TrimSpace(name),
		Date:     date,
		Category: strings.TrimSpace(category),
		Amount:   amount,
		Currency: strings.TrimSpace(currency),
	})
}

func validate(expense Expense) (Expense, error) {
	expense.Check(validator.StringLengthBetween("name", expense.Name, NameMinLength, NameMaxLength))
	expense.Check(validator.StringLengthBetween("category", expense.Category, expensecategory.CategoryNameMinLength, expensecategory.CategoryNameMaxLength))
	expense.Check(validator.OneOf("currency", expense.Currency, ValidCurrencies))
	expense.Check(validator.IsAmountPrecision("amount", expense.Amount))
	expense.Check(validator.IsNonZero("amount", expense.Amount))
	expense.Check(validator.IsTime("date", time.DateOnly, expense.Date))

	if err := expense.Validate(); err != nil {
		return Expense{}, err
	}

	return expense, nil
}
