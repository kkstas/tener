package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/kkstas/tjener/internal/database"
	"github.com/kkstas/tjener/internal/server"
)

func initApplication() *server.Application {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := database.CreateDynamoDBClient(ctx)
	tableName := os.Getenv("DDB_TABLE_NAME")

	exists, err := database.DDBTableExists(ctx, client, tableName)
	if err != nil {
		log.Fatalf("checking if DDB table exists failed: %#v", err)
	}
	if !exists {
		log.Fatalf("DynamoDB table %q not found", tableName)
	}

	return server.NewApplication(client, tableName)
}
