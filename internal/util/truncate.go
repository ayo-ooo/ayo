// Package util provides common utility functions shared across the codebase.
package util

import "strings"

// Truncate truncates a string to maxLen characters, adding "..." if truncated.
// It normalizes whitespace by replacing newlines with spaces and trimming.
func Truncate(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// TruncateRaw truncates without adding ellipsis or normalizing whitespace.
func TruncateRaw(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// TruncateTitle truncates for display titles, normalizing whitespace and using
// a single unicode ellipsis character (…) for compactness.
func TruncateTitle(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}
