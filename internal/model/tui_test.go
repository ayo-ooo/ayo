package model

import (
	"testing"

	"charm.land/catwalk/pkg/catwalk"
)

func TestTUI_InitialState(t *testing.T) {
	tui := NewTUI()

	if tui == nil {
		t.Fatal("NewTUI should not return nil")
	}
}

func TestTUI_HasProviders(t *testing.T) {
	tui := NewTUI()

	if tui.providers == nil {
		t.Error("Expected providers to be initialized")
	}
}

func TestTUI_State(t *testing.T) {
	tui := NewTUI()

	if tui.state != stateProviderSelect {
		t.Errorf("Expected initial state to be stateProviderSelect, got %v", tui.state)
	}
}

func TestProviderItem_FilterValue(t *testing.T) {
	p := ProviderItem{
		provider:  catwalk.Provider{Name: "OpenAI"},
		hasAPIKey: true,
	}

	if p.FilterValue() != "OpenAI" {
		t.Errorf("Expected FilterValue to return 'OpenAI', got %q", p.FilterValue())
	}
}

func TestModelItem_FilterValue(t *testing.T) {
	m := ModelItem{
		model: catwalk.Model{Name: "GPT-4o"},
	}

	if m.FilterValue() != "GPT-4o" {
		t.Errorf("Expected FilterValue to return 'GPT-4o', got %q", m.FilterValue())
	}
}
