package model_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kkstas/tjener/internal/model"
)

func TestCreateInMemoryExpense(t *testing.T) {
	ctx := context.Background()

	store := model.ExpenseInMemoryStore{}

	expenses, _ := store.Query(ctx)

	err := store.Create(ctx, model.Expense{})
	if err != nil {
		t.Fatalf("failed creating expense, %v", err)
	}
	newExpenses, err := store.Query(ctx)

	if err != nil {
		t.Fatalf("failed querying for expenses after creating expense, %v", err)
	}
	if (len(newExpenses) - 1) != len(expenses) {
		t.Errorf("expected one new expense added. got %d", len(newExpenses)-len(expenses))
	}
}

func TestDeleteInMemoryExpense(t *testing.T) {
	t.Run("deletes existing expense", func(t *testing.T) {
		ctx := context.Background()
		store := model.ExpenseInMemoryStore{}
		createdAt := "2024-08-24T00:12:00.288547471+02:00"

		_ = store.Create(ctx, model.Expense{CreatedAt: createdAt})

		expenses, _ := store.Query(ctx)
		if len(expenses) != 1 {
			t.Fatalf("expected one expense saved in the store, got %#v", expenses)
		}

		err := store.Delete(ctx, "2024-08-24T00:12:00.288547471+02:00")
		if err != nil {
			t.Fatalf("didn't expect an error while deleting expense but got one: %v", err)
		}

		expenses, _ = store.Query(ctx)
		if len(expenses) != 0 {
			t.Errorf("expected 0 expenses after deleting, got %d", len(expenses))
		}
	})

	t.Run("returns proper error when expense for deletion does not exist", func(t *testing.T) {
		ctx := context.Background()
		store := model.ExpenseInMemoryStore{}
		invalidCreatedAt := "invalidCreatedAt"

		err := store.Delete(ctx, invalidCreatedAt)
		if err == nil {
			t.Fatal("expected an error but didn't get one")
		}

		var notFoundErr *model.ExpenseNotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &model.ExpenseNotFoundError{CreatedAt: invalidCreatedAt})
		}
	})
}

func TestUpdateInMemoryExpense(t *testing.T) {
	t.Run("updates existing expense", func(t *testing.T) {
		ctx := context.Background()
		store := model.ExpenseInMemoryStore{}
		createdAt := "2024-08-24T00:12:00.288547471+02:00"

		expense := model.Expense{
			CreatedAt: createdAt,
			Name:      "old name",
		}

		_ = store.Create(ctx, expense)

		expense.Name = "new name"
		receivedExpense, err := store.Update(ctx, expense)
		if err != nil {
			t.Fatalf("didn't expect an error while updating expense but got one: %v", err)
		}
		newExpense, _ := store.FindOne(ctx, createdAt)

		if newExpense.Name != expense.Name || receivedExpense.Name != expense.Name {
			t.Error("expense update failed")
		}
	})

	t.Run("returns proper error when expense for update does not exist", func(t *testing.T) {
		ctx := context.Background()
		store := model.ExpenseInMemoryStore{}
		invalidCreatedAt := "invalidCreatedAt"

		_, err := store.Update(ctx, model.Expense{CreatedAt: invalidCreatedAt})
		if err == nil {
			t.Fatal("expected an error but didn't get one")
		}

		var notFoundErr *model.ExpenseNotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &model.ExpenseNotFoundError{CreatedAt: invalidCreatedAt})
		}
	})
}

func TestFindOneInMemoryExpense(t *testing.T) {
	t.Run("finds existing expense", func(t *testing.T) {
		ctx := context.Background()
		store := model.ExpenseInMemoryStore{}
		createdAt := "2024-08-24T00:12:00.288547471+02:00"

		expense := model.Expense{
			CreatedAt: createdAt,
			Name:      "old name",
		}

		_ = store.Create(ctx, expense)
		_, err := store.FindOne(ctx, createdAt)

		if err != nil {
			t.Errorf("didn't expect error but got one: %v", err)
		}
	})

	t.Run("returns proper error when expense for update does not exist", func(t *testing.T) {
		ctx := context.Background()
		store := model.ExpenseInMemoryStore{}
		invalidCreatedAt := "invalidCreatedAt"

		_, err := store.FindOne(ctx, invalidCreatedAt)
		if err == nil {
			t.Fatal("expected an error but didn't get one")
		}

		var notFoundErr *model.ExpenseNotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &model.ExpenseNotFoundError{CreatedAt: invalidCreatedAt})
		}
	})
}
