package model

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/kkstas/tjener/internal/database"
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

func TestPutItem(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
	if err != nil {
		t.Fatalf("could not create local test ddb table, %v", err)
	}
	defer removeDDB()

	store := NewExpenseStore(tableName, client)

	expenses, err := store.Query(ctx)
	if err != nil {
		t.Fatalf("could not query ddb table for expenses before putting expense, %v", err)
	}

	err = store.PutItem(ctx, Expense{
		PK: "asdf",
		SK: timestampNow(),
	})
	if err != nil {
		t.Fatalf("could not put item into ddb, %v", err)
	}
	newExpenses, err := store.Query(ctx)
	if err != nil {
		t.Fatalf("could not query ddb table for expenses after putting expense, %v", err)
	}
	if (len(newExpenses) - 1) != len(expenses) {
		t.Errorf("expected one new expense added. got %d", len(newExpenses)-len(expenses))
	}
}

func TestGetDateAgo(t *testing.T) {
	t.Run("returns datetime string with time at midnight", func(t *testing.T) {
		got := getDateDaysAgo(0)
		if !strings.HasPrefix(got[11:], "00:00:00") {
			t.Errorf("received string that is not valid RFC3339Nano from midnight - %q", got)
		}
	})

	t.Run("returns today at midnight", func(t *testing.T) {
		now := time.Now()
		loc, _ := time.LoadLocation("Europe/Warsaw")
		want := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).Format(time.RFC3339Nano)

		got := getDateDaysAgo(0)
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

}
