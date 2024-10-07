package expensecategory_test

import (
	"context"
	"testing"
	"time"

	"github.com/kkstas/tjener/internal/database"
	"github.com/kkstas/tjener/internal/model/expensecategory"
)

func TestDDBCreate(t *testing.T) {
	t.Run("adds category to the database", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
		if err != nil {
			t.Fatalf("failed creating local test ddb table, %v", err)
		}
		defer removeDDB()

		store := expensecategory.NewDDBStore(tableName, client)

		categories, err := store.FindAll(ctx, "activeVaultID")
		if err != nil {
			t.Fatalf("failed querying ddb table for expense categories before putting expense category, %v", err)
		}

		categoryFC, _ := expensecategory.New("some-name")
		err = store.Create(ctx, categoryFC, "userID", "activeVaultID")
		if err != nil {
			t.Fatalf("failed putting item into ddb, %v", err)
		}
		newCategories, err := store.FindAll(ctx, "activeVaultID")
		if err != nil {
			t.Fatalf("failed querying ddb table for expense categories after putting expense category, %v", err)
		}
		if (len(newCategories) - 1) != len(categories) {
			t.Errorf("expected one new expense category added. got %d", len(newCategories)-len(categories))
		}
	})
}

func TestDDBDelete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
	if err != nil {
		t.Fatalf("failed creating local test ddb table, %v", err)
	}
	defer removeDDB()

	store := expensecategory.NewDDBStore(tableName, client)

	categoryFC, _ := expensecategory.New("some-name")

	err = store.Create(ctx, categoryFC, "userID", "activeVaultID")
	if err != nil {
		t.Fatalf("failed putting item into ddb, %v", err)
	}

	categories, err := store.FindAll(ctx, "activeVaultID")
	if err != nil {
		t.Fatalf("failed querying ddb table for expense categories before deleting one, %v", err)
	}

	err = store.Delete(ctx, categoryFC.Name, "activeVaultID")
	if err != nil {
		t.Fatalf("failed deleting expense category: %v", err)
	}

	newCategories, err := store.FindAll(ctx, "activeVaultID")
	if err != nil {
		t.Fatalf("failed querying ddb table for expense categories after deleting one, %v", err)
	}
	if (len(newCategories) + 1) != len(categories) {
		t.Errorf("expected one expense category deleted. got %d", len(newCategories)-len(categories))
	}
}
