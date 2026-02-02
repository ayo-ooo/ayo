package plugins

import (
	"testing"
)

func TestConflictType_String(t *testing.T) {
	tests := []struct {
		ct   ConflictType
		want string
	}{
		{ConflictAgent, "agent"},
		{ConflictSkill, "skill"},
		{ConflictTool, "tool"},
		{ConflictType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.ct.String(); got != tt.want {
				t.Errorf("ConflictType.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolutionAction_String(t *testing.T) {
	tests := []struct {
		ra   ResolutionAction
		want string
	}{
		{ResolutionSkip, "skip"},
		{ResolutionReplace, "replace"},
		{ResolutionRename, "rename"},
		{ResolutionAction(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.ra.String(); got != tt.want {
				t.Errorf("ResolutionAction.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSuggestRename(t *testing.T) {
	tests := []struct {
		name         string
		conflictType ConflictType
		want         string
	}{
		{"@myagent", ConflictAgent, "@myagent-plugin"},
		{"myagent", ConflictAgent, "@myagent-plugin"},
		{"my-skill", ConflictSkill, "my-skill-plugin"},
		{"my-tool", ConflictTool, "my-tool-ext"},
		{"something", ConflictType(99), "something-alt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SuggestRename(tt.name, tt.conflictType)
			if got != tt.want {
				t.Errorf("SuggestRename(%q, %v) = %q, want %q", tt.name, tt.conflictType, got, tt.want)
			}
		})
	}
}

func TestValidateAgentRename_EmptyHandle(t *testing.T) {
	err := ValidateAgentRename("@old", "")
	if err == nil {
		t.Error("expected error for empty new handle")
	}
}

func TestValidateAgentRename_ReservedNamespace(t *testing.T) {
	tests := []string{
		"@ayo",
		"@ayo.test",
		"ayo",       // Will be normalized to @ayo
		"@ayo-test", // Also reserved
	}

	for _, newHandle := range tests {
		t.Run(newHandle, func(t *testing.T) {
			err := ValidateAgentRename("@old", newHandle)
			if err == nil {
				t.Errorf("expected error for reserved namespace %q", newHandle)
			}
		})
	}
}

func TestConflict_Fields(t *testing.T) {
	c := Conflict{
		Type:         ConflictAgent,
		Name:         "@test",
		ExistingPath: "/path/to/existing",
		NewPath:      "/path/to/new",
		Source:       "builtin",
	}

	if c.Type != ConflictAgent {
		t.Errorf("Type = %v, want %v", c.Type, ConflictAgent)
	}
	if c.Name != "@test" {
		t.Errorf("Name = %q, want %q", c.Name, "@test")
	}
	if c.ExistingPath != "/path/to/existing" {
		t.Errorf("ExistingPath = %q, want %q", c.ExistingPath, "/path/to/existing")
	}
	if c.NewPath != "/path/to/new" {
		t.Errorf("NewPath = %q, want %q", c.NewPath, "/path/to/new")
	}
	if c.Source != "builtin" {
		t.Errorf("Source = %q, want %q", c.Source, "builtin")
	}
}

func TestConflictResolution_Fields(t *testing.T) {
	cr := ConflictResolution{
		Conflict: Conflict{
			Type: ConflictAgent,
			Name: "@test",
		},
		Action:   ResolutionRename,
		RenameTo: "@test-alt",
	}

	if cr.Action != ResolutionRename {
		t.Errorf("Action = %v, want %v", cr.Action, ResolutionRename)
	}
	if cr.RenameTo != "@test-alt" {
		t.Errorf("RenameTo = %q, want %q", cr.RenameTo, "@test-alt")
	}
}

func TestDetectConflicts_EmptyManifest(t *testing.T) {
	manifest := &Manifest{
		Name:    "test-plugin",
		Version: "1.0.0",
		Agents:  []string{},
		Skills:  []string{},
		Tools:   []string{},
	}

	conflicts, err := DetectConflicts(manifest, "/tmp/test-plugin")
	if err != nil {
		t.Fatalf("DetectConflicts failed: %v", err)
	}

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts for empty manifest, got %d", len(conflicts))
	}
}

func TestDetectConflicts_BuiltinToolConflict(t *testing.T) {
	manifest := &Manifest{
		Name:    "test-plugin",
		Version: "1.0.0",
		Tools:   []string{"bash"}, // Built-in tool
	}

	conflicts, err := DetectConflicts(manifest, "/tmp/test-plugin")
	if err != nil {
		t.Fatalf("DetectConflicts failed: %v", err)
	}

	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict for builtin tool, got %d", len(conflicts))
	}

	c := conflicts[0]
	if c.Type != ConflictTool {
		t.Errorf("Type = %v, want %v", c.Type, ConflictTool)
	}
	if c.Name != "bash" {
		t.Errorf("Name = %q, want %q", c.Name, "bash")
	}
	if c.Source != "builtin" {
		t.Errorf("Source = %q, want %q", c.Source, "builtin")
	}
}
