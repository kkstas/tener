package server

import (
	"context"
	"net/http"

	"github.com/kkstas/tener/assets"
	"github.com/kkstas/tener/internal/model/expense"
	"github.com/kkstas/tener/internal/model/expensecategory"
	"github.com/kkstas/tener/internal/model/user"
)

type expenseStore interface {
	Create(ctx context.Context, expenseFC expense.Expense, userID, vaultID string) (expense.Expense, error)
	Delete(ctx context.Context, SK, vaultID string) error
	Update(ctx context.Context, expenseFU expense.Expense, vaultID string) error
	FindOne(ctx context.Context, SK, vaultID string) (expense.Expense, error)
	Query(ctx context.Context, from, to string, categories []string, vaultID string) ([]expense.Expense, error)
	GetMonthlySums(ctx context.Context, monthsAgo int, vaultID string) ([]expense.MonthlySum, error)
}

type expenseCategoryStore interface {
	Create(ctx context.Context, categoryFC expensecategory.Category, userID, vaultID string) error
	Delete(ctx context.Context, name, vaultID string) error
	FindAll(ctx context.Context, vaultID string) ([]expensecategory.Category, error)
}

type userStore interface {
	Create(ctx context.Context, userFC user.User) (user.User, error)
	Delete(ctx context.Context, id string) error
	FindOneByID(ctx context.Context, id string) (user.User, error)
	FindOneByEmail(ctx context.Context, email string) (user.User, error)
	FindAll(ctx context.Context) ([]user.User, error)
	FindAllByIDs(ctx context.Context, ids []string) (map[string]user.User, error)
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
	mux.HandleFunc("GET /register", app.toggleRegisterMiddleware(redirectIfLoggedIn(app.renderRegisterPage)))
	mux.HandleFunc("POST /register", app.toggleRegisterMiddleware(app.handleRegister))

	mux.HandleFunc("GET    /home", app.withUser(app.renderHomePage))
	mux.HandleFunc("GET    /expense/all", app.withUser(app.renderExpenses))
	mux.HandleFunc("POST   /expense/create", app.withUser(app.createSingleExpenseAndRenderExpenses))
	mux.HandleFunc("PUT    /expense/edit/{SK}", app.withUser(app.updateSingleExpenseAndRenderExpenses))
	mux.HandleFunc("DELETE /expense/{SK}", app.withUser(app.deleteSingleExpense))
	mux.HandleFunc("GET /expense/sums", app.withUser(app.getMonthlySums))

	mux.HandleFunc("GET    /expensecategories", app.withUser(app.renderExpenseCategoriesPage))
	mux.HandleFunc("POST   /expensecategories/create", app.withUser(app.createAndRenderSingleExpenseCategory))
	mux.HandleFunc("DELETE /expensecategories/{name}", app.withUser(app.deleteSingleExpenseCategory))

	app.Handler = secureHeaders(mux)

	return app
}
