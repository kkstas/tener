package validator

import "encoding/json"

type ValidationError struct {
	ErrMessages map[string][]string
	Err         error
}

func (e *ValidationError) Unwrap() error { return e.Err }
func (e *ValidationError) Error() string {
	val, _ := json.Marshal(e.ErrMessages)
	return string(val)
}
