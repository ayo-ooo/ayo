package cli

import (
	"testing"
)

func TestParseSchedule_Cron(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// 6-field cron (with seconds)
		{"0 0 * * * *", "0 0 * * * *"},
		{"0 30 9 * * *", "0 30 9 * * *"},
		{"0 0 2 * * MON", "0 0 2 * * MON"},
		{"*/5 * * * * *", "*/5 * * * * *"},
		
		// 5-field cron (standard)
		{"0 * * * *", "0 * * * *"},
		{"30 9 * * *", "30 9 * * *"},
		{"0 2 * * MON", "0 2 * * MON"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := ParseSchedule(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("got %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestParseSchedule_SimplePatterns(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hourly", "0 0 * * * *"},
		{"daily", "0 0 0 * * *"},
		{"weekly", "0 0 0 * * SUN"},
		{"monthly", "0 0 0 1 * *"},
		{"yearly", "0 0 0 1 1 *"},
		{"midnight", "0 0 0 * * *"},
		{"noon", "0 0 12 * * *"},
		
		// Case insensitive
		{"HOURLY", "0 0 * * * *"},
		{"Daily", "0 0 0 * * *"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := ParseSchedule(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("got %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestParseSchedule_EveryPatterns(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Basic intervals
		{"every minute", "0 * * * * *"},
		{"every hour", "0 0 * * * *"},
		{"every day", "0 0 0 * * *"},
		{"every week", "0 0 0 * * SUN"},
		{"every month", "0 0 0 1 * *"},
		{"every year", "0 0 0 1 1 *"},
		
		// With time
		{"every day at 9am", "0 0 9 * * *"},
		{"every day at 9:30am", "0 30 9 * * *"},
		{"every day at 2pm", "0 0 14 * * *"},
		{"every day at 14:00", "0 0 14 * * *"},
		
		// Day names
		{"every monday", "0 0 0 * * MON"},
		{"every monday at 3pm", "0 0 15 * * MON"},
		{"every friday at 5pm", "0 0 17 * * FRI"},
		{"every sunday", "0 0 0 * * SUN"},
		
		// Interval patterns
		{"every 5 minutes", "0 */5 * * * *"},
		{"every 15 minutes", "0 */15 * * * *"},
		{"every 2 hours", "0 0 */2 * * *"},
		{"every 6 hours", "0 0 */6 * * *"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := ParseSchedule(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("got %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestParseSchedule_Errors(t *testing.T) {
	tests := []string{
		"",                    // Empty
		"never",               // Unrecognized
		"every other day",     // Not supported
		"on tuesdays",         // Wrong format
		"every 7 minutes",     // 60 not divisible by 7
		"every day at 25:00",  // Invalid time
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseSchedule(input)
			if err == nil {
				t.Errorf("expected error for %q", input)
			}
		})
	}
}

func TestLooksLikeCron(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"0 0 * * * *", true},
		{"*/5 * * * *", true},
		{"0 0 2 * * MON", true},
		{"every hour", false},
		{"daily", false},
		{"0 0 * * *", true},  // 5 fields also valid
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := looksLikeCron(tc.input)
			if result != tc.expected {
				t.Errorf("looksLikeCron(%q) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}
