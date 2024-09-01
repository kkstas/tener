package validator

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

type Validator struct {
	ErrMessages map[string][]string
}

func NewValidator() Validator {
	return Validator{ErrMessages: make(map[string][]string)}
}

func (v *Validator) Check(ok bool, name, msg string) {
	if v.ErrMessages == nil {
		v.ErrMessages = make(map[string][]string)
	}
	if !ok {
		v.ErrMessages[name] = append(v.ErrMessages[name], msg)
	}
}

func (v *Validator) Validate() *ValidationError {
	if len(v.ErrMessages) > 0 {
		return &ValidationError{ErrMessages: v.ErrMessages}
	}
	return nil
}

func StringLengthBetween(name, val string, min, max int) (bool, string, string) {
	length := utf8.RuneCountInString(strings.TrimSpace(val))
	return length <= max && length >= min, name, fmt.Sprintf("must be between %d and %d characters long", min, max)
}

func OneOf[T comparable](name string, val T, arr []T) (bool, string, string) {
	return slices.Contains(arr, val), name, fmt.Sprintf("must be one of %v", arr)
}

func IsValidAmountPrecision(name string, amount float64) (bool, string, string) {
	if amount != roundToDecimalPlaces(amount, 2) {
		return false, name, "must have a precision of up to 2 decimal places"
	}
	return true, "", ""
}

func IsNonZero(name string, amount float64) (bool, string, string) {
	if amount == 0 {
		return false, name, "must be non-zero"
	}
	return true, "", ""
}

func IsValidDate(name, dateString string) (bool, string, string) {
	_, err := time.Parse("2006-01-02", dateString)
	if err != nil {
		return false, name, "must be valid date"
	}

	return true, "", ""
}

func roundToDecimalPlaces(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return math.Round(num*output) / output
}
