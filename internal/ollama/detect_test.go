package ollama

import (
	"testing"
)

func TestStatusString(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{StatusNotInstalled, "not installed"},
		{StatusInstalled, "installed (not running)"},
		{StatusRunning, "running"},
	}

	for _, tt := range tests {
		got := tt.status.String()
		if got != tt.want {
			t.Errorf("Status(%d).String() = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestIsCapableForChat(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		// Exact matches
		{"mistral:7b", true},
		{"llama3:8b", true},
		{"ministral:3b", true},

		// With variants
		{"llama3:8b-instruct", true},
		{"mistral:7b-instruct-q4_K_M", true},

		// Non-capable models
		{"unknown-model:1b", false},
		{"nomic-embed-text", false},

		// Edge cases
		{"", false},
		{"mistral", true}, // Base name match
	}

	for _, tt := range tests {
		got := IsCapableForChat(tt.name)
		if got != tt.want {
			t.Errorf("IsCapableForChat(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestGetBaseName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"mistral:7b", "mistral"},
		{"llama3:8b-instruct", "llama3"},
		{"phi3:mini", "phi3"},
		{"noname", "noname"},
		{"", ""},
	}

	for _, tt := range tests {
		got := getBaseName(tt.name)
		if got != tt.want {
			t.Errorf("getBaseName(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestGetSuggestedModels(t *testing.T) {
	models := GetSuggestedModels()
	if len(models) == 0 {
		t.Error("expected at least one suggested model")
	}

	// All suggested models should be capable
	for _, m := range models {
		if !IsCapableForChat(m.Name) {
			t.Errorf("suggested model %q should be capable", m.Name)
		}
	}

	// Should have the recommended model
	var hasRecommended bool
	for _, m := range models {
		if m.Name == RecommendedModel {
			hasRecommended = true
			break
		}
	}
	if !hasRecommended {
		t.Errorf("suggested models should include recommended model %q", RecommendedModel)
	}
}

func TestCapableModelsNotEmpty(t *testing.T) {
	if len(CapableModels) == 0 {
		t.Error("CapableModels should not be empty")
	}
}
