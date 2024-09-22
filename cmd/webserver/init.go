package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"

	"github.com/kkstas/tjener/internal/database"
	"github.com/kkstas/tjener/internal/model/expense"
	"github.com/kkstas/tjener/internal/model/expensecategory"
	"github.com/kkstas/tjener/internal/model/user"
	"github.com/kkstas/tjener/internal/server"
)

func initApplicationAndDDB() *server.Application {
	initLogger()

	tableName := os.Getenv("DDB_TABLE_NAME")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := database.CreateDynamoDBClient(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("creating DDB client failed")
		os.Exit(1)
	}

	createDDBTableIfNotExists(ctx, client, tableName)

	expenseStore := expense.NewDDBStore(tableName, client)
	expenseCategoryStore := expensecategory.NewDDBStore(tableName, client)
	userStore := &user.InMemoryStore{}

	newApp := server.NewApplication(expenseStore, expenseCategoryStore, userStore)
	newApp.Handler = loggingMiddleware(newApp.Handler)

	return newApp
}

func createDDBTableIfNotExists(ctx context.Context, client *dynamodb.Client, tableName string) {
	exists, err := database.DDBTableExists(ctx, client, tableName)
	if err != nil {
		log.Fatal().Err(err).Msg("checking if DDB table exists failed")
		os.Exit(1)
	}
	if exists {
		log.Info().Msgf("DynamoDB table '%s' exists", tableName)
		return
	}

	log.Printf("DynamoDB table '%s' does not exist. Creating...", tableName)
	if err := database.CreateDDBTable(ctx, client, tableName); err != nil {
		log.Fatal().Err(err).Msg("creating DynamoDB table failed")
		os.Exit(1)
	}
	log.Info().Msgf("DynamoDB table '%s' created successfully", tableName)
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)

		log.Info().
			Str("method", r.Method).
			Int("status", lrw.statusCode).
			Str("uri", r.RequestURI).
			Str("duration", time.Since(start).String()).
			Str("remote_addr", r.RemoteAddr).
			Msg("HTTP request processed")
	})
}

func initLogger() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
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
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}
}
