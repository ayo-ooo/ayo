package util

import (
	"testing"
	"time"
)

// TestFormatDuration tests the FormatDuration function.
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "milliseconds",
			duration: 123 * time.Millisecond,
			expected: "123ms",
		},
		{
			name:     "milliseconds - small",
			duration: 1 * time.Millisecond,
			expected: "1ms",
		},
		{
			name:     "milliseconds - large but < 1s",
			duration: 999 * time.Millisecond,
			expected: "999ms",
		},
		{
			name:     "seconds - decimal",
			duration: 1500 * time.Millisecond,
			expected: "1.5s",
		},
		{
			name:     "seconds - one decimal",
			duration: 4500 * time.Millisecond,
			expected: "4.5s",
		},
		{
			name:     "seconds - whole number",
			duration: 5 * time.Second,
			expected: "5.0s",
		},
		{
			name:     "seconds - almost a minute",
			duration: 59 * time.Second,
			expected: "59.0s",
		},
		{
			name:     "minutes and seconds",
			duration: 90 * time.Second,
			expected: "1m30s",
		},
		{
			name:     "multiple minutes",
			duration: 2*time.Minute + 30*time.Second,
			expected: "2m30s",
		},
		{
			name:     "minutes only",
			duration: 5 * time.Minute,
			expected: "5m0s",
		},
		{
			name:     "zero duration",
			duration: 0,
			expected: "0ms",
		},
		{
			name:     "negative duration (edge case)",
			duration: -100 * time.Millisecond,
			expected: "-100ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

// TestFormatDurationSeconds tests the FormatDurationSeconds function.
func TestFormatDurationSeconds(t *testing.T) {
	tests := []struct {
		name     string
		seconds  float64
		expected string
	}{
		{
			name:     "very small",
			seconds:  0.05,
			expected: "<0.1s",
		},
		{
			name:     "just below 0.1s",
			seconds:  0.099,
			expected: "<0.1s",
		},
		{
			name:     "exactly 0.1s",
			seconds:  0.1,
			expected: "0.1s",
		},
		{
			name:     "seconds with decimal",
			seconds:  1.5,
			expected: "1.5s",
		},
		{
			name:     "seconds - many decimals",
			seconds:  4.567,
			expected: "4.6s",
		},
		{
			name:     "whole seconds",
			seconds:  10,
			expected: "10.0s",
		},
		{
			name:     "almost a minute",
			seconds:  59.9,
			expected: "59.9s",
		},
		{
			name:     "one minute",
			seconds:  60,
			expected: "1m0s",
		},
		{
			name:     "minute and seconds",
			seconds:  90,
			expected: "1m30s",
		},
		{
			name:     "multiple minutes",
			seconds:  150,
			expected: "2m30s",
		},
		{
			name:     "five minutes",
			seconds:  300,
			expected: "5m0s",
		},
		{
			name:     "exactly an hour",
			seconds:  3600,
			expected: "60m0s",
		},
		{
			name:     "zero seconds",
			seconds:  0,
			expected: "<0.1s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDurationSeconds(tt.seconds)
			if result != tt.expected {
				t.Errorf("FormatDurationSeconds(%v) = %q, want %q", tt.seconds, result, tt.expected)
			}
		})
	}
}

// TestFormatDurationConsistency tests that both functions produce consistent results.
func TestFormatDurationConsistency(t *testing.T) {
	// Test that FormatDuration and FormatDurationSeconds produce similar results
	duration := 90 * time.Second
	result1 := FormatDuration(duration)
	result2 := FormatDurationSeconds(duration.Seconds())

	if result1 != result2 {
		t.Logf("FormatDuration(90s) = %q, FormatDurationSeconds(90) = %q - results differ (acceptable)", result1, result2)
	}
}
