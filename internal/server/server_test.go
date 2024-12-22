package server_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kkstas/tener/assets"
	"github.com/kkstas/tener/internal/auth"
	"github.com/kkstas/tener/internal/database"
	"github.com/kkstas/tener/internal/model/expense"
	"github.com/kkstas/tener/internal/model/expensecategory"
	"github.com/kkstas/tener/internal/model/user"
	"github.com/kkstas/tener/internal/server"
	u "github.com/kkstas/tener/internal/url"
)

func TestHomeHandler(t *testing.T) {
	t.Run("responds with html", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/home", nil)
		addTokenCookie(t, request)
		newTestApplication().ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)

		contentType := response.Header().Get("content-type")
		if !strings.HasPrefix(contentType, "text/html") {
			t.Errorf("invalid content type %q", contentType)
		}
	})
}

func TestHealthCheck(t *testing.T) {
	t.Run("returns status 200", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/health-check", nil)
		newTestApplication().ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
	})
}

func TestStaticCss(t *testing.T) {
	t.Run("returns css file content with status 200", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/assets/public/css/"+assets.OutCSSFilename(), nil)
		newTestApplication().ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)

		if response.Body.Len() == 0 {
			t.Errorf("response body is empty")
		}
	})
}

func TestCreateExpense(t *testing.T) {
	t.Run("returns 400 if there's no form params", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/expense/create", nil)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		addTokenCookie(t, request)
		newTestApplication().ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("returns 400 if amount is not valid float64", func(t *testing.T) {
		var param = url.Values{}
		param.Set("paymentMethod", expense.PaymentMethods[0])
		param.Set("amount", "1.9d9")
		param.Set("category", "food")
		param.Set("date", "2024-01-01")
		var payload = bytes.NewBufferString(param.Encode())

		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/expense/create", payload)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		addTokenCookie(t, request)
		newTestApplication().ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("returns 200", func(t *testing.T) {
		var param = url.Values{}
		param.Set("paymentMethod", expense.PaymentMethods[0])
		param.Set("amount", "1.99")
		param.Set("category", "food")
		param.Set("name", "some name")
		param.Set("date", "2024-01-01")
		var payload = bytes.NewBufferString(param.Encode())

		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/expense/create", payload)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		addTokenCookie(t, request)

		newTestApplication().ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusOK)
	})

	t.Run("allows comma and dot as a decimal separator", func(t *testing.T) {
		var param = url.Values{}
		param.Set("paymentMethod", expense.PaymentMethods[0])
		param.Set("amount", "24,95")
		param.Set("category", "food")
		param.Set("name", "some name")
		param.Set("date", "2024-01-01")
		var payload = bytes.NewBufferString(param.Encode())

		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/expense/create", payload)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		addTokenCookie(t, request)

		newTestApplication().ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusOK)
	})

	t.Run("returns status forbidden and does not exceed expense count limit", func(t *testing.T) {
		expenseLimit := 3
		app, cancel := newTestApplicationWithDDB(t, expenseLimit)
		defer cancel()

		createExpense := func() *httptest.ResponseRecorder {
			var param = url.Values{}
			param.Set("paymentMethod", expense.PaymentMethods[0])
			param.Set("amount", "1.99")
			param.Set("category", "food")
			param.Set("name", "some name")
			param.Set("date", "2024-01-01")
			var payload = bytes.NewBufferString(param.Encode())
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/expense/create", payload)
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			addTokenCookie(t, request)
			app.ServeHTTP(response, request)
			return response
		}

		for i := 0; i < expenseLimit; i++ {
			res := createExpense()
			if res.Code != http.StatusOK {
				t.Fatalf("failed to create expense - got response: %v", res)
			}
		}

		res := createExpense()
		assertStatus(t, res.Code, http.StatusForbidden)
	})
}

func TestUpdateExpense(t *testing.T) {
	t.Run("allows comma and dot as a decimal separator", func(t *testing.T) {
		store := expense.InMemoryStore{}
		SK := "2024-08-25::1725652252238"

		_, err := store.Create(
			context.Background(),
			expense.Expense{PK: "expense", SK: SK, Name: "name", Amount: 18.24, Category: "food", PaymentMethod: expense.PaymentMethods[0]},
			"userID",
			"activeVaultID",
		)

		if err != nil {
			t.Fatalf("didn't expect an error but got one: %v", err)
		}

		var param = url.Values{}
		param.Set("paymentMethod", expense.PaymentMethods[0])
		param.Set("amount", "24,95")
		param.Set("category", "food")
		param.Set("name", "some name")
		param.Set("date", "2024-01-01")
		var payload = bytes.NewBufferString(param.Encode())

		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPut, u.Create(context.Background(), "expense", "edit", SK), payload)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		addTokenCookie(t, request)

		logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
		server.NewApplication(logger, &store, &expensecategory.InMemoryStore{}, &user.InMemoryStore{}).ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusOK)
	})
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status: got %d, want %d", got, want)
	}
}

func newTestApplication() *server.Application {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	return server.NewApplication(logger, &expense.InMemoryStore{}, &expensecategory.InMemoryStore{}, &user.InMemoryStore{})
}

func newTestApplicationWithDDB(t testing.TB, expenseLimit int) (app *server.Application, cancelFunc func()) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
	if err != nil {
		t.Fatalf("failed creating local test ddb table: %v", err)
	}

	cancelFunc = func() {
		removeDDB()
		cancel()
	}

	store := expense.NewDDBStoreWithExpenseMonthLimit(tableName, client, expenseLimit)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	return server.NewApplication(logger, store, &expensecategory.InMemoryStore{}, &user.InMemoryStore{}), cancelFunc
}

func addTokenCookie(t testing.TB, r *http.Request) {
	t.Helper()
	userFC, isValid, errMessages := user.New(validFirstName, validLastName, validEmail, validPassword)
	if !isValid {
		t.Fatalf("didn't expect na error but got one: %v", errMessages)
	}
	userFC.Vaults = []string{"vaultID"}

	token, err := auth.CreateToken(userFC)
	if err != nil {
		t.Fatalf("didn't expect na error but got one: %v", err)
	}

	r.Header.Add("cookie", fmt.Sprintf("token=%s", token))
}
