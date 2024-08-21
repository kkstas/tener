package model

import (
	"fmt"
)

type ExpenseCategoryAlreadyExistsError struct {
	Name string
}

func (e *ExpenseCategoryAlreadyExistsError) Error() string {
	return fmt.Sprintf("expense category '%s' already exists", e.Name)
}
