package model

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const expensePK = "expense"

type ExpenseDDBStore struct {
	client    *dynamodb.Client
	tableName string
}

func getExpenseKey(sk string) map[string]types.AttributeValue {
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

func NewExpenseDDBStore(tableName string, client *dynamodb.Client) *ExpenseDDBStore {
	return &ExpenseDDBStore{
		tableName: tableName,
		client:    client,
	}
}

func (es *ExpenseDDBStore) Create(ctx context.Context, expenseFC Expense) error {
	item, err := attributevalue.MarshalMap(
		Expense{
			PK:        expensePK,
			CreatedAt: generateCurrentTimestamp(),
			Name:      expenseFC.Name,
			Amount:    expenseFC.Amount,
			Currency:  expenseFC.Currency,
			Category:  expenseFC.Category,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to marshal expense: %w", err)
	}

	_, err = es.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &es.tableName,
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(SK)"),
	})

	if err != nil {
		return fmt.Errorf("failed to put item into DynamoDB: %w", err)
	}

	return nil
}

func (es *ExpenseDDBStore) FindOne(ctx context.Context, SK string) (Expense, error) {
	expense := Expense{PK: expensePK, CreatedAt: SK}
	response, err := es.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &es.tableName,
		Key:       getExpenseKey(SK),
	})

	if err != nil {
		return Expense{}, fmt.Errorf("GetItem DynamoDB operation failed for SK='%s': %w", SK, err)
	}

	if len(response.Item) == 0 {
		return Expense{}, &ExpenseNotFoundError{SK: SK}
	}

	err = attributevalue.UnmarshalMap(response.Item, &expense)
	if err != nil {
		return Expense{}, fmt.Errorf("failed to unmarshal expense: %w", err)
	}

	return expense, nil
}

func (es *ExpenseDDBStore) Update(ctx context.Context, SK, name, category string, amount float64, currency string) (Expense, error) {
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
		Key:                       getExpenseKey(expense.CreatedAt),
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
	updatedExpense.CreatedAt = SK

	return updatedExpense, nil
}

func (es *ExpenseDDBStore) Delete(ctx context.Context, sk string) error {
	_, err := es.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &es.tableName,
		Key:       getExpenseKey(sk),
	})

	if err != nil {
		return fmt.Errorf("failed to delete expense with SK=%q from the table: %w", sk, err)
	}
	return nil
}

func (es *ExpenseDDBStore) Query(ctx context.Context) ([]Expense, error) {
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

func (es *ExpenseDDBStore) queryExpenses(ctx context.Context, expr expression.Expression) ([]Expense, error) {
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
