package model_test

import (
	"context"
	"errors"
	"strings"
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

	expenseName := "Some name"
	expenseDate := "2024-09-07"
	expenseCategory := "Some category"
	expenseAmount := 24.99

	expenseFC, err := model.NewExpenseFC(expenseName, expenseDate, expenseCategory, expenseAmount, model.ValidCurrencies[0])
	if err != nil {
		t.Fatalf("didn't expect an error but got one: %v", err)
	}

	expense, err := store.Create(ctx, expenseFC)
	if err != nil {
		t.Fatalf("failed putting item into ddb, %v", err)
	}

	foundExpense, err := store.FindOne(ctx, expense.SK)
	if err != nil {
		t.Errorf("didn't expect an error but got one: %v", err)
	}

	t.Run("creates new expense with correct data", func(t *testing.T) {
		assertEqual(t, foundExpense.Name, expenseName)
		assertEqual(t, strings.HasPrefix(foundExpense.SK, expenseDate), true)
		assertEqual(t, foundExpense.Date, expenseDate)
		assertEqual(t, foundExpense.Category, expenseCategory)
		assertEqual(t, foundExpense.Amount, expenseAmount)
		assertValidTime(t, time.RFC3339Nano, foundExpense.CreatedAt)
	})

	t.Run("creates expense with SK that consists of Date and CreatedAt values", func(t *testing.T) {
		split := strings.Split(foundExpense.SK, "::")
		date, createdAt := split[0], split[1]
		assertEqual(t, date, foundExpense.Date)
		assertEqual(t, createdAt, foundExpense.CreatedAt)
	})
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
		newExpense, err := store.FindOne(ctx, receivedExpense.SK)
		if err != nil {
			t.Fatalf("didn't expect an error while updating expense but got one: %v", err)
		}

		if newExpense.Name != expense.Name || receivedExpense.Name != expense.Name {
			t.Error("expense update failed")
		}
	})

	t.Run("assigns Date as first part of SK and keeps CreatedAt as second part when Date is updated", func(t *testing.T) {
		expenseFC := model.Expense{Date: "2024-04-04"}
		expense, err := store.Create(ctx, expenseFC)
		if err != nil {
			t.Fatalf("failed creating expense: %v", err)
		}

		newDate := "2024-09-09"
		expense.Date = newDate
		receivedExpense, err := store.Update(ctx, expense)
		if err != nil {
			t.Fatalf("didn't expect an error while updating expense but got one: %v", err)
		}
		newExpense, err := store.FindOne(ctx, receivedExpense.SK)
		if err != nil {
			t.Fatalf("didn't expect an error while searching for expense but got one: %v", err)
		}

		split := strings.Split(newExpense.SK, "::")
		dateFromSK, createdAtFromSK := split[0], split[1]

		assertEqual(t, dateFromSK, newDate)
		assertEqual(t, createdAtFromSK, expense.CreatedAt)
	})

	t.Run("keep the same SK when Date is not changed", func(t *testing.T) {
		expenseFC := model.Expense{Date: "2024-04-04", Name: "name"}
		expense, err := store.Create(ctx, expenseFC)
		if err != nil {
			t.Fatalf("failed creating expense: %v", err)
		}

		newName := "new name"
		expense.Name = newName
		receivedExpense, err := store.Update(ctx, expense)
		if err != nil {
			t.Fatalf("didn't expect an error while updating expense but got one: %v", err)
		}
		newExpense, err := store.FindOne(ctx, receivedExpense.SK)
		if err != nil {
			t.Fatalf("didn't expect an error while searching for expense but got one: %v", err)
		}

		assertEqual(t, newExpense.SK, expense.SK)
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

func assertEqual[T comparable](t testing.TB, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func assertValidTime(t testing.TB, layout, datestring string) {
	t.Helper()
	_, err := time.Parse(layout, datestring)
	if err != nil {
		t.Errorf("string '%s' is not valid datetime: %v", datestring, err)
	}
}
