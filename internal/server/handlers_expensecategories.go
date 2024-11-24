package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/kkstas/tener/internal/components"
	"github.com/kkstas/tener/internal/model/expense"
	"github.com/kkstas/tener/internal/model/expensecategory"
	"github.com/kkstas/tener/internal/model/user"
)

func (app *Application) renderExpenseCategoriesPage(w http.ResponseWriter, r *http.Request, u user.User) error {
	categories, err := app.expenseCategory.FindAll(r.Context(), u.ActiveVault)
	if err != nil {
		return fmt.Errorf("failed to query expense categories: %w", err)
	}

	users, err := app.user.FindAllByIDs(r.Context(), extractUserIDs([]expense.Expense{}, categories))
	if err != nil {
		return fmt.Errorf("failed to find matching users for expenses & expense categories: %w", err)
	}
	return app.renderTempl(w, r, components.ExpenseCategoriesPage(r.Context(), categories, u, users))
}

func (app *Application) createAndRenderSingleExpenseCategory(w http.ResponseWriter, r *http.Request, u user.User) error {
	name := r.FormValue("name")

	categoryFC, isValid, errMessages := expensecategory.New(name)
	if !isValid {
		return InvalidRequestData(errMessages)
	}

	err := app.expenseCategory.Create(r.Context(), categoryFC, u.ID, u.ActiveVault)
	if err != nil {
		var alreadyExistsErr *expensecategory.AlreadyExistsError
		if errors.As(err, &alreadyExistsErr) {
			return NewAPIError(http.StatusConflict, err)
		}
		return fmt.Errorf("failed to put item: %w", err)
	}

	return app.renderTempl(w, r, components.SingleExpenseCategory(r.Context(), categoryFC, u))
}

func (app *Application) deleteSingleExpenseCategory(w http.ResponseWriter, r *http.Request, u user.User) error {
	name := r.PathValue("name")

	err := app.expenseCategory.Delete(r.Context(), name, u.ActiveVault)
	if err != nil {
		return fmt.Errorf("failed deleting item: %w", err)
	}

	w.WriteHeader(http.StatusOK)
	return nil
}
