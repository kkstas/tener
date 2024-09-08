package components

import (
	"time"
)

func parseDate(date, layout string) string {
	parsedDate, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return date
	}
	readableDate := parsedDate.Format(layout)
	return readableDate
}
