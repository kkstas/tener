package model

import (
	"github.com/kkstas/tjener/pkg/validator"
)

type ExpenseCategory struct {
	PK                  string `dynamodbav:"PK"`
	Name                string `dynamodbav:"SK"`
	validator.Validator `dynamodbav:"-"`
}

func NewExpenseCategory(name string) (ExpenseCategory, error) {
	category := ExpenseCategory{
		PK:   expenseCategoryPK,
		Name: name,
	}
	category.Check(validator.StringLengthBetween("name", name, expenseCategoryMinLength, expenseCategoryMaxLength))
	if err := category.Validate(); err != nil {
		return ExpenseCategory{}, err
	}

	return category, nil
}
