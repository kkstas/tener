package model

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/kkstas/tjener/pkg/validator"
)

var ValidCurrencies = []string{"PLN", "USD", "EUR", "GBP", "CHF", "NOK", "SEK", "DKK", "HUF", "CZK", "CAD", "AUD", "JPY", "CNY", "TRY"}

const expensePK = "expense"

const (
	expenseNameMinLength     = 2
	expenseNameMaxLength     = 50
	expenseCategoryMinLength = 2
	expenseCategoryMaxLength = 50
)

type Expense struct {
	PK                  string  `dynamodbav:"PK"`
	SK                  string  `dynamodbav:"SK"`
	Name                string  `dynamodbav:"name"`
	Category            string  `dynamodbav:"category"`
	Amount              float64 `dynamodbav:"amount"`
	Currency            string  `dynamodbav:"currency"`
	validator.Validator `dynamodbav:"-"`
}

type ExpenseStore struct {
	client    *dynamodb.Client
	tableName string
}

func GetExpenseKey(sk string) map[string]types.AttributeValue {
	PK, err := attributevalue.Marshal(expensePK)
	if err != nil {
		panic(err)
	}
	SK, err := attributevalue.Marshal(sk)
	if err != nil {
		panic(err)
	}
	return map[string]types.AttributeValue{"PK": PK, "SK": SK}
}

func NewExpense(name, category string, amount float64, currency string) (Expense, error) {

	return newExpenseInternal(expensePK, generateCurrentTimestamp(), name, category, amount, currency)
}

func newExpenseInternal(PK, SK, name, category string, amount float64, currency string) (Expense, error) {
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

func (es *ExpenseStore) PutExpense(ctx context.Context, expenseFC Expense) error {
	item, err := attributevalue.MarshalMap(
		Expense{
			PK:       expensePK,
			SK:       generateCurrentTimestamp(),
			Name:     expenseFC.Name,
			Amount:   expenseFC.Amount,
			Currency: expenseFC.Currency,
			Category: expenseFC.Category,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to marshal expense: %w", err)
	}

	_, err = es.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &es.tableName,
		Item:      item,
	})

	if err != nil {
		return fmt.Errorf("failed to put item into DynamoDB: %w", err)
	}

	return nil
}

func (es *ExpenseStore) GetExpense(ctx context.Context, sk string) (Expense, bool, error) {
	expense := Expense{PK: expensePK, SK: sk}
	response, err := es.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &es.tableName,
		Key:       GetExpenseKey(sk),
	})

	if err != nil {
		return Expense{}, false, fmt.Errorf("GetItem DynamoDB operation failed for SK='%s': %w", sk, err)
	}

	if response.Item == nil || len(response.Item) == 0 {
		return Expense{}, false, nil
	}

	err = attributevalue.UnmarshalMap(response.Item, &expense)
	if err != nil {
		return Expense{}, true, fmt.Errorf("failed to unmarshal expense: %w", err)
	}

	return expense, true, nil
}

func (es *ExpenseStore) UpdateExpense(ctx context.Context, SK, name, category string, amount float64, currency string) (Expense, error) {
	var err error
	var response *dynamodb.UpdateItemOutput

	expense, err := newExpenseInternal(expensePK, SK, name, category, amount, currency)
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
		return Expense{}, fmt.Errorf("failed to build expression for update: %w", err)
	}

	response, err = es.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 &es.tableName,
		Key:                       GetExpenseKey(expense.SK),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueUpdatedNew,
	})

	if err != nil {
		return Expense{}, fmt.Errorf("failed to update expense: %w", err)
	}

	var updatedExpense Expense
	err = attributevalue.UnmarshalMap(response.Attributes, &updatedExpense)
	if err != nil {
		return Expense{}, fmt.Errorf("failed to unmarshall updated response: %w", err)
	}

	updatedExpense.PK = expensePK
	updatedExpense.SK = SK

	return updatedExpense, nil
}

func (es *ExpenseStore) DeleteExpense(ctx context.Context, sk string) error {
	_, err := es.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &es.tableName,
		Key:       GetExpenseKey(sk),
	})

	if err != nil {
		return fmt.Errorf("failed to delete expense with SK=%q from the table: %w", sk, err)
	}
	return nil
}

func (es *ExpenseStore) Query(ctx context.Context) ([]Expense, error) {
	keyCond := expression.
		Key("PK").Equal(expression.Value(expensePK)).
		And(expression.Key("SK").GreaterThanEqual(expression.Value(getTimestampDaysAgo(7))))

	exprBuilder := expression.NewBuilder()
	exprBuilder.WithKeyCondition(keyCond)

	expr, err := expression.NewBuilder().
		WithKeyCondition(keyCond).
		Build()

	if err != nil {
		return nil, fmt.Errorf("failed to build expression for query %w", err)
	}

	return es.queryExpenses(ctx, expr)
}

func (es *ExpenseStore) queryExpenses(ctx context.Context, expr expression.Expression) ([]Expense, error) {
	var expenses []Expense

	queryInput := dynamodb.QueryInput{
		TableName:                 &es.tableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
	}

	queryPaginator := dynamodb.NewQueryPaginator(es.client, &queryInput)

	for queryPaginator.HasMorePages() {
		response, err := queryPaginator.NextPage(ctx)

		if err != nil {
			return nil, fmt.Errorf("failed to query for expenses: %w", err)
		}

		var resExpenses []Expense
		err = attributevalue.UnmarshalListOfMaps(response.Items, &resExpenses)

		if err != nil {
			return expenses, fmt.Errorf("failed to unmarshal query response %w", err)
		}

		expenses = append(expenses, resExpenses...)
	}

	return expenses, nil
}
