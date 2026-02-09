package capabilities

import (
	"context"
	"strings"
	"testing"
)

func TestInferenceInputHash(t *testing.T) {
	input1 := InferenceInput{
		SystemPrompt: "You are a code reviewer.",
		SkillNames:   []string{"coding"},
	}

	input2 := InferenceInput{
		SystemPrompt: "You are a code reviewer.",
		SkillNames:   []string{"coding"},
	}

	input3 := InferenceInput{
		SystemPrompt: "You are a code reviewer.",
		SkillNames:   []string{"coding", "memory"},
	}

	// Same inputs should produce same hash
	hash1 := input1.Hash()
	hash2 := input2.Hash()
	if hash1 != hash2 {
		t.Error("same inputs should produce same hash")
	}

	// Different inputs should produce different hash
	hash3 := input3.Hash()
	if hash1 == hash3 {
		t.Error("different inputs should produce different hashes")
	}

	// Hash should be 64 characters (SHA256 hex)
	if len(hash1) != 64 {
		t.Errorf("expected 64-character hash, got %d", len(hash1))
	}
}

func TestParseCapabilities(t *testing.T) {
	tests := []struct {
		name     string
		response string
		wantLen  int
		wantErr  bool
	}{
		{
			name: "simple JSON",
			response: `[
				{"name": "code-review", "description": "Reviews code", "confidence": 0.95, "source": "system_prompt"}
			]`,
			wantLen: 1,
		},
		{
			name: "JSON in markdown code block",
			response: "```json\n" + `[
				{"name": "code-review", "description": "Reviews code", "confidence": 0.95, "source": "system_prompt"}
			]` + "\n```",
			wantLen: 1,
		},
		{
			name: "multiple capabilities",
			response: `[
				{"name": "code-review", "description": "Reviews code", "confidence": 0.95},
				{"name": "security", "description": "Security analysis", "confidence": 0.8},
				{"name": "testing", "description": "Test generation", "confidence": 0.6}
			]`,
			wantLen: 3,
		},
		{
			name:     "invalid JSON",
			response: "not valid json",
			wantErr:  true,
		},
		{
			name:     "empty array",
			response: "[]",
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caps, err := parseCapabilities(tt.response)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(caps) != tt.wantLen {
				t.Errorf("expected %d capabilities, got %d", tt.wantLen, len(caps))
			}
		})
	}
}

func TestParseCapabilitiesConfidenceNormalization(t *testing.T) {
	// Test that confidence is clamped to [0, 1]
	response := `[
		{"name": "high", "description": "Too high", "confidence": 1.5},
		{"name": "low", "description": "Too low", "confidence": -0.5},
		{"name": "normal", "description": "Normal", "confidence": 0.7}
	]`

	caps, err := parseCapabilities(response)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(caps) != 3 {
		t.Fatalf("expected 3 capabilities, got %d", len(caps))
	}

	// Find each capability by name
	for _, cap := range caps {
		switch cap.Name {
		case "high":
			if cap.Confidence != 1.0 {
				t.Errorf("expected confidence 1.0 for high, got %f", cap.Confidence)
			}
		case "low":
			if cap.Confidence != 0.0 {
				t.Errorf("expected confidence 0.0 for low, got %f", cap.Confidence)
			}
		case "normal":
			if cap.Confidence != 0.7 {
				t.Errorf("expected confidence 0.7 for normal, got %f", cap.Confidence)
			}
		}
	}
}

func TestBuildPrompt(t *testing.T) {
	input := InferenceInput{
		SystemPrompt:  "You are a code reviewer focusing on security.",
		SkillNames:    []string{"coding", "security"},
		SkillContents: []string{"Coding skill content", "Security skill content"},
		SchemaJSON:    `{"type": "object"}`,
	}

	prompt := buildPrompt(input)

	// Check key components are included
	if !strings.Contains(prompt, "You are a code reviewer focusing on security.") {
		t.Error("prompt should include system prompt")
	}
	if !strings.Contains(prompt, "- coding") {
		t.Error("prompt should include skill names")
	}
	if !strings.Contains(prompt, "- security") {
		t.Error("prompt should include skill names")
	}
	if !strings.Contains(prompt, `{"type": "object"}`) {
		t.Error("prompt should include schema")
	}
	if !strings.Contains(prompt, "JSON array") {
		t.Error("prompt should include output format instructions")
	}
}

func TestInferWithoutLLM(t *testing.T) {
	tests := []struct {
		name           string
		input          InferenceInput
		expectContains []string
	}{
		{
			name: "code review keywords",
			input: InferenceInput{
				SystemPrompt: "You are a code reviewer focusing on security and testing.",
			},
			expectContains: []string{"code-review", "security-analysis", "testing"},
		},
		{
			name: "data analysis keywords",
			input: InferenceInput{
				SystemPrompt: "You analyze data and write SQL queries.",
			},
			expectContains: []string{"data-analysis", "sql-queries"},
		},
		{
			name: "skills included",
			input: InferenceInput{
				SystemPrompt: "You are an assistant.",
				SkillNames:   []string{"memory", "debugging"},
			},
			expectContains: []string{"memory", "debugging"},
		},
		{
			name: "empty input gets default",
			input: InferenceInput{
				SystemPrompt: "",
			},
			expectContains: []string{"general-assistance"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InferWithoutLLM(tt.input)

			if result == nil {
				t.Fatal("expected non-nil result")
			}

			if result.ModelUsed != "heuristic" {
				t.Errorf("expected model 'heuristic', got %q", result.ModelUsed)
			}

			if result.InputHash == "" {
				t.Error("expected non-empty input hash")
			}

			// Check expected capabilities are present
			capNames := make(map[string]bool)
			for _, cap := range result.Capabilities {
				capNames[cap.Name] = true
			}

			for _, expected := range tt.expectContains {
				if !capNames[expected] {
					t.Errorf("expected capability %q not found in %v", expected, capNames)
				}
			}
		})
	}
}

func TestInferrer(t *testing.T) {
	// Mock LLM function
	mockLLM := func(ctx context.Context, prompt string) (string, error) {
		return `[
			{"name": "mocked-capability", "description": "From mock LLM", "confidence": 0.9, "source": "system_prompt"}
		]`, nil
	}

	inferrer := NewInferrer(mockLLM, "mock-model")

	input := InferenceInput{
		SystemPrompt: "Test prompt",
	}

	result, err := inferrer.Infer(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ModelUsed != "mock-model" {
		t.Errorf("expected model 'mock-model', got %q", result.ModelUsed)
	}

	if len(result.Capabilities) != 1 {
		t.Errorf("expected 1 capability, got %d", len(result.Capabilities))
	}

	if result.Capabilities[0].Name != "mocked-capability" {
		t.Errorf("expected 'mocked-capability', got %q", result.Capabilities[0].Name)
	}
}
