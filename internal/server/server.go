package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/rs/zerolog/log"

	"github.com/kkstas/tjener/assets"
	"github.com/kkstas/tjener/internal/components"
	"github.com/kkstas/tjener/internal/model"
	"github.com/kkstas/tjener/internal/url"
	"github.com/kkstas/tjener/pkg/validator"
)

type ExpenseStore interface {
	Create(ctx context.Context, expenseFC model.Expense) error
	Delete(ctx context.Context, SK string) error
	Update(ctx context.Context, expenseFU model.Expense) (model.Expense, error)
	FindOne(ctx context.Context, SK string) (model.Expense, error)
	Query(ctx context.Context) ([]model.Expense, error)
}

type ExpenseCategoryStore interface {
	Create(ctx context.Context, categoryFC model.ExpenseCategory) error
	Delete(ctx context.Context, name string) error
	Query(ctx context.Context) ([]model.ExpenseCategory, error)
}

type Application struct {
	expense         ExpenseStore
	expenseCategory ExpenseCategoryStore
	http.Handler
}

func NewApplication(expenseStore ExpenseStore, expenseCategoryStore ExpenseCategoryStore) *Application {
	app := new(Application)

	app.expense = expenseStore
	app.expenseCategory = expenseCategoryStore

	mux := http.NewServeMux()

	mux.HandleFunc("/", app.notFoundHandler)
	mux.HandleFunc("GET /health-check", app.healthCheck)
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets.Public))))

	mux.HandleFunc("GET /home", app.homeHandler)

	mux.HandleFunc("GET /expense/{SK}", app.showExpense)
	mux.HandleFunc("GET /expense/edit/{SK}", app.showEditableExpense)
	mux.HandleFunc("PUT /expense/edit/{SK}", app.updateExpense)
	mux.HandleFunc("DELETE /expense/{SK}", app.deleteExpense)

	mux.HandleFunc("GET /expense/create", app.createExpensePage)
	mux.HandleFunc("POST /expense/create", app.createExpense)

	mux.HandleFunc("GET /expensecategories", app.createExpenseCategoryPage)
	mux.HandleFunc("POST /expensecategories/create", app.createExpenseCategory)
	mux.HandleFunc("DELETE /expensecategories/{name}", app.deleteExpenseCategory)

	app.Handler = secureHeaders(mux)

	return app
}

func (app *Application) deleteExpense(w http.ResponseWriter, r *http.Request) {
	sk := r.PathValue("SK")

	err := app.expense.Delete(r.Context(), sk)
	if err != nil {
		var notFoundErr *model.ExpenseNotFoundError
		if errors.As(err, &notFoundErr) {
			sendErrorResponse(w, http.StatusNotFound, err.Error(), err)
			return
		}
		sendErrorResponse(w, http.StatusInternalServerError, "error while deleting item: "+err.Error(), err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (app *Application) deleteExpenseCategory(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	err := app.expenseCategory.Delete(r.Context(), name)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "deleting item failed: "+err.Error(), err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (app *Application) updateExpense(w http.ResponseWriter, r *http.Request) {
	createdAt := r.PathValue("SK")

	category := r.FormValue("category")
	currency := r.FormValue("currency")
	name := r.FormValue("name")
	amountRaw := r.FormValue("amount")
	amount, err := strconv.ParseFloat(amountRaw, 64)

	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "invalid amount value", err)
		return
	}

	expenseFU, err := model.NewExpenseFU(name, createdAt, category, amount, currency)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "", err)
	}

	expense, err := app.expense.Update(r.Context(), expenseFU)
	if err != nil {
		var notFoundErr *model.ExpenseNotFoundError
		if errors.As(err, &notFoundErr) {
			sendErrorResponse(w, http.StatusNotFound, err.Error(), err)
			return
		}
		sendErrorResponse(w, http.StatusInternalServerError, "error while putting item: "+err.Error(), err)
		return
	}

	app.renderTempl(w, r, components.SingleExpense(r.Context(), expense))
}

func (app *Application) showEditableExpense(w http.ResponseWriter, r *http.Request) {
	sk := r.PathValue("SK")

	expense, err := app.expense.FindOne(r.Context(), sk)

	if err != nil {
		var notFoundErr *model.ExpenseNotFoundError
		if errors.As(err, &notFoundErr) {
			sendErrorResponse(w, http.StatusNotFound, err.Error(), err)
			return
		}
		sendErrorResponse(w, http.StatusInternalServerError, "error while getting expense: "+err.Error(), err)
		return
	}

	categories, err := app.expenseCategory.Query(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to query expense categories: "+err.Error(), err)
		return
	}

	app.renderTempl(w, r, components.EditSingleExpense(r.Context(), expense, categories))
}

func (app *Application) showExpense(w http.ResponseWriter, r *http.Request) {
	sk := r.PathValue("SK")

	expense, err := app.expense.FindOne(r.Context(), sk)
	if err != nil {
		var notFoundErr *model.ExpenseNotFoundError
		if errors.As(err, &notFoundErr) {
			sendErrorResponse(w, http.StatusNotFound, err.Error(), err)
			return
		}
		sendErrorResponse(w, http.StatusInternalServerError, "error while getting expense: "+err.Error(), err)
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
	categories, err := app.expenseCategory.Query(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to query expense categories: "+err.Error(), err)
		return
	}

	app.renderTempl(w, r, components.CreateExpensePage(r.Context(), model.ValidCurrencies, categories))
}

func (app *Application) createExpenseCategoryPage(w http.ResponseWriter, r *http.Request) {
	categories, err := app.expenseCategory.Query(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to query expense categories: "+err.Error(), err)
		return
	}

	app.renderTempl(w, r, components.ExpenseCategoriesPage(r.Context(), categories))
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

	expense, err := model.NewExpenseFC(name, category, amount, currency)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	err = app.expense.Create(r.Context(), expense)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to put item: "+err.Error(), err)
		return
	}

	http.Redirect(w, r, url.Create(r.Context(), "home"), http.StatusSeeOther)
}

func (app *Application) createExpenseCategory(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")

	categoryFC, err := model.NewExpenseCategory(name)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	err = app.expenseCategory.Create(r.Context(), categoryFC)
	if err != nil {
		var alreadyExistsErr *model.ExpenseCategoryAlreadyExistsError
		if errors.As(err, &alreadyExistsErr) {
			sendErrorResponse(w, http.StatusConflict, err.Error(), err)
			return
		}

		sendErrorResponse(w, http.StatusInternalServerError, "failed to put item: "+err.Error(), err)
		return
	}

	app.renderTempl(w, r, components.SingleExpenseCategory(r.Context(), categoryFC))
}

func (app *Application) homeHandler(w http.ResponseWriter, r *http.Request) {
	expenses, err := app.expense.Query(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to query items: "+err.Error(), err)
		return
	}

	app.renderTempl(w, r, components.Page(r.Context(), expenses))
}

func (app *Application) renderTempl(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html")

	if err := component.Render(r.Context(), w); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error while generating template: "+err.Error(), err)
		return
	}
}

func sendErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	log.Error().Stack().Err(err).Msg("")
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(statusCode)

	var validationErr *validator.ValidationError
	if errors.As(err, &validationErr) {
		_ = json.NewEncoder(w).Encode(validationErr.ErrMessages)
		return
	}

	fmt.Fprintf(w, `{"message":%q}`, message)
}
