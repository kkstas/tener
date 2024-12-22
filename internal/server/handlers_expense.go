package server

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/kkstas/tener/internal/components"
	"github.com/kkstas/tener/internal/helpers"
	"github.com/kkstas/tener/internal/model/expense"
	"github.com/kkstas/tener/internal/model/expensecategory"
	"github.com/kkstas/tener/internal/model/user"
)

var (
	MonthlySumsLastMonthsCount = 6
)

func (app *Application) renderHomePage(w http.ResponseWriter, r *http.Request, u user.User) error {
	expenses := []expense.Expense{}
	categories := []expensecategory.Category{}
	monthlySums := []expense.MonthlySum{}

	expChan := make(chan []expense.Expense)
	catChan := make(chan []expensecategory.Category)
	sumsChan := make(chan []expense.MonthlySum)
	errChan := make(chan error)

	go func() {
		expenses, err := app.expense.Query(r.Context(), helpers.GetFirstDayOfCurrentMonth(), helpers.DaysAgo(0), []string{}, u.ActiveVault)
		if err != nil {
			errChan <- fmt.Errorf("failed to query expenses: %w", err)
			return
		}
		expChan <- expenses
	}()

	go func() {
		categories, err := app.expenseCategory.FindAll(r.Context(), u.ActiveVault)
		if err != nil {
			errChan <- fmt.Errorf("failed to find all expense categories: %w", err)
			return
		}
		catChan <- categories
	}()

	go func() {
		monthlySums, err := app.expense.GetMonthlySums(r.Context(), MonthlySumsLastMonthsCount, u.ActiveVault)
		if err != nil {
			errChan <- fmt.Errorf("failed to get monthly sums: %w", err)
			return
		}
		sumsChan <- monthlySums
	}()

	for i := 0; i < 3; i++ {
		select {
		case err := <-errChan:
			return err
		case result := <-expChan:
			expenses = result
		case result := <-catChan:
			categories = result
		case result := <-sumsChan:
			monthlySums = result
		}
	}

	users, err := app.user.FindAllByIDs(r.Context(), extractUserIDs(expenses, categories))
	if err != nil {
		return fmt.Errorf("failed to find matching users for expenses & expense categories: %w", err)
	}

	return app.renderTempl(
		w, r,
		components.Home(r.Context(), expenses, expense.PaymentMethods, categories, u, users, monthlySums),
	)
}

func (app *Application) getMonthlySumsJSON(w http.ResponseWriter, r *http.Request, u user.User) error {
	_, _, selectedCategories := queryFilters(r)
	monthlySums, err := app.expense.GetMonthlySums(r.Context(), MonthlySumsLastMonthsCount, u.ActiveVault)
	if err != nil {
		return fmt.Errorf("failed to find monthly sums: %w", err)
	}

	if len(selectedCategories) > 0 {
		filteredSums := []expense.MonthlySum{}
		for _, s := range monthlySums {
			if slices.Contains(selectedCategories, s.Category) {
				filteredSums = append(filteredSums, s)
			}
		}
		monthlySums = filteredSums
	}

	return writeJSON(w, http.StatusOK, expense.TransformToChartData(monthlySums))
}

func (app *Application) getExpensesJSON(w http.ResponseWriter, r *http.Request, u user.User) error {
	from, to, selectedCategories := queryFilters(r)

	expenses, err := app.expense.Query(r.Context(), from, to, selectedCategories, u.ActiveVault)
	if err != nil {
		return fmt.Errorf("failed to query expenses: %w", err)
	}

	categories, err := app.expenseCategory.FindAll(r.Context(), u.ActiveVault)
	if err != nil {
		return fmt.Errorf("failed to query expense categories: %w", err)
	}

	users, err := app.user.FindAllByIDs(r.Context(), extractUserIDs(expenses, categories))
	if err != nil {
		return fmt.Errorf("failed to find matching users for expenses & expense categories: %w", err)
	}

	return writeJSON(w, http.StatusOK, map[string]any{
		"expenses":   expenses,
		"categories": categories,
		"users":      users,
	})
}

func (app *Application) createSingleExpenseJSON(w http.ResponseWriter, r *http.Request, u user.User) error {
	from, to, selectedCategories := queryFilters(r)

	category := r.FormValue("category")
	paymentMethod := r.FormValue("paymentMethod")
	date := r.FormValue("date")
	name := r.FormValue("name")
	amountRaw := strings.Replace(r.FormValue("amount"), ",", ".", 1)

	amount, err := strconv.ParseFloat(amountRaw, 64)
	if err != nil {
		app.emitActionTrail("create_expense", false, &u, err, map[string]interface{}{"inputForm": r.Form})
		return InvalidRequestData(map[string][]string{"amount": {"must be a valid decimal number"}})
	}

	exp, isValid, errMessages := expense.New(name, date, category, amount, paymentMethod)
	if !isValid {
		validationErr := InvalidRequestData(errMessages)
		app.emitActionTrail("create_expense", false, &u, validationErr, map[string]interface{}{"inputForm": r.Form})
		return validationErr
	}

	_, err = app.expense.Create(r.Context(), exp, u.ID, u.ActiveVault)
	if err != nil {
		app.emitActionTrail("create_expense", false, &u, err, map[string]interface{}{"inputForm": r.Form})

		var maxCountErr *expense.MaxMonthExpenseCountExceededError
		if errors.As(err, &maxCountErr) {
			return NewAPIError(http.StatusForbidden, err)
		}

		return fmt.Errorf("failed to put item: %w", err)
	}

	app.emitActionTrail("create_expense", true, &u, nil, map[string]interface{}{"inputForm": r.Form})

	expenses, err := app.expense.Query(r.Context(), from, to, selectedCategories, u.ActiveVault)
	if err != nil {
		return fmt.Errorf("failed to query items: %w", err)
	}

	categories, err := app.expenseCategory.FindAll(r.Context(), u.ActiveVault)
	if err != nil {
		return fmt.Errorf("failed to query expense categories: %w", err)
	}

	users, err := app.user.FindAllByIDs(r.Context(), extractUserIDs(expenses, categories))
	if err != nil {
		return fmt.Errorf("failed to find matching users for expenses & expense categories: %w", err)
	}

	return writeJSON(w, http.StatusOK, map[string]any{
		"expenses":   expenses,
		"categories": categories,
		"users":      users,
	})
}

func (app *Application) updateSingleExpenseJSON(w http.ResponseWriter, r *http.Request, u user.User) error {
	from, to, selectedCategories := queryFilters(r)

	SK := r.PathValue("SK")
	category := strings.TrimSpace(r.FormValue("category"))
	paymentMethod := strings.TrimSpace(r.FormValue("paymentMethod"))
	date := r.FormValue("date")
	name := strings.TrimSpace(r.FormValue("name"))
	amountRaw := strings.Replace(r.FormValue("amount"), ",", ".", 1)
	amount, err := strconv.ParseFloat(amountRaw, 64)

	if err != nil {
		app.emitActionTrail("update_expense", false, &u, err, map[string]interface{}{"inputForm": r.Form})
		return InvalidRequestData(map[string][]string{"amount": {"invalid amount value"}})
	}

	expenseFU, isValid, errMessages := expense.NewFU(SK, name, date, category, amount, paymentMethod)
	if !isValid {
		validationErr := InvalidRequestData(errMessages)
		app.emitActionTrail("update_expense", false, &u, validationErr, map[string]interface{}{"inputForm": r.Form, "expenseFU": expenseFU})
		return validationErr
	}

	err = app.expense.Update(r.Context(), expenseFU, u.ActiveVault)
	if err != nil {
		app.emitActionTrail("update_expense", false, &u, err, map[string]interface{}{"inputForm": r.Form, "expenseFU": expenseFU})

		var notFoundErr *expense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return NewAPIError(http.StatusNotFound, err)
		}

		var maxCountErr *expense.MaxMonthExpenseCountExceededError
		if errors.As(err, &maxCountErr) {
			return NewAPIError(http.StatusForbidden, err)
		}

		return fmt.Errorf("failed to put item: %w", err)
	}

	app.emitActionTrail("update_expense", true, &u, nil, map[string]interface{}{"inputForm": r.Form})

	expenses, err := app.expense.Query(r.Context(), from, to, selectedCategories, u.ActiveVault)
	if err != nil {
		return fmt.Errorf("failed to query expenses: %w", err)
	}

	categories, err := app.expenseCategory.FindAll(r.Context(), u.ActiveVault)
	if err != nil {
		return fmt.Errorf("failed to query expense categories: %w", err)
	}

	users, err := app.user.FindAllByIDs(r.Context(), extractUserIDs(expenses, categories))
	if err != nil {
		return fmt.Errorf("failed to find matching users for expenses & expense categories: %w", err)
	}

	return writeJSON(w, http.StatusOK, map[string]any{
		"expenses":   expenses,
		"categories": categories,
		"users":      users,
	})
}

func (app *Application) deleteSingleExpenseJSON(w http.ResponseWriter, r *http.Request, u user.User) error {
	sk := r.PathValue("SK")

	err := app.expense.Delete(r.Context(), sk, u.ActiveVault)
	if err != nil {
		app.emitActionTrail("delete_expense", false, &u, err, map[string]interface{}{"SK": sk})
		var notFoundErr *expense.NotFoundError
		if errors.As(err, &notFoundErr) {
			return NewAPIError(http.StatusNotFound, err)
		}
		return fmt.Errorf("failed to delete item: %w", err)
	}

	app.emitActionTrail("delete_expense", true, &u, nil, map[string]interface{}{"SK": sk})

	from, to, selectedCategories := queryFilters(r)

	expenses, err := app.expense.Query(r.Context(), from, to, selectedCategories, u.ActiveVault)
	if err != nil {
		return fmt.Errorf("failed to query expenses: %w", err)
	}

	categories, err := app.expenseCategory.FindAll(r.Context(), u.ActiveVault)
	if err != nil {
		return fmt.Errorf("failed to query expense categories: %w", err)
	}

	users, err := app.user.FindAllByIDs(r.Context(), extractUserIDs(expenses, categories))
	if err != nil {
		return fmt.Errorf("failed to find matching users for expenses & expense categories: %w", err)
	}

	return writeJSON(w, http.StatusOK, map[string]any{
		"expenses":   expenses,
		"categories": categories,
		"users":      users,
	})
}
