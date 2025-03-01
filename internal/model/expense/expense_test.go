package expense_test

import (
	"testing"
	"time"

	"github.com/kkstas/tener/internal/model/expense"
	"github.com/kkstas/tener/internal/model/expensecategory"
)

func BenchmarkRFC3339(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = time.Now().Format(time.RFC3339)
	}
}

func BenchmarkRFC3339Nano(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = time.Now().Format(time.RFC3339Nano)
	}
}

func TestNew(t *testing.T) {
	validName := "name"
	validDate := "2024-01-01"
	validCategory := "food"
	validAmount := 24.99
	validPaymentMethod := expense.PaymentMethods[0]

	t.Run("creates valid expense", func(t *testing.T) {
		_, isValid, errMessages := expense.New(validName, validDate, validCategory, validAmount, validPaymentMethod)
		if !isValid {
			t.Errorf("didn't expect an error but got one: %v", errMessages)
		}
	})

	t.Run("returns an error when category is too short", func(t *testing.T) {
		tooShortCategory := string(make([]byte, expensecategory.CategoryNameMinLength-1))
		_, isValid, _ := expense.New(validName, validDate, tooShortCategory, validAmount, validPaymentMethod)

		if isValid {
			t.Error("expected an error but didn't get one")
		}
	})

	t.Run("returns an error when category is too long", func(t *testing.T) {
		tooLongCategory := string(make([]byte, expensecategory.CategoryNameMaxLength+1))

		_, isValid, _ := expense.New(validName, validDate, tooLongCategory, validAmount, validPaymentMethod)

		if isValid {
			t.Error("expected an error but didn't get one")
		}
	})

	t.Run("returns an error when amount is float with precision larger than two", func(t *testing.T) {
		_, isValid, _ := expense.New(validName, validDate, validCategory, 24.4234, validPaymentMethod)
		if isValid {
			t.Error("expected an error but didn't get one")
		}
	})

	t.Run("doesn't return an error when amount is float with precision lesser than or equal to two", func(t *testing.T) {
		_, isValid, errMessages := expense.New(validName, validDate, validCategory, 24.44, validPaymentMethod)
		if !isValid {
			t.Errorf("didn't expect an error but got one: %v", errMessages)
		}
		_, isValid, errMessages = expense.New(validName, validDate, validCategory, 24.4, validPaymentMethod)
		if !isValid {
			t.Errorf("didn't expect an error but got one: %v", errMessages)
		}
		_, isValid, errMessages = expense.New(validName, validDate, validCategory, 24, validPaymentMethod)
		if !isValid {
			t.Errorf("didn't expect an error but got one: %v", errMessages)
		}
	})

	t.Run("fails validation if paymentMethod is invalid", func(t *testing.T) {
		_, isValid, _ := expense.New(validName, validDate, validCategory, validAmount, "beans")
		if isValid {
			t.Error("expected expense to fail validation")
		}
	})

	t.Run("returns an error if date is invalid", func(t *testing.T) {
		_, isValid, _ := expense.New(validName, "202401-01", validCategory, validAmount, validPaymentMethod)
		if isValid {
			t.Error("expected expense to fail validation")
		}
	})
}
