package server

import (
	"errors"
	"net/http"

	"github.com/kkstas/tener/internal/components"
	"github.com/kkstas/tener/internal/model/expense"
	"github.com/kkstas/tener/internal/model/expensecategory"
	"github.com/kkstas/tener/internal/model/user"
)

func (app *Application) renderExpenseCategoriesPage(w http.ResponseWriter, r *http.Request, u user.User) {
	categories, err := app.expenseCategory.FindAll(r.Context(), u.ActiveVault)
	if err != nil {
		app.sendErrorResponse(w, http.StatusInternalServerError, "failed to query expense categories: "+err.Error(), err)
		return
	}

	users, err := app.user.FindAllByIDs(r.Context(), extractUserIDs([]expense.Expense{}, categories))
	if err != nil {
		app.sendErrorResponse(w,
			http.StatusInternalServerError,
			"failed to find matching users for expenses & expense categories: "+err.Error(),
			err)
		return
	}
	app.renderTempl(w, r, components.ExpenseCategoriesPage(r.Context(), categories, u, users))
}

func (app *Application) createAndRenderSingleExpenseCategory(w http.ResponseWriter, r *http.Request, u user.User) {
	name := r.FormValue("name")

	categoryFC, err := expensecategory.New(name)
	if err != nil {
		app.sendErrorResponse(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	err = app.expenseCategory.Create(r.Context(), categoryFC, u.ID, u.ActiveVault)
	if err != nil {
		var alreadyExistsErr *expensecategory.AlreadyExistsError
		if errors.As(err, &alreadyExistsErr) {
			app.sendErrorResponse(w, http.StatusConflict, err.Error(), err)
			return
		}

		app.sendErrorResponse(w, http.StatusInternalServerError, "failed to put item: "+err.Error(), err)
		return
	}

	app.renderTempl(w, r, components.SingleExpenseCategory(r.Context(), categoryFC, u))
}

func (app *Application) deleteSingleExpenseCategory(w http.ResponseWriter, r *http.Request, u user.User) {
	name := r.PathValue("name")

	err := app.expenseCategory.Delete(r.Context(), name, u.ActiveVault)
	if err != nil {
		app.sendErrorResponse(w, http.StatusInternalServerError, "deleting item failed: "+err.Error(), err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
