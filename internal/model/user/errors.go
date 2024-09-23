package user

import "fmt"

type NotFoundError struct {
	ID    string
	Email string
}

func (e *NotFoundError) Error() string {
	userDetail := ""
	if e.ID != "" {
		userDetail += fmt.Sprintf("ID='%s' ", e.ID)
	}
	if e.ID != "" {
		userDetail += fmt.Sprintf("Email='%s' ", e.Email)
	}
	return fmt.Sprintf("user with %s not found", userDetail)
}

type AlreadyExistsError struct {
	ID    string
	Email string
}

func (e *AlreadyExistsError) Error() string {
	userDetail := ""
	if e.ID != "" {
		userDetail += fmt.Sprintf("ID='%s' ", e.ID)
	}
	if e.ID != "" {
		userDetail += fmt.Sprintf("Email='%s' ", e.Email)
	}
	return fmt.Sprintf("user with '%s' already exists", userDetail)
}
