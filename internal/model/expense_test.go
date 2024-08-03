package model

import (
	"strings"
	"testing"
	"time"
)

func TestGetDateAgo(t *testing.T) {

	t.Run("returns datetime string with time at midnight", func(t *testing.T) {
		got := getDateDaysAgo(0)
		if !strings.HasPrefix(got[11:], "00:00:00") {
			t.Errorf("received string that is not valid RFC3339Nano from midnight - %q", got)
		}
	})

	t.Run("returns today at midnight", func(t *testing.T) {
		now := time.Now()
		loc, _ := time.LoadLocation("Europe/Warsaw")
		want := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).Format(time.RFC3339Nano)

		got := getDateDaysAgo(0)
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

}
