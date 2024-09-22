package auth_test

import (
	"os"
	"testing"

	"github.com/kkstas/tjener/internal/auth"
	"github.com/kkstas/tjener/internal/model/user"
)

func TestCreateToken(t *testing.T) {
	os.Setenv("TOKEN_SECRET", "gHg8v3-XKj9XO8M-6gpjzW0n1xn7UZTBICIY1FcjyPw")
	newUser, err := user.New("John", "Doe", "john@doe.com", "newPassword123!")
	if err != nil {
		t.Fatalf("didn't expect an error but got one: %v", err)
	}
	token, err := auth.CreateToken(newUser)
	if err != nil {
		t.Fatalf("didn't expect an error but got one: %v", err)
	}

	err = auth.ValidateToken(token)
	if err != nil {
		t.Errorf("didn't exepct an error, but got one: %v", err)
	}

	err = auth.ValidateToken("smh")
	if err == nil {
		t.Error("expected an error, but didn't get one")
	}
}
