package auth_test

import (
	"os"
	"testing"

	"github.com/kkstas/tener/internal/auth"
	"github.com/kkstas/tener/internal/model/user"
)

func TestCreateToken(t *testing.T) {
	os.Setenv("TOKEN_SECRET", "gHg8v3-XKj9XO8M-6gpjzW0n1xn7UZTBICIY1FcjyPw")
	newUser, isValid, errMessages := user.New("John", "Doe", "john@doe.com", "newPassword123!")
	if !isValid {
		t.Fatalf("didn't expect an error but got one: %v", errMessages)
	}
	token, err := auth.CreateToken(newUser)
	if err != nil {
		t.Fatalf("didn't expect an error but got one: %v", err)
	}

	_, err = auth.DecodeToken(token)
	if err != nil {
		t.Errorf("didn't exepct an error, but got one: %v", err)
	}

	_, err = auth.DecodeToken("smh")
	if err == nil {
		t.Error("expected an error, but didn't get one")
	}
}
