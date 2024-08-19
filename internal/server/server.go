package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/rs/zerolog/log"

	"github.com/kkstas/tjener/internal/components"
	"github.com/kkstas/tjener/internal/model"
	"github.com/kkstas/tjener/internal/url"
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

	mux.HandleFunc("/", app.notFoundHandler)
	mux.HandleFunc("GET /health-check", app.healthCheck)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		http.FileServer(http.FS(static.Css)).ServeHTTP(w, r)
	})))

	mux.HandleFunc("GET /home", app.homeHandler)

	mux.HandleFunc("GET /expense/create", app.createExpensePage)
	mux.HandleFunc("POST /expense/create", app.createExpense)
	mux.HandleFunc("GET /expense/{SK}", app.showExpense)
	mux.HandleFunc("GET /expense/edit/{SK}", app.showEditableExpense)
	mux.HandleFunc("PUT /expense/edit/{SK}", app.updateExpense)
	mux.HandleFunc("DELETE /expense/{SK}", app.deleteExpense)

	app.Handler = secureHeaders(mux)

	return app
}

func (app *Application) deleteExpense(w http.ResponseWriter, r *http.Request) {
	sk := r.PathValue("SK")

	err := app.expenseStore.DeleteExpense(r.Context(), sk)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error while deleting item:"+err.Error(), err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (app *Application) updateExpense(w http.ResponseWriter, r *http.Request) {
	sk := r.PathValue("SK")

	category := r.FormValue("category")
	currency := r.FormValue("currency")
	name := r.FormValue("name")
	amountRaw := r.FormValue("amount")
	amount, err := strconv.ParseFloat(amountRaw, 64)

	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "invalid amount value", err)
		return
	}
	expense, err := app.expenseStore.UpdateExpense(r.Context(), sk, name, category, amount, currency)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error while putting item:"+err.Error(), err)
		return
	}

	app.renderTempl(w, r, components.SingleExpense(r.Context(), expense))
}

func (app *Application) showEditableExpense(w http.ResponseWriter, r *http.Request) {
	sk := r.PathValue("SK")

	expense, found, err := app.expenseStore.GetExpense(r.Context(), sk)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error while getting expense:"+err.Error(), err)
		return
	}
	if !found {
		sendErrorResponse(w, http.StatusBadRequest, "no expense found for SK:"+sk, err)
		return
	}

	app.renderTempl(w, r, components.EditSingleExpense(r.Context(), expense))
}

func (app *Application) showExpense(w http.ResponseWriter, r *http.Request) {
	sk := r.PathValue("SK")

	expense, found, err := app.expenseStore.GetExpense(r.Context(), sk)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error while getting expense:"+err.Error(), err)
		return
	}
	if !found {
		sendErrorResponse(w, http.StatusNotFound, "no expense found for SK:"+sk, err)
		return
	}

	app.renderTempl(w, r, components.SingleExpense(r.Context(), expense))
}

func (app *Application) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (app *Application) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (app *Application) createExpensePage(w http.ResponseWriter, r *http.Request) {
	app.renderTempl(w, r, components.CreateExpensePage(r.Context(), model.ValidCurrencies))
}

func (app *Application) createExpense(w http.ResponseWriter, r *http.Request) {
	category := r.FormValue("category")
	currency := r.FormValue("currency")
	name := r.FormValue("name")
	amountRaw := r.FormValue("amount")
	amount, err := strconv.ParseFloat(amountRaw, 64)

	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "invalid amount value", err)
		return
	}

	expense, err := model.NewExpense(name, category, amount, currency)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	err = app.expenseStore.PutItem(r.Context(), expense)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to put item:"+err.Error(), err)
		return
	}

	http.Redirect(w, r, url.Create(r.Context(), "home"), http.StatusSeeOther)
}

func (app *Application) homeHandler(w http.ResponseWriter, r *http.Request) {
	expenses, err := app.expenseStore.Query(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to query items:"+err.Error(), err)
		return
	}

	app.renderTempl(w, r, components.Page(r.Context(), expenses))
}

func (app *Application) renderTempl(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html")

	if err := component.Render(r.Context(), w); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error while generating template:"+err.Error(), err)
		return
	}
}

func sendErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	log.Error().Stack().Err(err).Msg("invalid amount value")
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `{"message":%q}`, message)
}
