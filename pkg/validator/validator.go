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

func IsAmountPrecision(name string, amount float64) (bool, string, string) {
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

func IsTime(name, layout, dateString string) (bool, string, string) {
	_, err := time.Parse(layout, dateString)
	if err != nil {
		return false, name, "must be a valid date"
	}

	return true, "", ""
}

var (
	validEmailLocalChars  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!#$%&'*+-/=?^_`{|}~."
	validEmailDomainChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-."
)

func IsEmail(name, email string) (bool, string, string) {
	msg := "must be a valid email address"

	if len(email) == 0 || len(email) > 254 {
		return false, name, msg
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false, name, msg
	}
	local, domain := parts[0], parts[1]

	if len(local) == 0 || len(local) > 64 {
		return false, name, msg
	}
	if local[0] == '.' || local[len(local)-1] == '.' {
		return false, name, msg
	}
	if strings.Contains(local, "..") {
		return false, name, msg
	}
	for _, char := range local {
		if !strings.ContainsRune(validEmailLocalChars, char) {
			return false, name, msg
		}
	}

	if len(domain) == 0 || len(domain) > 255 {
		return false, name, msg
	}
	if strings.Count(domain, ".") < 1 || domain[0] == '.' || domain[len(domain)-1] == '.' {
		return false, name, msg
	}
	if strings.Contains(domain, "..") {
		return false, name, msg
	}
	domainParts := strings.Split(domain, ".")
	if len(domainParts[1]) < 2 {
		return false, name, msg
	}

	for _, char := range domain {
		if !strings.ContainsRune(validEmailDomainChars, char) {
			return false, name, msg
		}
	}

	return true, "", ""
}

func roundToDecimalPlaces(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return math.Round(num*output) / output
}
