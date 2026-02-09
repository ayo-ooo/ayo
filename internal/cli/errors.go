package cli

import (
	"fmt"
	"io"
	"strings"
)

// CLIError is a structured error for user-facing CLI output.
// It provides clear error messages with helpful suggestions for recovery.
type CLIError struct {
	// Brief is a short description of the error.
	Brief string

	// Details provides additional context (optional).
	Details string

	// Suggestion tells the user how to fix the problem (optional).
	Suggestion string

	// Code is the exit code for this error.
	Code ExitCode
}

// Error implements the error interface.
func (e *CLIError) Error() string {
	return e.Brief
}

// Print outputs the error to the writer with formatting.
func (e *CLIError) Print(w io.Writer) {
	fmt.Fprintf(w, "Error: %s\n", e.Brief)
	if e.Details != "" {
		fmt.Fprintf(w, "\n%s\n", e.Details)
	}
	if e.Suggestion != "" {
		fmt.Fprintf(w, "\nSuggestion: %s\n", e.Suggestion)
	}
}

// ExitCode returns the exit code for this error.
func (e *CLIError) ExitCode() ExitCode {
	return e.Code
}

// Common error constructors

// ErrNotFound creates an error for missing resources.
func ErrNotFound(resource, name string, available []string) *CLIError {
	err := &CLIError{
		Brief: fmt.Sprintf("%s '%s' not found", resource, name),
		Code:  ExitNotFound,
	}
	if len(available) > 0 {
		err.Details = fmt.Sprintf("Available %ss:\n  %s", strings.ToLower(resource), strings.Join(available, ", "))
	}
	return err
}

// ErrAgentNotFound creates an error for missing agents.
func ErrAgentNotFound(name string, available []string) *CLIError {
	err := ErrNotFound("Agent", name, available)
	err.Suggestion = "Use 'ayo agents list' to see all agents"
	return err
}

// ErrTriggerNotFound creates an error for missing triggers.
func ErrTriggerNotFound(id string) *CLIError {
	return &CLIError{
		Brief:      fmt.Sprintf("Trigger '%s' not found", id),
		Suggestion: "Use 'ayo trigger list' to see all triggers",
		Code:       ExitNotFound,
	}
}

// ErrFlowNotFound creates an error for missing flows.
func ErrFlowNotFound(name string) *CLIError {
	return &CLIError{
		Brief:      fmt.Sprintf("Flow '%s' not found", name),
		Suggestion: "Use 'ayo flows list' to see all flows",
		Code:       ExitNotFound,
	}
}

// ErrSessionNotFound creates an error for missing sessions.
func ErrSessionNotFound(id string) *CLIError {
	return &CLIError{
		Brief:      fmt.Sprintf("Session '%s' not found", id),
		Suggestion: "Use 'ayo sessions list' to see recent sessions",
		Code:       ExitNotFound,
	}
}

// ErrInvalidInput creates an error for invalid user input.
func ErrInvalidInput(message string) *CLIError {
	return &CLIError{
		Brief: message,
		Code:  ExitInvalidInput,
	}
}

// ErrInvalidInputWithSuggestion creates an error with a suggestion.
func ErrInvalidInputWithSuggestion(message, suggestion string) *CLIError {
	return &CLIError{
		Brief:      message,
		Suggestion: suggestion,
		Code:       ExitInvalidInput,
	}
}

// ErrMissingArgument creates an error for missing required arguments.
func ErrMissingArgument(name string) *CLIError {
	return &CLIError{
		Brief: fmt.Sprintf("Missing required argument: %s", name),
		Code:  ExitInvalidInput,
	}
}

// ErrDaemonNotRunning creates an error when the daemon isn't running.
func ErrDaemonNotRunning() *CLIError {
	return &CLIError{
		Brief:      "Service is not running",
		Suggestion: "Start with 'ayo service start'",
		Code:       ExitError,
	}
}

// ErrDaemonConnectionFailed creates an error when connection to daemon fails.
func ErrDaemonConnectionFailed(err error) *CLIError {
	return &CLIError{
		Brief:      "Failed to connect to service",
		Details:    err.Error(),
		Suggestion: "Try 'ayo service restart' or check 'ayo service status'",
		Code:       ExitError,
	}
}

// ErrNoProviders creates an error when no LLM providers are configured.
func ErrNoProviders() *CLIError {
	return &CLIError{
		Brief:      "No API providers configured",
		Suggestion: "Run 'ayo setup' to configure API keys",
		Code:       ExitError,
	}
}

// ErrPermissionDenied creates an error for permission issues.
func ErrPermissionDenied(resource string) *CLIError {
	return &CLIError{
		Brief: fmt.Sprintf("Permission denied: %s", resource),
		Code:  ExitError,
	}
}

// ErrTimeout creates an error for timeout conditions.
func ErrTimeout(operation string) *CLIError {
	return &CLIError{
		Brief: fmt.Sprintf("Operation timed out: %s", operation),
		Code:  ExitError,
	}
}

// Wrap wraps an error with CLI-friendly formatting.
// If the error is already a CLIError, it returns it unchanged.
// Otherwise, it creates a new CLIError with the error as the brief.
func Wrap(err error) *CLIError {
	if err == nil {
		return nil
	}
	if cliErr, ok := err.(*CLIError); ok {
		return cliErr
	}
	return &CLIError{
		Brief: err.Error(),
		Code:  ExitError,
	}
}

// WrapWithSuggestion wraps an error and adds a suggestion.
func WrapWithSuggestion(err error, suggestion string) *CLIError {
	cliErr := Wrap(err)
	if cliErr != nil {
		cliErr.Suggestion = suggestion
	}
	return cliErr
}
