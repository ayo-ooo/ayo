package model

import (
	"os"
	"testing"
)

func TestScanner_DetectsAnthropicKey(t *testing.T) {
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	scanner := NewScanner()
	providers := scanner.Scan()

	if !providers["anthropic"] {
		t.Error("Should detect anthropic provider when ANTHROPIC_API_KEY is set")
	}
}

func TestScanner_DetectsOpenAIKey(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "sk-test")
	defer os.Unsetenv("OPENAI_API_KEY")

	scanner := NewScanner()
	providers := scanner.Scan()

	if !providers["openai"] {
		t.Error("Should detect openai provider when OPENAI_API_KEY is set")
	}
}

func TestScanner_DetectsZAIKey(t *testing.T) {
	os.Setenv("ZAI_API_KEY", "zai-test")
	defer os.Unsetenv("ZAI_API_KEY")

	scanner := NewScanner()
	providers := scanner.Scan()

	if !providers["zai"] {
		t.Error("Should detect zai provider when ZAI_API_KEY is set")
	}
}

func TestScanner_DetectsOpenRouterKey(t *testing.T) {
	os.Setenv("OPENROUTER_API_KEY", "or-test")
	defer os.Unsetenv("OPENROUTER_API_KEY")

	scanner := NewScanner()
	providers := scanner.Scan()

	if !providers["openrouter"] {
		t.Error("Should detect openrouter provider when OPENROUTER_API_KEY is set")
	}
}

func TestScanner_DetectsGeminiKey(t *testing.T) {
	os.Setenv("GEMINI_API_KEY", "gemini-test")
	defer os.Unsetenv("GEMINI_API_KEY")

	scanner := NewScanner()
	providers := scanner.Scan()

	if !providers["gemini"] {
		t.Error("Should detect gemini provider when GEMINI_API_KEY is set")
	}
}

func TestScanner_DetectsGroqKey(t *testing.T) {
	os.Setenv("GROQ_API_KEY", "gsk-test")
	defer os.Unsetenv("GROQ_API_KEY")

	scanner := NewScanner()
	providers := scanner.Scan()

	if !providers["groq"] {
		t.Error("Should detect groq provider when GROQ_API_KEY is set")
	}
}

func TestScanner_DetectsMultipleKeys(t *testing.T) {
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test")
	os.Setenv("OPENAI_API_KEY", "sk-test")
	defer os.Unsetenv("ANTHROPIC_API_KEY")
	defer os.Unsetenv("OPENAI_API_KEY")

	scanner := NewScanner()
	providers := scanner.Scan()

	if len(providers) < 2 {
		t.Errorf("Should detect at least 2 providers, got %d", len(providers))
	}

	if !providers["anthropic"] || !providers["openai"] {
		t.Error("Should detect both anthropic and openai")
	}
}

func TestScanner_NoKeys(t *testing.T) {
	// Clear all known API keys
	keys := []string{
		"ANTHROPIC_API_KEY",
		"OPENAI_API_KEY",
		"ZAI_API_KEY",
		"OPENROUTER_API_KEY",
		"GEMINI_API_KEY",
		"GROQ_API_KEY",
	}
	for _, key := range keys {
		os.Unsetenv(key)
	}

	scanner := NewScanner()
	providers := scanner.Scan()

	if len(providers) != 0 {
		t.Errorf("Should detect 0 providers when no keys set, got %d", len(providers))
	}
}

func TestScanner_EmptyKeyNotDetected(t *testing.T) {
	os.Setenv("ANTHROPIC_API_KEY", "")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	scanner := NewScanner()
	providers := scanner.Scan()

	if providers["anthropic"] {
		t.Error("Should not detect provider when API key is empty")
	}
}

func TestScanner_WhitespaceOnlyKeyNotDetected(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "   ")
	defer os.Unsetenv("OPENAI_API_KEY")

	scanner := NewScanner()
	providers := scanner.Scan()

	if providers["openai"] {
		t.Error("Should not detect provider when API key is whitespace only")
	}
}
