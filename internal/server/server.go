package server

import (
	"context"
	"net/http"

	"github.com/kkstas/tjener/assets"
	"github.com/kkstas/tjener/internal/model/expense"
	"github.com/kkstas/tjener/internal/model/expensecategory"
)

type ExpenseStore interface {
	Create(ctx context.Context, expenseFC expense.Expense) (expense.Expense, error)
	Delete(ctx context.Context, SK string) error
	Update(ctx context.Context, expenseFU expense.Expense) error
	FindOne(ctx context.Context, SK string) (expense.Expense, error)
	Query(ctx context.Context) ([]expense.Expense, error)
	QueryByDateRange(ctx context.Context, from, to string) ([]expense.Expense, error)
}

type ExpenseCategoryStore interface {
	Create(ctx context.Context, categoryFC expensecategory.Category) error
	Delete(ctx context.Context, name string) error
	Query(ctx context.Context) ([]expensecategory.Category, error)
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

	mux.HandleFunc("GET /health-check", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", cacheControlMiddleware(http.FileServer(http.FS(assets.Public)))))

	mux.HandleFunc("GET    /home", app.renderHomePage)
	mux.HandleFunc("GET    /expense/all", app.renderExpenses)
	mux.HandleFunc("POST   /expense/create", app.createSingleExpenseAndRenderExpenses)
	mux.HandleFunc("PUT    /expense/edit/{SK}", app.updateSingleExpenseAndRenderExpenses)
	mux.HandleFunc("DELETE /expense/{SK}", app.deleteSingleExpense)

	mux.HandleFunc("GET    /expensecategories", app.renderExpenseCategoriesPage)
	mux.HandleFunc("POST   /expensecategories/create", app.createAndRenderSingleExpenseCategory)
	mux.HandleFunc("DELETE /expensecategories/{name}", app.deleteSingleExpenseCategory)

	app.Handler = secureHeaders(mux)

	return app
}
