package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/kkstas/tener/internal/database"
	"github.com/kkstas/tener/internal/model/expense"
	"github.com/kkstas/tener/internal/model/expensecategory"
	"github.com/kkstas/tener/internal/model/user"
	"github.com/kkstas/tener/internal/server"
)

func run(ctx context.Context, w io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	app, err := initApplication(ctx, w)
	if err != nil {
		return err
	}

	lambda.Start(httpadapter.New(app).ProxyWithContext)
	return nil
}

func initApplication(ctx context.Context, w io.Writer) (*server.Application, error) {
	logger := initLogger(w)

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	client, err := database.CreateDynamoDBClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating DDB client failed: %w", err)
	}

	tableName := os.Getenv("DDB_TABLE_NAME")

	exists, err := database.DDBTableExists(ctx, client, tableName)
	if err != nil {
		return nil, fmt.Errorf("checking if DDB table exists failed: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("DynamoDB table %q not found", tableName)
	}

	expenseStore := expense.NewDDBStore(tableName, client)
	expenseCategoryStore := expensecategory.NewDDBStore(tableName, client)
	userStore := user.NewDDBStore(tableName, client)

	return server.NewApplication(logger, expenseStore, expenseCategoryStore, userStore), nil
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
		level = slog.LevelInfo
	}

	return slog.New(slog.NewJSONHandler(
		w,
		&slog.HandlerOptions{Level: level},
	))
}
