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

	expense, err := store.Create(ctx, model.Expense{})
	if err != nil {
		t.Fatalf("failed creating expense: %v", err)
	}

	_, err = store.FindOne(ctx, expense.CreatedAt)
	if err != nil {
		t.Errorf("didn't expect an error but got one: %v", err)
	}
}

func TestDeleteInMemoryExpense(t *testing.T) {
	t.Run("deletes existing expense", func(t *testing.T) {
		ctx := context.Background()
		store := model.ExpenseInMemoryStore{}

		expense, err := store.Create(ctx, model.Expense{})
		if err != nil {
			t.Fatalf("failed creating expense: %v", err)
		}

		_, err = store.FindOne(ctx, expense.CreatedAt)
		if err != nil {
			t.Fatalf("failed finding expense after creation: %v", err)
		}

		err = store.Delete(ctx, expense.CreatedAt)
		if err != nil {
			t.Fatalf("failed deleting expense: %v", err)
		}

		_, err = store.FindOne(ctx, expense.CreatedAt)
		if err == nil {
			t.Fatal("expected error after trying to find deleted expense but didn't get one")
		}
		var notFoundErr *model.ExpenseNotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &model.ExpenseNotFoundError{CreatedAt: expense.CreatedAt})
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
	ctx := context.Background()
	store := model.ExpenseInMemoryStore{}
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
		newExpense, _ := store.FindOne(ctx, expense.CreatedAt)

		if newExpense.Name != expense.Name || receivedExpense.Name != expense.Name {
			t.Error("expense update failed")
		}
	})

	t.Run("returns proper error when expense for update does not exist", func(t *testing.T) {
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
	ctx := context.Background()
	store := model.ExpenseInMemoryStore{}
	t.Run("finds existing expense", func(t *testing.T) {
		expenseFC := model.Expense{}
		expense, err := store.Create(ctx, expenseFC)
		if err != nil {
			t.Fatalf("failed creating expense: %v", err)
		}

		_, err = store.FindOne(ctx, expense.CreatedAt)
		if err != nil {
			t.Errorf("didn't expect an error while finding expense but got one: %v", err)
		}
	})

	t.Run("returns proper error when expense for update does not exist", func(t *testing.T) {
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
