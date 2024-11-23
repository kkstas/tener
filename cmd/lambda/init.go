package main

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/kkstas/tener/internal/database"
	"github.com/kkstas/tener/internal/model/expense"
	"github.com/kkstas/tener/internal/model/expensecategory"
	"github.com/kkstas/tener/internal/model/user"
	"github.com/kkstas/tener/internal/server"
)

func initApplication() *server.Application {
	logger := initLogger()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := database.CreateDynamoDBClient(ctx)
	if err != nil {
		logger.Error("creating DDB client failed", "error", err)
	}

	tableName := os.Getenv("DDB_TABLE_NAME")

	exists, err := database.DDBTableExists(ctx, client, tableName)
	if err != nil {
		logger.Error("checking if DDB table exists failed", "error", err)
		os.Exit(1)
	}
	if !exists {
		logger.Error("DynamoDB table not found", "tableName", tableName, "error", err)
		os.Exit(1)
	}

	expenseStore := expense.NewDDBStore(tableName, client)
	expenseCategoryStore := expensecategory.NewDDBStore(tableName, client)
	userStore := user.NewDDBStore(tableName, client)

	return server.NewApplication(logger, expenseStore, expenseCategoryStore, userStore)
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
		level = slog.LevelInfo
	}

	return slog.New(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{Level: level},
	))
}
