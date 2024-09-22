package user

import "fmt"

type NotFoundError struct {
	ID    string
	Email string
	Err   error
}

func (e *NotFoundError) Unwrap() error { return e.Err }
func (e *NotFoundError) Error() string {
	userDetail := ""
	if e.ID != "" {
		userDetail += fmt.Sprintf("ID='%s' ", e.ID)
	}
	if e.ID != "" {
		userDetail += fmt.Sprintf("Email='%s' ", e.Email)
	}
	return fmt.Sprintf("user %s not found", userDetail)
}
