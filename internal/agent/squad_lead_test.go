package agent

import (
	"testing"

	"github.com/alexcabrera/ayo/internal/squads"
)

func TestCreateSquadLead(t *testing.T) {
	baseAyo := Agent{
		Handle:         "@ayo",
		Model:          "claude-3.5-sonnet",
		System:         "You are ayo.",
		CombinedSystem: "Base system prompt.",
		Config: Config{
			Delegates: map[string]string{
				"coding": "@crush",
			},
		},
	}

	constitution := &squads.Constitution{
		Raw:       "# Team Mission\nBuild great things.",
		SquadName: "test-squad",
	}

	lead, err := CreateSquadLead(baseAyo, constitution)
	if err != nil {
		t.Fatalf("CreateSquadLead failed: %v", err)
	}

	// Should be marked as squad lead
	if !lead.IsSquadLead {
		t.Error("expected IsSquadLead to be true")
	}

	// Should have squad name set
	if lead.SquadName != "test-squad" {
		t.Errorf("SquadName = %q, want %q", lead.SquadName, "test-squad")
	}

	// Should have sandbox enabled
	if lead.Config.Sandbox.Enabled == nil || !*lead.Config.Sandbox.Enabled {
		t.Error("expected sandbox to be enabled")
	}

	// Should have delegates cleared
	if lead.Config.Delegates != nil && len(lead.Config.Delegates) > 0 {
		t.Error("expected delegates to be cleared")
	}

	// Should have constitution injected into system prompt
	if lead.CombinedSystem == baseAyo.CombinedSystem {
		t.Error("expected CombinedSystem to be modified with constitution")
	}

	// Original agent should be unchanged
	if baseAyo.IsSquadLead {
		t.Error("original agent should not be modified")
	}
	if baseAyo.Config.Delegates == nil || len(baseAyo.Config.Delegates) == 0 {
		t.Error("original agent delegates should not be modified")
	}
}

func TestCreateSquadLead_NilConstitution(t *testing.T) {
	baseAyo := Agent{
		Handle:         "@ayo",
		CombinedSystem: "Base system prompt.",
	}

	lead, err := CreateSquadLead(baseAyo, nil)
	if err != nil {
		t.Fatalf("CreateSquadLead failed: %v", err)
	}

	// Should still be marked as squad lead
	if !lead.IsSquadLead {
		t.Error("expected IsSquadLead to be true")
	}

	// Squad name should be empty
	if lead.SquadName != "" {
		t.Errorf("SquadName = %q, want empty", lead.SquadName)
	}

	// Should have sandbox enabled
	if lead.Config.Sandbox.Enabled == nil || !*lead.Config.Sandbox.Enabled {
		t.Error("expected sandbox to be enabled")
	}
}

func TestAgent_IsLeadingSquad(t *testing.T) {
	tests := []struct {
		name  string
		agent Agent
		want  bool
	}{
		{
			name:  "squad lead",
			agent: Agent{IsSquadLead: true},
			want:  true,
		},
		{
			name:  "not squad lead",
			agent: Agent{IsSquadLead: false},
			want:  false,
		},
		{
			name:  "default (zero value)",
			agent: Agent{},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.agent.IsLeadingSquad(); got != tt.want {
				t.Errorf("IsLeadingSquad() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgent_LeadingSquad(t *testing.T) {
	tests := []struct {
		name  string
		agent Agent
		want  string
	}{
		{
			name:  "has squad name",
			agent: Agent{SquadName: "test-squad"},
			want:  "test-squad",
		},
		{
			name:  "empty squad name",
			agent: Agent{},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.agent.LeadingSquad(); got != tt.want {
				t.Errorf("LeadingSquad() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCreateSquadLead_PreservesOtherFields(t *testing.T) {
	baseAyo := Agent{
		Handle:         "@ayo",
		Dir:            "/path/to/ayo",
		Model:          "claude-3.5-sonnet",
		System:         "You are ayo.",
		CombinedSystem: "Base system prompt.",
		BuiltIn:        true,
	}

	constitution := &squads.Constitution{
		Raw:       "# Mission",
		SquadName: "test-squad",
	}

	lead, err := CreateSquadLead(baseAyo, constitution)
	if err != nil {
		t.Fatalf("CreateSquadLead failed: %v", err)
	}

	// These fields should be preserved
	if lead.Handle != baseAyo.Handle {
		t.Errorf("Handle = %q, want %q", lead.Handle, baseAyo.Handle)
	}
	if lead.Dir != baseAyo.Dir {
		t.Errorf("Dir = %q, want %q", lead.Dir, baseAyo.Dir)
	}
	if lead.Model != baseAyo.Model {
		t.Errorf("Model = %q, want %q", lead.Model, baseAyo.Model)
	}
	if lead.System != baseAyo.System {
		t.Errorf("System = %q, want %q", lead.System, baseAyo.System)
	}
	if lead.BuiltIn != baseAyo.BuiltIn {
		t.Errorf("BuiltIn = %v, want %v", lead.BuiltIn, baseAyo.BuiltIn)
	}
}

func TestCreateSquadLead_SetsRestrictedTools(t *testing.T) {
	baseAyo := Agent{
		Handle:         "@ayo",
		CombinedSystem: "Base system prompt.",
	}

	constitution := &squads.Constitution{
		Raw:       "# Mission",
		SquadName: "test-squad",
	}

	lead, err := CreateSquadLead(baseAyo, constitution)
	if err != nil {
		t.Fatalf("CreateSquadLead failed: %v", err)
	}

	// Should have restricted tools set
	if len(lead.RestrictedTools) == 0 {
		t.Error("expected RestrictedTools to be set")
	}

	// Should include the standard restricted tools
	for _, tool := range SquadLeadRestrictedTools {
		if !lead.IsToolRestricted(tool) {
			t.Errorf("expected tool %q to be restricted", tool)
		}
	}
}

func TestAgent_IsToolRestricted(t *testing.T) {
	tests := []struct {
		name       string
		agent      Agent
		toolName   string
		restricted bool
	}{
		{
			name:       "tool in restricted list",
			agent:      Agent{RestrictedTools: []string{"dispatch_squad", "invoke_agent"}},
			toolName:   "dispatch_squad",
			restricted: true,
		},
		{
			name:       "tool not in restricted list",
			agent:      Agent{RestrictedTools: []string{"dispatch_squad"}},
			toolName:   "bash",
			restricted: false,
		},
		{
			name:       "empty restricted list",
			agent:      Agent{},
			toolName:   "dispatch_squad",
			restricted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.agent.IsToolRestricted(tt.toolName); got != tt.restricted {
				t.Errorf("IsToolRestricted(%q) = %v, want %v", tt.toolName, got, tt.restricted)
			}
		})
	}
}

func TestAgent_FilterRestrictedTools(t *testing.T) {
	tests := []struct {
		name     string
		agent    Agent
		tools    []string
		expected []string
	}{
		{
			name:     "filters restricted tools",
			agent:    Agent{RestrictedTools: []string{"dispatch_squad", "invoke_agent"}},
			tools:    []string{"bash", "dispatch_squad", "edit", "invoke_agent", "view"},
			expected: []string{"bash", "edit", "view"},
		},
		{
			name:     "no restricted tools",
			agent:    Agent{},
			tools:    []string{"bash", "dispatch_squad", "edit"},
			expected: []string{"bash", "dispatch_squad", "edit"},
		},
		{
			name:     "all tools restricted",
			agent:    Agent{RestrictedTools: []string{"a", "b", "c"}},
			tools:    []string{"a", "b", "c"},
			expected: []string{},
		},
		{
			name:     "no tools provided",
			agent:    Agent{RestrictedTools: []string{"dispatch_squad"}},
			tools:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.agent.FilterRestrictedTools(tt.tools)
			if len(got) != len(tt.expected) {
				t.Errorf("FilterRestrictedTools() = %v, want %v", got, tt.expected)
				return
			}
			for i, tool := range got {
				if tool != tt.expected[i] {
					t.Errorf("FilterRestrictedTools()[%d] = %q, want %q", i, tool, tt.expected[i])
				}
			}
		})
	}
}

func TestAgent_RestrictToolsForSquadLead(t *testing.T) {
	t.Run("adds base restrictions", func(t *testing.T) {
		agent := Agent{IsSquadLead: true}
		agent.RestrictToolsForSquadLead()

		if len(agent.RestrictedTools) != len(SquadLeadRestrictedTools) {
			t.Errorf("RestrictedTools length = %d, want %d", len(agent.RestrictedTools), len(SquadLeadRestrictedTools))
		}
	})

	t.Run("adds additional restrictions", func(t *testing.T) {
		agent := Agent{IsSquadLead: true}
		agent.RestrictToolsForSquadLead("custom_tool", "another_tool")

		expectedLen := len(SquadLeadRestrictedTools) + 2
		if len(agent.RestrictedTools) != expectedLen {
			t.Errorf("RestrictedTools length = %d, want %d", len(agent.RestrictedTools), expectedLen)
		}

		if !agent.IsToolRestricted("custom_tool") {
			t.Error("expected custom_tool to be restricted")
		}
		if !agent.IsToolRestricted("another_tool") {
			t.Error("expected another_tool to be restricted")
		}
	})

	t.Run("does nothing for non-squad-lead", func(t *testing.T) {
		agent := Agent{IsSquadLead: false}
		agent.RestrictToolsForSquadLead("tool")

		if len(agent.RestrictedTools) != 0 {
			t.Errorf("RestrictedTools should be empty for non-squad-lead, got %v", agent.RestrictedTools)
		}
	})

	t.Run("does not duplicate restrictions", func(t *testing.T) {
		agent := Agent{IsSquadLead: true}
		agent.RestrictToolsForSquadLead(SquadLeadRestrictedTools[0])

		// Should not have duplicates
		count := 0
		for _, tool := range agent.RestrictedTools {
			if tool == SquadLeadRestrictedTools[0] {
				count++
			}
		}
		if count != 1 {
			t.Errorf("expected %q to appear once, got %d times", SquadLeadRestrictedTools[0], count)
		}
	})
}
