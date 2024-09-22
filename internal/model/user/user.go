package user

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/kkstas/tjener/internal/helpers"
	"github.com/kkstas/tjener/pkg/validator"
)

const (
	pk                 = "user"
	FirstNameMinLength = 2
	FirstNameMaxLength = 64
	LastNameMinLength  = 2
	LastNameMaxLength  = 64
)

type User struct {
	PK                  string   `dynamodbav:"PK"`
	ID                  string   `dynamodbav:"SK"`
	FirstName           string   `dynamodbav:"firstName"`
	LastName            string   `dynamodbav:"lastName"`
	Email               string   `dynamodbav:"Email"`
	Vaults              []string `dynamodbav:"vaults"`
	PasswordHash        string   `dynamodbav:"createdAt"`
	CreatedAt           string   `dynamodbav:"createdAt"`
	validator.Validator `dynamodbav:"-"`
}

func New(firstName, lastName, email, password string) (User, error) {
	currentTimestamp := helpers.GenerateCurrentTimestamp()
	id := uuid.New().String()

	passwordHash, err := hashPassword(password)
	if err != nil {
		return User{}, fmt.Errorf("failed hashing password: %w", err)
	}

	return validate(User{
		PK:           pk,
		ID:           id,
		FirstName:    strings.TrimSpace(firstName),
		LastName:     strings.TrimSpace(lastName),
		Email:        strings.TrimSpace(email),
		Vaults:       []string{},
		PasswordHash: passwordHash,
		CreatedAt:    currentTimestamp,
	})
}

func validate(user User) (User, error) {
	user.Check(validator.StringLengthBetween("firstName", user.FirstName, FirstNameMinLength, FirstNameMaxLength))
	user.Check(validator.StringLengthBetween("lastName", user.LastName, LastNameMinLength, LastNameMaxLength))

	if err := user.Validate(); err != nil {
		return User{}, err
	}

	return user, nil
}
