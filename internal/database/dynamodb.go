package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func CreateDynamoDBClient() *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())

	if err != nil {
		log.Fatalf("Unable to load SDK config, %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)
	return client
}

func CreateLocalTable() (*types.TableDescription, error) {
	client := CreateDynamoDBClient()
	tableName := os.Getenv("DDB_TABLE_NAME")

	var tableDesc *types.TableDescription

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

	table, err := client.CreateTable(context.TODO(), tableInput)
	if err != nil {
		log.Printf("Couldn't create table %v: %v\n", tableName, err)
		return nil, err
	}

	waiter := dynamodb.NewTableExistsWaiter(client)
	err = waiter.Wait(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}, 5*time.Minute)
	if err != nil {
		log.Printf("Wait for table exists failed: %v\n", err)
		return nil, err
	}

	tableDesc = table.TableDescription
	return tableDesc, nil
}
