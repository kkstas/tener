package server_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/kkstas/tjener/internal/model/expense"
	"github.com/kkstas/tjener/internal/model/expensecategory"
	"github.com/kkstas/tjener/internal/model/user"
	"github.com/kkstas/tjener/internal/server"
)

const (
	validFirstName = "John"
	validLastName  = "Doe"
	validEmail     = "test@example.us"
	validPassword  = "newPassword123!"
)

func TestLogin(t *testing.T) {
	t.Run("redirects and creates token cookie if email and password are valid", func(t *testing.T) {
		os.Setenv("TOKEN_SECRET", "gHg8v3-XKj9XO8M-6gpjzW0n1xn7UZTBICIY1FcjyPw")
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
		assertStatus(t, response.Code, http.StatusFound)

		var cookie *http.Cookie
		for _, c := range response.Result().Cookies() {
			if c.Name == "token" && c.Value != "" {
				cookie = c
			}
		}

		if cookie == nil {
			t.Error("expected set token cookie, but didn't find one")
			return
		}

		if !cookie.HttpOnly {
			t.Error("expected token cookie to have set HttpOnly to true")
		}
		if !cookie.Secure {
			t.Error("expected token cookie to have set Secure to true")
		}

		if cookie.SameSite != http.SameSiteLaxMode {
			t.Errorf("expected cookie to have set SameSiteLaxMode, got %v", cookie.SameSite)
		}

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

func TestRegister(t *testing.T) {
	t.Run("registers new user", func(t *testing.T) {
		var param = url.Values{}
		param.Set("firstName", validFirstName)
		param.Set("lastName", validLastName)
		param.Set("email", validEmail)
		param.Set("password", validPassword)
		param.Set("confirmPassword", validPassword)
		var payload = bytes.NewBufferString(param.Encode())

		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/register", payload)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		userStore := &user.InMemoryStore{}
		app := server.NewApplication(&expense.InMemoryStore{}, &expensecategory.InMemoryStore{}, userStore)
		app.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusFound)

		users, err := userStore.FindAll(context.Background())
		if err != nil {
			t.Fatalf("didn't expect an error but got one: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("expected find all user list's length equal to 1, got %d", len(users))
		}
	})

	t.Run("returns error when passwords don't match", func(t *testing.T) {
		var param = url.Values{}
		param.Set("firstName", validFirstName)
		param.Set("lastName", validLastName)
		param.Set("email", validEmail)
		param.Set("password", validPassword)
		param.Set("confirmPassword", validPassword+"X")
		var payload = bytes.NewBufferString(param.Encode())

		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/register", payload)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		userStore := &user.InMemoryStore{}
		server.NewApplication(&expense.InMemoryStore{}, &expensecategory.InMemoryStore{}, userStore).ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("returns error when password has less than 8 characters", func(t *testing.T) {
		tooShortPassword := string(make([]byte, user.PasswordMinLength-1))
		var param = url.Values{}
		param.Set("firstName", validFirstName)
		param.Set("lastName", validLastName)
		param.Set("email", validEmail)
		param.Set("password", tooShortPassword)
		param.Set("confirmPassword", tooShortPassword)
		var payload = bytes.NewBufferString(param.Encode())

		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/register", payload)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		userStore := &user.InMemoryStore{}
		server.NewApplication(&expense.InMemoryStore{}, &expensecategory.InMemoryStore{}, userStore).ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("returns error when user with provided email already exists", func(t *testing.T) {
		var param = url.Values{}
		param.Set("firstName", validFirstName)
		param.Set("lastName", validLastName)
		param.Set("email", validEmail)
		param.Set("password", validPassword)
		param.Set("confirmPassword", validPassword)
		var payload = bytes.NewBufferString(param.Encode())

		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/register", payload)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		userStore := &user.InMemoryStore{}
		app := server.NewApplication(&expense.InMemoryStore{}, &expensecategory.InMemoryStore{}, userStore)

		userFC, err := user.New(validFirstName, validLastName, validEmail, validPassword)
		if err != nil {
			t.Fatalf("didn't expect an error but got one: %v", err)
		}
		_, err = userStore.Create(context.Background(), userFC)
		if err != nil {
			t.Fatalf("didn't expect an error but got one: %v", err)
		}
		usersBefore, err := userStore.FindAll(context.Background())
		if err != nil {
			t.Fatalf("didn't expect an error but got one: %v", err)
		}

		app.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusBadRequest)

		usersAfter, err := userStore.FindAll(context.Background())
		if err != nil {
			t.Fatalf("didn't expect an error but got one: %v", err)
		}

		lenBefore := len(usersBefore)
		lenAfter := len(usersAfter)

		if lenBefore != lenAfter {
			t.Errorf("expected unchanged number of users, got before: %d and after %d", lenBefore, lenAfter)
		}
	})
}

func TestLogout(t *testing.T) {
	t.Run("overwrites token cookie and redirects to login", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/logout", nil)

		userStore := &user.InMemoryStore{}
		app := server.NewApplication(&expense.InMemoryStore{}, &expensecategory.InMemoryStore{}, userStore)

		app.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusSeeOther)

		var cookie *http.Cookie
		for _, c := range response.Result().Cookies() {
			if c.Name == "token" {
				cookie = c
			}
		}

		if cookie == nil {
			t.Error("expected set token cookie, but didn't find one")
			return
		}

		if cookie.Value != "" {
			t.Errorf("expected token cookie value to be empty string, but got %s", cookie.Value)
		}

		if !cookie.HttpOnly {
			t.Error("expected token cookie to have set HttpOnly to true")
		}
		if !cookie.Secure {
			t.Error("expected token cookie to have set Secure to true")
		}

		if cookie.SameSite != http.SameSiteLaxMode {
			t.Errorf("expected cookie to have set SameSiteLaxMode, got %v", cookie.SameSite)
		}

	})
}
