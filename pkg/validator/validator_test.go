package validator_test

import (
	"fmt"
	"testing"

	"github.com/kkstas/tjener/pkg/validator"
)

type TestStruct struct {
	Name string
	Age  int
	validator.Validator
}

func TestIsNonZero(t *testing.T) {
	t.Run("returns false if amount is zero", func(t *testing.T) {
		got, _, _ := validator.IsNonZero("name", 0)
		if got {
			t.Error("expected false for zero")
		}
		got, _, _ = validator.IsNonZero("name", 1)
		if !got {
			t.Error("expected true for non-zero value")
		}
	})
}

func TestIsValidAmountPrecision(t *testing.T) {
	t.Run("returns false if amount has invalid precision", func(t *testing.T) {
		got, _, _ := validator.IsValidAmountPrecision("name", 19.449)
		if got {
			t.Error("expected false for value with invalid precision")
		}
		got, _, _ = validator.IsValidAmountPrecision("name", 19.44)
		if !got {
			t.Error("expected true for value with valid precision")
		}
	})

}

func TestStringLengthBetween(t *testing.T) {
	cases := []struct {
		want bool
		name string
		val  string
		min  int
		max  int
	}{
		{true, "some-name-1", "", 0, 0},
		{false, "some-name-2", "", 1, 1},
		{true, "some-name-3", "hello", 5, 9},
		{false, "some-name-4", "hello", 1, 4},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%+v", c), func(t *testing.T) {
			got, name, _ := validator.StringLengthBetween(c.name, c.val, c.min, c.max)
			if got != c.want {
				t.Errorf("got %t want %t for val '%s' with min=%d and max=%d", got, c.want, c.val, c.min, c.max)
			}
			if name != c.name {
				t.Errorf("expected name %s as second return value, got %s", c.name, name)
			}
		})
	}
}

func TestOneOf(t *testing.T) {
	someSlice := []string{"one", "two", "three"}
	name := "smth"
	val := "two"
	ok, receivedName, _ := validator.OneOf(name, val, someSlice)

	if !ok {
		t.Errorf("expected ok=true, got %t for value %s in %v", ok, val, someSlice)
	}

	if name != receivedName {
		t.Errorf("expected name=%s, got %s for value %s in %v", name, receivedName, val, someSlice)
	}

	invalidVal := "invalidVal"

	ok, _, _ = validator.OneOf(name, invalidVal, someSlice)
	if ok {
		t.Errorf("expected ok=false, got %t for value %s in %v", ok, invalidVal, someSlice)
	}
}

func TestCheck(t *testing.T) {
	t.Run("returns map of errors when check's first parameter is false", func(t *testing.T) {
		someStruct := TestStruct{}

		someStruct.Check(false, "name", "one")
		someStruct.Check(false, "age", "two")

		err := someStruct.Validate()

		if err == nil {
			t.Fatal("expected validation errors but didn't get any")
		}

		got := len(err.ErrMessages)
		want := 2

		if got != want {
			t.Errorf("got map with length of %d, want %d", got, want)
		}
	})
}
