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
	t.Run("returns today", func(t *testing.T) {
		now := time.Now()
		loc, _ := time.LoadLocation("Europe/Warsaw")
		want, _, _ := strings.Cut(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).Format(time.RFC3339Nano), "T")

		got := getDateStringDaysAgo(0)
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestNewExpense(t *testing.T) {
	validName := "name"
	validDate := "2024-01-01"
	validCategory := "food"
	validAmount := 24.99
	validCurrency := "PLN"

	t.Run("creates valid expense", func(t *testing.T) {
		_, err := NewExpenseFC(validName, validDate, validCategory, validAmount, validCurrency)
		if err != nil {
			t.Errorf("didn't expect an error but got one: %v", err)
		}
	})

	t.Run("returns an error when category is too short", func(t *testing.T) {
		tooShortCategory := string(make([]byte, ExpenseCategoryMinLength-1))
		_, err := NewExpenseFC(validName, validDate, tooShortCategory, validAmount, validCurrency)

		if err == nil {
			t.Error("expected an error but didn't get one")
		}
		var validationErr *validator.ValidationError
		if !errors.As(err, &validationErr) {
			t.Errorf("expected %T, got %#v", validationErr, err)
		}
	})

	t.Run("returns an error when category is too long", func(t *testing.T) {
		tooLongCategory := string(make([]byte, ExpenseCategoryMaxLength+1))

		_, err := NewExpenseFC(validName, validDate, tooLongCategory, validAmount, validCurrency)

		if err == nil {
			t.Error("expected an error but didn't get one")
		}
		var validationErr *validator.ValidationError
		if !errors.As(err, &validationErr) {
			t.Errorf("expected %T, got %#v", validationErr, err)
		}
	})

	t.Run("returns an error when amount is float with precision larger than two", func(t *testing.T) {
		_, err := NewExpenseFC(validName, validDate, validCategory, 24.4234, validCurrency)
		var validationErr *validator.ValidationError
		if err == nil {
			t.Error("expected an error but didn't get one")
		}

		if !errors.As(err, &validationErr) {
			t.Errorf("expected %T, got %#v", validationErr, err)
		}
	})

	t.Run("doesn't return an error when amount is float with precision lesser than or equal to two", func(t *testing.T) {
		_, err := NewExpenseFC(validName, validDate, validCategory, 24.44, validCurrency)
		if err != nil {
			t.Errorf("didn't expect an error but got one: %v", err)
		}
		_, err = NewExpenseFC(validName, validDate, validCategory, 24.4, validCurrency)
		if err != nil {
			t.Errorf("didn't expect an error but got one: %v", err)
		}
		_, err = NewExpenseFC(validName, validDate, validCategory, 24, validCurrency)
		if err != nil {
			t.Errorf("didn't expect an error but got one: %v", err)
		}
	})

	t.Run("returns an error if currency is invalid", func(t *testing.T) {
		_, err := NewExpenseFC(validName, validDate, validCategory, validAmount, "memecoin")
		var validationErr *validator.ValidationError
		if !errors.As(err, &validationErr) {
			t.Errorf("expected %T, got %#v", validationErr, err)
		}
	})

	t.Run("returns an error if date is invalid", func(t *testing.T) {
		_, err := NewExpenseFC(validName, "202401-01", validCategory, validAmount, validCurrency)
		var validationErr *validator.ValidationError
		if !errors.As(err, &validationErr) {
			t.Errorf("expected %T, got %#v", validationErr, err)
		}
	})
}
