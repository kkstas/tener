package server

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/kkstas/tener/internal/auth"
	"github.com/kkstas/tener/internal/components"
	"github.com/kkstas/tener/internal/model/user"
	"github.com/kkstas/tener/internal/url"
	"github.com/kkstas/tener/pkg/validator"
)

func (app *Application) renderLoginPage(w http.ResponseWriter, r *http.Request) error {
	return app.renderTempl(w, r, components.LoginPage(r.Context()))
}

func (app *Application) handleLogin(w http.ResponseWriter, r *http.Request) error {
	email := r.FormValue("email")
	password := r.FormValue("password")

	if ok, _, _ := validator.IsEmail("email", email); !ok {
		return &validator.ValidationError{ErrMessages: map[string][]string{"email": {"invalid email"}}}
	}

	foundUser, err := app.user.FindOneByEmail(r.Context(), email)
	if err != nil {
		var notFoundErr *user.NotFoundError
		if errors.As(err, &notFoundErr) {
			return &validator.ValidationError{ErrMessages: map[string][]string{"email": {"user with that email does not exist"}}}
		}
		return err
	}

	if !user.CheckPassword(foundUser.PasswordHash, password) {
		return &validator.ValidationError{ErrMessages: map[string][]string{"password": {"invalid password"}}}
	}

	token, err := auth.CreateToken(foundUser)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(auth.TokenTTL),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	})

	w.Header().Set("HX-Redirect", url.Create(r.Context(), "home"))
	http.Redirect(w, r, url.Create(r.Context(), "home"), http.StatusOK)
	return nil
}

func (app *Application) renderRegisterPage(w http.ResponseWriter, r *http.Request) error {
	return app.renderTempl(w, r, components.RegisterPage(r.Context()))
}

func (app *Application) handleRegister(w http.ResponseWriter, r *http.Request) error {
	email := r.FormValue("email")
	firstName := r.FormValue("firstName")
	lastName := r.FormValue("lastName")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirmPassword")

	if password != confirmPassword {
		return &validator.ValidationError{ErrMessages: map[string][]string{"confirmPassword": {"passwords do not match"}}}
	}

	userFC, err := user.New(firstName, lastName, email, password)
	if err != nil {
		return err
	}

	_, err = app.user.FindOneByEmail(r.Context(), userFC.Email)
	if err == nil {
		return NewAPIError(http.StatusBadRequest, errors.New("user with that email already exists"))
	}
	var notFoundErr *user.NotFoundError
	if !errors.As(err, &notFoundErr) {
		return fmt.Errorf("failed to check if user with given email already exists: %w", err)
	}

	_, err = app.user.Create(r.Context(), userFC)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	w.Header().Set("HX-Redirect", url.Create(r.Context(), "login"))
	http.Redirect(w, r, url.Create(r.Context(), "login"), http.StatusOK)
	return nil
}

func (app *Application) handleLogout(w http.ResponseWriter, r *http.Request) error {
	clearTokenCookie(w)
	http.Redirect(w, r, url.Create(r.Context(), "login"), http.StatusSeeOther)
	return nil
}
