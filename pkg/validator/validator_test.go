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

func TestStringLengthBetween(t *testing.T) {
	cases := []struct {
		want bool
		val  string
		min  int
		max  int
	}{
		{true, "", 0, 0},
		{false, "", 1, 1},
		{true, "hello", 5, 9},
		{false, "hello", 1, 4},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%+v", c), func(t *testing.T) {
			got := validator.StringLengthBetween(c.val, c.min, c.max)
			if got != c.want {
				t.Errorf("got %t want %t for val '%s' with min=%d and max=%d", got, c.want, c.val, c.min, c.max)
			}
		})
	}

}

func TestCheck(t *testing.T) {
	t.Run("returns map of errors when check's first parameter is false", func(t *testing.T) {
		someStruct := TestStruct{}

		someStruct.Check(false, "name", "one")
		someStruct.Check(false, "age", "two")

		ok, errors := someStruct.Validate()

		if ok {
			t.Fatal("expected validation errors but didn't get any")
		}

		got := len(errors)
		want := 2

		if got != want {
			t.Errorf("got map with length of %d, want %d", got, want)
		}
	})
}
