package components

import (
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	t.Run("returns parsed date if provided date is valid YYYY-MM-DD", func(t *testing.T) {
		got := parseDate("2024-09-07", time.RFC3339)
		want := "2024-09-07T00:00:00Z"
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})

	t.Run("returns input date if provided date is not valid YYYY-MM-DD", func(t *testing.T) {
		input := "2024-0907"
		got := parseDate(input, time.RFC3339)
		if got != input {
			t.Errorf("got %s, want %s", got, input)
		}
	})
}
