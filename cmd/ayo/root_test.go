package main

import (
	"testing"
)

func TestParseInvocation(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedType   InvocationType
		expectedHandle string
		expectedPrompt []string
	}{
		{
			name:           "no args defaults to @ayo agent",
			args:           []string{},
			expectedType:   InvocationTypeAgent,
			expectedHandle: "@ayo",
			expectedPrompt: nil,
		},
		{
			name:           "nil args defaults to @ayo agent",
			args:           nil,
			expectedType:   InvocationTypeAgent,
			expectedHandle: "@ayo",
			expectedPrompt: nil,
		},
		{
			name:           "agent handle with @",
			args:           []string{"@myagent", "do", "something"},
			expectedType:   InvocationTypeAgent,
			expectedHandle: "@myagent",
			expectedPrompt: []string{"do", "something"},
		},
		{
			name:           "agent handle without prompt",
			args:           []string{"@myagent"},
			expectedType:   InvocationTypeAgent,
			expectedHandle: "@myagent",
			expectedPrompt: []string{},
		},
		{
			name:           "squad handle with #",
			args:           []string{"#frontend", "build", "feature"},
			expectedType:   InvocationTypeSquad,
			expectedHandle: "#frontend",
			expectedPrompt: []string{"build", "feature"},
		},
		{
			name:           "squad handle without prompt",
			args:           []string{"#frontend"},
			expectedType:   InvocationTypeSquad,
			expectedHandle: "#frontend",
			expectedPrompt: []string{},
		},
		{
			name:           "plain prompt defaults to @ayo",
			args:           []string{"tell", "me", "a", "joke"},
			expectedType:   InvocationTypeAgent,
			expectedHandle: "@ayo",
			expectedPrompt: []string{"tell", "me", "a", "joke"},
		},
		{
			name:           "single word prompt defaults to @ayo",
			args:           []string{"hello"},
			expectedType:   InvocationTypeAgent,
			expectedHandle: "@ayo",
			expectedPrompt: []string{"hello"},
		},
		{
			name:           "squad handle with hyphen",
			args:           []string{"#frontend-team", "deploy"},
			expectedType:   InvocationTypeSquad,
			expectedHandle: "#frontend-team",
			expectedPrompt: []string{"deploy"},
		},
		{
			name:           "squad handle with underscore",
			args:           []string{"#data_pipeline", "run"},
			expectedType:   InvocationTypeSquad,
			expectedHandle: "#data_pipeline",
			expectedPrompt: []string{"run"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseInvocation(tt.args)

			if result.Type != tt.expectedType {
				t.Errorf("Type = %v, want %v", result.Type, tt.expectedType)
			}
			if result.Handle != tt.expectedHandle {
				t.Errorf("Handle = %v, want %v", result.Handle, tt.expectedHandle)
			}
			if len(result.PromptArgs) != len(tt.expectedPrompt) {
				t.Errorf("PromptArgs length = %v, want %v", len(result.PromptArgs), len(tt.expectedPrompt))
			} else {
				for i, arg := range result.PromptArgs {
					if arg != tt.expectedPrompt[i] {
						t.Errorf("PromptArgs[%d] = %v, want %v", i, arg, tt.expectedPrompt[i])
					}
				}
			}
		})
	}
}

func TestLooksLikeSubcommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "lowercase word looks like subcommand",
			input:    "setup",
			expected: true,
		},
		{
			name:     "lowercase with hyphen",
			input:    "run-test",
			expected: true,
		},
		{
			name:     "lowercase with underscore",
			input:    "run_test",
			expected: true,
		},
		{
			name:     "contains space is prompt",
			input:    "tell me",
			expected: false,
		},
		{
			name:     "starts with quote",
			input:    `"hello`,
			expected: false,
		},
		{
			name:     "contains uppercase",
			input:    "Setup",
			expected: false,
		},
		{
			name:     "contains number",
			input:    "test123",
			expected: false,
		},
		{
			name:     "starts with @",
			input:    "@agent",
			expected: false,
		},
		{
			name:     "starts with #",
			input:    "#squad",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := looksLikeSubcommand(tt.input)
			if result != tt.expected {
				t.Errorf("looksLikeSubcommand(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
