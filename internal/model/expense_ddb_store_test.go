package model_test

import (
	"context"
	"testing"
	"time"

	"github.com/kkstas/tjener/internal/database"
	"github.com/kkstas/tjener/internal/model"
)

func TestCreateDDBExpense(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
	if err != nil {
		t.Fatalf("failed creating local test ddb table, %v", err)
	}
	defer removeDDB()

	store := model.NewExpenseDDBStore(tableName, client)

	expenses, err := store.Query(ctx)
	if err != nil {
		t.Fatalf("failed querying for expenses before putting expense, %v", err)
	}

	err = store.Create(ctx, model.Expense{})
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

func TestDeleteDDBExpense(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
	if err != nil {
		t.Fatalf("failed creating local test ddb table, %v", err)
	}
	defer removeDDB()

	store := model.NewExpenseDDBStore(tableName, client)

	err = store.Create(ctx, model.Expense{})
	if err != nil {
		t.Fatalf("failed putting item into ddb, %v", err)
	}

	expenses, err := store.Query(ctx)
	if err != nil {
		t.Fatalf("failed querying ddb table for expenses before deleting expense, %v", err)
	}

	err = store.Delete(ctx, expenses[0].CreatedAt)
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
