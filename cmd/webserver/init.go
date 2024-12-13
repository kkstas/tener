package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/kkstas/tener/internal/database"
	"github.com/kkstas/tener/internal/model/expense"
	"github.com/kkstas/tener/internal/model/expensecategory"
	"github.com/kkstas/tener/internal/model/user"
	"github.com/kkstas/tener/internal/server"
)

func run(ctx context.Context, w io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	app, err := initApplicationAndDDB(ctx, w)
	if err != nil {
		return fmt.Errorf("failed to initialize application and ddb: %w", err)
	}

	server := &http.Server{
		Addr:              ":8081",
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           app,
	}

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to ListenAndServe: %w", err)
	}

	return nil
}

func initApplicationAndDDB(ctx context.Context, w io.Writer) (*server.Application, error) {
	logger := initLogger(w)

	tableName := os.Getenv("DDB_TABLE_NAME")

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	client, err := database.CreateDynamoDBClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating DDB client failed: %w", err)
	}

	err = createDDBTableIfNotExists(ctx, logger, client, tableName)
	if err != nil {
		return nil, fmt.Errorf("application init failed: %w", err)
	}

	expenseStore := expense.NewDDBStore(tableName, client)
	expenseCategoryStore := expensecategory.NewDDBStore(tableName, client)
	userStore := user.NewDDBStore(tableName, client)

	newApp := server.NewApplication(logger, expenseStore, expenseCategoryStore, userStore)
	return newApp, nil
}

func createDDBTableIfNotExists(ctx context.Context, logger *slog.Logger, client *dynamodb.Client, tableName string) error {
	exists, err := database.DDBTableExists(ctx, client, tableName)
	if err != nil {
		return fmt.Errorf("checking if DDB table exists failed: %w", err)
	}
	if exists {
		logger.Info("DynamoDB table exists", "tableName", tableName)
		return nil
	}

	logger.Info("DynamoDB table '%s' does not exist. Creating...", "tableName", tableName)
	if err := database.CreateDDBTable(ctx, client, tableName); err != nil {
		return fmt.Errorf("creating DynamoDB table failed: %w", err)
	}
	logger.Info("DynamoDB table created successfully", "tableName", tableName)
	return nil
}

func initLogger(w io.Writer) *slog.Logger {
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
		w,
		&slog.HandlerOptions{Level: level},
	))
}
