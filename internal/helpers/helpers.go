package helpers

import (
	"fmt"
	"strings"
	"time"
)

const DefaultDaysAgo = 7

func GenerateCurrentTimestamp() string {
	loc, _ := time.LoadLocation("Europe/Warsaw")
	return time.Now().In(loc).Format(time.RFC3339Nano)
}

// Gets next day of YYYY-MM-DD date string
func NextDay(date string) (string, error) {
	parsedDate, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return "", err
	}
	nextDay := parsedDate.AddDate(0, 0, 1)
	return nextDay.Format(time.DateOnly), nil
}

// Counts difference in days of two YYYY-MM-DD date strings
func DaysBetween(from, to string) (int, error) {
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

// Returns YYYY-MM-DD date from given amount of days ago
func DaysAgo(days int) string {
	loc, _ := time.LoadLocation("Europe/Warsaw")
	now := setTimeToMidnight(time.Now().In(loc), loc)
	pastDate := now.Add(-(time.Duration(days) * 24 * time.Hour))
	date, _, _ := strings.Cut(pastDate.Format(time.RFC3339), "T")
	return date
}

// Returns YYYY-MM-DD date of num months ago from now
func MonthsAgo(num int) string {
	loc, _ := time.LoadLocation("Europe/Warsaw")
	now := time.Now().In(loc)
	monthAgo := now.AddDate(0, -num, 0)
	return monthAgo.Format("2006-01-02")
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

func GetFirstAndLastDayOfMonth(dateStr string) (from, to string, err error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse date %s: %w", dateStr, err)
	}

	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	lastDay := firstDay.AddDate(0, 1, 0).Add(-time.Hour * 24)

	from = firstDay.Format("2006-01-02")
	to = lastDay.Format("2006-01-02")
	return from, to, nil
}

func GetFirstDayOfCurrentMonth() string {
	date, err := time.Parse("2006-01-02", DaysAgo(0))
	if err != nil {
		fmt.Println("Error parsing date:", err)
		return ""
	}

	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	return firstDay.Format("2006-01-02")
}

func IsValidYYYYMM(dateString string) bool {
	layout := "2006-01"
	_, err := time.Parse(layout, dateString)
	return err == nil
}
