package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"

	"github.com/kkstas/tener/internal/helpers"
	"github.com/kkstas/tener/internal/model/expense"
	"github.com/kkstas/tener/internal/model/expensecategory"
	"github.com/kkstas/tener/internal/model/user"
)

func (app *Application) renderTempl(w http.ResponseWriter, r *http.Request, component templ.Component) error {
	w.Header().Set("Content-Type", "text/html")

	if err := component.Render(r.Context(), w); err != nil {
		return fmt.Errorf("failed to generate template: %w", err)
	}
	return nil
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

func (app *Application) emitActionTrail(action string, success bool, actor *user.User, err error, data interface{}) {
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

	app.logger.Info("actionTrail", "actionTrail", line)
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
