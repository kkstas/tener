package model

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/kkstas/tjener/internal/database"
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

func TestPutExpense(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
	if err != nil {
		t.Fatalf("failed creating local test ddb table, %v", err)
	}
	defer removeDDB()

	store := NewExpenseStore(tableName, client)

	expenses, err := store.Query(ctx)
	if err != nil {
		t.Fatalf("failed querying ddb table for expenses before putting expense, %v", err)
	}

	err = store.Create(ctx, Expense{})
	if err != nil {
		t.Fatalf("failed putting item into ddb, %v", err)
	}
	newExpenses, err := store.Query(ctx)
	if err != nil {
		t.Fatalf("failed querying ddb table for expenses after putting expense, %v", err)
	}
	if (len(newExpenses) - 1) != len(expenses) {
		t.Errorf("expected one new expense added. got %d", len(newExpenses)-len(expenses))
	}
}

func TestGetDateAgo(t *testing.T) {
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

func TestCreateExpense(t *testing.T) {
	t.Run("creates valid expense", func(t *testing.T) {
		_, err := NewExpense("name", "food", 24.99, "PLN")
		if err != nil {
			t.Errorf("didn't expect an error but got one: %v", err)
		}
	})

	t.Run("returns an error when category is too short", func(t *testing.T) {
		tooShortCategory := string(make([]byte, expenseCategoryMinLength-1))
		_, err := NewExpense("some name", tooShortCategory, 24.99, "PLN")

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

		_, err := NewExpense("some name", tooLongCategory, 24.99, "PLN")

		if err == nil {
			t.Error("expected an error but didn't get one")
		}
		var validationErr *validator.ValidationError
		if !errors.As(err, &validationErr) {
			t.Errorf("expected %T, got %#v", validationErr, err)
		}
	})

	t.Run("returns an error when amount is float with precision larger than two", func(t *testing.T) {
		_, err := NewExpense("", "food", 24.4234, "PLN")
		var validationErr *validator.ValidationError
		if err == nil {
			t.Error("expected an error but didn't get one")
		}

		if !errors.As(err, &validationErr) {
			t.Errorf("expected %T, got %#v", validationErr, err)
		}
	})

	t.Run("returns an error if currency is invalid", func(t *testing.T) {
		_, err := NewExpense("", "food", 24.99, "memecoin")
		var validationErr *validator.ValidationError
		if !errors.As(err, &validationErr) {
			t.Errorf("expected %T, got %#v", validationErr, err)
		}
	})
}

func TestDeleteExpense(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
	if err != nil {
		t.Fatalf("failed creating local test ddb table, %v", err)
	}
	defer removeDDB()

	store := NewExpenseStore(tableName, client)

	err = store.Create(ctx, Expense{})
	if err != nil {
		t.Fatalf("failed putting item into ddb, %v", err)
	}

	expenses, err := store.Query(ctx)
	if err != nil {
		t.Fatalf("failed querying ddb table for expenses before deleting expense, %v", err)
	}

	err = store.Delete(ctx, expenses[0].SK)
	if err != nil {
		t.Fatalf("failed deleting expense: %v", err)
	}

	newExpenses, err := store.Query(ctx)
	if err != nil {
		t.Fatalf("failed querying ddb table for expenses after deleting expense, %v", err)
	}
	if (len(newExpenses) + 1) != len(expenses) {
		t.Errorf("expected one expense deleted. got %d", len(newExpenses)-len(expenses))
	}
}
