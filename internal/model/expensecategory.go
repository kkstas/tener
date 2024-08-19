package model

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const expenseCategoryPK = "expensecategory"

type ExpenseCategory struct {
	PK string `dynamodbav:"PK"       json:"PK"`
	SK string `dynamodbav:"SK"       json:"SK"`
}

type ExpenseCategoryStore struct {
	client    *dynamodb.Client
	tableName string
}

func (c *ExpenseCategory) GetKey() map[string]types.AttributeValue {
	PK, err := attributevalue.Marshal(c.PK)
	if err != nil {
		panic(err)
	}
	SK, err := attributevalue.Marshal(c.SK)
	if err != nil {
		panic(err)
	}
	return map[string]types.AttributeValue{"PK": PK, "SK": SK}
}

func NewExpenseCategory(name string) (ExpenseCategory, error) {
	if err := validateCategory(name); err != nil {
		return ExpenseCategory{}, err
	}

	return ExpenseCategory{
		PK: expenseCategoryPK,
		SK: name,
	}, nil
}

func NewExpenseCategoryStore(tableName string, client *dynamodb.Client) *ExpenseCategoryStore {
	return &ExpenseCategoryStore{
		tableName: tableName,
		client:    client,
	}
}

func (cs *ExpenseCategoryStore) CreateExpenseCategory(ctx context.Context, categoryFC ExpenseCategory) error {
	item, err := attributevalue.MarshalMap(
		ExpenseCategory{
			PK: expenseCategoryPK,
			SK: categoryFC.SK,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to marshal expense category: %w", err)
	}

	_, err = cs.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &cs.tableName,
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to put expense category into DynamoDB: %w", err)
	}

	return nil
}

func (cs *ExpenseCategoryStore) DeleteExpenseCategory(ctx context.Context, name string) error {
	categoryFD := ExpenseCategory{PK: expenseCategoryPK, SK: name}
	_, err := cs.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &cs.tableName,
		Key:       categoryFD.GetKey(),
	})
	if err != nil {
		return fmt.Errorf("failed to delete expense category with name=%q from the table: %w", name, err)
	}

	return nil
}

func (cs *ExpenseCategoryStore) Query(ctx context.Context) ([]ExpenseCategory, error) {
	keyCond := expression.
		Key("PK").Equal(expression.Value(expenseCategoryPK))

	exprBuilder := expression.NewBuilder()
	exprBuilder.WithKeyCondition(keyCond)

	expr, err := expression.NewBuilder().
		WithKeyCondition(keyCond).
		Build()

	if err != nil {
		return nil, fmt.Errorf("failed to build expression for expense category query %w", err)
	}

	return cs.queryExpenseCategories(ctx, expr)

}

func (cs *ExpenseCategoryStore) queryExpenseCategories(ctx context.Context, expr expression.Expression) ([]ExpenseCategory, error) {
	var categories []ExpenseCategory

	queryInput := dynamodb.QueryInput{
		TableName:                 &cs.tableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
	}

	queryPaginator := dynamodb.NewQueryPaginator(cs.client, &queryInput)

	for queryPaginator.HasMorePages() {
		response, err := queryPaginator.NextPage(ctx)

		if err != nil {
			return nil, fmt.Errorf("failed to query for expense categories: %w", err)
		}

		var resCategories []ExpenseCategory
		err = attributevalue.UnmarshalListOfMaps(response.Items, &resCategories)
		if err != nil {
			return categories, fmt.Errorf("failed to unmarshal query response for expense categories: %w", err)
		}

		categories = append(categories, resCategories...)
	}

	return categories, nil
}
