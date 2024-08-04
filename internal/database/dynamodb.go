package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func CreateDynamoDBClient(ctx context.Context) *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)
	return client
}

func DDBTableExists(ctx context.Context, client *dynamodb.Client, tableName string) (bool, error) {
	_, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: &tableName,
	})

	if err != nil {
		var notFoundErr *types.ResourceNotFoundException
		if ok := errors.As(err, &notFoundErr); !ok {
			return false, fmt.Errorf("failed to describe table %s: %v\n", tableName, err)
		}
		return false, nil
	}

	return true, nil
}

func CreateDDBTable(ctx context.Context, client *dynamodb.Client, tableName string) error {
	tableInput := &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("SK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("PK"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("SK"),
				KeyType:       types.KeyTypeRange,
			},
		},
		TableName: aws.String(tableName),
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
	}

	_, err := client.CreateTable(ctx, tableInput)
	if err != nil {
		return fmt.Errorf("could not create table %v: %w", tableName, err)
	}

	waiter := dynamodb.NewTableExistsWaiter(client)
	err = waiter.Wait(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("wait for table exists failed: %w", err)
	}
	return nil
}
