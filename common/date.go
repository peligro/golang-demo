package common

import "time"

// FormatDate formatea time.Time a "DD/MM/YYYY"
func FormatDate(t time.Time) string {
	return t.Format("02/01/2006")
}

// FormatTime formatea time.Time a "HH:MM:SS" (24h)
func FormatTime(t time.Time) string {
	return t.Format("15:04:05")
}

// FormatDateTime combina date y time
func FormatDateTime(t time.Time) string {
	return FormatDate(t) + " " + FormatTime(t)
}
