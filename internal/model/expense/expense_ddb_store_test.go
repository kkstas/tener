package expense_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/kkstas/tjener/internal/database"
	"github.com/kkstas/tjener/internal/model/expense"
)

const (
	validDDBExpenseName      = "Some name"
	validDDBExpenseDate      = "2024-09-07"
	validDDBExpenseCategory  = "Some category"
	validDDBExpenseCategory2 = "Other category"
	validDDBExpenseAmount    = 24.99
)

func TestDDBCreate(t *testing.T) {
	vaultID := "activeVaultID"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
	if err != nil {
		t.Fatalf("failed creating local test ddb table: %v", err)
	}
	defer removeDDB()
	store := expense.NewDDBStore(tableName, client)

	expense := createDefaultDDBExpenseHelper(t, ctx, store)

	foundExpense, err := store.FindOne(ctx, expense.SK, vaultID)
	if err != nil {
		t.Fatalf("didn't expect an error but got one: %v", err)
	}

	t.Run("creates new expense with correct data", func(t *testing.T) {
		assertEqual(t, foundExpense.Name, validDDBExpenseName)
		assertEqual(t, strings.HasPrefix(foundExpense.SK, validDDBExpenseDate), true)
		assertEqual(t, foundExpense.Date, validDDBExpenseDate)
		assertEqual(t, foundExpense.Category, validDDBExpenseCategory)
		assertEqual(t, foundExpense.Amount, validDDBExpenseAmount)
		assertValidTime(t, time.RFC3339Nano, foundExpense.CreatedAt)
	})

	t.Run("creates expense with SK that consists of Date and CreatedAt values", func(t *testing.T) {
		split := strings.Split(foundExpense.SK, "::")
		date, createdAt := split[0], split[1]
		assertEqual(t, date, foundExpense.Date)
		assertEqual(t, createdAt, foundExpense.CreatedAt)
	})

	t.Run("creates monthly sum for given month & category", func(t *testing.T) {
		monthlySums, err := store.GetMonthlySums(ctx, 100, vaultID)
		if err != nil {
			t.Fatalf("didn't expect an error but got one: %v", err)
		}

		want := 1
		got := len(monthlySums)
		if got != want {
			t.Errorf("expected %d monthly sum(s), got %d", want, got)
		}
	})
}

func TestDDBDelete(t *testing.T) {
	t.Run("deletes existing expense", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
		if err != nil {
			t.Fatalf("failed creating local test ddb table, %v", err)
		}
		defer removeDDB()

		store := expense.NewDDBStore(tableName, client)
		exp := createDefaultDDBExpenseHelper(t, ctx, store)

		_, err = store.FindOne(ctx, exp.SK, "activeVaultID")
		if err != nil {
			t.Fatalf("failed finding expense after creation: %v", err)
		}

		err = store.Delete(ctx, exp.SK, "activeVaultID")
		if err != nil {
			t.Fatalf("failed deleting expense: %v", err)
		}

		_, err = store.FindOne(ctx, exp.SK, "activeVaultID")
		if err == nil {
			t.Fatal("expected error after trying to find deleted expense but didn't get one")
		}
		var notFoundErr *expense.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &expense.NotFoundError{SK: exp.SK})
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

		store := expense.NewDDBStore(tableName, client)

		err = store.Delete(ctx, invalidSK, "activeVaultID")
		if err == nil {
			t.Fatal("expected an error but didn't get one")
		}

		var notFoundErr *expense.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &expense.NotFoundError{SK: invalidSK})
		}
	})
}

func TestDDBUpdate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
	if err != nil {
		t.Fatalf("failed creating local test ddb table, %v", err)
	}
	defer removeDDB()

	store := expense.NewDDBStore(tableName, client)

	t.Run("updates existing expense", func(t *testing.T) {
		expense := createDefaultDDBExpenseHelper(t, ctx, store)

		expense.Name = validDDBExpenseName
		err = store.Update(ctx, expense, "activeVaultID")
		if err != nil {
			t.Fatalf("didn't expect an error while updating expense but got one: %v", err)
		}
		newExpense, err := store.FindOne(ctx, expense.SK, "activeVaultID")
		if err != nil {
			t.Fatalf("didn't expect an error while updating expense but got one: %v", err)
		}

		if newExpense.Name != expense.Name {
			t.Error("expense update failed")
		}
	})

	t.Run("assigns Date as first part of SK and keeps CreatedAt as second part when Date is updated", func(t *testing.T) {
		expense := createDefaultDDBExpenseHelper(t, ctx, store)
		newDate := "2024-09-09"
		expense.Date = newDate
		err = store.Update(ctx, expense, "activeVaultID")
		if err != nil {
			t.Fatalf("didn't expect an error while updating expense but got one: %v", err)
		}

		newExpense, err := store.FindOne(ctx, expense.Date+"::"+expense.CreatedAt, "activeVaultID")
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
		expense := createDefaultDDBExpenseHelper(t, ctx, store)
		newName := "new name"
		expense.Name = newName
		err = store.Update(ctx, expense, "activeVaultID")
		if err != nil {
			t.Fatalf("didn't expect an error while updating expense but got one: %v", err)
		}
		newExpense, err := store.FindOne(ctx, expense.Date+"::"+expense.CreatedAt, "activeVaultID")
		if err != nil {
			t.Fatalf("didn't expect an error while searching for expense but got one: %v", err)
		}

		assertEqual(t, newExpense.SK, expense.SK)
		assertEqual(t, newExpense.Name, newName)
	})

	t.Run("returns proper error when expense for update does not exist", func(t *testing.T) {
		invalidSK := "invalidSK"

		err := store.Update(ctx, expense.Expense{SK: invalidSK}, "activeVaultID")
		if err == nil {
			t.Fatal("expected an error but didn't get one")
		}

		var notFoundErr *expense.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &expense.NotFoundError{SK: invalidSK})
		}
	})
}

func TestDDBFindOne(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
	if err != nil {
		t.Fatalf("failed creating local test ddb table, %v", err)
	}
	defer removeDDB()

	store := expense.NewDDBStore(tableName, client)

	t.Run("finds existing expense", func(t *testing.T) {
		expense := createDefaultDDBExpenseHelper(t, ctx, store)

		_, err = store.FindOne(ctx, expense.SK, "activeVaultID")
		if err != nil {
			t.Errorf("didn't expect an error while finding expense but got one: %v", err)
		}
	})

	t.Run("returns proper error when searched expense does not exist", func(t *testing.T) {
		invalidSK := "invalidSK"

		_, err := store.FindOne(ctx, invalidSK, "activeVaultID")
		if err == nil {
			t.Fatal("expected an error but didn't get one")
		}

		var notFoundErr *expense.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &expense.NotFoundError{SK: invalidSK})
		}
	})
}

func TestDDBQuery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
	if err != nil {
		t.Fatalf("failed creating local test ddb table, %v", err)
	}
	defer removeDDB()
	store := expense.NewDDBStore(tableName, client)

	createDDBExpenseHelper(t,
		ctx,
		store,
		validDDBExpenseName,
		"2024-01-15",
		validDDBExpenseCategory,
		validDDBExpenseAmount,
		expense.PaymentMethods[0],
	)
	createDDBExpenseHelper(t,
		ctx,
		store,
		validDDBExpenseName,
		"2024-01-16",
		validDDBExpenseCategory2,
		validDDBExpenseAmount,
		expense.PaymentMethods[0],
	)
	createDDBExpenseHelper(t,
		ctx,
		store,
		validDDBExpenseName,
		"2024-01-17",
		validDDBExpenseCategory2,
		validDDBExpenseAmount,
		expense.PaymentMethods[0],
	)
	createDDBExpenseHelper(t,
		ctx,
		store,
		validDDBExpenseName,
		"2024-01-18",
		validDDBExpenseCategory,
		validDDBExpenseAmount,
		expense.PaymentMethods[0],
	)

	t.Run("returns expenses that are greater or equal than 'from', and lesser or equal than 'to'", func(t *testing.T) {
		expenses, err := store.Query(ctx, "2024-01-15", "2024-01-18", []string{}, "activeVaultID")
		if err != nil {
			t.Fatalf("didn't expect an error while querying by date range, but got one: %v", err)
		}
		if len(expenses) != 4 {
			t.Errorf("expected 4 expenses returned, got %d", len(expenses))
		}

		expenses, err = store.Query(ctx, "2024-01-15", "2024-01-16", []string{}, "activeVaultID")
		if err != nil {
			t.Fatalf("didn't expect an error while querying by date range, but got one: %v", err)
		}
		if len(expenses) != 2 {
			t.Errorf("expected 2 expenses returned, got %d", len(expenses))
		}

		expenses, err = store.Query(ctx, "2024-01-15", "2024-01-15", []string{}, "activeVaultID")
		if err != nil {
			t.Fatalf("didn't expect an error while querying by date range, but got one: %v", err)
		}
		if len(expenses) != 1 {
			t.Errorf("expected 1 expense returned, got %d", len(expenses))
		}
	})

	t.Run("returns error when date range is above one year", func(t *testing.T) {
		_, err := store.Query(ctx, "2023-01-01", "2024-01-02", []string{}, "activeVaultID")
		if err == nil {
			t.Error("expected and error but didn't get one")
		}
	})

	t.Run("returns expenses with filtered categories", func(t *testing.T) {
		expenses, err := store.Query(ctx,
			"2024-01-15",
			"2024-01-18",
			[]string{validDDBExpenseCategory, validDDBExpenseCategory2},
			"activeVaultID")
		if err != nil {
			t.Fatalf("didn't expect an error while querying, but got one: %v", err)
		}
		got := len(expenses)
		want := 4
		if got != want {
			t.Errorf("expected %d expenses returned, got %d", want, got)
		}

		expenses, err = store.Query(ctx,
			"2024-01-15",
			"2024-01-18",
			[]string{validDDBExpenseCategory},
			"activeVaultID")
		if err != nil {
			t.Fatalf("didn't expect an error while querying, but got one: %v", err)
		}
		got = len(expenses)
		want = 2
		if got != want {
			t.Errorf("expected %d expenses returned, got %d", want, got)
		}
	})
}

func createDefaultDDBExpenseHelper(t testing.TB, ctx context.Context, store *expense.DDBStore) expense.Expense {
	t.Helper()
	return createDDBExpenseHelper(
		t,
		ctx,
		store,
		validDDBExpenseName,
		validDDBExpenseDate,
		validDDBExpenseCategory,
		validDDBExpenseAmount,
		expense.PaymentMethods[0],
	)
}

func createDDBExpenseHelper(
	t testing.TB,
	ctx context.Context,
	store *expense.DDBStore,
	name,
	date,
	category string,
	amount float64,
	paymentMethod string,
) expense.Expense {
	t.Helper()
	expenseFC, err := expense.New(name, date, category, amount, paymentMethod)
	if err != nil {
		t.Fatalf("didn't expect an error while creating NewExpenseFC but got one: %v", err)
	}
	expense, err := store.Create(ctx, expenseFC, "userID", "activeVaultID")
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
