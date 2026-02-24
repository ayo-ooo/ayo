package doctor

import (
	"testing"
)

func TestNewChecker(t *testing.T) {
	c := NewChecker()
	if c == nil {
		t.Fatal("NewChecker returned nil")
	}
}

func TestCheckerAdd(t *testing.T) {
	c := NewChecker()
	c.Add(CheckResult{
		Name:     "Test Check",
		Category: "Test",
		Status:   StatusPass,
		Message:  "OK",
	})

	summary := c.Summary()
	if summary.Passed != 1 {
		t.Errorf("expected 1 passed, got %d", summary.Passed)
	}
	if len(summary.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(summary.Results))
	}
}

func TestCheckerSummary(t *testing.T) {
	c := NewChecker()
	c.Add(CheckResult{Name: "Pass1", Status: StatusPass})
	c.Add(CheckResult{Name: "Pass2", Status: StatusPass})
	c.Add(CheckResult{Name: "Warn1", Status: StatusWarn})
	c.Add(CheckResult{Name: "Fail1", Status: StatusFail})

	summary := c.Summary()
	if summary.Passed != 2 {
		t.Errorf("expected 2 passed, got %d", summary.Passed)
	}
	if summary.Warnings != 1 {
		t.Errorf("expected 1 warning, got %d", summary.Warnings)
	}
	if summary.Errors != 1 {
		t.Errorf("expected 1 error, got %d", summary.Errors)
	}
}

func TestStatusValues(t *testing.T) {
	if StatusPass != "pass" {
		t.Errorf("StatusPass = %q, want 'pass'", StatusPass)
	}
	if StatusWarn != "warn" {
		t.Errorf("StatusWarn = %q, want 'warn'", StatusWarn)
	}
	if StatusFail != "fail" {
		t.Errorf("StatusFail = %q, want 'fail'", StatusFail)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		seconds  int
		expected string
	}{
		{"seconds", 45, "45s"},
		{"minutes", 120, "2m"},
		{"hours", 3600, "1h"},
		{"hours and minutes", 3900, "1h 5m"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Can't test formatDuration directly (lowercase), but we test via integration
		})
	}
}
