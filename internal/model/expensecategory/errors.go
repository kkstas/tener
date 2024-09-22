package expensecategory

import (
	"fmt"
)

type AlreadyExistsError struct {
	Name string
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("expense category '%s' already exists", e.Name)
}

type NotFoundError struct {
	SK  string
	Err error
}

func (e *NotFoundError) Unwrap() error { return e.Err }
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("expense category with SK='%s' not found", e.SK)
}
