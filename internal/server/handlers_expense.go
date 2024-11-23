package server

import (
	"errors"
	"fmt"
	"net/http"
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

func (app *Application) renderHomePage(w http.ResponseWriter, r *http.Request, u user.User) {
	var expenses []expense.Expense
	var categories []expensecategory.Category
	var monthlySums []expense.MonthlySum

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
			app.sendErrorResponse(w, http.StatusInternalServerError, err.Error(), err)
			return
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
		app.sendErrorResponse(w,
			http.StatusInternalServerError,
			"failed to find matching users for expenses & expense categories: "+err.Error(),
			err)
		return
	}

	app.renderTempl(
		w, r,
		components.Page(r.Context(), expenses, expense.PaymentMethods, categories, u, users, monthlySums),
	)
}

func (app *Application) getMonthlySums(w http.ResponseWriter, r *http.Request, u user.User) {
	monthlySums, err := app.expense.GetMonthlySums(r.Context(), MonthlySumsLastMonthsCount, u.ActiveVault)
	if err != nil {
		app.sendErrorResponse(w,
			http.StatusInternalServerError,
			"failed to find monthly sums: "+err.Error(),
			err)
		return
	}

	app.renderTempl(
		w, r,
		components.MonthlySumsChart(r.Context(), expense.TransformToChartData(monthlySums)),
	)
}

func (app *Application) renderExpenses(w http.ResponseWriter, r *http.Request, u user.User) {
	from, to, selectedCategories := queryFilters(r)

	expenses, err := app.expense.Query(r.Context(), from, to, selectedCategories, u.ActiveVault)
	if err != nil {
		app.sendErrorResponse(w, http.StatusInternalServerError, "failed to query items: "+err.Error(), err)
		return
	}

	categories, err := app.expenseCategory.FindAll(r.Context(), u.ActiveVault)
	if err != nil {
		app.sendErrorResponse(w,
			http.StatusInternalServerError,
			"failed to query expense categories: "+err.Error(),
			err)
		return
	}

	users, err := app.user.FindAllByIDs(r.Context(), extractUserIDs(expenses, categories))
	if err != nil {
		app.sendErrorResponse(w,
			http.StatusInternalServerError,
			"failed to find matching users for expenses & expense categories: "+err.Error(),
			err)
		return
	}

	app.renderTempl(w, r, components.Expenses(r.Context(), expenses, expense.PaymentMethods, categories, users))
}

func (app *Application) createSingleExpenseAndRenderExpenses(w http.ResponseWriter, r *http.Request, u user.User) {
	from, to, selectedCategories := queryFilters(r)

	category := r.FormValue("category")
	paymentMethod := r.FormValue("paymentMethod")
	date := r.FormValue("date")
	name := r.FormValue("name")
	amountRaw := strings.Replace(r.FormValue("amount"), ",", ".", 1)
	amount, err := strconv.ParseFloat(amountRaw, 64)

	if err != nil {
		app.emitActionTrail("create_expense", false, &u, err, map[string]interface{}{"inputForm": r.Form})
		app.sendErrorResponse(w, http.StatusBadRequest, "amount must be a valid decimal number", err)
		return
	}

	exp, err := expense.New(name, date, category, amount, paymentMethod)
	if err != nil {
		app.emitActionTrail("create_expense", false, &u, err, map[string]interface{}{"inputForm": r.Form})
		app.sendErrorResponse(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	_, err = app.expense.Create(r.Context(), exp, u.ID, u.ActiveVault)
	if err != nil {
		app.emitActionTrail("create_expense", false, &u, err, map[string]interface{}{"inputForm": r.Form})
		app.sendErrorResponse(w, http.StatusInternalServerError, "failed to put item: "+err.Error(), err)
		return
	}

	expenses, err := app.expense.Query(r.Context(), from, to, selectedCategories, u.ActiveVault)
	if err != nil {
		app.emitActionTrail("create_expense", false, &u, err, map[string]interface{}{"inputForm": r.Form})
		app.sendErrorResponse(w, http.StatusInternalServerError, "failed to query items: "+err.Error(), err)
		return
	}

	categories, err := app.expenseCategory.FindAll(r.Context(), u.ActiveVault)
	if err != nil {
		app.emitActionTrail("create_expense", false, &u, err, map[string]interface{}{"inputForm": r.Form})
		app.sendErrorResponse(w,
			http.StatusInternalServerError,
			"failed to query expense categories: "+err.Error(),
			err)
		return
	}

	users, err := app.user.FindAllByIDs(r.Context(), extractUserIDs(expenses, categories))
	if err != nil {
		app.emitActionTrail("create_expense", false, &u, err, map[string]interface{}{"inputForm": r.Form})
		app.sendErrorResponse(w,
			http.StatusInternalServerError,
			"failed to find matching users for expenses & expense categories: "+err.Error(),
			err)
		return
	}

	app.emitActionTrail("create_expense", true, &u, nil, map[string]interface{}{"inputForm": r.Form})
	app.renderTempl(
		w, r,
		components.Expenses(r.Context(), expenses, expense.PaymentMethods, categories, users),
	)
}

func (app *Application) updateSingleExpenseAndRenderExpenses(w http.ResponseWriter, r *http.Request, u user.User) {
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
		app.sendErrorResponse(w, http.StatusBadRequest, "invalid amount value", err)
		return
	}

	expenseFU, err := expense.NewFU(SK, name, date, category, amount, paymentMethod)
	if err != nil {
		app.emitActionTrail("update_expense", false, &u, err, map[string]interface{}{"inputForm": r.Form, "expenseFU": expenseFU})
		app.sendErrorResponse(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	err = app.expense.Update(r.Context(), expenseFU, u.ActiveVault)
	if err != nil {
		app.emitActionTrail("update_expense", false, &u, err, map[string]interface{}{"inputForm": r.Form, "expenseFU": expenseFU})
		var notFoundErr *expense.NotFoundError
		if errors.As(err, &notFoundErr) {
			app.sendErrorResponse(w, http.StatusNotFound, err.Error(), err)
			return
		}
		app.sendErrorResponse(w,
			http.StatusInternalServerError,
			"error while putting item: "+err.Error(),
			err)
		return
	}

	expenses, err := app.expense.Query(r.Context(), from, to, selectedCategories, u.ActiveVault)
	if err != nil {
		app.emitActionTrail("update_expense", false, &u, err, map[string]interface{}{"inputForm": r.Form})
		app.sendErrorResponse(w,
			http.StatusInternalServerError,
			"failed to query items: "+err.Error(),
			err)
		return
	}

	categories, err := app.expenseCategory.FindAll(r.Context(), u.ActiveVault)
	if err != nil {
		app.emitActionTrail("update_expense", false, &u, err, map[string]interface{}{"inputForm": r.Form})
		app.sendErrorResponse(w,
			http.StatusInternalServerError,
			"failed to query expense categories: "+err.Error(),
			err)
		return
	}

	users, err := app.user.FindAllByIDs(r.Context(), extractUserIDs(expenses, categories))
	if err != nil {
		app.emitActionTrail("update_expense", false, &u, err, map[string]interface{}{"inputForm": r.Form})
		app.sendErrorResponse(w,
			http.StatusInternalServerError,
			"failed to find matching users for expenses & expense categories: "+err.Error(),
			err)
		return
	}

	app.emitActionTrail("update_expense", true, &u, nil, map[string]interface{}{"inputForm": r.Form})
	w.WriteHeader(201)
	app.renderTempl(
		w, r,
		components.Expenses(r.Context(), expenses, expense.PaymentMethods, categories, users),
	)
}

func (app *Application) deleteSingleExpense(w http.ResponseWriter, r *http.Request, u user.User) {
	sk := r.PathValue("SK")

	err := app.expense.Delete(r.Context(), sk, u.ActiveVault)
	if err != nil {
		app.emitActionTrail("delete_expense", false, &u, err, map[string]interface{}{"SK": sk})
		var notFoundErr *expense.NotFoundError
		if errors.As(err, &notFoundErr) {
			app.sendErrorResponse(w, http.StatusNotFound, err.Error(), err)
			return
		}
		app.sendErrorResponse(w,
			http.StatusInternalServerError,
			"error while deleting item: "+err.Error(),
			err)
		return
	}

	app.emitActionTrail("delete_expense", true, &u, nil, map[string]interface{}{"SK": sk})
	w.WriteHeader(http.StatusOK)
}
