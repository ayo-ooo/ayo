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

	t.Run("no bash or agent_call returns empty", func(t *testing.T) {
		prompt := BuildToolsPrompt([]string{"other_tool"})
		if prompt != "" {
			t.Errorf("expected empty prompt when bash not allowed, got %q", prompt)
		}
	})

	t.Run("default includes agent_call", func(t *testing.T) {
		prompt := BuildToolsPrompt(nil)
		if !strings.Contains(prompt, "<agent_call>") {
			t.Error("expected prompt to contain <agent_call> section")
		}
	})

	t.Run("agent_call lists builtin agents dynamically", func(t *testing.T) {
		prompt := BuildToolsPrompt([]string{"agent_call"})
		if !strings.Contains(prompt, "@ayo") {
			t.Error("expected prompt to list @ayo agent")
		}
	})

	t.Run("agent_call describes tool usage", func(t *testing.T) {
		prompt := BuildToolsPrompt([]string{"agent_call"})
		if !strings.Contains(prompt, "specialized agents") {
			t.Error("expected prompt to describe agent_call purpose")
		}
	})
}
