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

const pkPrefix = "expensecategory"

type DDBStore struct {
	client    *dynamodb.Client
	tableName string
}

func (c *Category) getKey(vaultID string) map[string]types.AttributeValue {
	PK, err := attributevalue.Marshal(buildPK(vaultID))
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

func (cs *DDBStore) Create(ctx context.Context, categoryFC Category, userID, vaultID string) error {
	pk := buildPK(vaultID)
	item, err := attributevalue.MarshalMap(
		Category{
			PK:        pk,
			Name:      categoryFC.Name,
			CreatedBy: userID,
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
			return &AlreadyExistsError{PK: pk, Name: categoryFC.Name}
		}

		return fmt.Errorf("failed to put expense category into DynamoDB: %w", err)
	}

	return nil
}

func (cs *DDBStore) Delete(ctx context.Context, name, vaultID string) error {
	categoryFD := Category{Name: name}
	_, err := cs.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &cs.tableName,
		Key:       categoryFD.getKey(vaultID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete expense category with PK='%s' name='%s' from table: %w", buildPK(vaultID), name, err)
	}

	return nil
}

func (cs *DDBStore) FindAll(ctx context.Context, vaultID string) ([]Category, error) {
	pk := buildPK(vaultID)
	keyCond := expression.Key("PK").Equal(expression.Value(pk))

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
