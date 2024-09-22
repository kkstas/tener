package user

import (
	"testing"
)

func TestCheckPassword(t *testing.T) {
	validPassword := "validPassword"
	invalidPassword := "invalidPassword"

	hashedValidPassword, err := hashPassword(validPassword)
	if err != nil {
		t.Fatalf("didn't expect an error but got one: %v", err)
	}

	if CheckPassword(hashedValidPassword, validPassword) != true {
		t.Error("didn't return true for correct password")
	}
	if CheckPassword(hashedValidPassword, invalidPassword) != false {
		t.Error("didn't return false for incorrect password")
	}
}
