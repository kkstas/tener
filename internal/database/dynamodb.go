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

func DDBTableExists(client *dynamodb.Client, tableName string) bool {
	_, err := client.DescribeTable(context.Background(), &dynamodb.DescribeTableInput{
		TableName: &tableName,
	})

	if err != nil {
		var notFoundErr *types.ResourceNotFoundException
		if ok := errors.As(err, &notFoundErr); !ok {
			log.Fatalf("failed to describe table %s: %v\n", tableName, err)
		}
		return false
	}

	return true
}

func CreateDDBTableIfNotExists(ctx context.Context, client *dynamodb.Client, tableName string) {
	if DDBTableExists(client, tableName) {
		fmt.Printf("DynamoDB table %q exists.\n", tableName)
		return
	}
	fmt.Printf("DynamoDB table %q does not exist. Creating...\n", tableName)

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
		fmt.Printf("Couldn't create table %v: %v\n", tableName, err)
		return
	}

	waiter := dynamodb.NewTableExistsWaiter(client)
	err = waiter.Wait(context.Background(), &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}, 5*time.Minute)
	if err != nil {
		log.Fatalf("wait for table exists failed: %v\n", err)
	}

	fmt.Printf("Table %q created successfully.\n", tableName)
}
