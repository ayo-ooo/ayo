// Package pipe provides utilities for detecting pipeline execution
// and managing chain context between agents.
package pipe

import (
	"encoding/json"
	"io"
	"os"

	"golang.org/x/term"
)

// ChainContextEnvVar is the environment variable used to pass chain context.
const ChainContextEnvVar = "AYO_CHAIN_CONTEXT"

// ChainContext contains metadata about the current chain execution.
type ChainContext struct {
	// Depth is the position in the chain (1 = first agent after initial input)
	Depth int `json:"depth"`

	// Source is the handle of the agent that produced the input
	Source string `json:"source,omitempty"`

	// SourceDescription is a human-readable description of the source agent
	SourceDescription string `json:"source_description,omitempty"`
}

// IsStdinPiped returns true if stdin is receiving piped input.
func IsStdinPiped() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	// Check if stdin is a pipe or has data
	return (stat.Mode()&os.ModeCharDevice) == 0 || stat.Size() > 0
}

// IsStdoutPiped returns true if stdout is being piped to another process.
func IsStdoutPiped() bool {
	return !term.IsTerminal(int(os.Stdout.Fd()))
}

// ReadStdin reads all available data from stdin.
// Returns empty string if stdin is not piped or has no data.
func ReadStdin() (string, error) {
	if !IsStdinPiped() {
		return "", nil
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetChainContext retrieves chain context from the environment.
// Returns nil if not in a chain.
func GetChainContext() *ChainContext {
	envVal := os.Getenv(ChainContextEnvVar)
	if envVal == "" {
		return nil
	}

	var ctx ChainContext
	if err := json.Unmarshal([]byte(envVal), &ctx); err != nil {
		return nil
	}
	return &ctx
}
