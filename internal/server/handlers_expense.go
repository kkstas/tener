package server

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/kkstas/tjener/internal/components"
	"github.com/kkstas/tjener/internal/model"
)

func (app *Application) renderHomePage(w http.ResponseWriter, r *http.Request) {
	expenses, err := app.expense.Query(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to query items: "+err.Error(), err)
		return
	}

	categories, err := app.expenseCategory.Query(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to query expense categories: "+err.Error(), err)
		return
	}

	app.renderTempl(w, r, components.Page(r.Context(), expenses, model.ValidCurrencies, categories))
}

func (app *Application) renderSingleExpense(w http.ResponseWriter, r *http.Request) {
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

func (app *Application) createAndRenderSingleExpense(w http.ResponseWriter, r *http.Request) {
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

	expense, err := model.NewExpenseFC(name, date, category, amount, currency)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	createdExpense, err := app.expense.Create(r.Context(), expense)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to put item: "+err.Error(), err)
		return
	}

	app.renderTempl(w, r, components.SingleExpense(r.Context(), createdExpense))
}

func (app *Application) renderSingleEditableExpense(w http.ResponseWriter, r *http.Request) {
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

	app.renderTempl(w, r, components.EditSingleExpense(r.Context(), expense, model.ValidCurrencies, categories))
}

func (app *Application) updateAndRenderSingleExpense(w http.ResponseWriter, r *http.Request) {
	createdAt := r.PathValue("SK")
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

	expenseFU, err := model.NewExpenseFU(name, createdAt, date, category, amount, currency)
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

func (app *Application) deleteSingleExpense(w http.ResponseWriter, r *http.Request) {
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
