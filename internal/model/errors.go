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

type ExpenseNotFoundError struct {
	SK  string
	Err error
}

func (e *ExpenseNotFoundError) Unwrap() error { return e.Err }
func (e *ExpenseNotFoundError) Error() string {
	return fmt.Sprintf("expense with SK='%s' not found", e.SK)
}
