package run

import (
	"encoding/json"
	"fmt"
)

// CommandResult holds the result of a command execution (bash or external tool).
type CommandResult struct {
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr,omitempty"`
	ExitCode  int    `json:"exit_code"`
	TimedOut  bool   `json:"timed_out,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
	Error     string `json:"error,omitempty"`
}

// String returns the result as JSON (always structured).
func (r CommandResult) String() string {
	data, err := json.Marshal(r)
	if err != nil {
		return fmt.Sprintf(`{"error":"marshal error: %v"}`, err)
	}
	return string(data)
}

// SmartString returns stdout on success, JSON on error.
// More user-friendly for external tool output.
func (r CommandResult) SmartString() string {
	// For successful runs, just return stdout
	if r.ExitCode == 0 && r.Error == "" && !r.TimedOut {
		if r.Stdout != "" {
			return r.Stdout
		}
		return "[command completed successfully with no output]"
	}

	// For errors, return structured JSON
	return r.String()
}
