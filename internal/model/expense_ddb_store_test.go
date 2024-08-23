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
		t.Fatalf("failed creating local test ddb table, %v", err)
	}
	defer removeDDB()

	store := model.NewExpenseDDBStore(tableName, client)

	expenses, err := store.Query(ctx)
	if err != nil {
		t.Fatalf("failed querying for expenses before putting expense, %v", err)
	}

	err = store.Create(ctx, model.Expense{})
	if err != nil {
		t.Fatalf("failed putting item into ddb, %v", err)
	}
	newExpenses, err := store.Query(ctx)
	if err != nil {
		t.Fatalf("failed querying ddb table for expenses after putting expense, %v", err)
	}
	if (len(newExpenses) - 1) != len(expenses) {
		t.Errorf("expected one new expense added. got %d", len(newExpenses)-len(expenses))
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

		err = store.Create(ctx, model.Expense{})
		if err != nil {
			t.Fatalf("failed putting item into ddb, %v", err)
		}

		expenses, err := store.Query(ctx)
		if err != nil {
			t.Fatalf("failed querying ddb table for expenses before deleting expense, %v", err)
		}

		err = store.Delete(ctx, expenses[0].CreatedAt)
		if err != nil {
			t.Fatalf("failed deleting expense: %v", err)
		}

		newExpenses, err := store.Query(ctx)
		if err != nil {
			t.Fatalf("failed querying ddb table for expenses after deleting expense, %v", err)
		}
		if (len(newExpenses) + 1) != len(expenses) {
			t.Errorf("expected one expense deleted. got %d", len(newExpenses)-len(expenses))
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

		invalidCreatedAt := "asdf"

		store := model.NewExpenseDDBStore(tableName, client)

		err = store.Delete(ctx, invalidCreatedAt)
		if err == nil {
			t.Fatal("expected an error but didn't get one")
		}

		var notFoundErr *model.ExpenseNotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &model.ExpenseNotFoundError{CreatedAt: invalidCreatedAt})
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
		err = store.Create(ctx, expenseFC)
		if err != nil {
			t.Fatalf("failed putting item into ddb, %v", err)
		}

		expenses, err := store.Query(ctx)
		if err != nil {
			t.Fatalf("failed querying ddb table for expenses: %v", err)
		}

		expense := expenses[0]

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
		err = store.Create(ctx, expenseFC)
		if err != nil {
			t.Fatalf("failed putting item into ddb, %v", err)
		}

		expenses, err := store.Query(ctx)
		if err != nil {
			t.Fatalf("failed querying ddb table for expenses: %v", err)
		}
		expense := expenses[0]

		_, err = store.FindOne(ctx, expense.CreatedAt)
		if err != nil {
			t.Errorf("didn't expect an error while finding expense but got one: %v", err)
		}
	})

	t.Run("returns proper error when searched expense does not exist", func(t *testing.T) {
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
