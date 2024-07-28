package server

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/kkstas/tjener/internal/database"
)

type Application struct {
	ddb *dynamodb.Client
	http.Handler
}

func NewApplication() *Application {
	app := new(Application)
	app.ddb = database.CreateDynamoDBClient()

	mux := http.NewServeMux()

	mux.Handle("GET /hello", http.HandlerFunc(handlerGetHello))
	mux.Handle("POST /hello", http.HandlerFunc(handlerPostHello))
	mux.Handle("/", http.HandlerFunc(notFound))

	app.Handler = mux

	return app

}

func handlerGetHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "invoked GET /hello")
}

func handlerPostHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "invoked POST hello")
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
