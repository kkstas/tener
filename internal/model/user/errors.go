package user

import "fmt"

type NotFoundError struct {
	ID  string
	Err error
}

func (e *NotFoundError) Unwrap() error { return e.Err }
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("user with ID='%s' not found", e.ID)
}
