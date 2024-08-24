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

const expenseCategoryPK = "expensecategory"

type ExpenseCategoryStore struct {
	client    *dynamodb.Client
	tableName string
}

func (c *ExpenseCategory) getKey() map[string]types.AttributeValue {
	PK, err := attributevalue.Marshal(c.PK)
	if err != nil {
		panic(err)
	}
	SK, err := attributevalue.Marshal(c.Name)
	if err != nil {
		panic(err)
	}
	return map[string]types.AttributeValue{"PK": PK, "SK": SK}
}

func NewExpenseCategoryStore(tableName string, client *dynamodb.Client) *ExpenseCategoryStore {
	return &ExpenseCategoryStore{
		tableName: tableName,
		client:    client,
	}
}

func (cs *ExpenseCategoryStore) Create(ctx context.Context, categoryFC ExpenseCategory) error {
	item, err := attributevalue.MarshalMap(
		ExpenseCategory{
			PK:   expenseCategoryPK,
			Name: categoryFC.Name,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to marshal expense category: %w", err)
	}

	_, err = cs.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &cs.tableName,
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(SK)"),
	})
	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return &ExpenseCategoryAlreadyExistsError{Name: categoryFC.Name}
		}

		return fmt.Errorf("failed to put expense category into DynamoDB: %w", err)
	}

	return nil
}

func (cs *ExpenseCategoryStore) Delete(ctx context.Context, name string) error {
	categoryFD := ExpenseCategory{PK: expenseCategoryPK, Name: name}
	_, err := cs.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &cs.tableName,
		Key:       categoryFD.getKey(),
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
