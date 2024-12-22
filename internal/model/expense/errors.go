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

type MaxMonthExpenseCountExceededError struct {
	Month string
	Vault string
	Err   error
}

func (e *MaxMonthExpenseCountExceededError) Unwrap() error { return e.Err }
func (e *MaxMonthExpenseCountExceededError) Error() string {
	return fmt.Sprintf("maximum expense count exceeded for month %s in vault %s", e.Month, e.Vault)
}
