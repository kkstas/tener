package user_test

import (
	"testing"

	"github.com/kkstas/tener/internal/model/user"
)

const (
	validFirstName = "John"
	validLastName  = "Doe"
	validEmail     = "john@doe.com"
	validPassword  = "newPassword123!"
)

func TestNew(t *testing.T) {
	t.Run("creates new user with hashed password", func(t *testing.T) {
		newUser, isValid, errMessages := user.New(validFirstName, validLastName, validEmail, validPassword)
		if !isValid {
			t.Fatalf("didn't expect an error but got one: %v", errMessages)
		}
		assertEqual(t, newUser.FirstName, validFirstName)
		assertEqual(t, newUser.LastName, validLastName)
		assertEqual(t, newUser.Email, validEmail)
		assertEqual(t, user.CheckPassword(newUser.PasswordHash, validPassword), true)
	})

	t.Run("returns error if name length is invalid", func(t *testing.T) {
		tooShortFirstName := string(make([]byte, user.FirstNameMinLength-1))
		tooLongFirstName := string(make([]byte, user.FirstNameMaxLength+1))
		tooShortLastName := string(make([]byte, user.LastNameMinLength-1))
		tooLongLastName := string(make([]byte, user.LastNameMaxLength+1))

		_, isValid, _ := user.New(tooShortFirstName, validLastName, validEmail, validPassword)
		if isValid {
			t.Error("expected an error for too short first name but didn't get one")
		}
		_, isValid, _ = user.New(tooLongFirstName, validLastName, validEmail, validPassword)
		if isValid {
			t.Error("expected an error for too long first name but didn't get one")
		}
		_, isValid, _ = user.New(validFirstName, tooShortLastName, validEmail, validPassword)
		if isValid {
			t.Error("expected an error for too short last name but didn't get one")
		}
		_, isValid, _ = user.New(validFirstName, tooLongLastName, validEmail, validPassword)
		if isValid {
			t.Error("expected an error for too long last name but didn't get one")
		}
	})

	t.Run("returns error if password length is invalid", func(t *testing.T) {
		tooShortPassword := string(make([]byte, user.PasswordMinLength-1))
		tooLongPassword := string(make([]byte, user.PasswordMaxLength+1))

		_, isValid, _ := user.New(validFirstName, validLastName, validEmail, tooShortPassword)
		if isValid {
			t.Error("expected an error for too short password but didn't get one")
		}
		_, isValid, _ = user.New(validFirstName, validLastName, validEmail, tooLongPassword)
		if isValid {
			t.Error("expected an error for too long password but didn't get one")
		}
	})

	t.Run("returns error for invalid email", func(t *testing.T) {
		invalidEmails := []string{
			"plainaddress",
			"@missinglocalpart.com",
			"user@.domain.com",
			"user@domain..com",
			"user@domain",
			"user@domain.c",
			"user@domain..com",
			"user@ domain.com",
			"user@domain .com",
			"user@domain.com ",
			" user@domain.com",
			"user@@domain.com",
			"userdomain.com",
			"user@domain.com..",
			"user@domaincom.",
		}
		for _, c := range invalidEmails {
			_, isValid, _ := user.New(validFirstName, validLastName, c, validPassword)
			if isValid {
				t.Errorf("expected an error for invalid email %s but didn't get one", c)
			}
		}
	})
}

func assertEqual[T comparable](t testing.TB, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
