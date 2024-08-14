package model

import (
	"context"
	"fmt"
	"log"
	"math"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type ExpenseCategoryIsTooShortError struct {
	Name string
}

func (e *ExpenseCategoryIsTooShortError) Error() string {
	return fmt.Sprintf("expense category '%s' is too short", e.Name)
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

type InvalidAmountPrecisionError struct {
	Amount float64
}

func (e *InvalidAmountPrecisionError) Error() string {
	return fmt.Sprintf("amount '%f' has too large precision", e.Amount)
}

type ExpenseNotFoundError struct {
	PK string
	SK string
}

func (e *ExpenseNotFoundError) Error() string {
	return fmt.Sprintf("no item found with PK: %v, SK: %v", e.PK, e.SK)
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

func (e Expense) GetKey() map[string]types.AttributeValue {
	PK, err := attributevalue.Marshal(e.PK)
	if err != nil {
		panic(err)
	}
	SK, err := attributevalue.Marshal(e.SK)
	if err != nil {
		panic(err)
	}
	return map[string]types.AttributeValue{"PK": PK, "SK": SK}
}

func CreateExpense(name, category string, amount float64, currency string) (Expense, error) {
	return newExpense("expense", timestampNow(), name, category, amount, currency)
}

func newExpense(PK, SK, name, category string, amount float64, currency string) (Expense, error) {
	if err := validateCategory(category); err != nil {
		return Expense{}, err
	}
	if err := validateAmount(amount); err != nil {
		return Expense{}, err
	}
	if err := validateCurrency(currency); err != nil {
		return Expense{}, &InvalidCurrencyError{currency}
	}

	return Expense{
		PK:       PK,
		SK:       SK,
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

func (m *ExpenseStore) GetExpense(ctx context.Context, PK, SK string) (Expense, bool, error) {
	expense := Expense{PK: PK, SK: SK}
	response, err := m.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &m.tableName,
		Key:       expense.GetKey(),
	})

	if err != nil {
		log.Printf("Couldn't get info about %v. Here's why: %v\n", PK, err)
		return Expense{}, false, err
	}

	if response.Item == nil || len(response.Item) == 0 {
		return Expense{}, false, nil
	}

	err = attributevalue.UnmarshalMap(response.Item, &expense)
	if err != nil {
		log.Printf("Couldn't unmarshal response. Here's why: %v\n", err)
		return Expense{}, true, err
	}

	return expense, true, nil
}

func (m *ExpenseStore) UpdateExpense(ctx context.Context, PK, SK, name, category string, amount float64, currency string) (Expense, error) {
	var err error
	var response *dynamodb.UpdateItemOutput

	expense, err := newExpense(PK, SK, name, category, amount, currency)
	if err != nil {
		return Expense{}, err
	}

	update := expression.
		Set(expression.Name("name"), expression.Value(expense.Name)).
		Set(expression.Name("category"), expression.Value(expense.Category)).
		Set(expression.Name("amount"), expression.Value(expense.Amount)).
		Set(expression.Name("currency"), expression.Value(expense.Currency))

	expr, err := expression.NewBuilder().WithUpdate(update).Build()

	if err != nil {
		return Expense{}, fmt.Errorf("couldn't build expression for update. Here's why: %v", err)
	}
	response, err = m.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 &m.tableName,
		Key:                       expense.GetKey(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueUpdatedNew,
	})
	if err != nil {
		return Expense{}, fmt.Errorf("couldn't update expense %v. Here's why: %v", expense, err)
	}

	var updatedExpense Expense
	err = attributevalue.UnmarshalMap(response.Attributes, &updatedExpense)
	if err != nil {
		return Expense{}, fmt.Errorf("couldn't unmarshall update response. Here's why: %v", err)
	}
	updatedExpense.PK = PK
	updatedExpense.SK = SK

	return updatedExpense, nil
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

func validateCurrency(curr string) error {
	if !slices.Contains(validCurrencies, curr) {
		return &InvalidCurrencyError{curr}
	}
	return nil
}

func validateCategory(category string) error {
	if len(category) <= 1 {
		return &ExpenseCategoryIsTooShortError{category}
	}
	return nil
}

func validateAmount(amount float64) error {
	if amount == 0 {
		return &ExpenseAmountIsZeroError{}
	}
	if amount != toFixed(amount, 2) {
		return &InvalidAmountPrecisionError{amount}
	}
	return nil
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(int(num*output)) / output
}
