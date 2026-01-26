package tools

import (
	"testing"

	"github.com/alexcabrera/ayo/internal/config"
)

func TestIsCategory(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"planning is category", "planning", true},
		{"shell is category", "shell", true},
		{"bash is not category", "bash", false},
		{"todo is not category", "todo", false},
		{"random is not category", "random", false},
		{"empty is not category", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCategory(tt.input)
			if got != tt.expected {
				t.Errorf("IsCategory(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDefaultForCategory(t *testing.T) {
	tests := []struct {
		name     string
		cat      Category
		expected string
	}{
		{"planning default", CategoryPlanning, "todo"},
		{"shell default", CategoryShell, "bash"},
		{"search has no default", CategorySearch, ""},
		{"unknown category", Category("unknown"), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultForCategory(tt.cat)
			if got != tt.expected {
				t.Errorf("DefaultForCategory(%q) = %q, want %q", tt.cat, got, tt.expected)
			}
		})
	}
}

func TestResolveToolName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		cfg      *config.Config
		expected string
	}{
		{
			name:     "category with no config",
			input:    "planning",
			cfg:      nil,
			expected: "todo",
		},
		{
			name:     "category with empty config",
			input:    "planning",
			cfg:      &config.Config{},
			expected: "todo",
		},
		{
			name:  "category with override",
			input: "planning",
			cfg: &config.Config{
				DefaultTools: map[string]string{"planning": "plan"},
			},
			expected: "plan",
		},
		{
			name:     "shell category default",
			input:    "shell",
			cfg:      nil,
			expected: "bash",
		},
		{
			name:  "shell category override",
			input: "shell",
			cfg: &config.Config{
				DefaultTools: map[string]string{"shell": "nushell"},
			},
			expected: "nushell",
		},
		{
			name:     "non-category tool",
			input:    "memory",
			cfg:      nil,
			expected: "memory",
		},
		{
			name:  "non-category with alias",
			input: "search",
			cfg: &config.Config{
				DefaultTools: map[string]string{"search": "searxng"},
			},
			expected: "searxng",
		},
		{
			name:     "concrete tool passthrough",
			input:    "bash",
			cfg:      nil,
			expected: "bash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveToolName(tt.input, tt.cfg)
			if got != tt.expected {
				t.Errorf("ResolveToolName(%q, cfg) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestListCategories(t *testing.T) {
	cats := ListCategories()

	// Verify expected categories exist
	if cats[CategoryPlanning] != "todo" {
		t.Errorf("expected planning -> todo, got %q", cats[CategoryPlanning])
	}
	if cats[CategoryShell] != "bash" {
		t.Errorf("expected shell -> bash, got %q", cats[CategoryShell])
	}

	// Verify returned map is a copy (mutation shouldn't affect original)
	cats[CategoryPlanning] = "modified"
	if DefaultForCategory(CategoryPlanning) == "modified" {
		t.Error("ListCategories should return a copy, not the original map")
	}
}
