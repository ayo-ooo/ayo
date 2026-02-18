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
