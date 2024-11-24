package expensecategory

import (
	"github.com/kkstas/tener/pkg/validator"
)

const (
	CategoryNameMinLength = 2
	CategoryNameMaxLength = 50
)

type Category struct {
	PK                  string `dynamodbav:"PK"`
	Name                string `dynamodbav:"SK"`
	CreatedBy           string `dynamodbav:"createdBy"`
	validator.Validator `dynamodbav:"-"`
}

func New(name string) (category Category, isValid bool, errMessages validator.ErrMessages) {
	category = Category{
		Name: name,
	}
	category.Check(validator.StringLengthBetween("name", name, CategoryNameMinLength, CategoryNameMaxLength))
	if isValid, errMessages = category.Validate(); !isValid {
		return Category{}, false, errMessages
	}

	return category, true, nil
}
