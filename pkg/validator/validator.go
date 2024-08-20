package validator

import (
	"unicode/utf8"
)

type Validator struct {
	Errors map[string][]string
}

func NewValidator() Validator {
	return Validator{Errors: make(map[string][]string)}
}

func (v *Validator) Check(ok bool, name, msg string) {
	if v.Errors == nil {
		v.Errors = make(map[string][]string)
	}
	if !ok {
		v.Errors[name] = append(v.Errors[name], msg)
	}
}

func (v *Validator) Validate() (bool, map[string][]string) {
	if len(v.Errors) > 0 {
		return false, v.Errors
	}
	return true, nil
}

func StringLengthBetween(val string, min, max int) bool {
	length := utf8.RuneCountInString(val)
	return length <= max && length >= min
}
