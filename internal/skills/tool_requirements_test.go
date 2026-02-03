package skills

import (
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
			name:     "memory tool (no requirements)",
			tools:    []string{"memory"},
			expected: []string{},
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
		name        string
		tools       []string
		expectCount int
	}{
		{
			name:        "no tools",
			tools:       []string{},
			expectCount: 0,
		},
		{
			name:        "bash only",
			tools:       []string{"bash"},
			expectCount: 0,
		},
		{
			name:        "memory tool",
			tools:       []string{"memory"},
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetToolRequirementsForTools(tt.tools)

			if len(got) != tt.expectCount {
				t.Errorf("GetToolRequirementsForTools(%v) returned %d items, want %d", tt.tools, len(got), tt.expectCount)
			}
		})
	}
}

func TestGetToolSkillRequirements(t *testing.T) {
	reqs := GetToolSkillRequirements()

	// Currently no tool-skill requirements exist after agent_call removal
	if len(reqs) != 0 {
		t.Errorf("GetToolSkillRequirements() returned %d items, expected 0", len(reqs))
	}
}
