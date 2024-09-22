package server

import (
	"context"
	"net/http"

	"github.com/kkstas/tjener/assets"
	"github.com/kkstas/tjener/internal/model/expense"
	"github.com/kkstas/tjener/internal/model/expensecategory"
	"github.com/kkstas/tjener/internal/model/user"
)

type expenseStore interface {
	Create(ctx context.Context, expenseFC expense.Expense) (expense.Expense, error)
	Delete(ctx context.Context, SK string) error
	Update(ctx context.Context, expenseFU expense.Expense) error
	FindOne(ctx context.Context, SK string) (expense.Expense, error)
	Query(ctx context.Context) ([]expense.Expense, error)
	QueryByDateRange(ctx context.Context, from, to string) ([]expense.Expense, error)
}

type expenseCategoryStore interface {
	Create(ctx context.Context, categoryFC expensecategory.Category) error
	Delete(ctx context.Context, name string) error
	Query(ctx context.Context) ([]expensecategory.Category, error)
}

type userStore interface {
	Create(ctx context.Context, userFC user.User) (user.User, error)
	Delete(ctx context.Context, id string) error
	FindOneByID(ctx context.Context, id string) (user.User, error)
	FindOneByEmail(ctx context.Context, email string) (user.User, error)
	FindAll(ctx context.Context) ([]user.User, error)
}

type Application struct {
	expense         expenseStore
	expenseCategory expenseCategoryStore
	user            userStore
	http.Handler
}

func NewApplication(expenseStore expenseStore, expenseCategoryStore expenseCategoryStore, userStore userStore) *Application {
	app := new(Application)

	app.expense = expenseStore
	app.expenseCategory = expenseCategoryStore
	app.user = userStore

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health-check", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", cacheControlMiddleware(http.FileServer(http.FS(assets.Public)))))

	mux.HandleFunc("GET /login", redirectIfLoggedIn(app.renderLoginPage))
	mux.HandleFunc("POST /login", app.handleLogin)
	mux.HandleFunc("GET /logout", app.handleLogout)
	mux.HandleFunc("GET /register", redirectIfLoggedIn(app.renderRegisterPage))
	mux.HandleFunc("POST /register", app.handleRegister)

	mux.HandleFunc("GET    /home", withUser(app.renderHomePage))
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
