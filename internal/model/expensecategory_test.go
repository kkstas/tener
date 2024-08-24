package model_test

import (
	"testing"

	"github.com/kkstas/tjener/internal/model"
)

func TestNewExpenseCategory(t *testing.T) {
	t.Run("returns error when expense category has invalid length", func(t *testing.T) {
		_, err := model.NewExpenseCategory("a")
		if err == nil {
			t.Error("expected error when category name is too short")
		}

		_, err = model.NewExpenseCategory(string(make([]byte, 100)))
		if err == nil {
			t.Error("expected error when category name is too long")
		}
	})

	t.Run("creates expense category when category name is valid", func(t *testing.T) {
		_, err := model.NewExpenseCategory("food")
		if err != nil {
			t.Errorf("didn't expect error: %v", err)
		}
	})
}
