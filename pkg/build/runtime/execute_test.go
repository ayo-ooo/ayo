package runtime

import (
	"strings"
	"testing"
)

func TestLoadSkills(t *testing.T) {
	// Create a test embedded filesystem with skills
	type FS struct {
		skills string
	}

	// Test with skills directory
	t.Run("with skills", func(t *testing.T) {
		// Note: This test is limited because we can't easily create
		// an embed.FS for testing. The loadSkills function is
		// tested indirectly through integration tests.
		t.Skip("Cannot easily create embed.FS for unit testing")
	})

	// Test with no skills
	t.Run("no skills directory", func(t *testing.T) {
		// Create empty FS
		// Note: embed.FS cannot be created programmatically
		t.Skip("Cannot easily create embed.FS for unit testing")
	})
}

func TestCombineSystemPromptAndSkills(t *testing.T) {
	tests := []struct {
		name          string
		systemPrompt  string
		skills        map[string]string
		expectedParts []string
	}{
		{
			name:         "no skills",
			systemPrompt: "You are a helpful assistant.",
			skills:       map[string]string{},
			expectedParts: []string{"You are a helpful assistant."},
		},
		{
			name:         "empty system prompt with skills",
			systemPrompt: "",
			skills: map[string]string{
				"coding.md": "You can code in Go.",
			},
			expectedParts: []string{"## Skills", "### coding", "You can code in Go."},
		},
		{
			name:         "system prompt with skills",
			systemPrompt: "You are a helpful assistant.",
			skills: map[string]string{
				"coding.md":   "You can code in Go.",
				"writing.md":  "You can write well.",
				"research.md": "You can research effectively.",
			},
			expectedParts: []string{
				"You are a helpful assistant.",
				"## Skills",
				"### coding",
				"You can code in Go.",
				"### writing",
				"You can write well.",
				"### research",
				"You can research effectively.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := combineSystemPromptAndSkills(tt.systemPrompt, tt.skills)

			// Check that all expected parts are present
			for _, part := range tt.expectedParts {
				if !contains(result, part) {
					t.Errorf("Expected result to contain %q, but it doesn't", part)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestExtractToolDescription tests extracting descriptions from tool scripts.
func TestExtractToolDescription(t *testing.T) {
	tests := []struct {
		name       string
		script     string
		toolName   string
		wantPrefix string
	}{
		{
			name:       "script with comment",
			script:     "# This tool calculates prime numbers\necho calculating...",
			toolName:   "primes",
			wantPrefix: "This tool calculates prime numbers",
		},
		{
			name:       "script without comment",
			script:     "echo calculating...",
			toolName:   "primes",
			wantPrefix: "Execute the primes tool",
		},
		{
			name:       "script with multiple comments",
			script:     "# Calculate primes\n# This is efficient\necho calculating...",
			toolName:   "primes",
			wantPrefix: "Calculate primes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractToolDescription(tt.script, tt.toolName)
			if !strings.HasPrefix(result, tt.wantPrefix) {
				t.Errorf("extractToolDescription() = %v, want prefix %v", result, tt.wantPrefix)
			}
		})
	}
}

// TestCreateBashTool tests the bash tool creation.
func TestCreateBashTool(t *testing.T) {
	tool := createBashTool()

	info := tool.Info()
	if info.Name != "bash" {
		t.Errorf("tool name = %v, want bash", info.Name)
	}
	if info.Description == "" {
		t.Error("tool description should not be empty")
	}
}
