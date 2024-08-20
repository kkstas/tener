package model

import (
	"slices"
	"unicode/utf8"
)

var ValidCurrencies = []string{"PLN", "USD", "EUR", "GBP", "CHF", "NOK", "SEK", "DKK", "HUF", "CZK", "CAD", "AUD", "JPY", "CNY", "TRY"}

const (
	expenseCategoryMinLength = 2
	expenseCategoryMaxLength = 50
)

func validateCurrency(curr string) error {
	if !slices.Contains(ValidCurrencies, curr) {
		return &InvalidCurrencyError{curr}
	}
	return nil
}

func validateCategory(category string) error {
	categoryLength := utf8.RuneCountInString(category)
	if categoryLength < expenseCategoryMinLength || categoryLength > expenseCategoryMaxLength {
		return &InvalidExpenseCategoryLengthError{expenseCategoryMinLength, expenseCategoryMaxLength}
	}
	return nil
}

func validateAmount(amount float64) error {
	if amount == 0 {
		return &ExpenseAmountIsZeroError{}
	}
	if amount != roundToDecimalPlaces(amount, 2) {
		return &InvalidAmountPrecisionError{amount}
	}
	return nil
}
