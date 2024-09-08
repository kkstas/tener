package model_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kkstas/tjener/internal/model"
)

const (
	validInMemoryExpenseName     = "Some name"
	validInMemoryExpenseDate     = "2024-09-07"
	validInMemoryExpenseCategory = "Some category"
	validInMemoryExpenseAmount   = 24.99
)

func TestCreateInMemoryExpense(t *testing.T) {
	ctx := context.Background()
	store := &model.ExpenseInMemoryStore{}

	expense := createDefaultInMemoryExpenseHelper(t, ctx, store)

	_, err := store.FindOne(ctx, expense.SK)
	if err != nil {
		t.Errorf("didn't expect an error but got one: %v", err)
	}
}

func TestDeleteInMemoryExpense(t *testing.T) {
	t.Run("deletes existing expense", func(t *testing.T) {
		ctx := context.Background()
		store := &model.ExpenseInMemoryStore{}

		expense := createDefaultInMemoryExpenseHelper(t, ctx, store)

		_, err := store.FindOne(ctx, expense.SK)
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
		ctx := context.Background()
		store := model.ExpenseInMemoryStore{}
		invalidSK := "invalidSK"

		err := store.Delete(ctx, invalidSK)
		if err == nil {
			t.Fatal("expected an error but didn't get one")
		}

		var notFoundErr *model.ExpenseNotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &model.ExpenseNotFoundError{SK: invalidSK})
		}
	})
}

func TestUpdateInMemoryExpense(t *testing.T) {
	ctx := context.Background()
	store := &model.ExpenseInMemoryStore{}
	t.Run("updates existing expense", func(t *testing.T) {
		expense := createDefaultInMemoryExpenseHelper(t, ctx, store)
		expense.Name = "new name"
		err := store.Update(ctx, expense)
		if err != nil {
			t.Fatalf("didn't expect an error while updating expense but got one: %v", err)
		}
		newExpense, _ := store.FindOne(ctx, expense.SK)

		if newExpense.Name != expense.Name {
			t.Error("expense update failed")
		}
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

func TestFindOneInMemoryExpense(t *testing.T) {
	ctx := context.Background()
	store := &model.ExpenseInMemoryStore{}
	t.Run("finds existing expense", func(t *testing.T) {
		expense := createDefaultInMemoryExpenseHelper(t, ctx, store)
		_, err := store.FindOne(ctx, expense.SK)
		if err != nil {
			t.Errorf("didn't expect an error while finding expense but got one: %v", err)
		}
	})

	t.Run("returns proper error when expense for update does not exist", func(t *testing.T) {
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

func TestQueryInMemoryByDateRange(t *testing.T) {
	ctx := context.Background()
	store := &model.ExpenseInMemoryStore{}

	createInMemoryExpenseHelper(t, ctx, store, validInMemoryExpenseName, "2024-01-15", validInMemoryExpenseCategory, validInMemoryExpenseAmount, model.ValidCurrencies[0])
	createInMemoryExpenseHelper(t, ctx, store, validInMemoryExpenseName, "2024-01-16", validInMemoryExpenseCategory, validInMemoryExpenseAmount, model.ValidCurrencies[0])
	createInMemoryExpenseHelper(t, ctx, store, validInMemoryExpenseName, "2024-01-17", validInMemoryExpenseCategory, validInMemoryExpenseAmount, model.ValidCurrencies[0])
	createInMemoryExpenseHelper(t, ctx, store, validInMemoryExpenseName, "2024-01-18", validInMemoryExpenseCategory, validInMemoryExpenseAmount, model.ValidCurrencies[0])

	t.Run("returns expenses that are greater or equal than 'from', and lesser or equal than 'to'", func(t *testing.T) {
		expenses, err := store.QueryByDateRange(ctx, "2024-01-15", "2024-01-18")
		if err != nil {
			t.Fatalf("didn't expect an error while querying by date range, but got one: %v", err)
		}
		if len(expenses) != 4 {
			t.Errorf("expected 4 expenses returned, got %d", len(expenses))
		}

		expenses, err = store.QueryByDateRange(ctx, "2024-01-15", "2024-01-16")
		if err != nil {
			t.Fatalf("didn't expect an error while querying by date range, but got one: %v", err)
		}
		if len(expenses) != 2 {
			t.Errorf("expected 2 expenses returned, got %d", len(expenses))
		}

		expenses, err = store.QueryByDateRange(ctx, "2024-01-15", "2024-01-15")
		if err != nil {
			t.Fatalf("didn't expect an error while querying by date range, but got one: %v", err)
		}
		if len(expenses) != 1 {
			t.Errorf("expected 1 expense returned, got %d", len(expenses))
		}
	})

	t.Run("returns error when date range is above one year", func(t *testing.T) {
		_, err := store.QueryByDateRange(ctx, "2023-01-01", "2024-01-02")
		if err == nil {
			t.Error("expected and error but didn't get one")
		}
	})
}

func createDefaultInMemoryExpenseHelper(t testing.TB, ctx context.Context, store *model.ExpenseInMemoryStore) model.Expense {
	t.Helper()
	return createInMemoryExpenseHelper(t, ctx, store, validInMemoryExpenseName, validInMemoryExpenseDate, validInMemoryExpenseCategory, validInMemoryExpenseAmount, model.ValidCurrencies[0])
}

func createInMemoryExpenseHelper(t testing.TB, ctx context.Context, store *model.ExpenseInMemoryStore, name, date, category string, amount float64, currency string) model.Expense {
	t.Helper()
	expenseFC, err := model.NewExpenseFC(name, date, category, amount, currency)
	if err != nil {
		t.Fatalf("didn't expect an error while creating NewExpenseFC but got one: %v", err)
	}
	expense, err := store.Create(ctx, expenseFC)
	if err != nil {
		t.Fatalf("didn't expect an error while putting expense into in memory store but got one: %v", err)
	}
	return expense
}
