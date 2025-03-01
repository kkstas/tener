package expense_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kkstas/tener/internal/model/expense"
)

const (
	validInMemoryExpenseName      = "Some name"
	validInMemoryExpenseDate      = "2024-09-07"
	validInMemoryExpenseCategory  = "Some category"
	validInMemoryExpenseCategory2 = "Other category"
	validInMemoryExpenseAmount    = 24.99
)

func TestInMemoryCreate(t *testing.T) {
	ctx := context.Background()
	store := &expense.InMemoryStore{}

	expense := createDefaultInMemoryExpenseHelper(t, ctx, store)

	_, err := store.FindOne(ctx, expense.SK, "activeVaultID")
	if err != nil {
		t.Errorf("didn't expect an error but got one: %v", err)
	}
}

func TestInMemoryDelete(t *testing.T) {
	t.Run("deletes existing expense", func(t *testing.T) {
		ctx := context.Background()
		store := &expense.InMemoryStore{}

		exp := createDefaultInMemoryExpenseHelper(t, ctx, store)

		_, err := store.FindOne(ctx, exp.SK, "activeVaultID")
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
		ctx := context.Background()
		store := expense.InMemoryStore{}
		invalidSK := "invalidSK"

		err := store.Delete(ctx, invalidSK, "activeVaultID")
		if err == nil {
			t.Fatal("expected an error but didn't get one")
		}

		var notFoundErr *expense.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &expense.NotFoundError{SK: invalidSK})
		}
	})
}

func TestInMemoryUpdate(t *testing.T) {
	ctx := context.Background()
	store := &expense.InMemoryStore{}
	t.Run("updates existing expense", func(t *testing.T) {
		expense := createDefaultInMemoryExpenseHelper(t, ctx, store)
		expense.Name = "new name"
		err := store.Update(ctx, expense, "activeVaultID")
		if err != nil {
			t.Fatalf("didn't expect an error while updating expense but got one: %v", err)
		}
		newExpense, _ := store.FindOne(ctx, expense.SK, "activeVaultID")

		if newExpense.Name != expense.Name {
			t.Error("expense update failed")
		}
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

func TestInMemoryFindOne(t *testing.T) {
	ctx := context.Background()
	store := &expense.InMemoryStore{}
	t.Run("finds existing expense", func(t *testing.T) {
		expense := createDefaultInMemoryExpenseHelper(t, ctx, store)
		_, err := store.FindOne(ctx, expense.SK, "activeVaultID")
		if err != nil {
			t.Errorf("didn't expect an error while finding expense but got one: %v", err)
		}
	})

	t.Run("returns proper error when expense for update does not exist", func(t *testing.T) {
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

func TestInMemoryQuery(t *testing.T) {
	ctx := context.Background()
	store := &expense.InMemoryStore{}

	createInMemoryExpenseHelper(
		t,
		ctx,
		store,
		validInMemoryExpenseName,
		"2024-01-15",
		validInMemoryExpenseCategory,
		validInMemoryExpenseAmount,
		expense.PaymentMethods[0],
	)
	createInMemoryExpenseHelper(
		t,
		ctx,
		store,
		validInMemoryExpenseName,
		"2024-01-16",
		validInMemoryExpenseCategory2,
		validInMemoryExpenseAmount,
		expense.PaymentMethods[0],
	)
	createInMemoryExpenseHelper(
		t,
		ctx,
		store,
		validInMemoryExpenseName,
		"2024-01-17",
		validInMemoryExpenseCategory2,
		validInMemoryExpenseAmount,
		expense.PaymentMethods[0],
	)
	createInMemoryExpenseHelper(
		t,
		ctx,
		store,
		validInMemoryExpenseName,
		"2024-01-18",
		validInMemoryExpenseCategory,
		validInMemoryExpenseAmount,
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
			[]string{validInMemoryExpenseCategory, validInMemoryExpenseCategory2},
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
			[]string{validInMemoryExpenseCategory},
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

func createDefaultInMemoryExpenseHelper(t testing.TB, ctx context.Context, store *expense.InMemoryStore) expense.Expense {
	t.Helper()
	return createInMemoryExpenseHelper(
		t,
		ctx,
		store,
		validInMemoryExpenseName,
		validInMemoryExpenseDate,
		validInMemoryExpenseCategory,
		validInMemoryExpenseAmount,
		expense.PaymentMethods[0],
	)
}

func createInMemoryExpenseHelper(
	t testing.TB,
	ctx context.Context,
	store *expense.InMemoryStore,
	name,
	date,
	category string,
	amount float64,
	paymentMethod string,
) expense.Expense {
	t.Helper()
	expenseFC, isValid, errMessages := expense.New(name, date, category, amount, paymentMethod)
	if !isValid {
		t.Fatalf("didn't expect an error while creating NewExpenseFC but got one: %v", errMessages)
	}
	expense, err := store.Create(ctx, expenseFC, "userID", "activeVaultID")
	if err != nil {
		t.Fatalf("didn't expect an error while putting expense into in memory store but got one: %v", err)
	}
	return expense
}
