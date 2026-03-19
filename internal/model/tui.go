package model

import (
	"fmt"
	"slices"
	"sort"
)

// TUI provides a terminal UI for model selection
type TUI struct {
	AvailableProviders map[string]bool
	SelectedProvider   string
	SelectedModel      string
	providers          []string
	models             map[string][]string
}

// NewTUI creates a new model selection TUI
func NewTUI(availableProviders map[string]bool) *TUI {
	// Build sorted list of available providers
	var providers []string
	for p, available := range availableProviders {
		if available {
			providers = append(providers, p)
		}
	}
	sort.Strings(providers)

	return &TUI{
		AvailableProviders: availableProviders,
		providers:          providers,
		models:             getProviderModels(),
	}
}

// ProviderList returns a sorted list of available providers
func (t *TUI) ProviderList() []string {
	return t.providers
}

// SelectProvider selects a provider by index
func (t *TUI) SelectProvider(idx int) error {
	if idx < 0 || idx >= len(t.providers) {
		return fmt.Errorf("invalid provider index: %d", idx)
	}
	t.SelectedProvider = t.providers[idx]
	return nil
}

// ModelList returns the list of models for the selected provider
func (t *TUI) ModelList() []string {
	if t.SelectedProvider == "" {
		return nil
	}
	return t.models[t.SelectedProvider]
}

// SelectModel selects a model by index
func (t *TUI) SelectModel(idx int) error {
	models := t.ModelList()
	if idx < 0 || idx >= len(models) {
		return fmt.Errorf("invalid model index: %d", idx)
	}
	t.SelectedModel = models[idx]
	return nil
}

// getProviderModels returns the default models for each provider
func getProviderModels() map[string][]string {
	return map[string][]string{
		"anthropic": {
			"claude-3-5-sonnet-20241022",
			"claude-3-5-haiku-20241022",
			"claude-3-opus-20240229",
			"claude-3-sonnet-20240229",
			"claude-3-haiku-20240307",
		},
		"openai": {
			"gpt-4o",
			"gpt-4o-mini",
			"gpt-4-turbo",
			"gpt-4",
			"gpt-3.5-turbo",
		},
		"zai": {
			"zai-v4",
		},
		"openrouter": {
			"anthropic/claude-3.5-sonnet",
			"anthropic/claude-3-opus",
			"openai/gpt-4o",
			"openai/gpt-4-turbo",
			"google/gemini-pro-1.5",
		},
		"gemini": {
			"gemini-1.5-pro",
			"gemini-1.5-flash",
			"gemini-1.0-pro",
		},
		"groq": {
			"llama-3.3-70b-versatile",
			"llama-3.1-70b-versatile",
			"llama-3.1-8b-instant",
			"mixtral-8x7b-32768",
		},
	}
}

// HasProviders returns true if there are any available providers
func (t *TUI) HasProviders() bool {
	return len(t.providers) > 0
}

// IsComplete returns true if both provider and model are selected
func (t *TUI) IsComplete() bool {
	return t.SelectedProvider != "" && t.SelectedModel != ""
}

// Run runs the TUI and returns the selected provider and model
// This is a stub that should be replaced with actual Bubble Tea implementation
func (t *TUI) Run() (provider, model string, err error) {
	// Stub implementation - in a real implementation, this would launch
	// a Bubble Tea TUI for interactive selection
	if !t.HasProviders() {
		return "", "", fmt.Errorf("no providers available")
	}

	// Default to first provider and first model
	if err := t.SelectProvider(0); err != nil {
		return "", "", err
	}

	models := t.ModelList()
	if len(models) == 0 {
		return t.SelectedProvider, "", fmt.Errorf("no models available for provider %s", t.SelectedProvider)
	}

	t.SelectedModel = models[0]

	// Sort providers to get deterministic "first" provider
	providers := slices.Clone(t.providers)
	sort.Strings(providers)
	if len(providers) > 0 {
		t.SelectedProvider = providers[0]
		t.SelectedModel = t.models[t.SelectedProvider][0]
	}

	return t.SelectedProvider, t.SelectedModel, nil
}
