package user_test

import (
	"context"
	"testing"
	"time"

	"github.com/kkstas/tjener/internal/database"
	"github.com/kkstas/tjener/internal/model/user"
)

func TestDDBCreate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
	if err != nil {
		t.Fatalf("failed creating local test ddb table, %v", err)
	}
	defer removeDDB()
	store := user.NewDDBStore(tableName, client)

	t.Run("creates new user", func(t *testing.T) {
		usersBefore, err := store.FindAll(ctx)
		assertNoError(t, err)

		userFC, err := user.New(validFirstName, validLastName, "johndoe123@email.com", validPassword)
		assertNoError(t, err)

		_, err = store.Create(ctx, userFC)
		assertNoError(t, err)

		usersAfter, err := store.FindAll(ctx)
		assertNoError(t, err)

		if len(usersBefore)+1 != len(usersAfter) {
			t.Errorf("expected one user more after creating user than before, got %d", len(usersAfter)-len(usersBefore))
		}
	})

	t.Run("does not create new user if email is already taken", func(t *testing.T) {
		userFC, err := user.New(validFirstName, validLastName, validEmail, validPassword)
		assertNoError(t, err)

		_, err = store.Create(ctx, userFC)
		assertNoError(t, err)

		secondUserFC, err := user.New(validFirstName, validLastName, validEmail, validPassword)
		assertNoError(t, err)

		_, err = store.Create(ctx, secondUserFC)

		if err == nil {
			t.Error("expected an error but didn't get one")
		}
	})

	t.Run("does not create new user if user with that ID already exists", func(t *testing.T) {
		userFC, err := user.New(validFirstName, validLastName, "john847234doe@email.de", validPassword)
		assertNoError(t, err)

		_, err = store.Create(ctx, userFC)
		assertNoError(t, err)

		secondUserFC, err := user.New(validFirstName, validLastName, "doe123842doe@email.eu", validPassword)
		assertNoError(t, err)
		secondUserFC.ID = userFC.ID

		_, err = store.Create(ctx, secondUserFC)

		if err == nil {
			t.Error("expected an error but didn't get one")
		}
	})
}

func TestDDBFindOneByEmail(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
	if err != nil {
		t.Fatalf("failed creating local test ddb table, %v", err)
	}
	defer removeDDB()
	store := user.NewDDBStore(tableName, client)

	t.Run("finds created user by email", func(t *testing.T) {
		userFC, err := user.New(validFirstName, validLastName, validEmail, validPassword)
		assertNoError(t, err)

		createdUser, err := store.Create(ctx, userFC)
		assertNoError(t, err)

		_, err = store.FindOneByEmail(ctx, createdUser.Email)
		assertNoError(t, err)
	})
}

func TestDDBFindOneByID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
	if err != nil {
		t.Fatalf("failed creating local test ddb table, %v", err)
	}
	defer removeDDB()
	store := user.NewDDBStore(tableName, client)

	t.Run("finds created user by ID", func(t *testing.T) {
		userFC, err := user.New(validFirstName, validLastName, validEmail, validPassword)
		assertNoError(t, err)

		createdUser, err := store.Create(ctx, userFC)
		assertNoError(t, err)

		_, err = store.FindOneByID(ctx, createdUser.ID)
		assertNoError(t, err)
	})
}

func assertNoError(t testing.TB, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("got an error but didn't expect one: %v", err)
	}
}
