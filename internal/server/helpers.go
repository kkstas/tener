package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/kkstas/tjener/pkg/validator"
	"github.com/rs/zerolog/log"
)

func (app *Application) renderTempl(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html")

	if err := component.Render(r.Context(), w); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error while generating template: "+err.Error(), err)
		return
	}
}

func sendErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(statusCode)

	var validationErr *validator.ValidationError
	if errors.As(err, &validationErr) {
		_ = json.NewEncoder(w).Encode(validationErr.ErrMessages)
		return
	}

	log.Error().Stack().Err(err).Msg("")
	fmt.Fprintf(w, `{"message":%q}`, message)
}
