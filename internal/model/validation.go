package model

import "slices"

var ValidCurrencies = []string{"PLN", "USD", "EUR", "GBP", "CHF", "NOK", "SEK", "DKK", "HUF", "CZK", "CAD", "AUD", "JPY", "CNY", "TRY"}

func validateCurrency(curr string) error {
	if !slices.Contains(ValidCurrencies, curr) {
		return &InvalidCurrencyError{curr}
	}
	return nil
}

func validateCategory(category string) error {
	if len(category) <= 1 {
		return &ExpenseCategoryIsTooShortError{category}
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
