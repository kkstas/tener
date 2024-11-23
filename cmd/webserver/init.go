package main

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/kkstas/tener/internal/database"
	"github.com/kkstas/tener/internal/model/expense"
	"github.com/kkstas/tener/internal/model/expensecategory"
	"github.com/kkstas/tener/internal/model/user"
	"github.com/kkstas/tener/internal/server"
)

func initApplicationAndDDB() *server.Application {
	logger := initLogger()

	tableName := os.Getenv("DDB_TABLE_NAME")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := database.CreateDynamoDBClient(ctx)
	if err != nil {
		logger.Error("creating DDB client failed", "error", err)
		os.Exit(1)
	}

	createDDBTableIfNotExists(ctx, logger, client, tableName)

	expenseStore := expense.NewDDBStore(tableName, client)
	expenseCategoryStore := expensecategory.NewDDBStore(tableName, client)
	userStore := user.NewDDBStore(tableName, client)

	newApp := server.NewApplication(logger, expenseStore, expenseCategoryStore, userStore)
	return newApp
}

func createDDBTableIfNotExists(ctx context.Context, logger *slog.Logger, client *dynamodb.Client, tableName string) {
	exists, err := database.DDBTableExists(ctx, client, tableName)
	if err != nil {
		logger.Error("checking if DDB table exists failed", "error", err)
		os.Exit(1)
	}
	if exists {
		logger.Info("DynamoDB table exists", "tableName", tableName)
		return
	}

	logger.Info("DynamoDB table '%s' does not exist. Creating...", "tableName", tableName)
	if err := database.CreateDDBTable(ctx, client, tableName); err != nil {
		logger.Error("creating DynamoDB table failed", "error", err)
		os.Exit(1)
	}
	logger.Info("DynamoDB table created successfully", "tableName", tableName)
}

func initLogger() *slog.Logger {
	envLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	var level slog.Level

	switch envLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelDebug
	}

	return slog.New(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{Level: level},
	))
}
