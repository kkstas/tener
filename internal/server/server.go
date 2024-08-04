package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/kkstas/tjener/internal/components"
	"github.com/kkstas/tjener/internal/model"
	"github.com/kkstas/tjener/static"
)

type Application struct {
	ddb          *dynamodb.Client
	expenseStore *model.ExpenseStore
	http.Handler
}

func NewApplication(ddb *dynamodb.Client, tableName string) *Application {
	app := new(Application)
	app.ddb = ddb

	app.expenseStore = model.NewExpenseStore(tableName, app.ddb)

	mux := http.NewServeMux()

	mux.Handle("GET /health-check", http.HandlerFunc(app.healthCheck))
	mux.Handle("GET /home", templ.Handler(components.Page()))
	mux.Handle("/", http.HandlerFunc(app.notFound))

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.Static))))

	mux.Handle("POST /put-item", http.HandlerFunc(app.putItem))
	mux.Handle("GET /query", http.HandlerFunc(app.queryItems))

	app.Handler = mux

	return app
}

func (app *Application) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (app *Application) notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (app *Application) putItem(w http.ResponseWriter, r *http.Request) {
	err := app.expenseStore.PutItem(r.Context(), model.Expense{
		Name:     "some expense",
		Category: "junk food",
		Amount:   2932.42,
	})

	if err != nil {
		fmt.Printf("error while put item %#v\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusCreated)
}

func (app *Application) queryItems(w http.ResponseWriter, r *http.Request) {
	expenses, err := app.expenseStore.Query(r.Context())
	if err != nil {
		fmt.Printf("error while query items %v\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
	err = json.NewEncoder(w).Encode(expenses)
	if err != nil {
		fmt.Printf("error while encoding %v\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

}
