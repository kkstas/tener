package model

import (
	"time"
)

func generateCurrentTimestamp() string {
	loc, _ := time.LoadLocation("Europe/Warsaw")
	return time.Now().In(loc).Format(time.RFC3339Nano)
}

func getTimestampDaysAgo(days int) string {
	loc, _ := time.LoadLocation("Europe/Warsaw")
	now := setTimeToMidnight(time.Now(), loc)
	pastDate := now.Add(-(time.Duration(days) * 24 * time.Hour))
	return pastDate.Format(time.RFC3339Nano)
}

func setTimeToMidnight(t time.Time, loc *time.Location) time.Time {
	return time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		0, 0, 0, 0,
		loc,
	)
}