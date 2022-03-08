package xtime

import "time"

const (
	Fmt   = "2006-01-02 15:04:05"
	Day   = 24 * time.Hour
	Month = 30 * Day
)

func SameYear(date1, date2 time.Time) bool {
	y1, _, _ := date1.Date()
	y2, _, _ := date2.Date()
	return y1 == y2
}

func SameMonth(date1, date2 time.Time) bool {
	y1, m1, _ := date1.Date()
	y2, m2, _ := date2.Date()
	return y1 == y2 && m1 == m2
}

func SameDay(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
