package model

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/kkstas/tjener/pkg/validator"
)

func BenchmarkRFC3339(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = time.Now().Format(time.RFC3339)
	}
}

func BenchmarkRFC3339Nano(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = time.Now().Format(time.RFC3339Nano)
	}
}

func TestTimestampDaysAgo(t *testing.T) {
	t.Run("returns datetime string with time at midnight", func(t *testing.T) {
		got := getTimestampDaysAgo(0)
		if !strings.HasPrefix(got[11:], "00:00:00") {
			t.Errorf("received string that is not valid RFC3339Nano from midnight - %q", got)
		}
	})

	t.Run("returns today at midnight", func(t *testing.T) {
		now := time.Now()
		loc, _ := time.LoadLocation("Europe/Warsaw")
		want := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).Format(time.RFC3339Nano)

		got := getTimestampDaysAgo(0)
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

}

func TestNewExpense(t *testing.T) {
	t.Run("creates valid expense", func(t *testing.T) {
		_, err := NewExpenseFC("name", "food", 24.99, "PLN")
		if err != nil {
			t.Errorf("didn't expect an error but got one: %v", err)
		}
	})

	t.Run("returns an error when category is too short", func(t *testing.T) {
		tooShortCategory := string(make([]byte, expenseCategoryMinLength-1))
		_, err := NewExpenseFC("some name", tooShortCategory, 24.99, "PLN")

		if err == nil {
			t.Error("expected an error but didn't get one")
		}
		var validationErr *validator.ValidationError
		if !errors.As(err, &validationErr) {
			t.Errorf("expected %T, got %#v", validationErr, err)
		}
	})

	t.Run("returns an error when category is too long", func(t *testing.T) {
		tooLongCategory := string(make([]byte, expenseCategoryMaxLength+1))

		_, err := NewExpenseFC("some name", tooLongCategory, 24.99, "PLN")

		if err == nil {
			t.Error("expected an error but didn't get one")
		}
		var validationErr *validator.ValidationError
		if !errors.As(err, &validationErr) {
			t.Errorf("expected %T, got %#v", validationErr, err)
		}
	})

	t.Run("returns an error when amount is float with precision larger than two", func(t *testing.T) {
		_, err := NewExpenseFC("", "food", 24.4234, "PLN")
		var validationErr *validator.ValidationError
		if err == nil {
			t.Error("expected an error but didn't get one")
		}

		if !errors.As(err, &validationErr) {
			t.Errorf("expected %T, got %#v", validationErr, err)
		}
	})

	t.Run("returns an error if currency is invalid", func(t *testing.T) {
		_, err := NewExpenseFC("", "food", 24.99, "memecoin")
		var validationErr *validator.ValidationError
		if !errors.As(err, &validationErr) {
			t.Errorf("expected %T, got %#v", validationErr, err)
		}
	})
}