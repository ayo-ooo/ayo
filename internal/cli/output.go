// Package cli provides CLI output helpers for consistent JSON and quiet mode support.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
)

// ExitCode defines standard exit codes for CLI commands.
type ExitCode int

const (
	ExitSuccess      ExitCode = 0
	ExitError        ExitCode = 1
	ExitInvalidInput ExitCode = 2
	ExitNotFound     ExitCode = 3
)

// Output provides structured output helpers for CLI commands.
type Output struct {
	JSON  bool
	Quiet bool
}

// Result represents a standard JSON response structure.
type Result struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// Print outputs data in the appropriate format.
// In JSON mode, outputs data as JSON.
// In quiet mode, outputs nothing.
// Otherwise, outputs the text.
func (o *Output) Print(data any, text string) {
	if o.JSON {
		o.printJSON(Result{Success: true, Data: data})
		return
	}
	if o.Quiet {
		return
	}
	fmt.Println(text)
}

// Printf outputs formatted text (non-JSON, non-quiet only).
func (o *Output) Printf(format string, args ...any) {
	if o.JSON || o.Quiet {
		return
	}
	fmt.Printf(format, args...)
}

// PrintData outputs structured data with a formatted text fallback.
func (o *Output) PrintData(data any, format string, args ...any) {
	if o.JSON {
		o.printJSON(Result{Success: true, Data: data})
		return
	}
	if o.Quiet {
		return
	}
	fmt.Printf(format, args...)
}

// PrintSuccess outputs a success message.
func (o *Output) PrintSuccess(text string) {
	if o.JSON {
		o.printJSON(Result{Success: true, Message: text})
		return
	}
	if o.Quiet {
		return
	}
	fmt.Println(text)
}

// PrintError outputs an error.
func (o *Output) PrintError(err error) {
	if o.JSON {
		o.printJSON(Result{Success: false, Error: err.Error()})
		return
	}
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}

// PrintResult outputs a full result structure.
func (o *Output) PrintResult(r Result) {
	if o.JSON {
		o.printJSON(r)
		return
	}
	if r.Success {
		if !o.Quiet && r.Message != "" {
			fmt.Println(r.Message)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", r.Error)
	}
}

// printJSON outputs data as JSON to stdout.
func (o *Output) printJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// MarshalJSON outputs data as JSON, returning the bytes.
func MarshalJSON(data any) ([]byte, error) {
	return json.MarshalIndent(Result{Success: true, Data: data}, "", "  ")
}
