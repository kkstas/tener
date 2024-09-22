package server

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/kkstas/tjener/internal/auth"
	"github.com/kkstas/tjener/internal/components"
	"github.com/kkstas/tjener/internal/model/user"
	"github.com/kkstas/tjener/internal/url"
	"github.com/kkstas/tjener/pkg/validator"
)

func (app *Application) renderLoginPage(w http.ResponseWriter, r *http.Request) {
	app.renderTempl(w, r, components.LoginPage(r.Context()))
}

func (app *Application) handleLogin(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	if ok, _, _ := validator.IsEmail("email", email); !ok {
		sendErrorResponse(w, http.StatusBadRequest, "invalid email", nil)
		return
	}

	foundUser, err := app.user.FindOneByEmail(r.Context(), email)
	if err != nil {
		var notFoundErr *user.NotFoundError
		if errors.As(err, &notFoundErr) {
			sendErrorResponse(w, http.StatusNotFound, "user with that email does not exist", nil)
			return
		}
		sendErrorResponse(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	if !user.CheckPassword(foundUser.PasswordHash, password) {
		sendErrorResponse(w, http.StatusUnauthorized, "invalid password", err)
		return
	}

	token, err := auth.CreateToken(foundUser)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	})

	http.Redirect(w, r, url.Create(r.Context(), "home"), http.StatusFound)
}

func (app *Application) renderRegisterPage(w http.ResponseWriter, r *http.Request) {
	app.renderTempl(w, r, components.RegisterPage(r.Context()))
}

func (app *Application) handleRegister(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	firstName := r.FormValue("firstName")
	lastName := r.FormValue("lastName")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirmPassword")

	if password != confirmPassword {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"passwords do not match"}`)
		return
	}

	userFC, err := user.New(firstName, lastName, email, password)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Bad Request", err)
		return
	}

	_, err = app.user.FindOneByEmail(r.Context(), userFC.Email)
	if err == nil {
		sendErrorResponse(w, http.StatusBadRequest, "User with that email already exists", err)
		return
	}
	var notFoundErr *user.NotFoundError
	if !errors.As(err, &notFoundErr) {
		sendErrorResponse(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	_, err = app.user.Create(r.Context(), userFC)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Internal Server Error", err)
	}

	http.Redirect(w, r, url.Create(r.Context(), "login"), http.StatusFound)
}
