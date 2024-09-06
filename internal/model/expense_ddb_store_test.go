package model_test

import (
	"context"
	"errors"
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
		t.Fatalf("failed creating local test ddb table: %v", err)
	}
	defer removeDDB()

	store := model.NewExpenseDDBStore(tableName, client)

	expense, err := store.Create(ctx, model.Expense{})
	if err != nil {
		t.Fatalf("failed putting item into ddb, %v", err)
	}

	_, err = store.FindOne(ctx, expense.SK)
	if err != nil {
		t.Errorf("didn't expect an error but got one: %v", err)
	}
}

func TestDeleteDDBExpense(t *testing.T) {
	t.Run("deletes existing expense", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
		if err != nil {
			t.Fatalf("failed creating local test ddb table, %v", err)
		}
		defer removeDDB()

		store := model.NewExpenseDDBStore(tableName, client)

		expense, err := store.Create(ctx, model.Expense{})
		if err != nil {
			t.Fatalf("failed creating expense: %v", err)
		}

		_, err = store.FindOne(ctx, expense.SK)
		if err != nil {
			t.Fatalf("failed finding expense after creation: %v", err)
		}

		err = store.Delete(ctx, expense.SK)
		if err != nil {
			t.Fatalf("failed deleting expense: %v", err)
		}

		_, err = store.FindOne(ctx, expense.SK)
		if err == nil {
			t.Fatal("expected error after trying to find deleted expense but didn't get one")
		}
		var notFoundErr *model.ExpenseNotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &model.ExpenseNotFoundError{SK: expense.SK})
		}
	})

	t.Run("returns proper error when expense for deletion does not exist", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
		if err != nil {
			t.Fatalf("failed creating local test ddb table, %v", err)
		}
		defer removeDDB()

		invalidSK := "invalidSK"

		store := model.NewExpenseDDBStore(tableName, client)

		err = store.Delete(ctx, invalidSK)
		if err == nil {
			t.Fatal("expected an error but didn't get one")
		}

		var notFoundErr *model.ExpenseNotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &model.ExpenseNotFoundError{SK: invalidSK})
		}
	})
}

func TestUpdateDDBExpense(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
	if err != nil {
		t.Fatalf("failed creating local test ddb table, %v", err)
	}
	defer removeDDB()

	store := model.NewExpenseDDBStore(tableName, client)

	t.Run("updates existing expense", func(t *testing.T) {
		expenseFC := model.Expense{}
		expense, err := store.Create(ctx, expenseFC)
		if err != nil {
			t.Fatalf("failed creating expense: %v", err)
		}

		expense.Name = "new name"
		receivedExpense, err := store.Update(ctx, expense)
		if err != nil {
			t.Fatalf("didn't expect an error while updating expense but got one: %v", err)
		}
		newExpense, _ := store.FindOne(ctx, receivedExpense.SK)

		if newExpense.Name != expense.Name || receivedExpense.Name != expense.Name {
			t.Error("expense update failed")
		}
	})

	t.Run("updates SK in existing expense", func(t *testing.T) {
		expenseFC := model.Expense{}
		expense, err := store.Create(ctx, expenseFC)
		if err != nil {
			t.Fatalf("failed creating expense: %v", err)
		}

		expense.Date = "2024-09-09"
		receivedExpense, err := store.Update(ctx, expense)
		if err != nil {
			t.Fatalf("didn't expect an error while updating expense but got one: %v", err)
		}
		newExpense, _ := store.FindOne(ctx, expense.SK)

		if newExpense.Name != expense.Name || receivedExpense.Name != expense.Name {
			t.Error("expense update failed")
		}
	})

	t.Run("returns proper error when expense for update does not exist", func(t *testing.T) {
		invalidSK := "invalidSK"

		_, err := store.Update(ctx, model.Expense{SK: invalidSK})
		if err == nil {
			t.Fatal("expected an error but didn't get one")
		}

		var notFoundErr *model.ExpenseNotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &model.ExpenseNotFoundError{SK: invalidSK})
		}
	})
}

func TestFindOneDDBExpense(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
	if err != nil {
		t.Fatalf("failed creating local test ddb table, %v", err)
	}
	defer removeDDB()

	store := model.NewExpenseDDBStore(tableName, client)

	t.Run("finds existing expense", func(t *testing.T) {
		expenseFC := model.Expense{}
		expense, err := store.Create(ctx, expenseFC)
		if err != nil {
			t.Fatalf("failed creating expense: %v", err)
		}

		_, err = store.FindOne(ctx, expense.SK)
		if err != nil {
			t.Errorf("didn't expect an error while finding expense but got one: %v", err)
		}
	})

	t.Run("returns proper error when searched expense does not exist", func(t *testing.T) {
		invalidSK := "invalidSK"

		_, err := store.FindOne(ctx, invalidSK)
		if err == nil {
			t.Fatal("expected an error but didn't get one")
		}

		var notFoundErr *model.ExpenseNotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &model.ExpenseNotFoundError{SK: invalidSK})
		}
	})
}
