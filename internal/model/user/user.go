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
	PK                  string   `dynamodbav:"PK"           json:"-"`
	ID                  string   `dynamodbav:"SK"           json:"id"`
	FirstName           string   `dynamodbav:"firstName"    json:"firstName"`
	LastName            string   `dynamodbav:"lastName"     json:"lastName"`
	Email               string   `dynamodbav:"email"        json:"email"`
	Vaults              []string `dynamodbav:"vaults"       json:"vaults"`
	PasswordHash        string   `dynamodbav:"passwordHash" json:"-"`
	CreatedAt           string   `dynamodbav:"createdAt"    json:"createdAt"`
	validator.Validator `dynamodbav:"-" json:"-"`
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
		Email:        email,
		Vaults:       []string{},
		PasswordHash: passwordHash,
		CreatedAt:    currentTimestamp,
	})
}

func validate(user User) (User, error) {
	user.Check(validator.StringLengthBetween("firstName", user.FirstName, FirstNameMinLength, FirstNameMaxLength))
	user.Check(validator.StringLengthBetween("lastName", user.LastName, LastNameMinLength, LastNameMaxLength))
	user.Check(validator.IsEmail("email", user.Email))

	if err := user.Validate(); err != nil {
		return User{}, err
	}

	return user, nil
}
