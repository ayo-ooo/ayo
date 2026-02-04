package debug

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestSetEnabled(t *testing.T) {
	// Save initial state
	origEnabled := IsEnabled()
	defer SetEnabled(origEnabled)

	SetEnabled(true)
	if !IsEnabled() {
		t.Error("expected debug to be enabled")
	}

	SetEnabled(false)
	if IsEnabled() {
		t.Error("expected debug to be disabled")
	}
}

func TestLog_Disabled(t *testing.T) {
	origEnabled := IsEnabled()
	defer SetEnabled(origEnabled)

	var buf bytes.Buffer
	SetOutput(&buf)
	defer SetOutput(nil)

	SetEnabled(false)
	Log("test message", "key", "value")

	if buf.Len() > 0 {
		t.Error("expected no output when debug is disabled")
	}
}

func TestLog_Enabled(t *testing.T) {
	origEnabled := IsEnabled()
	defer SetEnabled(origEnabled)

	var buf bytes.Buffer
	SetOutput(&buf)
	defer SetOutput(nil)

	SetEnabled(true)
	Log("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "[DEBUG]") {
		t.Error("expected [DEBUG] in output")
	}
	if !strings.Contains(output, "test message") {
		t.Error("expected message in output")
	}
	if !strings.Contains(output, "key=value") {
		t.Error("expected key=value in output")
	}
}

func TestLogf(t *testing.T) {
	origEnabled := IsEnabled()
	defer SetEnabled(origEnabled)

	var buf bytes.Buffer
	SetOutput(&buf)
	defer SetOutput(nil)

	SetEnabled(true)
	Logf("count is %d", 42)

	output := buf.String()
	if !strings.Contains(output, "count is 42") {
		t.Error("expected formatted message in output")
	}
}

func TestWithComponent(t *testing.T) {
	origEnabled := IsEnabled()
	defer SetEnabled(origEnabled)

	var buf bytes.Buffer
	SetOutput(&buf)
	defer SetOutput(nil)

	SetEnabled(true)
	logger := WithComponent("daemon")
	logger.Log("started", "port", 8080)

	output := buf.String()
	if !strings.Contains(output, "[daemon]") {
		t.Error("expected [daemon] component in output")
	}
	if !strings.Contains(output, "started") {
		t.Error("expected message in output")
	}
}

func TestDump(t *testing.T) {
	origEnabled := IsEnabled()
	defer SetEnabled(origEnabled)

	var buf bytes.Buffer
	SetOutput(&buf)
	defer SetOutput(nil)

	SetEnabled(true)
	data := map[string]int{"a": 1, "b": 2}
	Dump("test data", data)

	output := buf.String()
	if !strings.Contains(output, "test data") {
		t.Error("expected label in output")
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		contains string
	}{
		{"string", "hello", "hello"},
		{"string with space", "hello world", `"hello world"`},
		{"int", 42, "42"},
		{"duration", 5 * time.Second, "5s"},
		{"error", testError("oops"), `"oops"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.input)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("formatValue(%v) = %q, expected to contain %q", tt.input, result, tt.contains)
			}
		})
	}
}

type testError string

func (e testError) Error() string {
	return string(e)
}
