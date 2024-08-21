package validator

import "fmt"

type ValidationError struct {
	ErrMessages map[string][]string
	Err         error
}

func (e *ValidationError) Unwrap() error { return e.Err }
func (e *ValidationError) Error() string { return fmt.Sprintf("%+v", e.ErrMessages) }
