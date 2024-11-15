package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/rs/zerolog/log"

	"github.com/kkstas/tener/internal/helpers"
	"github.com/kkstas/tener/internal/model/expense"
	"github.com/kkstas/tener/internal/model/expensecategory"
	"github.com/kkstas/tener/internal/model/user"
	"github.com/kkstas/tener/pkg/validator"
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

func sendFormErrorResponse(w http.ResponseWriter, statusCode int, messages map[string][]string) {
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(messages)
}

func queryFilters(r *http.Request) (from, to string, selectedCategories []string) {
	from = r.FormValue("from")
	to = r.FormValue("to")

	if from == "" {
		from = helpers.GetFirstDayOfCurrentMonth()
	}
	if to == "" {
		to = helpers.DaysAgo(0)
	}

	categories := r.FormValue("categories")
	if categories != "" {
		selectedCategories = strings.Split(categories, ";")
	}

	return from, to, selectedCategories
}

func extractUserIDs(expenses []expense.Expense, categories []expensecategory.Category) []string {
	uniqueUserIDs := make(map[string]bool)

	if len(expenses) > 0 {
		for _, expense := range expenses {
			uniqueUserIDs[expense.CreatedBy] = true
		}
	}

	if len(categories) > 0 {
		for _, category := range categories {
			uniqueUserIDs[category.CreatedBy] = true
		}
	}

	userIDs := make([]string, 0, len(uniqueUserIDs))

	for key := range uniqueUserIDs {
		userIDs = append(userIDs, key)
	}

	return userIDs
}

func clearTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	})
}

type logLine struct {
	Success   bool        `json:"success"`
	Action    string      `json:"action"`
	ActorName string      `json:"actorName"`
	Data      interface{} `json:"data"`
	ActorID   string      `json:"actorID"`
	ErrorMsg  string      `json:"error"`
}

func emitActionTrail(action string, success bool, actor *user.User, err error, data interface{}) {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}

	line := logLine{
		ActorID:   actor.ID,
		ActorName: actor.FirstName + " " + actor.LastName,
		Action:    action,
		Success:   success,
		Data:      data,
		ErrorMsg:  errMsg,
	}
	logBytes, _ := json.Marshal(line)
	log.Info().RawJSON("message", logBytes).Msg("")
}
