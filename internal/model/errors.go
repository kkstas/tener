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
	PK  string
	SK  string
	Err error
}

func (e *ExpenseNotFoundError) Unwrap() error { return e.Err }
func (e *ExpenseNotFoundError) Error() string {
	return fmt.Sprintf("expense with PK='%s' & SK='%s' not found", e.PK, e.SK)
}
