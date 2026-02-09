package guardrails

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrap(t *testing.T) {
	g := Default()
	ctx := WrapContext{
		TrustLevel: "sandboxed",
		SessionID:  "sess-123",
		AgentName:  "test-agent",
	}

	result := g.Wrap("You are a helpful assistant.", ctx)

	// Check markers are present
	if !strings.Contains(result, "[AGENT_PROMPT_START]") {
		t.Error("expected [AGENT_PROMPT_START] marker")
	}
	if !strings.Contains(result, "[AGENT_PROMPT_END]") {
		t.Error("expected [AGENT_PROMPT_END] marker")
	}

	// Check agent prompt is in the middle
	if !strings.Contains(result, "You are a helpful assistant.") {
		t.Error("expected agent prompt in output")
	}

	// Check trust level was templated
	if !strings.Contains(result, "Trust level: sandboxed") {
		t.Error("expected trust level to be templated in suffix")
	}

	// Check session ID was templated
	if !strings.Contains(result, "Session ID: sess-123") {
		t.Error("expected session ID to be templated in suffix")
	}

	// Check order: prefix comes before agent prompt, suffix comes after
	// Use unique substrings to avoid matching prefixes/suffixes that mention markers
	prefixIdx := strings.Index(result, "CRITICAL SECURITY RULES")
	startMarkerIdx := strings.Index(result, "\n\n[AGENT_PROMPT_START]\n\n")
	if startMarkerIdx == -1 {
		startMarkerIdx = strings.Index(result, "[AGENT_PROMPT_START]\n\n") // might be at start
	}
	agentIdx := strings.Index(result, "You are a helpful assistant.")
	endMarkerIdx := strings.Index(result, "\n\n[AGENT_PROMPT_END]\n\n")
	suffixIdx := strings.Index(result, "REMINDER")

	if prefixIdx >= startMarkerIdx {
		t.Errorf("prefix should come before [AGENT_PROMPT_START]: prefixIdx=%d, startMarkerIdx=%d", prefixIdx, startMarkerIdx)
	}
	if startMarkerIdx >= agentIdx {
		t.Errorf("[AGENT_PROMPT_START] should come before agent prompt: startMarkerIdx=%d, agentIdx=%d", startMarkerIdx, agentIdx)
	}
	if agentIdx >= endMarkerIdx {
		t.Errorf("agent prompt should come before [AGENT_PROMPT_END]: agentIdx=%d, endMarkerIdx=%d", agentIdx, endMarkerIdx)
	}
	if endMarkerIdx >= suffixIdx {
		t.Errorf("[AGENT_PROMPT_END] should come before suffix: endMarkerIdx=%d, suffixIdx=%d", endMarkerIdx, suffixIdx)
	}
}

func TestWrapSimple(t *testing.T) {
	g := Default()
	result := g.WrapSimple("Test prompt")

	// Should have markers
	if !strings.Contains(result, "[AGENT_PROMPT_START]") {
		t.Error("expected [AGENT_PROMPT_START] marker")
	}
	if !strings.Contains(result, "[AGENT_PROMPT_END]") {
		t.Error("expected [AGENT_PROMPT_END] marker")
	}

	// Template variables should NOT be rendered (still have {{ }})
	if !strings.Contains(result, "{{ .TrustLevel }}") {
		t.Error("expected unrendered template in simple wrap")
	}
}

func TestWrapWithCustomGuardrails(t *testing.T) {
	g := &Guardrails{
		Prefix: "CUSTOM PREFIX: Trust level is {{ .TrustLevel }}",
		Suffix: "CUSTOM SUFFIX: Agent was {{ .AgentName }}",
	}

	ctx := WrapContext{
		TrustLevel: "privileged",
		AgentName:  "my-agent",
	}

	result := g.Wrap("Agent prompt here", ctx)

	if !strings.Contains(result, "CUSTOM PREFIX: Trust level is privileged") {
		t.Error("expected custom prefix with rendered template")
	}
	if !strings.Contains(result, "CUSTOM SUFFIX: Agent was my-agent") {
		t.Error("expected custom suffix with rendered template")
	}
}

func TestWrapWithEmptyGuardrails(t *testing.T) {
	g := &Guardrails{
		Prefix: "",
		Suffix: "",
	}

	result := g.WrapSimple("Agent prompt")

	// Should still have markers even with empty prefix/suffix
	if !strings.Contains(result, "[AGENT_PROMPT_START]") {
		t.Error("expected [AGENT_PROMPT_START] marker")
	}
	if !strings.Contains(result, "Agent prompt") {
		t.Error("expected agent prompt")
	}

	// Should not have extra empty lines from missing prefix
	if strings.HasPrefix(result, "\n\n") {
		t.Error("should not start with empty lines when prefix is empty")
	}
}

func TestLoad(t *testing.T) {
	t.Run("defaults when no custom files", func(t *testing.T) {
		tmpDir := t.TempDir()
		g := Load(tmpDir)

		if g.Prefix != DefaultPrefix {
			t.Error("expected default prefix when no custom file")
		}
		if g.Suffix != DefaultSuffix {
			t.Error("expected default suffix when no custom file")
		}
	})

	t.Run("custom prefix from file", func(t *testing.T) {
		tmpDir := t.TempDir()
		guardrailsDir := filepath.Join(tmpDir, "guardrails")
		if err := os.MkdirAll(guardrailsDir, 0755); err != nil {
			t.Fatal(err)
		}

		customPrefix := "My custom prefix instructions"
		if err := os.WriteFile(filepath.Join(guardrailsDir, "prefix.txt"), []byte(customPrefix), 0644); err != nil {
			t.Fatal(err)
		}

		g := Load(tmpDir)

		if g.Prefix != customPrefix {
			t.Errorf("expected custom prefix %q, got %q", customPrefix, g.Prefix)
		}
		if g.Suffix != DefaultSuffix {
			t.Error("expected default suffix when no custom suffix file")
		}
	})

	t.Run("custom suffix from file", func(t *testing.T) {
		tmpDir := t.TempDir()
		guardrailsDir := filepath.Join(tmpDir, "guardrails")
		if err := os.MkdirAll(guardrailsDir, 0755); err != nil {
			t.Fatal(err)
		}

		customSuffix := "My custom suffix instructions"
		if err := os.WriteFile(filepath.Join(guardrailsDir, "suffix.txt"), []byte(customSuffix), 0644); err != nil {
			t.Fatal(err)
		}

		g := Load(tmpDir)

		if g.Prefix != DefaultPrefix {
			t.Error("expected default prefix when no custom prefix file")
		}
		if g.Suffix != customSuffix {
			t.Errorf("expected custom suffix %q, got %q", customSuffix, g.Suffix)
		}
	})

	t.Run("both custom files", func(t *testing.T) {
		tmpDir := t.TempDir()
		guardrailsDir := filepath.Join(tmpDir, "guardrails")
		if err := os.MkdirAll(guardrailsDir, 0755); err != nil {
			t.Fatal(err)
		}

		customPrefix := "Custom prefix"
		customSuffix := "Custom suffix"
		if err := os.WriteFile(filepath.Join(guardrailsDir, "prefix.txt"), []byte(customPrefix), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(guardrailsDir, "suffix.txt"), []byte(customSuffix), 0644); err != nil {
			t.Fatal(err)
		}

		g := Load(tmpDir)

		if g.Prefix != customPrefix {
			t.Errorf("expected custom prefix %q, got %q", customPrefix, g.Prefix)
		}
		if g.Suffix != customSuffix {
			t.Errorf("expected custom suffix %q, got %q", customSuffix, g.Suffix)
		}
	})
}

func TestLegacyGuardrailsContent(t *testing.T) {
	// Verify legacy guardrails has expected content
	if !strings.Contains(LegacyGuardrails, "<guardrails>") {
		t.Error("legacy guardrails should have XML-style tags")
	}
	if !strings.Contains(LegacyGuardrails, "No malicious code") {
		t.Error("legacy guardrails should mention malicious code")
	}
	if !strings.Contains(LegacyGuardrails, "No credential exposure") {
		t.Error("legacy guardrails should mention credential exposure")
	}
}

func TestTemplateRenderingErrors(t *testing.T) {
	g := &Guardrails{
		Prefix: "Valid prefix",
		Suffix: "Bad template {{ .Unknown | invalid }}",
	}

	ctx := WrapContext{TrustLevel: "sandboxed"}
	result := g.Wrap("Agent prompt", ctx)

	// Should still work, returning original template on error
	if !strings.Contains(result, "Agent prompt") {
		t.Error("wrap should succeed even with template errors")
	}
}
