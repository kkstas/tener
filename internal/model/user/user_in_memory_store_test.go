package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kkstas/tener/internal/model/user"
)

func TestInMemoryCreate(t *testing.T) {
	ctx := context.Background()
	store := &user.InMemoryStore{}

	someUser := createDefaultInMemoryUserHelper(t, ctx, store)

	_, err := store.FindOneByID(ctx, someUser.ID)
	if err != nil {
		t.Errorf("didn't expect an error but got one: %v", err)
	}
}

func TestInMemoryDelete(t *testing.T) {
	t.Run("deletes existing user", func(t *testing.T) {
		ctx := context.Background()
		store := &user.InMemoryStore{}

		someUser := createDefaultInMemoryUserHelper(t, ctx, store)

		_, err := store.FindOneByID(ctx, someUser.ID)
		if err != nil {
			t.Fatalf("failed finding user after creation: %v", err)
		}

		err = store.Delete(ctx, someUser.ID)
		if err != nil {
			t.Fatalf("failed deleting user: %v", err)
		}

		_, err = store.FindOneByID(ctx, someUser.ID)
		if err == nil {
			t.Fatal("expected error after trying to find deleted user but didn't get one")
		}
		var notFoundErr *user.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &user.NotFoundError{ID: someUser.ID})
		}
	})

	t.Run("returns proper error when user for deletion does not exist", func(t *testing.T) {
		ctx := context.Background()
		store := user.InMemoryStore{}
		invalidID := "invalidID"

		err := store.Delete(ctx, invalidID)
		if err == nil {
			t.Fatal("expected an error but didn't get one")
		}

		var notFoundErr *user.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &user.NotFoundError{ID: invalidID})
		}
	})
}

func TestInMemoryUpdate(t *testing.T) {
	ctx := context.Background()
	store := &user.InMemoryStore{}
	t.Run("updates existing user", func(t *testing.T) {
		someUser := createDefaultInMemoryUserHelper(t, ctx, store)
		someUser.FirstName = "newname"
		err := store.Update(ctx, someUser)
		if err != nil {
			t.Fatalf("didn't expect an error while updating user but got one: %v", err)
		}
		updatedUser, _ := store.FindOneByID(ctx, someUser.ID)

		if updatedUser.FirstName != someUser.FirstName {
			t.Error("user update failed")
		}
	})

	t.Run("returns proper error when user for update does not exist", func(t *testing.T) {
		invalidID := "invalidID"

		err := store.Update(ctx, user.User{ID: invalidID})
		if err == nil {
			t.Fatal("expected an error but didn't get one")
		}

		var notFoundErr *user.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &user.NotFoundError{ID: invalidID})
		}
	})
}

func TestInMemoryFindOneByID(t *testing.T) {
	ctx := context.Background()
	store := &user.InMemoryStore{}
	t.Run("finds existing user", func(t *testing.T) {
		someUser := createDefaultInMemoryUserHelper(t, ctx, store)
		_, err := store.FindOneByID(ctx, someUser.ID)
		if err != nil {
			t.Errorf("didn't expect an error while finding user but got one: %v", err)
		}
	})

	t.Run("returns proper error when user for update does not exist", func(t *testing.T) {
		invalidID := "invalidID"

		_, err := store.FindOneByID(ctx, invalidID)
		if err == nil {
			t.Fatal("expected an error but didn't get one")
		}

		var notFoundErr *user.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &user.NotFoundError{ID: invalidID})
		}
	})
}

func TestInMemoryFindAllByIDs(t *testing.T) {
	ctx := context.Background()
	store := &user.InMemoryStore{}
	t.Run("finds created users by IDs", func(t *testing.T) {
		userFC, isValid, _ := user.New(validFirstName, validLastName, validEmail, validPassword)
		assertEqual(t, true, isValid)
		createdUser, err := store.Create(ctx, userFC)
		assertNoError(t, err)

		userFC2, isValid, _ := user.New(validFirstName, validLastName, "howdy@howdy.com", validPassword)
		assertEqual(t, true, isValid)
		createdUser2, err := store.Create(ctx, userFC2)
		assertNoError(t, err)

		res, err := store.FindAllByIDs(ctx, []string{createdUser.ID, createdUser2.ID})
		assertNoError(t, err)

		if len(res) != 2 {
			t.Errorf("expected response length of 2, got %d", len(res))
		}
	})

	t.Run("returns no error if received empty user ID slice", func(t *testing.T) {
		_, err := store.FindAllByIDs(ctx, []string{})
		if err != nil {
			t.Error("didn't expect an error but got one")
		}
	})
}

func TestInMemoryFindOneByEmail(t *testing.T) {
	ctx := context.Background()
	store := &user.InMemoryStore{}
	t.Run("finds existing user by email", func(t *testing.T) {
		someUser := createDefaultInMemoryUserHelper(t, ctx, store)
		_, err := store.FindOneByEmail(ctx, someUser.Email)
		if err != nil {
			t.Errorf("didn't expect an error while finding user but got one: %v", err)
		}
	})

	t.Run("returns proper error when user for update does not exist", func(t *testing.T) {
		invalidEmail := "invalidEmail"

		_, err := store.FindOneByEmail(ctx, invalidEmail)
		if err == nil {
			t.Fatal("expected an error but didn't get one")
		}

		var notFoundErr *user.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %#v, want %#v", err, &user.NotFoundError{Email: invalidEmail})
		}
	})
}
func createDefaultInMemoryUserHelper(t testing.TB, ctx context.Context, store *user.InMemoryStore) user.User {
	t.Helper()
	return createInMemoryUserHelper(t, ctx, store, validFirstName, validLastName, validEmail, validPassword)
}

func createInMemoryUserHelper(
	t testing.TB,
	ctx context.Context,
	store *user.InMemoryStore,
	firstName string,
	lastName string,
	email string,
	password string,
) user.User {
	t.Helper()
	newUser, isValid, _ := user.New(firstName, lastName, email, password)
	assertEqual(t, true, isValid)
	createdUser, err := store.Create(ctx, newUser)
	if err != nil {
		t.Fatalf("didn't expect an error while putting user into in memory store but got one: %v", err)
	}
	return createdUser
}
