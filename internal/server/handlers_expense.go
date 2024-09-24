package server

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/kkstas/tjener/internal/components"
	"github.com/kkstas/tjener/internal/helpers"
	"github.com/kkstas/tjener/internal/model/expense"
	"github.com/kkstas/tjener/internal/model/user"
)

func (app *Application) renderHomePage(w http.ResponseWriter, r *http.Request, u user.User) {
	expenses, err := app.expense.Query(r.Context(), helpers.MonthAgo(), helpers.DaysAgo(0), u.ActiveVault)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to query items: "+err.Error(), err)
		return
	}

	categories, err := app.expenseCategory.FindAll(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to query expense categories: "+err.Error(), err)
		return
	}

	app.renderTempl(w, r, components.Page(r.Context(), expenses, expense.ValidCurrencies, categories))
}

func (app *Application) renderExpenses(w http.ResponseWriter, r *http.Request, u user.User) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if from == "" {
		from = helpers.MonthAgo()
	}
	if to == "" {
		to = helpers.DaysAgo(0)
	}

	expenses, err := app.expense.Query(r.Context(), from, to, u.ActiveVault)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to query items: "+err.Error(), err)
		return
	}

	categories, err := app.expenseCategory.FindAll(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to query expense categories: "+err.Error(), err)
		return
	}

	app.renderTempl(w, r, components.Expenses(r.Context(), expenses, expense.ValidCurrencies, categories))
}

func (app *Application) createSingleExpenseAndRenderExpenses(w http.ResponseWriter, r *http.Request, u user.User) {
	from, to := queryDatesRange(r)

	category := r.FormValue("category")
	currency := r.FormValue("currency")
	date := r.FormValue("date")
	name := r.FormValue("name")
	amountRaw := strings.Replace(r.FormValue("amount"), ",", ".", 1)
	amount, err := strconv.ParseFloat(amountRaw, 64)

	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "amount must be a valid decimal number", err)
		return
	}

	exp, err := expense.New(name, date, category, amount, currency)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	_, err = app.expense.Create(r.Context(), exp, u.ActiveVault)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to put item: "+err.Error(), err)
		return
	}

	expenses, err := app.expense.Query(r.Context(), from, to, u.ActiveVault)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to query items: "+err.Error(), err)
		return
	}

	categories, err := app.expenseCategory.FindAll(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to query expense categories: "+err.Error(), err)
		return
	}

	app.renderTempl(w, r, components.Expenses(r.Context(), expenses, expense.ValidCurrencies, categories))
}

func (app *Application) updateSingleExpenseAndRenderExpenses(w http.ResponseWriter, r *http.Request, u user.User) {
	from, to := queryDatesRange(r)

	SK := r.PathValue("SK")
	category := strings.TrimSpace(r.FormValue("category"))
	currency := strings.TrimSpace(r.FormValue("currency"))
	date := r.FormValue("date")
	name := strings.TrimSpace(r.FormValue("name"))
	amountRaw := strings.Replace(r.FormValue("amount"), ",", ".", 1)
	amount, err := strconv.ParseFloat(amountRaw, 64)

	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "invalid amount value", err)
		return
	}

	expenseFU, err := expense.NewFU(SK, name, date, category, amount, currency)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	err = app.expense.Update(r.Context(), expenseFU, u.ActiveVault)
	if err != nil {
		var notFoundErr *expense.NotFoundError
		if errors.As(err, &notFoundErr) {
			sendErrorResponse(w, http.StatusNotFound, err.Error(), err)
			return
		}
		sendErrorResponse(w, http.StatusInternalServerError, "error while putting item: "+err.Error(), err)
		return
	}

	expenses, err := app.expense.Query(r.Context(), from, to, u.ActiveVault)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to query items: "+err.Error(), err)
		return
	}

	categories, err := app.expenseCategory.FindAll(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to query expense categories: "+err.Error(), err)
		return
	}

	app.renderTempl(w, r, components.Expenses(r.Context(), expenses, expense.ValidCurrencies, categories))
}

func (app *Application) deleteSingleExpense(w http.ResponseWriter, r *http.Request, u user.User) {
	sk := r.PathValue("SK")

	err := app.expense.Delete(r.Context(), sk, u.ActiveVault)
	if err != nil {
		var notFoundErr *expense.NotFoundError
		if errors.As(err, &notFoundErr) {
			sendErrorResponse(w, http.StatusNotFound, err.Error(), err)
			return
		}
		sendErrorResponse(w, http.StatusInternalServerError, "error while deleting item: "+err.Error(), err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
