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
		{"plan is category", "plan", true},
		{"shell is category", "shell", true},
		{"search is category", "search", true},
		{"bash is not category", "bash", false},
		{"todo is not category", "todo", false},
		{"random is not category", "random", false},
		{"empty is not category", "", false},
		{"planning is not category (old name)", "planning", false},
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
		{"plan has no default", CategoryPlan, ""},
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
			name:     "plan category with no config returns empty",
			input:    "plan",
			cfg:      nil,
			expected: "",
		},
		{
			name:     "plan category with empty config returns empty",
			input:    "plan",
			cfg:      &config.Config{},
			expected: "",
		},
		{
			name:  "plan category with override returns configured tool",
			input: "plan",
			cfg: &config.Config{
				DefaultTools: map[string]string{"plan": "ticket"},
			},
			expected: "ticket",
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
			name:     "search category with no config returns empty",
			input:    "search",
			cfg:      nil,
			expected: "",
		},
		{
			name:  "search category with override returns configured tool",
			input: "search",
			cfg: &config.Config{
				DefaultTools: map[string]string{"search": "searxng"},
			},
			expected: "searxng",
		},
		{
			name:     "non-category tool passthrough",
			input:    "memory",
			cfg:      nil,
			expected: "memory",
		},
		{
			name:     "concrete tool passthrough",
			input:    "bash",
			cfg:      nil,
			expected: "bash",
		},
		{
			name:     "todo tool passthrough (not a category)",
			input:    "todo",
			cfg:      nil,
			expected: "todo",
		},
		{
			name:     "old planning name passthrough (not a category anymore)",
			input:    "planning",
			cfg:      nil,
			expected: "planning",
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
	if _, ok := cats[CategoryPlan]; !ok {
		t.Error("expected plan category to exist")
	}
	if cats[CategoryPlan] != "" {
		t.Errorf("expected plan to have no default, got %q", cats[CategoryPlan])
	}
	if cats[CategoryShell] != "bash" {
		t.Errorf("expected shell -> bash, got %q", cats[CategoryShell])
	}
	if _, ok := cats[CategorySearch]; !ok {
		t.Error("expected search category to exist")
	}
	if cats[CategorySearch] != "" {
		t.Errorf("expected search to have no default, got %q", cats[CategorySearch])
	}

	// Verify all three categories are returned
	if len(cats) != 3 {
		t.Errorf("expected 3 categories, got %d", len(cats))
	}

	// Verify returned map is a copy (mutation shouldn't affect original)
	cats[CategoryShell] = "modified"
	if DefaultForCategory(CategoryShell) == "modified" {
		t.Error("ListCategories should return a copy, not the original map")
	}
}

func TestGetExecutionContext(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		expected ExecutionContext
	}{
		{"memory is host", "memory", ExecHost},
		{"agent_call is host", "agent_call", ExecHost},
		{"delegate is host", "delegate", ExecHost},
		{"todo is host", "todo", ExecHost},
		{"bash is sandbox", "bash", ExecSandbox},
		{"file_request is bridge", "file_request", ExecBridge},
		{"publish is bridge", "publish", ExecBridge},
		{"unknown defaults to sandbox", "unknown_tool", ExecSandbox},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetExecutionContext(tt.toolName)
			if got != tt.expected {
				t.Errorf("GetExecutionContext(%q) = %q, want %q", tt.toolName, got, tt.expected)
			}
		})
	}
}

func TestIsHostTool(t *testing.T) {
	if !IsHostTool("memory") {
		t.Error("memory should be host tool")
	}
	if IsHostTool("bash") {
		t.Error("bash should not be host tool")
	}
}

func TestIsSandboxTool(t *testing.T) {
	if !IsSandboxTool("bash") {
		t.Error("bash should be sandbox tool")
	}
	if IsSandboxTool("memory") {
		t.Error("memory should not be sandbox tool")
	}
}

func TestIsBridgeTool(t *testing.T) {
	if !IsBridgeTool("file_request") {
		t.Error("file_request should be bridge tool")
	}
	if !IsBridgeTool("publish") {
		t.Error("publish should be bridge tool")
	}
	if IsBridgeTool("bash") {
		t.Error("bash should not be bridge tool")
	}
}
