package util

import (
	"fmt"
	"time"
)

// FormatDuration formats a duration for human display.
// Returns formats like "123ms", "4.5s", "2m30s" depending on magnitude.
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	secs := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%ds", minutes, secs)
}

// FormatDurationSeconds formats seconds (float64) for human display.
// Returns formats like "<0.1s", "4.5s", "2m30s" depending on magnitude.
func FormatDurationSeconds(seconds float64) string {
	if seconds < 0.1 {
		return "<0.1s"
	}
	if seconds < 60 {
		return fmt.Sprintf("%.1fs", seconds)
	}
	minutes := int(seconds) / 60
	secs := int(seconds) % 60
	return fmt.Sprintf("%dm%ds", minutes, secs)
}
