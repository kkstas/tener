package database

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func CreateLocalTestDDBTable(ctx context.Context) (string, *dynamodb.Client, func(), error) {
	client, err := CreateLocalDynamoDBClient(ctx)
	if err != nil {
		return "", nil, nil, err
	}

	tableName := randomString(16)
	if err := CreateDDBTable(ctx, client, tableName); err != nil {
		return tableName, nil, nil, fmt.Errorf("error while creating DDB table: %w", err)
	}

	removeDDB := func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		err := deleteDDBTable(cleanupCtx, client, tableName)
		if err != nil {
			log.Fatalf("failed to delete table %q: %v", tableName, err)
		}
	}

	return tableName, client, removeDDB, nil
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func deleteDDBTable(ctx context.Context, client *dynamodb.Client, tableName string) error {
	exists, err := DDBTableExists(ctx, client, tableName)

	if err != nil {
		return fmt.Errorf("checking if DDB table exists failed: %w", err)
	}

	if !exists {
		return fmt.Errorf("DynamoDB table %q does not exist, nothing to delete", tableName)
	}

	_, err = client.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return fmt.Errorf("couldn't delete table %q: %w", tableName, err)
	}

	waiter := dynamodb.NewTableNotExistsWaiter(client)
	err = waiter.Wait(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("wait for table deletion failed: %w", err)
	}
	return nil
}

func CreateLocalDynamoDBClient(ctx context.Context) (*dynamodb.Client, error) {
	customEndpoint := "http://localhost:8000"
	region := "local"

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("x", "y", "z")),
	)
	if err != nil {
		return nil, err
	}

	options := dynamodb.Options{
		EndpointResolver: dynamodb.EndpointResolverFromURL(customEndpoint),
		Credentials:      cfg.Credentials,
		Region:           cfg.Region,
	}

	client := dynamodb.New(options)
	return client, nil
}
