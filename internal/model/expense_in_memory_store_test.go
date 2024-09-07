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

	_, err = store.FindOne(ctx, expense.SK)
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
	store := model.ExpenseInMemoryStore{}
	t.Run("updates existing expense", func(t *testing.T) {
		expenseFC := model.Expense{}
		expense, err := store.Create(ctx, expenseFC)
		if err != nil {
			t.Fatalf("failed creating expense: %v", err)
		}

		expense.Name = "new name"
		err = store.Update(ctx, expense)
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
	store := model.ExpenseInMemoryStore{}
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
