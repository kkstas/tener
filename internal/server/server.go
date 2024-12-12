package server

import (
	"context"
	"log/slog"
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
	logger          *slog.Logger
	http.Handler
}

func NewApplication(logger *slog.Logger, expenseStore expenseStore, expenseCategoryStore expenseCategoryStore, userStore userStore) *Application {
	app := new(Application)

	app.logger = logger

	app.expense = expenseStore
	app.expenseCategory = expenseCategoryStore
	app.user = userStore

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health-check", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", cacheControlMiddleware(http.FileServer(http.FS(assets.Public)))))

	mux.HandleFunc("GET  /login", app.make(redirectIfLoggedIn(app.renderLoginPage)))
	mux.HandleFunc("POST /login", app.make(app.handleLogin))
	mux.HandleFunc("GET  /logout", app.make(app.handleLogout))
	mux.HandleFunc("GET  /register", app.make(app.toggleRegisterMiddleware(redirectIfLoggedIn(app.renderRegisterPage))))
	mux.HandleFunc("POST /register", app.make(app.toggleRegisterMiddleware(app.handleRegister)))

	mux.HandleFunc("GET    /home", app.make(app.withUser(app.renderHomePage)))
	mux.HandleFunc("GET    /expense/all", app.make(app.withUser(app.getExpensesJSON)))
	mux.HandleFunc("POST   /expense/create", app.make(app.withUser(app.createSingleExpenseJSON)))
	mux.HandleFunc("PUT    /expense/edit/{SK}", app.make(app.withUser(app.updateSingleExpenseJSON)))
	mux.HandleFunc("DELETE /expense/{SK}", app.make(app.withUser(app.deleteSingleExpenseJSON)))
	mux.HandleFunc("GET    /expense/sums", app.make(app.withUser(app.getMonthlySumsJSON)))

	mux.HandleFunc("GET    /expensecategories", app.make(app.withUser(app.renderExpenseCategoriesPage)))
	mux.HandleFunc("POST   /expensecategories/create", app.make(app.withUser(app.createAndRenderSingleExpenseCategory)))
	mux.HandleFunc("DELETE /expensecategories/{name}", app.make(app.withUser(app.deleteSingleExpenseCategory)))

	app.Handler = app.logHTTP(secureHeaders(mux))

	return app
}
