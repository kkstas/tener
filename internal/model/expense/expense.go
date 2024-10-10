package expense

import (
	"strings"
	"time"

	"github.com/kkstas/tjener/internal/helpers"
	"github.com/kkstas/tjener/internal/model/expensecategory"
	"github.com/kkstas/tjener/pkg/validator"
)

var PaymentMethods = []string{"Cash", "Credit Card", "Debit Card"}

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
	PaymentMethod       string  `dynamodbav:"paymentMethod"`
	CreatedAt           string  `dynamodbav:"createdAt"`
	CreatedBy           string  `dynamodbav:"createdBy"`
	validator.Validator `dynamodbav:"-"`
}

func New(name, date, category string, amount float64, paymentMethod string) (Expense, error) {
	currentTimestamp := helpers.GenerateCurrentTimestamp()
	return validate(Expense{
		SK:            buildSK(date, currentTimestamp),
		Name:          strings.TrimSpace(name),
		Date:          date,
		Category:      strings.TrimSpace(category),
		Amount:        amount,
		PaymentMethod: paymentMethod,
		CreatedAt:     currentTimestamp,
	})
}

func NewFU(sk, name, date, category string, amount float64, paymentMethod string) (Expense, error) {
	return validate(Expense{
		SK:            sk,
		Name:          strings.TrimSpace(name),
		Date:          date,
		Category:      strings.TrimSpace(category),
		Amount:        amount,
		PaymentMethod: paymentMethod,
	})
}

func validate(expense Expense) (Expense, error) {
	expense.Check(validator.StringLengthBetween("name", expense.Name, NameMinLength, NameMaxLength))
	expense.Check(validator.StringLengthBetween(
		"category",
		expense.Category,
		expensecategory.CategoryNameMinLength,
		expensecategory.CategoryNameMaxLength,
	))
	expense.Check(validator.OneOf("paymentMethod", expense.PaymentMethod, PaymentMethods))
	expense.Check(validator.IsAmountPrecision("amount", expense.Amount))
	expense.Check(validator.IsNonZero("amount", expense.Amount))
	expense.Check(validator.IsTime("date", time.DateOnly, expense.Date))

	if err := expense.Validate(); err != nil {
		return Expense{}, err
	}

	return expense, nil
}
