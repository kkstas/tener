package user

import (
	"strings"

	"github.com/google/uuid"

	"github.com/kkstas/tener/internal/helpers"
	"github.com/kkstas/tener/pkg/validator"
)

const (
	userPK             = "user"
	FirstNameMinLength = 2
	FirstNameMaxLength = 64
	LastNameMinLength  = 2
	LastNameMaxLength  = 64
	PasswordMinLength  = 8
	PasswordMaxLength  = 64
)

type User struct {
	PK                  string   `dynamodbav:"PK"           json:"-"`
	ID                  string   `dynamodbav:"SK"           json:"id"`
	FirstName           string   `dynamodbav:"firstName"    json:"firstName"`
	LastName            string   `dynamodbav:"lastName"     json:"lastName"`
	Email               string   `dynamodbav:"email"        json:"email"`
	ActiveVault         string   `dynamodbav:"activeVault"  json:"activeVault"`
	Vaults              []string `dynamodbav:"vaults"       json:"vaults"`
	PasswordHash        string   `dynamodbav:"passwordHash" json:"-"`
	CreatedAt           string   `dynamodbav:"createdAt"    json:"createdAt"`
	validator.Validator `dynamodbav:"-" json:"-"`
}

func New(firstName, lastName, email, password string) (user User, isValid bool, errMessages validator.ErrMessages) {
	currentTimestamp := helpers.GenerateCurrentTimestamp()
	id := uuid.New().String()

	passwordHash, err := hashPassword(password)
	if err != nil {
		return User{}, false, map[string][]string{"password": {"failed hashing password"}}
	}

	return validate(password, User{
		PK:           userPK,
		ID:           id,
		FirstName:    strings.TrimSpace(firstName),
		LastName:     strings.TrimSpace(lastName),
		Email:        email,
		Vaults:       []string{},
		PasswordHash: passwordHash,
		CreatedAt:    currentTimestamp,
	})
}

func validate(password string, user User) (User, bool, validator.ErrMessages) {
	user.Check(validator.StringLengthBetween("firstName", user.FirstName, FirstNameMinLength, FirstNameMaxLength))
	user.Check(validator.StringLengthBetween("lastName", user.LastName, LastNameMinLength, LastNameMaxLength))
	user.Check(validator.IsEmail("email", user.Email))
	user.Check(validator.StringLengthBetween("password", password, PasswordMinLength, PasswordMaxLength))

	if isValid, errMessages := user.Validate(); !isValid {
		return User{}, false, errMessages
	}

	return user, true, nil
}
