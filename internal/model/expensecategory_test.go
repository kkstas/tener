package model

import (
	"context"
	"testing"
	"time"

	"github.com/kkstas/tjener/internal/database"
)

func TestCreateExpenseCategory(t *testing.T) {
	t.Run("adds category to the database", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
		if err != nil {
			t.Fatalf("failed creating local test ddb table, %v", err)
		}
		defer removeDDB()

		store := NewExpenseCategoryStore(tableName, client)

		categories, err := store.Query(ctx)
		if err != nil {
			t.Fatalf("failed querying ddb table for expense categories before putting expense category, %v", err)
		}

		err = store.CreateExpenseCategory(ctx, "asdf")
		if err != nil {
			t.Fatalf("failed putting item into ddb, %v", err)
		}
		newCategories, err := store.Query(ctx)
		if err != nil {
			t.Fatalf("failed querying ddb table for expense categories after putting expense category, %v", err)
		}
		if (len(newCategories) - 1) != len(categories) {
			t.Errorf("expected one new expense category added. got %d", len(newCategories)-len(categories))
		}
	})
}

func TestDeleteExpenseCategory(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
	if err != nil {
		t.Fatalf("failed creating local test ddb table, %v", err)
	}
	defer removeDDB()

	store := NewExpenseCategoryStore(tableName, client)

	expenseCategoryName := "some-name"

	err = store.CreateExpenseCategory(ctx, expenseCategoryName)
	if err != nil {
		t.Fatalf("failed putting item into ddb, %v", err)
	}

	categories, err := store.Query(ctx)
	if err != nil {
		t.Fatalf("failed querying ddb table for expense categories before deleting one, %v", err)
	}

	err = store.DeleteExpenseCategory(ctx, expenseCategoryName)
	if err != nil {
		t.Fatalf("failed deleting expense category: %v", err)
	}

	newCategories, err := store.Query(ctx)
	if err != nil {
		t.Fatalf("failed querying ddb table for expense categories after deleting one, %v", err)
	}
	if (len(newCategories) + 1) != len(categories) {
		t.Errorf("expected one expense category deleted. got %d", len(newCategories)-len(categories))
	}
}
