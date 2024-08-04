package main

import (
	"context"
	"os"

	"github.com/kkstas/tjener/internal/database"
	"github.com/kkstas/tjener/internal/server"
)

func initApplicationAndDDB() *server.Application {
	ddbTableName := os.Getenv("DDB_TABLE_NAME")
	ctx := context.Background()
	client := database.CreateDynamoDBClient(ctx)
	database.CreateDDBTableIfNotExists(ctx, client, ddbTableName)
	return server.NewApplication(client, ddbTableName)
}
