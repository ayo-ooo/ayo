// Package guardrails provides security guardrails for agent invocation.
//
// The sandwich pattern wraps untrusted agent prompts with system instructions:
//   - PREFIX establishes context and rules BEFORE the agent prompt
//   - SUFFIX reinforces rules and provides final authoritative instructions
//
// This makes prompt injection much harder because the agent's prompt is treated
// as content sandwiched between system instructions, not as instructions itself.
package guardrails

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Guardrails contains the PREFIX and SUFFIX used to sandwich agent prompts.
type Guardrails struct {
	Prefix string // System instructions BEFORE agent prompt
	Suffix string // System instructions AFTER agent prompt
}

// WrapContext provides template data for PREFIX/SUFFIX rendering.
type WrapContext struct {
	TrustLevel        string // Agent's trust level (sandboxed, privileged, unrestricted)
	SessionID         string // Current session identifier
	AgentName         string // Name of the agent being invoked
	AgentHandle       string // Handle of the agent (e.g., "@coder")
	OrchestratorAgent string // Name of the orchestrating agent (e.g., "@ayo")
	SessionRoom       string // Matrix room for the session (deprecated, use TicketsDir)
	TicketsDir        string // Path to tickets directory (e.g., "/workspace/.tickets")
}

// Wrap wraps an agent prompt with PREFIX and SUFFIX guardrails.
// The agent prompt is placed between markers to make it clear it's untrusted.
func (g *Guardrails) Wrap(agentPrompt string, ctx WrapContext) string {
	prefix := g.renderTemplate(g.Prefix, ctx)
	suffix := g.renderTemplate(g.Suffix, ctx)

	parts := make([]string, 0, 3)
	if prefix != "" {
		parts = append(parts, prefix)
	}
	parts = append(parts, "[AGENT_PROMPT_START]")
	parts = append(parts, agentPrompt)
	parts = append(parts, "[AGENT_PROMPT_END]")
	if suffix != "" {
		parts = append(parts, suffix)
	}

	return strings.Join(parts, "\n\n")
}

// WrapSimple wraps an agent prompt without template rendering.
// Use this when context is not available or not needed.
func (g *Guardrails) WrapSimple(agentPrompt string) string {
	parts := make([]string, 0, 3)
	if g.Prefix != "" {
		parts = append(parts, g.Prefix)
	}
	parts = append(parts, "[AGENT_PROMPT_START]")
	parts = append(parts, agentPrompt)
	parts = append(parts, "[AGENT_PROMPT_END]")
	if g.Suffix != "" {
		parts = append(parts, g.Suffix)
	}

	return strings.Join(parts, "\n\n")
}

// renderTemplate renders a template string with the given context.
// If rendering fails, returns the original string unchanged.
func (g *Guardrails) renderTemplate(text string, ctx WrapContext) string {
	if !strings.Contains(text, "{{") {
		return text
	}

	tmpl, err := template.New("guardrail").Parse(text)
	if err != nil {
		return text
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return text
	}

	return buf.String()
}

// Load returns Guardrails with custom templates from the config directory,
// falling back to defaults if custom files don't exist.
func Load(configDir string) *Guardrails {
	g := &Guardrails{
		Prefix: DefaultPrefix,
		Suffix: DefaultSuffix,
	}

	guardrailsDir := filepath.Join(configDir, "guardrails")

	// Load custom prefix if exists
	if prefixPath := filepath.Join(guardrailsDir, "prefix.txt"); fileExists(prefixPath) {
		if content, err := os.ReadFile(prefixPath); err == nil {
			g.Prefix = strings.TrimSpace(string(content))
		}
	}

	// Load custom suffix if exists
	if suffixPath := filepath.Join(guardrailsDir, "suffix.txt"); fileExists(suffixPath) {
		if content, err := os.ReadFile(suffixPath); err == nil {
			g.Suffix = strings.TrimSpace(string(content))
		}
	}

	return g
}

// Default returns Guardrails with default PREFIX and SUFFIX.
func Default() *Guardrails {
	return &Guardrails{
		Prefix: DefaultPrefix,
		Suffix: DefaultSuffix,
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
