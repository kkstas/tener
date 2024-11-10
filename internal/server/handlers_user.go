package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/kkstas/tener/internal/auth"
	"github.com/kkstas/tener/internal/components"
	"github.com/kkstas/tener/internal/model/user"
	"github.com/kkstas/tener/internal/url"
	"github.com/kkstas/tener/pkg/validator"
)

func (app *Application) renderLoginPage(w http.ResponseWriter, r *http.Request) {
	app.renderTempl(w, r, components.LoginPage(r.Context()))
}

func (app *Application) handleLogin(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	if ok, _, _ := validator.IsEmail("email", email); !ok {
		sendFormErrorResponse(w, http.StatusBadRequest, map[string][]string{"email": {"invalid email"}})
		return
	}

	foundUser, err := app.user.FindOneByEmail(r.Context(), email)
	if err != nil {
		var notFoundErr *user.NotFoundError
		if errors.As(err, &notFoundErr) {
			sendFormErrorResponse(w, http.StatusNotFound, map[string][]string{"email": {"user with that email does not exist"}})
			return
		}
		sendErrorResponse(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	if !user.CheckPassword(foundUser.PasswordHash, password) {
		sendFormErrorResponse(w, http.StatusUnauthorized, map[string][]string{"password": {"invalid password"}})
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
		Expires:  time.Now().Add(auth.TokenTTL),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	})

	w.Header().Set("HX-Redirect", url.Create(r.Context(), "home"))
	http.Redirect(w, r, url.Create(r.Context(), "home"), http.StatusOK)
	return
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
		sendFormErrorResponse(w, http.StatusBadRequest, map[string][]string{"confirmPassword": {"passwords do not match"}})
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
		return
	}

	w.Header().Set("HX-Redirect", url.Create(r.Context(), "login"))
	http.Redirect(w, r, url.Create(r.Context(), "login"), http.StatusOK)
	return
}

func (app *Application) handleLogout(w http.ResponseWriter, r *http.Request) {
	clearTokenCookie(w)
	http.Redirect(w, r, url.Create(r.Context(), "login"), http.StatusSeeOther)
}
