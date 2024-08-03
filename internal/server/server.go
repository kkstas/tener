package server

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/kkstas/tjener/internal/components"
	"github.com/kkstas/tjener/static"
)

type Application struct {
	ddb *dynamodb.Client
	http.Handler
}

func NewApplication(ddb *dynamodb.Client) *Application {
	app := new(Application)
	app.ddb = ddb
	mux := http.NewServeMux()

	mux.Handle("GET /health-check", http.HandlerFunc(healthCheck))
	mux.Handle("GET /home", templ.Handler(components.Page()))
	mux.Handle("/", http.HandlerFunc(notFound))

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.Static))))

	app.Handler = mux

	return app
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
