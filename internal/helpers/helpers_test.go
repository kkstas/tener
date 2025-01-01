package helpers

import (
	"strings"
	"testing"
	"time"
)

func TestDateStringDaysAgo(t *testing.T) {
	t.Run("returns today", func(t *testing.T) {
		loc, _ := time.LoadLocation("Europe/Warsaw")
		now := time.Now().In(loc)
		want, _, _ := strings.Cut(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).Format(time.RFC3339Nano), "T")

		got := DaysAgo(0)
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestNextDay(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{input: "2024-01-31", want: "2024-02-01"},
		{input: "2023-12-31", want: "2024-01-01"},
		{input: "2024-11-11", want: "2024-11-12"},
	}

	for _, c := range cases {
		t.Run("returns next day", func(t *testing.T) {
			got, err := NextDay(c.input)
			if err != nil {
				t.Fatalf("didn't expect an error but got one: %v", err)
			}
			if got != c.want {
				t.Errorf("got '%s', want '%s'", got, c.want)
			}
		})
	}

	t.Run("should return error on string with invalid date layout", func(t *testing.T) {
		_, err := NextDay("2024-0404")
		if err == nil {
			t.Error("expected error but didn't get one")
		}
	})
}

func TestDaysBetween(t *testing.T) {
	cases := []struct {
		from string
		to   string
		want int
	}{
		{from: "2024-01-31", to: "2024-01-01", want: -30},
		{from: "2023-12-31", to: "2024-01-01", want: 1},
		{from: "2024-01-01", to: "2024-01-01", want: 0},
		{from: "2023-01-01", to: "2024-01-01", want: 365},
	}

	for _, c := range cases {
		t.Run("returns next day", func(t *testing.T) {
			got, err := DaysBetween(c.from, c.to)
			if err != nil {
				t.Fatalf("didn't expect an error but got one: %v", err)
			}
			if got != c.want {
				t.Errorf("got '%d', want '%d'", got, c.want)
			}
		})
	}

	t.Run("should return error on strings with invalid date layout", func(t *testing.T) {
		_, err := DaysBetween("2024-0404", "2024-04-05")
		if err == nil {
			t.Error("expected error but didn't get one")
		}
		_, err = DaysBetween("2024-04-04", "202404-05")
		if err == nil {
			t.Error("expected error but didn't get one")
		}
	})
}

func TestGetFirstAndLastDayOfMonth(t *testing.T) {
	gotFrom, gotTo, err := GetFirstAndLastDayOfMonth("2024-01-01")
	if err != nil {
		t.Fatalf("didn't expect an error but got one: %v", err)
	}
	wantFrom := "2024-01-01"
	wantTo := "2024-01-31"

	if gotFrom != wantFrom {
		t.Errorf("got %s want %s", gotFrom, wantFrom)
	}
	if gotTo != wantTo {
		t.Errorf("got %s want %s", gotTo, wantTo)
	}
}
