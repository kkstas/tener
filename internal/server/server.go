package server

import (
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
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.Static))))
	mux.Handle("/", http.HandlerFunc(app.notFoundHandler))

	mux.Handle("GET /home", http.HandlerFunc(app.homeHandler))

	mux.Handle("POST /expense/create", http.HandlerFunc(app.createExpense))

	mux.Handle("GET /expense/{PK}/{SK}", http.HandlerFunc(app.showExpense))
	mux.Handle("GET /expense/edit/{PK}/{SK}", http.HandlerFunc(app.showEditableExpense))
	mux.Handle("PUT /expense/edit/{PK}/{SK}", http.HandlerFunc(app.updateExpense))

	app.Handler = mux

	return app
}

func (app *Application) updateExpense(w http.ResponseWriter, r *http.Request) {
	pk := r.PathValue("PK")
	sk := r.PathValue("SK")

	if pk == "" || sk == "" {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid PK '%s' or SK '%s'", pk, sk))
		return
	}
	category := r.FormValue("category")
	currency := r.FormValue("currency")
	name := r.FormValue("name")
	amountRaw := r.FormValue("amount")
	amount, err := strconv.ParseFloat(amountRaw, 64)

	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "invalid amount value")
		return
	}
	expense, err := app.expenseStore.UpdateExpense(r.Context(), pk, sk, name, category, amount, currency)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error while putting item:"+err.Error())
		return
	}

	renderTempl(w, r, components.SingleExpense(expense))
}

func (app *Application) showEditableExpense(w http.ResponseWriter, r *http.Request) {
	pk := r.PathValue("PK")
	sk := r.PathValue("SK")

	if pk == "" || sk == "" {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid PK '%s' or SK '%s'", pk, sk))
		return
	}

	expense, found, err := app.expenseStore.GetExpense(r.Context(), pk, sk)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error while getting expense:"+err.Error())
		return
	}
	if !found {
		sendErrorResponse(w, http.StatusBadRequest, "no expense found for PK:"+pk+"SK:"+sk)
		return
	}

	renderTempl(w, r, components.EditSingleExpense(expense))
}

func (app *Application) showExpense(w http.ResponseWriter, r *http.Request) {
	pk := r.PathValue("PK")
	sk := r.PathValue("SK")

	if pk == "" || sk == "" {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid PK '%s' or SK '%s'", pk, sk))
		return
	}

	expense, found, err := app.expenseStore.GetExpense(r.Context(), pk, sk)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error while getting expense:"+err.Error())
		return
	}
	if !found {
		sendErrorResponse(w, http.StatusNotFound, fmt.Sprintf("no expense found for PK: %s & SK: %s", pk, sk))
		return
	}

	renderTempl(w, r, components.SingleExpense(expense))
}

func (app *Application) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (app *Application) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (app *Application) createExpense(w http.ResponseWriter, r *http.Request) {
	category := r.FormValue("category")
	currency := r.FormValue("currency")
	name := r.FormValue("name")
	amountRaw := r.FormValue("amount")
	amount, err := strconv.ParseFloat(amountRaw, 64)

	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "invalid amount value")
		return
	}

	expense, err := model.NewExpense(name, category, amount, currency)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	err = app.expenseStore.PutItem(r.Context(), expense)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to put item:"+err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)

	expenses, err := app.expenseStore.Query(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to while query for items:"+err.Error())
		return
	}

	renderTempl(w, r, components.ExpensesContainer(expenses))
}

func (app *Application) homeHandler(w http.ResponseWriter, r *http.Request) {
	expenses, err := app.expenseStore.Query(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to query items:"+err.Error())
		return
	}

	renderTempl(w, r, components.Page(expenses))
}

func renderTempl(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html")
	if err := component.Render(r.Context(), w); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error while generating template:"+err.Error())
		return
	}
}

func sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `{"message":%q}`, message)
}
