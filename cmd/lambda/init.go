package main

import (
	"context"
	"log"
	"os"

	"github.com/kkstas/tjener/internal/database"
	"github.com/kkstas/tjener/internal/server"
)

func initApplication() *server.Application {
	client := database.CreateDynamoDBClient(context.Background())
	tableName := os.Getenv("DDB_TABLE_NAME")

	if !database.DDBTableExists(client, tableName) {
		log.Fatalf("DynamoDB table %q not found", tableName)
	}

	return server.NewApplication(client, tableName)
}
