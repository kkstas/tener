package model

import (
	"strconv"
	"strings"
	"time"
)

func generateSK(date string) string {
	return date + "::" + strconv.FormatInt(time.Now().UnixMilli(), 10)
}

func getTimestampDaysAgo(days int) string {
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
