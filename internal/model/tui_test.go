package model

import (
	"testing"
)

func TestTUI_InitialState(t *testing.T) {
	availableProviders := map[string]bool{
		"anthropic": true,
		"openai":    true,
	}

	tui := NewTUI(availableProviders)

	if tui == nil {
		t.Fatal("NewTUI should not return nil")
	}

	if len(tui.AvailableProviders) != 2 {
		t.Errorf("Expected 2 available providers, got %d", len(tui.AvailableProviders))
	}
}

func TestTUI_ProviderSelection(t *testing.T) {
	availableProviders := map[string]bool{
		"anthropic": true,
		"openai":    true,
	}

	tui := NewTUI(availableProviders)
	tui.SelectProvider(0)

	if tui.SelectedProvider != "anthropic" {
		t.Errorf("Expected selected provider 'anthropic', got %q", tui.SelectedProvider)
	}
}

func TestTUI_ModelSelection(t *testing.T) {
	availableProviders := map[string]bool{
		"anthropic": true,
	}

	tui := NewTUI(availableProviders)
	tui.SelectProvider(0)
	tui.SelectModel(0)

	if tui.SelectedModel == "" {
		t.Error("Expected a model to be selected")
	}
}

func TestTUI_Confirmation(t *testing.T) {
	availableProviders := map[string]bool{
		"anthropic": true,
	}

	tui := NewTUI(availableProviders)
	tui.SelectProvider(0)
	tui.SelectModel(0)

	if !tui.IsComplete() {
		t.Error("TUI should be complete after provider and model selection")
	}
}

func TestTUI_ProviderList(t *testing.T) {
	availableProviders := map[string]bool{
		"anthropic":  true,
		"openai":     true,
		"openrouter": true,
	}

	tui := NewTUI(availableProviders)
	providers := tui.ProviderList()

	if len(providers) != 3 {
		t.Errorf("Expected 3 providers in list, got %d", len(providers))
	}
}

func TestTUI_ModelList(t *testing.T) {
	availableProviders := map[string]bool{
		"anthropic": true,
	}

	tui := NewTUI(availableProviders)
	tui.SelectProvider(0)
	models := tui.ModelList()

	if len(models) == 0 {
		t.Error("Expected models to be available for selected provider")
	}
}

func TestTUI_NoProvidersAvailable(t *testing.T) {
	availableProviders := map[string]bool{}

	tui := NewTUI(availableProviders)

	if len(tui.AvailableProviders) != 0 {
		t.Error("Should have no available providers")
	}

	providers := tui.ProviderList()
	if len(providers) != 0 {
		t.Error("Provider list should be empty")
	}
}
