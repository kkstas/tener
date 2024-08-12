package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/kkstas/tjener/internal/database"
	"github.com/kkstas/tjener/internal/server"
)

func initApplicationAndDDB() *server.Application {
	tableName := os.Getenv("DDB_TABLE_NAME")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := database.CreateDynamoDBClient(ctx)
	if err != nil {
		log.Fatalf("creating DDB client failed: %v", err)
	}

	createDDBTableIfNotExists(ctx, client, tableName)
	return server.NewApplication(client, tableName)
}

func createDDBTableIfNotExists(ctx context.Context, client *dynamodb.Client, tableName string) {
	exists, err := database.DDBTableExists(ctx, client, tableName)
	if err != nil {
		log.Fatalf("checking if DDB table exists failed: %#v", err)
	}
	if exists {
		fmt.Printf("DynamoDB table %q exists.\n", tableName)
		return
	}

	fmt.Printf("DynamoDB table %q does not exist. Creating...\n", tableName)
	if err := database.CreateDDBTable(ctx, client, tableName); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("DynamoDB table %q created successfully.\n", tableName)
}
