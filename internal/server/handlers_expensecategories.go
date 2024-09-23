package server

import (
	"errors"
	"net/http"

	"github.com/kkstas/tjener/internal/components"
	"github.com/kkstas/tjener/internal/model/expensecategory"
)

func (app *Application) renderExpenseCategoriesPage(w http.ResponseWriter, r *http.Request) {
	categories, err := app.expenseCategory.FindAll(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "failed to query expense categories: "+err.Error(), err)
		return
	}

	app.renderTempl(w, r, components.ExpenseCategoriesPage(r.Context(), categories))
}

func (app *Application) createAndRenderSingleExpenseCategory(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")

	categoryFC, err := expensecategory.New(name)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	err = app.expenseCategory.Create(r.Context(), categoryFC)
	if err != nil {
		var alreadyExistsErr *expensecategory.AlreadyExistsError
		if errors.As(err, &alreadyExistsErr) {
			sendErrorResponse(w, http.StatusConflict, err.Error(), err)
			return
		}

		sendErrorResponse(w, http.StatusInternalServerError, "failed to put item: "+err.Error(), err)
		return
	}

	app.renderTempl(w, r, components.SingleExpenseCategory(r.Context(), categoryFC))
}

func (app *Application) deleteSingleExpenseCategory(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	err := app.expenseCategory.Delete(r.Context(), name)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "deleting item failed: "+err.Error(), err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
