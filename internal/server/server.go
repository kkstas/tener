package server

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/kkstas/tjener/internal/components"
	"github.com/kkstas/tjener/internal/database"
	"github.com/kkstas/tjener/static"
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
	mux.Handle("GET /app", templ.Handler(components.Page()))
	mux.Handle("/", http.HandlerFunc(notFound))

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.Static))))

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
