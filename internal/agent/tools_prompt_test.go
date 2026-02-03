package agent

import (
	"strings"
	"testing"
)

func TestBuildToolsPrompt(t *testing.T) {
	t.Run("default includes bash", func(t *testing.T) {
		prompt := BuildToolsPrompt(nil)
		if !strings.Contains(prompt, "<bash>") {
			t.Error("expected prompt to contain <bash> section")
		}
		if !strings.Contains(prompt, "description") {
			t.Error("expected prompt to mention description parameter")
		}
	})

	t.Run("explicit bash", func(t *testing.T) {
		prompt := BuildToolsPrompt([]string{"bash"})
		if !strings.Contains(prompt, "<tools>") {
			t.Error("expected prompt to contain <tools> wrapper")
		}
	})

	t.Run("no bash returns empty", func(t *testing.T) {
		prompt := BuildToolsPrompt([]string{"other_tool"})
		if prompt != "" {
			t.Errorf("expected empty prompt when bash not allowed, got %q", prompt)
		}
	})

	t.Run("memory tool section", func(t *testing.T) {
		prompt := BuildToolsPrompt([]string{"memory"})
		if !strings.Contains(prompt, "<memory>") {
			t.Error("expected prompt to contain <memory> section")
		}
	})
}
