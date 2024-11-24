package expensecategory_test

import (
	"testing"

	"github.com/kkstas/tener/internal/model/expensecategory"
)

func TestNew(t *testing.T) {
	t.Run("returns error when expense category has invalid length", func(t *testing.T) {
		_, isValid, _ := expensecategory.New("a")
		if isValid {
			t.Error("expected error when category name is too short")
		}

		_, isValid, _ = expensecategory.New(string(make([]byte, 100)))
		if isValid {
			t.Error("expected error when category name is too long")
		}
	})

	t.Run("creates expense category when category name is valid", func(t *testing.T) {
		_, isValid, errMessages := expensecategory.New("food")
		if !isValid {
			t.Errorf("didn't expect error: %v", errMessages)
		}
	})
}
