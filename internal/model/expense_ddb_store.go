package model

import (
	"context"
	"errors"
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

func (es *ExpenseDDBStore) marshalExpense(PK, SK, name, date string, amount float64, currency, category, createdAt string) (Expense, map[string]types.AttributeValue, error) {
	newExpense := Expense{
		PK:        PK,
		SK:        SK,
		Name:      name,
		Date:      date,
		Amount:    amount,
		Currency:  currency,
		Category:  category,
		CreatedAt: createdAt,
	}
	item, err := attributevalue.MarshalMap(newExpense)
	return newExpense, item, err
}

func (es *ExpenseDDBStore) Create(ctx context.Context, expenseFC Expense) (Expense, error) {
	newExpense, item, err := es.marshalExpense(expensePK,
		buildSK(expenseFC.Date, expenseFC.CreatedAt),
		expenseFC.Name,
		expenseFC.Date,
		expenseFC.Amount,
		expenseFC.Currency,
		expenseFC.Category,
		expenseFC.CreatedAt,
	)

	if err != nil {
		return Expense{}, fmt.Errorf("failed to marshal expense: %w", err)
	}

	_, err = es.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &es.tableName,
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(SK)"),
	})

	if err != nil {
		return Expense{}, fmt.Errorf("failed to put item into DynamoDB: %w", err)
	}

	return newExpense, nil
}

func (es *ExpenseDDBStore) FindOne(ctx context.Context, SK string) (Expense, error) {
	expense := Expense{PK: expensePK, SK: SK}
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

func (es *ExpenseDDBStore) Update(ctx context.Context, expenseFU Expense) error {
	foundExpense, err := es.FindOne(ctx, expenseFU.SK)
	if err != nil {
		return fmt.Errorf("failed to find expense for update: %w", err)
	}

	expenseFU.CreatedAt = foundExpense.CreatedAt

	if expenseFU.SK == buildSK(expenseFU.Date, foundExpense.CreatedAt) {
		return es.updateWithoutNewSK(ctx, expenseFU)
	}
	return es.updateWithNewSK(ctx, expenseFU)
}

func (es *ExpenseDDBStore) updateWithNewSK(ctx context.Context, expenseFU Expense) error {
	deleteItem := types.TransactWriteItem{
		Delete: &types.Delete{
			TableName:           aws.String(es.tableName),
			Key:                 getExpenseKey(expenseFU.SK),
			ConditionExpression: aws.String("attribute_exists(SK)"),
		},
	}

	expense, item, err := es.marshalExpense(
		expensePK,
		buildSK(expenseFU.Date, expenseFU.CreatedAt),
		expenseFU.Name,
		expenseFU.Date,
		expenseFU.Amount,
		expenseFU.Currency,
		expenseFU.Category,
		expenseFU.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to marshal expense: %w", err)
	}

	putItem := types.TransactWriteItem{
		Put: &types.Put{
			TableName: aws.String(es.tableName),
			Item:      item,
		},
	}

	_, err = es.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{deleteItem, putItem},
	})

	if err != nil {
		var transactionErr *types.TransactionCanceledException

		if errors.As(err, &transactionErr) {
			for _, reason := range transactionErr.CancellationReasons {
				if reason.Code != nil && *reason.Code == "ConditionalCheckFailed" {
					return &ExpenseNotFoundError{SK: expense.SK}
				}
			}
		}
		return fmt.Errorf("failed to update expense atomically: %w", err)
	}

	return nil
}

func (es *ExpenseDDBStore) updateWithoutNewSK(ctx context.Context, expenseFU Expense) error {
	update := expression.
		Set(expression.Name("name"), expression.Value(expenseFU.Name)).
		Set(expression.Name("category"), expression.Value(expenseFU.Category)).
		Set(expression.Name("amount"), expression.Value(expenseFU.Amount)).
		Set(expression.Name("currency"), expression.Value(expenseFU.Currency))

	expr, err := expression.NewBuilder().WithUpdate(update).Build()

	if err != nil {
		return fmt.Errorf("failed to build expression for update: %w", err)
	}

	_, err = es.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 &es.tableName,
		Key:                       getExpenseKey(expenseFU.SK),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueUpdatedNew,
		ConditionExpression:       aws.String("attribute_exists(SK)"),
	})

	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return &ExpenseNotFoundError{SK: expenseFU.SK}
		}
		return fmt.Errorf("failed to update expense: %w", err)
	}

	return nil
}

func (es *ExpenseDDBStore) Delete(ctx context.Context, SK string) error {
	if _, err := es.FindOne(ctx, SK); err != nil {
		return &ExpenseNotFoundError{SK: SK}
	}

	_, err := es.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &es.tableName,
		Key:       getExpenseKey(SK),
	})

	if err != nil {
		return fmt.Errorf("failed to delete expense with SK=%q from the table: %w", SK, err)
	}

	return nil
}

func (es *ExpenseDDBStore) Query(ctx context.Context) ([]Expense, error) {
	keyCond := expression.
		Key("PK").Equal(expression.Value(expensePK)).
		And(expression.Key("SK").GreaterThanEqual(expression.Value(getDateStringDaysAgo(31))))

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
