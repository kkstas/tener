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

const (
	validExpenseName     = "Some name"
	validExpenseDate     = "2024-09-07"
	validExpenseCategory = "Some category"
	validExpenseAmount   = 24.99
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

	expense := createDefaultExpenseHelper(t, ctx, store)

	foundExpense, err := store.FindOne(ctx, expense.SK)
	if err != nil {
		t.Fatalf("didn't expect an error but got one: %v", err)
	}

	t.Run("creates new expense with correct data", func(t *testing.T) {
		assertEqual(t, foundExpense.Name, validExpenseName)
		assertEqual(t, strings.HasPrefix(foundExpense.SK, validExpenseDate), true)
		assertEqual(t, foundExpense.Date, validExpenseDate)
		assertEqual(t, foundExpense.Category, validExpenseCategory)
		assertEqual(t, foundExpense.Amount, validExpenseAmount)
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
		expense := createDefaultExpenseHelper(t, ctx, store)

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
		expense := createDefaultExpenseHelper(t, ctx, store)

		expense.Name = validExpenseName
		err = store.Update(ctx, expense)
		if err != nil {
			t.Fatalf("didn't expect an error while updating expense but got one: %v", err)
		}
		newExpense, err := store.FindOne(ctx, expense.SK)
		if err != nil {
			t.Fatalf("didn't expect an error while updating expense but got one: %v", err)
		}

		if newExpense.Name != expense.Name {
			t.Error("expense update failed")
		}
	})

	t.Run("assigns Date as first part of SK and keeps CreatedAt as second part when Date is updated", func(t *testing.T) {
		expense := createDefaultExpenseHelper(t, ctx, store)
		newDate := "2024-09-09"
		expense.Date = newDate
		err = store.Update(ctx, expense)
		if err != nil {
			t.Fatalf("didn't expect an error while updating expense but got one: %v", err)
		}

		newExpense, err := store.FindOne(ctx, expense.Date+"::"+expense.CreatedAt)
		if err != nil {
			t.Fatalf("didn't expect an error while searching for expense but got one: %v", err)
		}

		split := strings.Split(newExpense.SK, "::")
		dateFromSK, createdAtFromSK := split[0], split[1]

		assertEqual(t, dateFromSK, newDate)
		assertEqual(t, createdAtFromSK, expense.CreatedAt)
		assertEqual(t, newExpense.CreatedAt, expense.CreatedAt)
	})

	t.Run("keep the same SK when Date is not changed", func(t *testing.T) {
		expense := createDefaultExpenseHelper(t, ctx, store)
		newName := "new name"
		expense.Name = newName
		err = store.Update(ctx, expense)
		if err != nil {
			t.Fatalf("didn't expect an error while updating expense but got one: %v", err)
		}
		newExpense, err := store.FindOne(ctx, expense.Date+"::"+expense.CreatedAt)
		if err != nil {
			t.Fatalf("didn't expect an error while searching for expense but got one: %v", err)
		}

		assertEqual(t, newExpense.SK, expense.SK)
		assertEqual(t, newExpense.Name, newName)
	})

	t.Run("returns proper error when expense for update does not exist", func(t *testing.T) {
		invalidSK := "invalidSK"

		err := store.Update(ctx, model.Expense{SK: invalidSK})
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
		expense := createDefaultExpenseHelper(t, ctx, store)

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

func createDefaultExpenseHelper(t testing.TB, ctx context.Context, store *model.ExpenseDDBStore) model.Expense {
	t.Helper()
	return createExpenseHelper(t, ctx, store, validExpenseName, validExpenseDate, validExpenseCategory, validExpenseAmount, model.ValidCurrencies[0])
}

func createExpenseHelper(t testing.TB, ctx context.Context, store *model.ExpenseDDBStore, name, date, category string, amount float64, currency string) model.Expense {
	t.Helper()
	expenseFC, err := model.NewExpenseFC(name, date, category, amount, currency)
	if err != nil {
		t.Fatalf("didn't expect an error while creating NewExpenseFC but got one: %v", err)
	}
	expense, err := store.Create(ctx, expenseFC)
	if err != nil {
		t.Fatalf("didn't expect an error while putting expense into DDB but got one: %v", err)
	}
	return expense
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
