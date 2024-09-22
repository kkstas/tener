package server_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/kkstas/tjener/internal/model/expense"
	"github.com/kkstas/tjener/internal/model/expensecategory"
	"github.com/kkstas/tjener/internal/model/user"
	"github.com/kkstas/tjener/internal/server"
)

func TestHandleLogin(t *testing.T) {
	t.Run("returns 200 if email and password are valid", func(t *testing.T) {
		email := "john.doe@gmail.com"
		password := "newPassword123!"

		var param = url.Values{}
		param.Set("email", email)
		param.Set("password", password)
		var payload = bytes.NewBufferString(param.Encode())

		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/login", payload)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		userStore := &user.InMemoryStore{}
		app := server.NewApplication(&expense.InMemoryStore{}, &expensecategory.InMemoryStore{}, userStore)

		userFC, err := user.New("John", "Doe", email, password)
		if err != nil {
			t.Fatalf("didn't expect an error but got one: %v", err)
		}
		_, err = userStore.Create(context.Background(), userFC)
		if err != nil {
			t.Fatalf("didn't expect an error but got one: %v", err)
		}

		app.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusOK)
	})

	t.Run("returns 401 if password is invalid", func(t *testing.T) {
		email := "john.doe@gmail.com"
		validPassword := "validPassword123!"
		invalidPassword := "invalidPassword321!"

		var param = url.Values{}
		param.Set("email", email)
		param.Set("password", invalidPassword)
		var payload = bytes.NewBufferString(param.Encode())

		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/login", payload)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		userStore := &user.InMemoryStore{}
		app := server.NewApplication(&expense.InMemoryStore{}, &expensecategory.InMemoryStore{}, userStore)

		userFC, err := user.New("John", "Doe", email, validPassword)
		if err != nil {
			t.Fatalf("didn't expect an error but got one: %v", err)
		}
		_, err = userStore.Create(context.Background(), userFC)
		if err != nil {
			t.Fatalf("didn't expect an error but got one: %v", err)
		}

		app.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusUnauthorized)
	})

	t.Run("returns 400 if email is invalid", func(t *testing.T) {
		var param = url.Values{}
		param.Set("email", "invalid-email.com")
		param.Set("password", "howdy")
		var payload = bytes.NewBufferString(param.Encode())

		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/login", payload)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		newTestApplication().ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("returns 404 if user with that email does not exist", func(t *testing.T) {
		var param = url.Values{}
		param.Set("email", "idontexist@gmail.com")
		param.Set("password", "howdy")
		var payload = bytes.NewBufferString(param.Encode())

		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/login", payload)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		newTestApplication().ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusNotFound)
	})
}
