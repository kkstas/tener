package expense

import "fmt"

type NotFoundError struct {
	SK  string
	Err error
}

func (e *NotFoundError) Unwrap() error { return e.Err }
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("expense with SK='%s' not found", e.SK)
}
