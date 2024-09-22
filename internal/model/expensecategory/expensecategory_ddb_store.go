package expensecategory

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

const PK = "expensecategory"

type DDBStore struct {
	client    *dynamodb.Client
	tableName string
}

func (c *Category) getKey() map[string]types.AttributeValue {
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

func NewDDBStore(tableName string, client *dynamodb.Client) *DDBStore {
	return &DDBStore{
		tableName: tableName,
		client:    client,
	}
}

func (cs *DDBStore) Create(ctx context.Context, categoryFC Category) error {
	item, err := attributevalue.MarshalMap(
		Category{
			PK:   PK,
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
			return &AlreadyExistsError{Name: categoryFC.Name}
		}

		return fmt.Errorf("failed to put expense category into DynamoDB: %w", err)
	}

	return nil
}

func (cs *DDBStore) Delete(ctx context.Context, name string) error {
	categoryFD := Category{PK: PK, Name: name}
	_, err := cs.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &cs.tableName,
		Key:       categoryFD.getKey(),
	})
	if err != nil {
		return fmt.Errorf("failed to delete expense category with name=%q from the table: %w", name, err)
	}

	return nil
}

func (cs *DDBStore) Query(ctx context.Context) ([]Category, error) {
	keyCond := expression.
		Key("PK").Equal(expression.Value(PK))

	exprBuilder := expression.NewBuilder()
	exprBuilder.WithKeyCondition(keyCond)

	expr, err := expression.NewBuilder().
		WithKeyCondition(keyCond).
		Build()

	if err != nil {
		return nil, fmt.Errorf("failed to build expression for expense category query %w", err)
	}

	return cs.query(ctx, expr)

}

func (cs *DDBStore) query(ctx context.Context, expr expression.Expression) ([]Category, error) {
	var categories []Category

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

		var resCategories []Category
		err = attributevalue.UnmarshalListOfMaps(response.Items, &resCategories)
		if err != nil {
			return categories, fmt.Errorf("failed to unmarshal query response for expense categories: %w", err)
		}

		categories = append(categories, resCategories...)
	}

	return categories, nil
}
