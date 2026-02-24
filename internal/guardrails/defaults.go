// Package guardrails provides default security prompts for agent sandwiching.
package guardrails

import (
	"github.com/alexcabrera/ayo/internal/prompts"
)

// DefaultPrefix returns the prefix prompt, loading from prompts directory
// or falling back to embedded default.
func DefaultPrefix() string {
	return prompts.Default().LoadWithEmbeddedFallback(prompts.PathSandwichPrefix)
}

// DefaultSuffix returns the suffix prompt, loading from prompts directory
// or falling back to embedded default.
func DefaultSuffix() string {
	return prompts.Default().LoadWithEmbeddedFallback(prompts.PathSandwichSuffix)
}

// LegacyGuardrails returns the legacy guardrails prompt, loading from prompts
// directory or falling back to embedded default.
func LegacyGuardrails() string {
	return prompts.Default().LoadWithEmbeddedFallback(prompts.PathGuardrailsDefault)
}
