package user_test

import (
	"context"
	"errors"
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

	t.Run("does not create new user if user with that email already exists", func(t *testing.T) {
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

		var alreadyExistsErr *user.AlreadyExistsError
		if ok := errors.As(err, &alreadyExistsErr); !ok {
			t.Errorf("expected AlreadyExistsError thrown, but instead got: %#v", err)
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
		var alreadyExistsErr *user.AlreadyExistsError
		if ok := errors.As(err, &alreadyExistsErr); !ok {
			t.Errorf("expected AlreadyExistsError thrown, but instead got: %#v", err)
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

	t.Run("returns NotFoundError if user with that email does not exist", func(t *testing.T) {
		_, err = store.FindOneByEmail(ctx, "invalidemail@email.com")

		if err == nil {
			t.Error("expected an error but didn't get one")
		}

		var notFoundErr *user.NotFoundError
		if ok := errors.As(err, &notFoundErr); !ok {
			t.Errorf("expected NotFoundError thrown, but instead got: %#v", err)
		}
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

	t.Run("returns NotFoundError if user with that ID does not exist", func(t *testing.T) {
		_, err = store.FindOneByID(ctx, "invalidID")

		if err == nil {
			t.Error("expected an error but didn't get one")
		}

		var notFoundErr *user.NotFoundError
		if ok := errors.As(err, &notFoundErr); !ok {
			t.Errorf("expected NotFoundError thrown, but instead got: %#v", err)
		}
	})
}

func TestDDBFindAllByIDs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
	if err != nil {
		t.Fatalf("failed creating local test ddb table, %v", err)
	}
	defer removeDDB()
	store := user.NewDDBStore(tableName, client)

	t.Run("finds users by IDs", func(t *testing.T) {
		userFC, err := user.New(validFirstName, validLastName, validEmail, validPassword)
		assertNoError(t, err)
		createdUser, err := store.Create(ctx, userFC)
		assertNoError(t, err)

		userFC2, err := user.New(validFirstName, validLastName, "howdy@howdy.com", validPassword)
		assertNoError(t, err)
		createdUser2, err := store.Create(ctx, userFC2)
		assertNoError(t, err)

		res, err := store.FindAllByIDs(ctx, []string{createdUser.ID, createdUser2.ID})
		assertNoError(t, err)

		if len(res) != 2 {
			t.Errorf("expected response length of 2, got %d", len(res))
		}
	})

	t.Run("returns error if received empty user ID slice", func(t *testing.T) {
		_, err := store.FindAllByIDs(ctx, []string{})
		if err == nil {
			t.Error("expected an error but didn't get one")
		}
	})
}

func TestDDBDelete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName, client, removeDDB, err := database.CreateLocalTestDDBTable(ctx)
	if err != nil {
		t.Fatalf("failed creating local test ddb table, %v", err)
	}
	defer removeDDB()
	store := user.NewDDBStore(tableName, client)

	t.Run("returns NotFoundError when there is no user with that ID", func(t *testing.T) {
		err := store.Delete(ctx, "invalidID")

		if err == nil {
			t.Error("expected an error but didn't get one")
		}

		var notFoundErr *user.NotFoundError
		if ok := errors.As(err, &notFoundErr); !ok {
			t.Errorf("expected NotFoundError thrown, but instead got: %#v", err)
		}
	})

	t.Run("deletes created user", func(t *testing.T) {
		userFC, err := user.New(validFirstName, validLastName, validEmail, validPassword)
		assertNoError(t, err)

		createdUser, err := store.Create(ctx, userFC)
		assertNoError(t, err)

		_, err = store.FindOneByID(ctx, createdUser.ID)
		assertNoError(t, err)

		err = store.Delete(ctx, createdUser.ID)
		assertNoError(t, err)

		_, err = store.FindOneByID(ctx, createdUser.ID)
		if err == nil {
			t.Error("expected an error but didn't get one")
		}
		var notFoundErr *user.NotFoundError
		if ok := errors.As(err, &notFoundErr); !ok {
			t.Errorf("expected NotFoundError thrown, but instead got: %#v", err)
		}
	})
}

func assertNoError(t testing.TB, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("got an error but didn't expect one: %v", err)
	}
}
