package skills

import (
	"sort"
	"testing"
)

func TestGetRequiredSkillsForTools(t *testing.T) {
	tests := []struct {
		name     string
		tools    []string
		expected []string
	}{
		{
			name:     "no tools",
			tools:    []string{},
			expected: []string{},
		},
		{
			name:     "bash only (no requirements)",
			tools:    []string{"bash"},
			expected: []string{},
		},
		{
			name:     "agent_call requires agent-discovery",
			tools:    []string{"agent_call"},
			expected: []string{"agent-discovery"},
		},
		{
			name:     "bash and agent_call",
			tools:    []string{"bash", "agent_call"},
			expected: []string{"agent-discovery"},
		},
		{
			name:     "unknown tool",
			tools:    []string{"unknown"},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRequiredSkillsForTools(tt.tools)
			
			// Sort for comparison
			sort.Strings(got)
			sort.Strings(tt.expected)
			
			if len(got) != len(tt.expected) {
				t.Errorf("GetRequiredSkillsForTools(%v) = %v, want %v", tt.tools, got, tt.expected)
				return
			}
			
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("GetRequiredSkillsForTools(%v) = %v, want %v", tt.tools, got, tt.expected)
					return
				}
			}
		})
	}
}

func TestGetToolRequirementsForTools(t *testing.T) {
	tests := []struct {
		name          string
		tools         []string
		expectCount   int
		expectToolIn  string
	}{
		{
			name:        "no tools",
			tools:       []string{},
			expectCount: 0,
		},
		{
			name:         "agent_call",
			tools:        []string{"agent_call"},
			expectCount:  1,
			expectToolIn: "agent_call",
		},
		{
			name:        "bash only",
			tools:       []string{"bash"},
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetToolRequirementsForTools(tt.tools)
			
			if len(got) != tt.expectCount {
				t.Errorf("GetToolRequirementsForTools(%v) returned %d items, want %d", tt.tools, len(got), tt.expectCount)
				return
			}
			
			if tt.expectToolIn != "" && len(got) > 0 {
				found := false
				for _, info := range got {
					if info.ToolName == tt.expectToolIn {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetToolRequirementsForTools(%v) missing tool %s", tt.tools, tt.expectToolIn)
				}
			}
		})
	}
}

func TestGetToolSkillRequirements(t *testing.T) {
	reqs := GetToolSkillRequirements()
	
	if len(reqs) == 0 {
		t.Error("GetToolSkillRequirements() returned empty, expected at least one requirement")
	}
	
	// Verify agent_call requirement exists
	found := false
	for _, req := range reqs {
		if req.ToolName == "agent_call" {
			found = true
			if len(req.RequiredSkills) == 0 {
				t.Error("agent_call requirement has no required skills")
			}
			if req.Reason == "" {
				t.Error("agent_call requirement has no reason")
			}
		}
	}
	
	if !found {
		t.Error("GetToolSkillRequirements() missing agent_call requirement")
	}
}
