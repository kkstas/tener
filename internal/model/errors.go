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

type ExpenseCategoryNotFoundError struct {
	SK  string
	Err error
}

func (e *ExpenseCategoryNotFoundError) Unwrap() error { return e.Err }
func (e *ExpenseCategoryNotFoundError) Error() string {
	return fmt.Sprintf("expense category with SK='%s' not found", e.SK)
}
