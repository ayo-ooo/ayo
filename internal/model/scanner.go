package model

import (
	"os"
	"strings"
)

// Scanner scans the environment for available API keys.
type Scanner struct{}

// NewScanner creates a new environment scanner.
func NewScanner() *Scanner {
	return &Scanner{}
}

// providerEnvMap maps provider names to their environment variable names.
var providerEnvMap = map[string]string{
	"anthropic":  "ANTHROPIC_API_KEY",
	"openai":     "OPENAI_API_KEY",
	"zai":        "ZAI_API_KEY",
	"openrouter": "OPENROUTER_API_KEY",
	"gemini":     "GEMINI_API_KEY",
	"groq":       "GROQ_API_KEY",
}

// Scan checks for available API keys and returns a map of available providers.
func (s *Scanner) Scan() map[string]bool {
	providers := make(map[string]bool)

	for provider, envVar := range providerEnvMap {
		value := os.Getenv(envVar)
		value = strings.TrimSpace(value)
		if value != "" {
			providers[provider] = true
		}
	}

	return providers
}
