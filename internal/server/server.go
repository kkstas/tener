package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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

	mux.Handle("POST /expense/create", http.HandlerFunc(app.putItem))
	mux.Handle("GET /expense/query", http.HandlerFunc(app.queryItems))

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
	category := r.FormValue("category")
	currency := r.FormValue("currency")
	name := r.FormValue("name")
	amountRaw := r.FormValue("amount")
	amount, err := strconv.ParseFloat(amountRaw, 64)

	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid amount value")
		return
	}

	expense, err := model.CreateExpense(name, category, amount, currency)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	err = app.expenseStore.PutItem(r.Context(), expense)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("error while putting item %v", err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (app *Application) queryItems(w http.ResponseWriter, r *http.Request) {
	expenses, err := app.expenseStore.Query(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("error while query items %v\n", err.Error()))
		return
	}

	w.Header().Add("content-type", "application/json")
	err = json.NewEncoder(w).Encode(expenses)

	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("error while encoding %v", err.Error()))
		return
	}
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `{"message":%q}`, message)
}
