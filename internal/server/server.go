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
	mux.Handle("GET /home", http.HandlerFunc(app.homeHandler))
	mux.Handle("/", http.HandlerFunc(app.notFound))

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.Static))))

	mux.Handle("POST /expense/create", http.HandlerFunc(app.postCreateExpense))
	mux.Handle("GET /expense/query", http.HandlerFunc(app.queryItems))

	mux.Handle("GET /expense/{PK}/{SK}", http.HandlerFunc(app.getExpense))
	mux.Handle("GET /expense/edit/{PK}/{SK}", http.HandlerFunc(app.editSingleExpenseTemplate))
	mux.Handle("PUT /expense/edit/{PK}/{SK}", http.HandlerFunc(app.editSingleExpense))

	app.Handler = mux

	return app
}

func (app *Application) editSingleExpense(w http.ResponseWriter, r *http.Request) {
	pk := r.PathValue("PK")
	sk := r.PathValue("SK")

	if pk == "" || sk == "" {
		writeErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid PK '%s' or SK '%s'", pk, sk))
		return
	}
	category := r.FormValue("category")
	currency := r.FormValue("currency")
	name := r.FormValue("name")
	amountRaw := r.FormValue("amount")
	amount, err := strconv.ParseFloat(amountRaw, 64)

	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid amount value")
		return
	}
	expense, err := app.expenseStore.UpdateExpense(r.Context(), pk, sk, name, category, amount, currency)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("error while putting item %v", err.Error()))
		return
	}

	renderTempl(w, r, components.SingleExpense(expense))
}

func (app *Application) editSingleExpenseTemplate(w http.ResponseWriter, r *http.Request) {
	pk := r.PathValue("PK")
	sk := r.PathValue("SK")

	if pk == "" || sk == "" {
		writeErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid PK '%s' or SK '%s'", pk, sk))
		return
	}

	expense, found, err := app.expenseStore.GetExpense(r.Context(), pk, sk)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("error while getting expense %v", err.Error()))
		return
	}
	if !found {
		writeErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("no expense found for PK: %s SK: %s", pk, sk))
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Add("content-type", "application/json")
		err = json.NewEncoder(w).Encode(expense)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	renderTempl(w, r, components.EditSingleExpense(expense))
}

func (app *Application) getExpense(w http.ResponseWriter, r *http.Request) {
	pk := r.PathValue("PK")
	sk := r.PathValue("SK")

	if pk == "" || sk == "" {
		writeErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid PK '%s' or SK '%s'", pk, sk))
		return
	}

	expense, found, err := app.expenseStore.GetExpense(r.Context(), pk, sk)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("error while getting expense %v", err.Error()))
		return
	}
	if !found {
		writeErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("no expense found for PK: %s SK: %s", pk, sk))
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Add("content-type", "application/json")
		err = json.NewEncoder(w).Encode(expense)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	renderTempl(w, r, components.SingleExpense(expense))
}

func (app *Application) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (app *Application) notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (app *Application) postCreateExpense(w http.ResponseWriter, r *http.Request) {
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

	expenses, err := app.expenseStore.Query(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("error while querying for items %v", err.Error()))
		return
	}

	renderTempl(w, r, components.ExpensesContainer(expenses))
}

func (app *Application) homeHandler(w http.ResponseWriter, r *http.Request) {
	expenses, err := app.expenseStore.Query(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("error while querying for items %v", err.Error()))
		return
	}

	renderTempl(w, r, components.Page(expenses))
}

func (app *Application) queryItems(w http.ResponseWriter, r *http.Request) {
	expenses, err := app.expenseStore.Query(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("error while querying for items %v", err.Error()))
		return
	}

	w.Header().Add("content-type", "application/json")
	err = json.NewEncoder(w).Encode(expenses)

	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("error while encoding %v", err.Error()))
		return
	}
}

func renderTempl(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html")
	err := component.Render(r.Context(), w)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("error while generating template %v", err.Error()))
		return
	}
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `{"message":%q}`, message)
}
