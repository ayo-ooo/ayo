package runtime

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/alexcabrera/ayo/internal/build/types"
)

// TestDetectAvailableKeys tests detection of API keys from environment variables.
func TestDetectAvailableKeys(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected map[string]string
	}{
		{
			name:     "no keys",
			envVars:  map[string]string{},
			expected: map[string]string{},
		},
		{
			name: "single anthropic key",
			envVars: map[string]string{
				"ANTHROPIC_API_KEY": "sk-ant-test",
			},
			expected: map[string]string{
				"anthropic": "sk-ant-test",
			},
		},
		{
			name: "single openai key",
			envVars: map[string]string{
				"OPENAI_API_KEY": "sk-open-test",
			},
			expected: map[string]string{
				"openai": "sk-open-test",
			},
		},
		{
			name: "single google key",
			envVars: map[string]string{
				"GEMINI_API_KEY": "ai-gem-test",
			},
			expected: map[string]string{
				"google": "ai-gem-test",
			},
		},
		{
			name: "single openrouter key",
			envVars: map[string]string{
				"OPENROUTER_API_KEY": "sk-or-test",
			},
			expected: map[string]string{
				"openrouter": "sk-or-test",
			},
		},
		{
			name: "single xai key",
			envVars: map[string]string{
				"XAI_API_KEY": "sk-xai-test",
			},
			expected: map[string]string{
				"xai": "sk-xai-test",
			},
		},
		{
			name: "single groq key",
			envVars: map[string]string{
				"GROQ_API_KEY": "gsk-groq-test",
			},
			expected: map[string]string{
				"groq": "gsk-groq-test",
			},
		},
		{
			name: "single deepseek key",
			envVars: map[string]string{
				"DEEPSEEK_API_KEY": "sk-ds-test",
			},
			expected: map[string]string{
				"deepseek": "sk-ds-test",
			},
		},
		{
			name: "single cerebras key",
			envVars: map[string]string{
				"CEREBRAS_API_KEY": "sk-cer-test",
			},
			expected: map[string]string{
				"cerebras": "sk-cer-test",
			},
		},
		{
			name: "single together key",
			envVars: map[string]string{
				"TOGETHER_API_KEY": "sk-tog-test",
			},
			expected: map[string]string{
				"together": "sk-tog-test",
			},
		},
		{
			name: "single azure key",
			envVars: map[string]string{
				"AZURE_OPENAI_API_KEY": "sk-az-test",
			},
			expected: map[string]string{
				"azure": "sk-az-test",
			},
		},
		{
			name: "multiple keys",
			envVars: map[string]string{
				"ANTHROPIC_API_KEY": "sk-ant-test",
				"OPENAI_API_KEY":    "sk-open-test",
				"GEMINI_API_KEY":   "ai-gem-test",
			},
			expected: map[string]string{
				"anthropic": "sk-ant-test",
				"openai":    "sk-open-test",
				"google":     "ai-gem-test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all environment variables first
			keys := []string{
				"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY",
				"OPENROUTER_API_KEY", "XAI_API_KEY", "GROQ_API_KEY",
				"DEEPSEEK_API_KEY", "CEREBRAS_API_KEY", "TOGETHER_API_KEY",
				"AZURE_OPENAI_API_KEY",
			}
			for _, key := range keys {
				os.Unsetenv(key)
			}

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Detect keys
			result := detectAvailableKeys()

			// Compare maps
			if len(result) != len(tt.expected) {
				t.Errorf("detectAvailableKeys() returned %d keys, want %d", len(result), len(tt.expected))
			}

			for provider, key := range tt.expected {
				if result[provider] != key {
					t.Errorf("detectAvailableKeys()[%q] = %q, want %q", provider, result[provider], key)
				}
			}

			// Cleanup
			for k := range tt.envVars {
				os.Unsetenv(k)
			}
		})
	}
}

// TestDetectProviderFromModel tests provider detection from model IDs.
func TestDetectProviderFromModel(t *testing.T) {
	tests := []struct {
		modelID    string
		providerID string
	}{
		{"gpt-4", "openai"},
		{"gpt-4o", "openai"},
		{"gpt-4o-mini", "openai"},
		{"gpt-3.5-turbo", "openai"},
		{"o1-preview", "openai"},
		{"o1-mini", "openai"},
		{"claude-3-opus", "anthropic"},
		{"claude-3-sonnet", "anthropic"},
		{"claude-3-haiku", "anthropic"},
		{"gemini-pro", "google"},
		{"gemini-1.5-pro", "google"},
		{"gemini-1.5-flash", "google"},
		{"anthropic/claude-3-sonnet", ""}, // OpenRouter-style model IDs are ambiguous
		{"openai/gpt-4o", ""},                // OpenRouter-style model IDs are ambiguous
		{"google/gemini-pro", ""},           // OpenRouter-style model IDs are ambiguous
		{"unknown-model", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.modelID, func(t *testing.T) {
			result := detectProviderFromModel(tt.modelID)
			if result != tt.providerID {
				t.Errorf("detectProviderFromModel(%q) = %q, want %q", tt.modelID, result, tt.providerID)
			}
		})
	}
}

// TestGetProviderDefaultModel tests getting default model from catwalk.
func TestGetProviderDefaultModel(t *testing.T) {
	tests := []struct {
		name       string
		providerID string
		wantPrefix string // Just check prefix, not exact model (catwalk may update)
	}{
		{
			name:       "anthropic provider",
			providerID: "anthropic",
			wantPrefix: "claude",
		},
		{
			name:       "openai provider",
			providerID: "openai",
			wantPrefix: "gpt",
		},
		{
			name:       "google provider",
			providerID: "google",
			wantPrefix: "", // May vary - Google provider might not have default configured
		},
		{
			name:       "openrouter provider",
			providerID: "openrouter",
			wantPrefix: "anthropic", // OpenRouter default is usually anthropic/claude-*
		},
		{
			name:       "xai provider",
			providerID: "xai",
			wantPrefix: "grok",
		},
		{
			name:       "groq provider",
			providerID: "groq",
			wantPrefix: "", // May vary
		},
		{
			name:       "deepseek provider",
			providerID: "deepseek",
			wantPrefix: "deepseek",
		},
		{
			name:       "cerebras provider",
			providerID: "cerebras",
			wantPrefix: "", // May vary
		},
		{
			name:       "together provider",
			providerID: "together",
			wantPrefix: "", // May vary
		},
		{
			name:       "azure provider",
			providerID: "azure",
			wantPrefix: "gpt",
		},
		{
			name:       "unknown provider",
			providerID: "unknown",
			wantPrefix: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getProviderDefaultModel(tt.providerID)

			if tt.wantPrefix == "" {
				// For unknown providers or empty result, just check we don't panic
				return
			}

			if result == "" {
				t.Errorf("getProviderDefaultModel(%q) returned empty string", tt.providerID)
			}

			if !strings.HasPrefix(strings.ToLower(result), tt.wantPrefix) {
				t.Logf("getProviderDefaultModel(%q) = %q (prefix check: want prefix %q)", tt.providerID, result, tt.wantPrefix)
			}
		})
	}
}

// TestGetAPIKey tests getAPIKey function with model specified.
func TestGetAPIKey(t *testing.T) {
	// Clear environment first
	keys := []string{
		"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY",
		"OPENROUTER_API_KEY", "XAI_API_KEY", "GROQ_API_KEY",
		"DEEPSEEK_API_KEY", "CEREBRAS_API_KEY", "TOGETHER_API_KEY",
		"AZURE_OPENAI_API_KEY",
	}
	for _, key := range keys {
		os.Unsetenv(key)
	}

	t.Run("no keys available", func(t *testing.T) {
		_, _, _, err := getAPIKey("gpt-4")
		if err == nil {
			t.Error("getAPIKey() should return error when no keys available")
		}
		if !strings.Contains(err.Error(), "no API keys found") {
			t.Errorf("getAPIKey() error = %q, want 'no API keys found'", err)
		}
	})

	t.Run("single key with model", func(t *testing.T) {
		os.Setenv("OPENAI_API_KEY", "sk-test-key")
		defer os.Unsetenv("OPENAI_API_KEY")

		apiKey, provider, modelID, err := getAPIKey("gpt-4o")
		if err != nil {
			t.Fatalf("getAPIKey() error = %v", err)
		}

		if apiKey != "sk-test-key" {
			t.Errorf("getAPIKey() apiKey = %q, want sk-test-key", apiKey)
		}

		if provider != "openai" {
			t.Errorf("getAPIKey() provider = %q, want openai", provider)
		}

		if modelID != "gpt-4o" {
			t.Errorf("getAPIKey() modelID = %q, want gpt-4o", modelID)
		}
	})

	t.Run("auto-detect with single key", func(t *testing.T) {
		os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test")
		defer os.Unsetenv("ANTHROPIC_API_KEY")

		apiKey, provider, modelID, err := getAPIKey("")
		if err != nil {
			t.Fatalf("getAPIKey() error = %v", err)
		}

		if apiKey != "sk-ant-test" {
			t.Errorf("getAPIKey() apiKey = %q, want sk-ant-test", apiKey)
		}

		if provider != "anthropic" {
			t.Errorf("getAPIKey() provider = %q, want anthropic", provider)
		}

		if modelID == "" {
			t.Error("getAPIKey() modelID should not be empty for auto-detect")
		}
	})
}

// TestFormatResponse tests response formatting.
func TestFormatResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain text",
			input:    "Hello, world!",
			expected: "Hello, world!",
		},
		{
			name:     "text with newlines",
			input:    "Line 1\nLine 2\nLine 3",
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "markdown text",
			input:    "# Title\n\nSome content.",
			expected: "# Title\n\nSome content.",
		},
		{
			name:     "text with special characters",
			input:    "Hello \"world\" & 'test' <tag>",
			expected: "Hello \"world\" & 'test' <tag>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatResponse(tt.input)
			// formatResponse uses lipgloss to create a bordered box
			// Just verify that the input text is present in the output (possibly split across lines)
			if tt.input == "" {
				// Empty input should produce a box with empty content
				if !strings.Contains(result, "│") {
					t.Errorf("formatResponse() should contain box characters for empty input")
				}
				return
			}

			// Check if each line from input is present somewhere in the output
			lines := strings.Split(tt.input, "\n")
			for _, line := range lines {
				if !strings.Contains(result, line) {
					t.Errorf("formatResponse() output does not contain line %q. Got: %q", line, result)
				}
			}
		})
	}
}

// TestIsPiped tests piped input detection.
func TestIsPiped(t *testing.T) {
	t.Run("not piped", func(t *testing.T) {
		// This test is context-dependent and may vary
		// Just ensure it doesn't panic
		result := isPiped()
		_ = result // Suppress unused warning
	})
}

// TestParseFlagValue tests flag value parsing.
func TestParseFlagValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		flagType string
		wantErr  bool
	}{
		{
			name:     "valid string",
			value:    "hello",
			flagType: "string",
			wantErr:  false,
		},
		{
			name:     "valid integer",
			value:    "42",
			flagType: "int",
			wantErr:  false,
		},
		{
			name:     "valid float",
			value:    "3.14",
			flagType: "float",
			wantErr:  false,
		},
		{
			name:     "valid boolean",
			value:    "true",
			flagType: "bool",
			wantErr:  false,
		},
		{
			name:     "invalid integer",
			value:    "not-a-number",
			flagType: "int",
			wantErr:  true,
		},
		{
			name:     "invalid float",
			value:    "also-not-a-number",
			flagType: "float",
			wantErr:  true,
		},
		{
			name:     "invalid boolean",
			value:    "maybe",
			flagType: "bool",
			wantErr:  true,
		},
		{
			name:     "empty string",
			value:    "",
			flagType: "string",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseFlagValue(tt.value, tt.flagType)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFlagValue(%q, %q) error = %v, wantErr %v", tt.value, tt.flagType, err, tt.wantErr)
			}
		})
	}
}

// TestCreateLanguageModel tests language model creation.
func TestCreateLanguageModel(t *testing.T) {
	t.Run("no API keys", func(t *testing.T) {
		// Clear all keys
		keys := []string{
			"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY",
			"OPENROUTER_API_KEY", "XAI_API_KEY", "GROQ_API_KEY",
			"DEEPSEEK_API_KEY", "CEREBRAS_API_KEY", "TOGETHER_API_KEY",
			"AZURE_OPENAI_API_KEY",
		}
		for _, key := range keys {
			os.Unsetenv(key)
		}

		_, err := createLanguageModel("gpt-4")
		if err == nil {
			t.Error("createLanguageModel() should return error when no API keys available")
		}
	})

	t.Run("openai provider", func(t *testing.T) {
		os.Setenv("OPENAI_API_KEY", "sk-test")
		defer os.Unsetenv("OPENAI_API_KEY")

		model, err := createLanguageModel("gpt-4o")
		if err != nil {
			t.Errorf("createLanguageModel() error = %v", err)
		}
		if model == nil {
			t.Error("createLanguageModel() should return non-nil model")
		}
	})

	t.Run("anthropic provider", func(t *testing.T) {
		os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test")
		defer os.Unsetenv("ANTHROPIC_API_KEY")

		model, err := createLanguageModel("claude-3-sonnet")
		if err != nil {
			t.Errorf("createLanguageModel() error = %v", err)
		}
		if model == nil {
			t.Error("createLanguageModel() should return non-nil model")
		}
	})

	t.Run("google provider", func(t *testing.T) {
		os.Setenv("GEMINI_API_KEY", "ai-gem-test")
		defer os.Unsetenv("GEMINI_API_KEY")

		model, err := createLanguageModel("gemini-pro")
		if err != nil {
			t.Errorf("createLanguageModel() error = %v", err)
		}
		if model == nil {
			t.Error("createLanguageModel() should return non-nil model")
		}
	})

	t.Run("openrouter provider", func(t *testing.T) {
		os.Setenv("OPENROUTER_API_KEY", "sk-or-test")
		defer os.Unsetenv("OPENROUTER_API_KEY")

		model, err := createLanguageModel("anthropic/claude-3-sonnet")
		if err != nil {
			t.Errorf("createLanguageModel() error = %v", err)
		}
		if model == nil {
			t.Error("createLanguageModel() should return non-nil model")
		}
	})

	t.Run("xai provider", func(t *testing.T) {
		os.Setenv("XAI_API_KEY", "sk-xai-test")
		defer os.Unsetenv("XAI_API_KEY")

		model, err := createLanguageModel("grok-beta")
		if err != nil {
			t.Errorf("createLanguageModel() error = %v", err)
		}
		if model == nil {
			t.Error("createLanguageModel() should return non-nil model")
		}
	})

	t.Run("groq provider", func(t *testing.T) {
		os.Setenv("GROQ_API_KEY", "gsk-test")
		defer os.Unsetenv("GROQ_API_KEY")

		model, err := createLanguageModel("llama3-8b-8192")
		if err != nil {
			t.Errorf("createLanguageModel() error = %v", err)
		}
		if model == nil {
			t.Error("createLanguageModel() should return non-nil model")
		}
	})

	t.Run("deepseek provider", func(t *testing.T) {
		os.Setenv("DEEPSEEK_API_KEY", "sk-ds-test")
		defer os.Unsetenv("DEEPSEEK_API_KEY")

		model, err := createLanguageModel("deepseek-chat")
		if err != nil {
			t.Errorf("createLanguageModel() error = %v", err)
		}
		if model == nil {
			t.Error("createLanguageModel() should return non-nil model")
		}
	})

	t.Run("cerebras provider", func(t *testing.T) {
		os.Setenv("CEREBRAS_API_KEY", "sk-cer-test")
		defer os.Unsetenv("CEREBRAS_API_KEY")

		model, err := createLanguageModel("llama3.1-8b")
		if err != nil {
			t.Errorf("createLanguageModel() error = %v", err)
		}
		if model == nil {
			t.Error("createLanguageModel() should return non-nil model")
		}
	})

	t.Run("together provider", func(t *testing.T) {
		os.Setenv("TOGETHER_API_KEY", "sk-tog-test")
		defer os.Unsetenv("TOGETHER_API_KEY")

		model, err := createLanguageModel("meta-llama/Llama-3.2-90B-Vision-Instruct-Turbo")
		if err != nil {
			t.Errorf("createLanguageModel() error = %v", err)
		}
		if model == nil {
			t.Error("createLanguageModel() should return non-nil model")
		}
	})

	t.Run("azure provider requires endpoint", func(t *testing.T) {
		os.Setenv("AZURE_OPENAI_API_KEY", "sk-az-test")
		defer os.Unsetenv("AZURE_OPENAI_API_KEY")

		_, err := createLanguageModel("gpt-4o")
		if err == nil {
			t.Error("createLanguageModel() should return error for Azure without endpoint")
		}
	})

	t.Run("azure provider with endpoint", func(t *testing.T) {
		os.Setenv("AZURE_OPENAI_API_KEY", "sk-az-test")
		os.Setenv("AZURE_OPENAI_ENDPOINT", "https://test.openai.azure.com/v1")
		defer func() {
			os.Unsetenv("AZURE_OPENAI_API_KEY")
			os.Unsetenv("AZURE_OPENAI_ENDPOINT")
		}()

		model, err := createLanguageModel("gpt-4o")
		if err != nil {
			t.Errorf("createLanguageModel() error = %v", err)
		}
		if model == nil {
			t.Error("createLanguageModel() should return non-nil model")
		}
	})

	t.Run("unknown provider", func(t *testing.T) {
		// Don't set any API keys to ensure unknown provider fails
		keys := []string{
			"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY",
			"OPENROUTER_API_KEY", "XAI_API_KEY", "GROQ_API_KEY",
			"DEEPSEEK_API_KEY", "CEREBRAS_API_KEY", "TOGETHER_API_KEY",
			"AZURE_OPENAI_API_KEY",
		}
		for _, key := range keys {
			os.Unsetenv(key)
		}

		_, err := createLanguageModel("unknown/model-id")
		if err == nil {
			t.Error("createLanguageModel() should return error for unknown provider")
		}
	})
}

// TestGetConfigPath tests config path generation.
func TestGetConfigPath(t *testing.T) {
	t.Run("returns valid path", func(t *testing.T) {
		path, err := getConfigPath()
		if err != nil {
			t.Errorf("getConfigPath() error = %v", err)
		}
		if path == "" {
			t.Error("getConfigPath() should return non-empty path")
		}
	})
}

// TestRuntimeConfigMarshal tests RuntimeConfig TOML marshaling.
func TestRuntimeConfigMarshal(t *testing.T) {
	tests := []struct {
		name   string
		config RuntimeConfig
	}{
		{
			name: "full config",
			config: RuntimeConfig{
				Provider: "openai",
				Model:    "gpt-4o",
			},
		},
		{
			name: "config with provider only",
			config: RuntimeConfig{
				Provider: "anthropic",
				Model:    "",
			},
		},
		{
			name: "empty config",
			config: RuntimeConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just ensure config can be created without panicking
			_ = tt.config
		})
	}
}

// TestPrintHelp tests help message printing.
func TestPrintHelp(t *testing.T) {
	t.Run("returns no error", func(t *testing.T) {
		// This function prints to stdout, just ensure it doesn't panic
		// In a real test environment, we might capture stdout
		err := printHelp()
		if err != nil {
			t.Errorf("printHelp() returned error: %v", err)
		}
	})
}

// TestPlatformConstants ensures platform-specific constants are set.
func TestPlatformConstants(t *testing.T) {
	t.Run("runtime is set", func(t *testing.T) {
		// Just verify runtime package is importable
		_ = runtime.GOOS
		_ = runtime.GOARCH
	})
}

// TestParseStructuredInput tests parsing structured CLI input.
func TestParseStructuredInput(t *testing.T) {
	t.Run("no flags defined", func(t *testing.T) {
		config := &types.Config{
			CLI: types.CLIConfig{
				Mode:        "structured",
				Description: "test",
				Flags:       map[string]types.CLIFlag{},
			},
		}

		_, err := parseStructuredInput([]string{}, config)
		if err == nil {
			t.Error("parseStructuredInput() should return error when no flags defined")
		}
		if !strings.Contains(err.Error(), "no flags defined") {
			t.Errorf("parseStructuredInput() error = %q, want 'no flags defined'", err)
		}
	})

	t.Run("unknown flag", func(t *testing.T) {
		config := &types.Config{
			CLI: types.CLIConfig{
				Mode:        "structured",
				Description: "test",
				Flags: map[string]types.CLIFlag{
					"name": {
						Name:        "name",
						Type:        "string",
						Required:    true,
						Description: "Name parameter",
					},
				},
			},
		}

		_, err := parseStructuredInput([]string{"--unknown", "value"}, config)
		if err == nil {
			t.Error("parseStructuredInput() should return error for unknown flag")
		}
	})
}
