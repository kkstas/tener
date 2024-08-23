package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"

	"github.com/kkstas/tjener/internal/database"
	"github.com/kkstas/tjener/internal/model"
	"github.com/kkstas/tjener/internal/server"
)

func initApplication() *server.Application {
	initLogger()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := database.CreateDynamoDBClient(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("creating DDB client failed")
	}

	tableName := os.Getenv("DDB_TABLE_NAME")

	exists, err := database.DDBTableExists(ctx, client, tableName)
	if err != nil {
		log.Fatal().Err(err).Msg("checking if DDB table exists failed")
	}
	if !exists {
		log.Fatal().Err(err).Msgf("DynamoDB table %q not found", tableName)
	}

	expenseStore := model.NewExpenseDDBStore(tableName, client)
	expenseCategoryStore := model.NewExpenseCategoryStore(tableName, client)

	return server.NewApplication(expenseStore, expenseCategoryStore)
}

func initLogger() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	level := strings.ToLower(os.Getenv("LOG_LEVEL"))

	switch level {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}
}
