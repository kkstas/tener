package model

import (
	"fmt"
	"strings"
	"time"
)

func buildSK(date, createdAt string) string {
	return date + "::" + createdAt
}

func generateCurrentTimestamp() string {
	loc, _ := time.LoadLocation("Europe/Warsaw")
	return time.Now().In(loc).Format(time.RFC3339Nano)
}

// Gets next day of YYYY-MM-DD date string
func nextDay(date string) (string, error) {
	parsedDate, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return "", err
	}
	nextDay := parsedDate.AddDate(0, 0, 1)
	return nextDay.Format(time.DateOnly), nil
}

// Counts difference in days of two YYYY-MM-DD date strings
func daysBetween(from, to string) (int, error) {
	startDate, err := time.Parse(time.DateOnly, from)
	if err != nil {
		return 0, fmt.Errorf("failed to parse 'from' date: %w", err)
	}
	endDate, err := time.Parse(time.DateOnly, to)
	if err != nil {
		return 0, fmt.Errorf("failed to parse 'to' date: %w", err)
	}
	diff := endDate.Sub(startDate).Hours() / 24

	return int(diff), nil
}

func dateStringDaysAgo(days int) string {
	loc, _ := time.LoadLocation("Europe/Warsaw")
	now := setTimeToMidnight(time.Now(), loc)
	pastDate := now.Add(-(time.Duration(days) * 24 * time.Hour))
	date, _, _ := strings.Cut(pastDate.Format(time.RFC3339), "T")
	return date
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
