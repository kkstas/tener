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
	CreatedAt string
	Err       error
}

func (e *ExpenseNotFoundError) Unwrap() error { return e.Err }
func (e *ExpenseNotFoundError) Error() string {
	return fmt.Sprintf("expense with CreatedAt='%s' not found", e.CreatedAt)
}
