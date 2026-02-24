package daemon

import (
	"testing"
)

func TestExpandCronAlias(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"@hourly", "0 * * * *"},
		{"@daily", "0 0 * * *"},
		{"@DAILY", "0 0 * * *"}, // Case insensitive
		{"@midnight", "0 0 * * *"},
		{"@weekly", "0 0 * * 0"},
		{"@monthly", "0 0 1 * *"},
		{"@yearly", "0 0 1 1 *"},
		{"@annually", "0 0 1 1 *"},
		{"@weekdays", "0 9 * * 1-5"},
		{"@weekends", "0 9 * * 0,6"},
		{"0 9 * * *", "0 9 * * *"},    // Non-alias unchanged
		{"*/15 * * * *", "*/15 * * * *"}, // Non-alias unchanged
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ExpandCronAlias(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandCronAlias(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateCronExpression(t *testing.T) {
	validCases := []string{
		"0 * * * *",
		"0 9 * * *",
		"*/15 * * * *",
		"0 9 * * 1-5",
		"0 0 1,15 * *",
		"30 8 * * 1",
		"@hourly",
		"@daily",
		"@weekly",
		"@monthly",
		"@yearly",
		"@weekdays",
		"@weekends",
		"@every 5m",
		"0 0 9 * * *",  // 6-field with seconds
	}

	for _, expr := range validCases {
		t.Run("valid_"+expr, func(t *testing.T) {
			err := ValidateCronExpression(expr)
			if err != nil {
				t.Errorf("ValidateCronExpression(%q) unexpected error: %v", expr, err)
			}
		})
	}

	invalidCases := []struct {
		expr    string
		wantMsg string
	}{
		{"@unknown", "unknown cron alias"},
		{"0 9 * *", "expected 5 fields"},
		{"0 9 * * * * *", "expected 5 fields"},
	}

	for _, tc := range invalidCases {
		t.Run("invalid_"+tc.expr, func(t *testing.T) {
			err := ValidateCronExpression(tc.expr)
			if err == nil {
				t.Errorf("ValidateCronExpression(%q) expected error", tc.expr)
				return
			}
			cronErr, ok := err.(*CronError)
			if !ok {
				t.Errorf("expected CronError, got %T", err)
				return
			}
			if cronErr.Expression != tc.expr {
				t.Errorf("CronError.Expression = %q, want %q", cronErr.Expression, tc.expr)
			}
		})
	}
}

func TestParseCronSchedule(t *testing.T) {
	tests := []struct {
		input       string
		expected    string
		expectError bool
	}{
		{"@daily", "0 0 * * *", false},
		{"@hourly", "0 * * * *", false},
		{"0 9 * * *", "0 9 * * *", false},
		{"@unknown", "", true},
		{"0 9 * *", "", true}, // Too few fields
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseCronSchedule(tt.input)
			if tt.expectError {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseCronSchedule(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetCronAliases(t *testing.T) {
	aliases := GetCronAliases()
	
	// Should have all expected aliases
	expected := []string{"@hourly", "@daily", "@weekly", "@monthly", "@yearly", "@weekdays", "@weekends"}
	for _, alias := range expected {
		if _, ok := aliases[alias]; !ok {
			t.Errorf("missing alias: %s", alias)
		}
	}
	
	// Should be a copy (modifying doesn't affect original)
	aliases["@custom"] = "1 2 3 4 5"
	if _, ok := cronAliases["@custom"]; ok {
		t.Error("GetCronAliases should return a copy")
	}
}

func TestCronHelp(t *testing.T) {
	help := CronHelp()
	
	// Should contain key sections
	sections := []string{
		"Cron Expression Syntax",
		"minute",
		"hour",
		"@hourly",
		"@daily",
		"Examples",
	}
	
	for _, section := range sections {
		if !containsString(help, section) {
			t.Errorf("CronHelp() missing section: %s", section)
		}
	}
}

func TestCronError_Error(t *testing.T) {
	err := &CronError{
		Expression: "0 9 * *",
		Message:    "expected 5 fields, got 4",
		Suggestion: "Example: '0 9 * * *'",
	}
	
	errStr := err.Error()
	if !containsString(errStr, "0 9 * *") {
		t.Error("error should contain expression")
	}
	if !containsString(errStr, "expected 5 fields") {
		t.Error("error should contain message")
	}
	if !containsString(errStr, "Example") {
		t.Error("error should contain suggestion")
	}
}

func TestValidateCronFields(t *testing.T) {
	validFields := [][]string{
		{"0", "9", "*", "*", "*"},
		{"*/15", "*", "*", "*", "*"},
		{"0", "9", "*", "*", "1-5"},
		{"0", "0", "1,15", "*", "*"},
		{"30", "8", "*", "*", "MON"},
	}
	
	for _, fields := range validFields {
		if err := validateCronFields(fields); err != nil {
			t.Errorf("validateCronFields(%v) unexpected error: %v", fields, err)
		}
	}
	
	invalidFields := [][]string{
		{"invalid", "9", "*", "*", "*"},
		{"0", "9@", "*", "*", "*"},
	}
	
	for _, fields := range invalidFields {
		if err := validateCronFields(fields); err == nil {
			t.Errorf("validateCronFields(%v) expected error", fields)
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
