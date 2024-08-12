package model

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type ExpenseNameIsTooShortError struct {
	Name string
}

func (e *ExpenseNameIsTooShortError) Error() string {
	return fmt.Sprintf("expense name '%s' is too short", e.Name)
}

type ExpenseAmountIsZeroError struct{}

func (e *ExpenseAmountIsZeroError) Error() string {
	return "expense amount cannot be zero"
}

type InvalidCurrencyError struct {
	Currency string
}

func (e *InvalidCurrencyError) Error() string {
	return fmt.Sprintf("currency '%s' is invalid", e.Currency)
}

type Expense struct {
	PK       string  `dynamodbav:"PK"       json:"PK"`
	SK       string  `dynamodbav:"SK"       json:"SK"`
	Name     string  `dynamodbav:"name"     json:"name"`
	Category string  `dynamodbav:"category" json:"category"`
	Amount   float64 `dynamodbav:"amount"   json:"amount"`
	Currency string  `dynamodbav:"currency" json:"currency"`
}

type ExpenseStore struct {
	client    *dynamodb.Client
	tableName string
}

func CreateExpense(name, category string, amount float64, currency string) (Expense, error) {
	if len(category) <= 1 {
		return Expense{}, &ExpenseNameIsTooShortError{name}
	}

	if amount == 0 {
		return Expense{}, &ExpenseAmountIsZeroError{}

	}

	if !isCurrencyValid(currency) {
		return Expense{}, &InvalidCurrencyError{currency}
	}

	return Expense{
		PK:       "expense",
		SK:       timestampNow(),
		Name:     name,
		Category: category,
		Amount:   amount,
		Currency: currency,
	}, nil
}

func NewExpenseStore(tableName string, client *dynamodb.Client) *ExpenseStore {
	return &ExpenseStore{
		tableName: tableName,
		client:    client,
	}
}

func (m *ExpenseStore) PutItem(ctx context.Context, expenseFC Expense) error {
	item, err := attributevalue.MarshalMap(
		Expense{
			PK:       "expense",
			SK:       timestampNow(),
			Name:     expenseFC.Name,
			Amount:   expenseFC.Amount,
			Currency: expenseFC.Currency,
			Category: expenseFC.Category,
		},
	)

	if err != nil {
		return err
	}

	_, err = m.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &m.tableName,
		Item:      item,
	})

	if err != nil {
		return err
	}

	return nil
}

func (m *ExpenseStore) Query(ctx context.Context) ([]Expense, error) {
	keyCond := expression.
		Key("PK").Equal(expression.Value("expense")).
		And(expression.Key("SK").GreaterThanEqual(expression.Value(getDateDaysAgo(7))))

	exprBuilder := expression.NewBuilder()
	exprBuilder.WithKeyCondition(keyCond)

	expr, err := expression.NewBuilder().
		WithKeyCondition(keyCond).
		Build()

	if err != nil {
		return nil, fmt.Errorf("couldn't build expression for query %w", err)
	}

	return m.queryItems(ctx, expr)

}

func (m *ExpenseStore) QueryByCategory(ctx context.Context, category string) ([]Expense, error) {
	keyCond := expression.
		Key("PK").Equal(expression.Value("expense")).
		And(expression.Key("SK").GreaterThanEqual(expression.Value(getDateDaysAgo(7))))

	filterCond := expression.Name("category").Equal(expression.Value(category))

	exprBuilder := expression.NewBuilder()
	exprBuilder.WithKeyCondition(keyCond)

	expr, err := expression.NewBuilder().
		WithKeyCondition(keyCond).
		WithFilter(filterCond).
		Build()

	if err != nil {
		return nil, fmt.Errorf("couldn't build expression for query %w", err)
	}

	return m.queryItems(ctx, expr)
}

func (m *ExpenseStore) queryItems(ctx context.Context, expr expression.Expression) ([]Expense, error) {
	var expenses []Expense
	var response *dynamodb.QueryOutput
	var err error

	queryInput := dynamodb.QueryInput{
		TableName:                 &m.tableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
	}

	queryPaginator := dynamodb.NewQueryPaginator(m.client, &queryInput)

	for queryPaginator.HasMorePages() {
		response, err = queryPaginator.NextPage(ctx)

		if err != nil {
			return nil, fmt.Errorf("couldn't query for expenses %w", err)
		}

		var resExpenses []Expense
		err = attributevalue.UnmarshalListOfMaps(response.Items, &resExpenses)

		if err != nil {
			return expenses, fmt.Errorf("couldn't unmarshal query response %w", err)
		}

		expenses = append(expenses, resExpenses...)
	}

	return expenses, err
}

func timestampNow() string {
	loc, _ := time.LoadLocation("Europe/Warsaw")
	return time.Now().In(loc).Format(time.RFC3339Nano)
}

func getDateDaysAgo(days int) string {
	loc, _ := time.LoadLocation("Europe/Warsaw")
	now := resetToMidnight(time.Now(), loc)
	pastDate := now.Add(-(time.Duration(days) * 24 * time.Hour))
	return pastDate.Format(time.RFC3339Nano)
}

func resetToMidnight(t time.Time, loc *time.Location) time.Time {
	return time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		0, 0, 0, 0,
		loc,
	)
}

var validCurrencies = []string{"PLN", "USD", "EUR", "GBP", "CHF", "NOK", "SEK", "DKK", "HUF", "CZK", "CAD", "AUD", "JPY", "CNY", "TRY"}

func isCurrencyValid(curr string) bool {
	return slices.Contains(validCurrencies, curr)
}
