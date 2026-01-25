package agent

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/alexcabrera/ayo/internal/config"
)

func TestDefaultAgent(t *testing.T) {
	if DefaultAgent != "@ayo" {
		t.Errorf("DefaultAgent = %q, want @ayo", DefaultAgent)
	}
}

func TestLoadCombinesPrefixAgentSuffix(t *testing.T) {
	home := t.TempDir()
	cfg := config.Config{
		AgentsDir:    filepath.Join(home, "ayo", "agents"),
		SystemPrefix: filepath.Join(home, "ayo", "prompts", "system-prefix.md"),
		SystemSuffix: filepath.Join(home, "ayo", "prompts", "system-suffix.md"),
		DefaultModel: "gpt-4.1",
	}

	mustWrite(t, cfg.SystemPrefix, "PREFIX")
	mustWrite(t, cfg.SystemSuffix, "SUFFIX")

	agentDir := filepath.Join(cfg.AgentsDir, "@alice")
	mustWrite(t, filepath.Join(agentDir, "system.md"), "AGENT")
	writeAgentConfig(t, agentDir, Config{Model: "model-1"})

	ag, err := Load(cfg, "@alice")
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	// Should start with environment block
	if !strings.HasPrefix(ag.CombinedSystem, "<environment>") {
		t.Fatalf("combined should start with <environment>, got:\n%s", ag.CombinedSystem)
	}
	// Should contain the expected content in order: env, guardrails, prefix, agent, suffix
	for _, expected := range []string{"</environment>", "<guardrails>", "PREFIX", "AGENT", "SUFFIX"} {
		if !strings.Contains(ag.CombinedSystem, expected) {
			t.Fatalf("combined should contain %q, got:\n%s", expected, ag.CombinedSystem)
		}
	}
	// Should include datetime in environment block
	if !strings.Contains(ag.CombinedSystem, "datetime:") {
		t.Fatalf("combined should include datetime, got:\n%s", ag.CombinedSystem)
	}
}

func TestLoadHandlesMissingPrefixSuffix(t *testing.T) {
	home := t.TempDir()
	cfg := config.Config{
		AgentsDir:    filepath.Join(home, "ayo", "agents"),
		SystemPrefix: filepath.Join(home, "ayo", "prompts", "prefix_missing.md"),
		SystemSuffix: filepath.Join(home, "ayo", "prompts", "suffix_missing.md"),
		DefaultModel: "gpt-4.1",
	}

	agentDir := filepath.Join(cfg.AgentsDir, "@carol")
	mustWrite(t, filepath.Join(agentDir, "system.md"), "AGENT")
	writeAgentConfig(t, agentDir, Config{})

	ag, err := Load(cfg, "@carol")
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	// Should start with environment block
	if !strings.HasPrefix(ag.CombinedSystem, "<environment>") {
		t.Fatalf("combined should start with <environment>, got:\n%s", ag.CombinedSystem)
	}
	// Should contain AGENT (no prefix/suffix since files don't exist)
	if !strings.Contains(ag.CombinedSystem, "AGENT") {
		t.Fatalf("combined should contain AGENT, got:\n%s", ag.CombinedSystem)
	}
	// Verify datetime is present with valid format in environment block
	pattern := regexp.MustCompile(`datetime: \d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} [A-Z]+`)
	if !pattern.MatchString(ag.CombinedSystem) {
		t.Fatalf("combined should include datetime with format, got:\n%s", ag.CombinedSystem)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func writeAgentConfig(t *testing.T, dir string, cfg Config) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir agent: %v", err)
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func TestIsReservedNamespace(t *testing.T) {
	tests := []struct {
		handle   string
		reserved bool
	}{
		{"@ayo", true},
		{"ayo", true},
		{"@ayo.helper", true},
		{"ayo.helper", true},
		{"@ayo.anything", true},
		{"ayo.", true},
		{"@myagent", false},
		{"myagent", false},
		{"@ayohelper", false}, // no dot, not reserved (unless exactly "ayo")
		{"@my.ayo.agent", false},
		{"@helper", false},
		{"@builtin.helper", false}, // old namespace no longer reserved
	}

	for _, tt := range tests {
		t.Run(tt.handle, func(t *testing.T) {
			got := IsReservedNamespace(tt.handle)
			if got != tt.reserved {
				t.Errorf("IsReservedNamespace(%q) = %v, want %v", tt.handle, got, tt.reserved)
			}
		})
	}
}

func TestSaveRejectsReservedNamespace(t *testing.T) {
	home := t.TempDir()
	cfg := config.Config{
		AgentsDir:    filepath.Join(home, "ayo", "agents"),
		DefaultModel: "gpt-4.1",
	}

	reservedHandles := []string{
		"@ayo.myagent",
		"ayo.test",
		"@ayo.helper",
		"@ayo",
	}

	for _, handle := range reservedHandles {
		t.Run(handle, func(t *testing.T) {
			_, err := Save(cfg, handle, Config{}, "system prompt")
			if err != ErrReservedNamespace {
				t.Errorf("Save(%q) = %v, want ErrReservedNamespace", handle, err)
			}
		})
	}
}

func TestSaveAllowsNonReservedNamespace(t *testing.T) {
	home := t.TempDir()
	cfg := config.Config{
		AgentsDir:    filepath.Join(home, "ayo", "agents"),
		DefaultModel: "gpt-4.1",
	}

	allowedHandles := []string{
		"@myagent",
		"helper",
		"@test.agent",
	}

	for _, handle := range allowedHandles {
		t.Run(handle, func(t *testing.T) {
			ag, err := Save(cfg, handle, Config{}, "system prompt")
			if err != nil {
				t.Fatalf("Save(%q) error: %v", handle, err)
			}
			if ag.Handle == "" {
				t.Errorf("Save(%q) returned empty handle", handle)
			}
		})
	}
}

func TestLoadWithInputSchema(t *testing.T) {
	home := t.TempDir()
	cfg := config.Config{
		AgentsDir:    filepath.Join(home, "ayo", "agents"),
		DefaultModel: "gpt-4.1",
	}

	agentDir := filepath.Join(cfg.AgentsDir, "@schema-agent")
	mustWrite(t, filepath.Join(agentDir, "system.md"), "Agent with input schema")
	writeAgentConfig(t, agentDir, Config{})

	// Write input schema
	schema := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"count": {"type": "integer"}
		},
		"required": ["name"]
	}`
	mustWrite(t, filepath.Join(agentDir, "input.jsonschema"), schema)

	ag, err := Load(cfg, "@schema-agent")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if !ag.HasInputSchema() {
		t.Error("Expected agent to have input schema")
	}

	if ag.InputSchema == nil {
		t.Fatal("InputSchema is nil")
	}

	if ag.InputSchema.Type != "object" {
		t.Errorf("InputSchema.Type = %q, want \"object\"", ag.InputSchema.Type)
	}
}

func TestLoadWithoutInputSchema(t *testing.T) {
	home := t.TempDir()
	cfg := config.Config{
		AgentsDir:    filepath.Join(home, "ayo", "agents"),
		DefaultModel: "gpt-4.1",
	}

	agentDir := filepath.Join(cfg.AgentsDir, "@no-schema-agent")
	mustWrite(t, filepath.Join(agentDir, "system.md"), "Agent without input schema")
	writeAgentConfig(t, agentDir, Config{})

	ag, err := Load(cfg, "@no-schema-agent")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if ag.HasInputSchema() {
		t.Error("Expected agent to NOT have input schema")
	}
}

func TestValidateInput(t *testing.T) {
	home := t.TempDir()
	cfg := config.Config{
		AgentsDir:    filepath.Join(home, "ayo", "agents"),
		DefaultModel: "gpt-4.1",
	}

	agentDir := filepath.Join(cfg.AgentsDir, "@validate-agent")
	mustWrite(t, filepath.Join(agentDir, "system.md"), "Agent with validation")
	writeAgentConfig(t, agentDir, Config{})

	// Write input schema requiring name (string) and optional count (integer)
	schema := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"count": {"type": "integer"}
		},
		"required": ["name"]
	}`
	mustWrite(t, filepath.Join(agentDir, "input.jsonschema"), schema)

	ag, err := Load(cfg, "@validate-agent")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid input with required field",
			input:   `{"name": "test"}`,
			wantErr: false,
		},
		{
			name:    "valid input with all fields",
			input:   `{"name": "test", "count": 5}`,
			wantErr: false,
		},
		{
			name:    "invalid - not JSON",
			input:   "hello world",
			wantErr: true,
		},
		{
			name:    "invalid - missing required field",
			input:   `{"count": 5}`,
			wantErr: true,
		},
		{
			name:    "invalid - wrong type for count",
			input:   `{"name": "test", "count": "five"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ag.ValidateInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInput() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				// Check it's the right error type
				var validationErr *InputValidationError
				if !errors.As(err, &validationErr) {
					t.Errorf("Expected InputValidationError, got %T", err)
				}
			}
		})
	}
}

func TestValidateInputNoSchema(t *testing.T) {
	home := t.TempDir()
	cfg := config.Config{
		AgentsDir:    filepath.Join(home, "ayo", "agents"),
		DefaultModel: "gpt-4.1",
	}

	agentDir := filepath.Join(cfg.AgentsDir, "@no-schema")
	mustWrite(t, filepath.Join(agentDir, "system.md"), "Agent without schema")
	writeAgentConfig(t, agentDir, Config{})

	ag, err := Load(cfg, "@no-schema")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Any input should pass when there's no schema
	inputs := []string{
		"hello world",
		`{"any": "json"}`,
		"123",
		"",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			if err := ag.ValidateInput(input); err != nil {
				t.Errorf("ValidateInput(%q) = %v, want nil", input, err)
			}
		})
	}
}

func TestLoadOutputSchema(t *testing.T) {
	home := t.TempDir()
	cfg := config.Config{
		AgentsDir:    filepath.Join(home, "ayo", "agents"),
		DefaultModel: "gpt-4.1",
	}

	agentDir := filepath.Join(cfg.AgentsDir, "@output-schema-agent")
	mustWrite(t, filepath.Join(agentDir, "system.md"), "Agent with output schema")
	writeAgentConfig(t, agentDir, Config{})

	// Write output schema
	schema := `{
		"type": "object",
		"properties": {
			"result": {"type": "string"},
			"count": {"type": "integer"}
		},
		"required": ["result"]
	}`
	mustWrite(t, filepath.Join(agentDir, "output.jsonschema"), schema)

	ag, err := Load(cfg, "@output-schema-agent")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if !ag.HasOutputSchema() {
		t.Error("HasOutputSchema() = false, want true")
	}

	if ag.OutputSchema == nil {
		t.Fatal("OutputSchema is nil")
	}

	if ag.OutputSchema.Type != "object" {
		t.Errorf("OutputSchema.Type = %q, want %q", ag.OutputSchema.Type, "object")
	}

	if len(ag.OutputSchema.Required) != 1 || ag.OutputSchema.Required[0] != "result" {
		t.Errorf("OutputSchema.Required = %v, want [result]", ag.OutputSchema.Required)
	}
}

func TestValidateOutput(t *testing.T) {
	home := t.TempDir()
	cfg := config.Config{
		AgentsDir:    filepath.Join(home, "ayo", "agents"),
		DefaultModel: "gpt-4.1",
	}

	agentDir := filepath.Join(cfg.AgentsDir, "@validate-output")
	mustWrite(t, filepath.Join(agentDir, "system.md"), "Agent with output validation")
	writeAgentConfig(t, agentDir, Config{})

	// Write output schema
	schema := `{
		"type": "object",
		"properties": {
			"status": {
				"type": "string",
				"enum": ["success", "failure"]
			},
			"message": {"type": "string"}
		},
		"required": ["status", "message"]
	}`
	mustWrite(t, filepath.Join(agentDir, "output.jsonschema"), schema)

	ag, err := Load(cfg, "@validate-output")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	tests := []struct {
		name    string
		output  string
		wantErr bool
	}{
		{
			name:    "valid output",
			output:  `{"status": "success", "message": "done"}`,
			wantErr: false,
		},
		{
			name:    "valid with failure status",
			output:  `{"status": "failure", "message": "error occurred"}`,
			wantErr: false,
		},
		{
			name:    "invalid - not JSON",
			output:  "hello world",
			wantErr: true,
		},
		{
			name:    "invalid - missing required field",
			output:  `{"status": "success"}`,
			wantErr: true,
		},
		{
			name:    "invalid - wrong enum value",
			output:  `{"status": "pending", "message": "waiting"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ag.ValidateOutput(tt.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOutput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateOutputNoSchema(t *testing.T) {
	home := t.TempDir()
	cfg := config.Config{
		AgentsDir:    filepath.Join(home, "ayo", "agents"),
		DefaultModel: "gpt-4.1",
	}

	agentDir := filepath.Join(cfg.AgentsDir, "@no-output-schema")
	mustWrite(t, filepath.Join(agentDir, "system.md"), "Agent without output schema")
	writeAgentConfig(t, agentDir, Config{})

	ag, err := Load(cfg, "@no-output-schema")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if ag.HasOutputSchema() {
		t.Error("HasOutputSchema() = true, want false")
	}

	// Any output should pass when there's no schema
	outputs := []string{
		"hello world",
		`{"any": "json"}`,
		"123",
		"",
	}

	for _, output := range outputs {
		t.Run(output, func(t *testing.T) {
			if err := ag.ValidateOutput(output); err != nil {
				t.Errorf("ValidateOutput(%q) = %v, want nil", output, err)
			}
		})
	}
}

func TestSchemaCompatibility(t *testing.T) {
	// Test schema compatibility tiers

	t.Run("exact match", func(t *testing.T) {
		output := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"name": {Type: "string"},
				"age":  {Type: "integer"},
			},
			Required: []string{"name", "age"},
		}
		input := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"name": {Type: "string"},
				"age":  {Type: "integer"},
			},
			Required: []string{"name", "age"},
		}

		source := Agent{OutputSchema: output}
		target := Agent{InputSchema: input}

		tier := source.CanChainTo(&target)
		if tier != CompatibilityExact {
			t.Errorf("expected CompatibilityExact, got %v", tier)
		}
	})

	t.Run("structural match - output superset", func(t *testing.T) {
		output := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"name":  {Type: "string"},
				"age":   {Type: "integer"},
				"email": {Type: "string"},
			},
			Required: []string{"name", "age", "email"},
		}
		input := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"name": {Type: "string"},
				"age":  {Type: "integer"},
			},
			Required: []string{"name", "age"},
		}

		source := Agent{OutputSchema: output}
		target := Agent{InputSchema: input}

		tier := source.CanChainTo(&target)
		if tier != CompatibilityStructural {
			t.Errorf("expected CompatibilityStructural, got %v", tier)
		}
	})

	t.Run("freeform - no input schema", func(t *testing.T) {
		output := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"name": {Type: "string"},
			},
		}

		source := Agent{OutputSchema: output}
		target := Agent{InputSchema: nil} // Freeform

		tier := source.CanChainTo(&target)
		if tier != CompatibilityFreeform {
			t.Errorf("expected CompatibilityFreeform, got %v", tier)
		}
	})

	t.Run("no output schema", func(t *testing.T) {
		input := &Schema{Type: "object"}

		source := Agent{OutputSchema: nil}
		target := Agent{InputSchema: input}

		tier := source.CanChainTo(&target)
		if tier != CompatibilityNone {
			t.Errorf("expected CompatibilityNone, got %v", tier)
		}
	})

	t.Run("incompatible - missing required field", func(t *testing.T) {
		output := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"name": {Type: "string"},
			},
			Required: []string{"name"},
		}
		input := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"name": {Type: "string"},
				"age":  {Type: "integer"},
			},
			Required: []string{"name", "age"},
		}

		source := Agent{OutputSchema: output}
		target := Agent{InputSchema: input}

		tier := source.CanChainTo(&target)
		if tier != CompatibilityNone {
			t.Errorf("expected CompatibilityNone, got %v", tier)
		}
	})

	t.Run("incompatible - type mismatch", func(t *testing.T) {
		output := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"age": {Type: "string"}, // String, not integer
			},
			Required: []string{"age"},
		}
		input := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"age": {Type: "integer"},
			},
			Required: []string{"age"},
		}

		source := Agent{OutputSchema: output}
		target := Agent{InputSchema: input}

		tier := source.CanChainTo(&target)
		if tier != CompatibilityNone {
			t.Errorf("expected CompatibilityNone, got %v", tier)
		}
	})
}

func TestIsChainable(t *testing.T) {
	t.Run("with input schema", func(t *testing.T) {
		ag := Agent{InputSchema: &Schema{Type: "object"}}
		if !ag.IsChainable() {
			t.Error("expected IsChainable() = true")
		}
	})

	t.Run("with output schema", func(t *testing.T) {
		ag := Agent{OutputSchema: &Schema{Type: "object"}}
		if !ag.IsChainable() {
			t.Error("expected IsChainable() = true")
		}
	})

	t.Run("with both schemas", func(t *testing.T) {
		ag := Agent{
			InputSchema:  &Schema{Type: "object"},
			OutputSchema: &Schema{Type: "object"},
		}
		if !ag.IsChainable() {
			t.Error("expected IsChainable() = true")
		}
	})

	t.Run("without schemas", func(t *testing.T) {
		ag := Agent{}
		if ag.IsChainable() {
			t.Error("expected IsChainable() = false")
		}
	})
}

func TestCompatibilityTierString(t *testing.T) {
	tests := []struct {
		tier CompatibilityTier
		want string
	}{
		{CompatibilityExact, "exact"},
		{CompatibilityStructural, "structural"},
		{CompatibilityFreeform, "freeform"},
		{CompatibilityNone, "none"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.tier.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGuardrailsEnabled(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name     string
		cfg      Config
		handle   string
		expected bool
	}{
		{
			name:     "default nil is enabled",
			cfg:      Config{},
			handle:   "@myagent",
			expected: true,
		},
		{
			name:     "explicit true is enabled",
			cfg:      Config{Guardrails: &trueVal},
			handle:   "@myagent",
			expected: true,
		},
		{
			name:     "explicit false is disabled",
			cfg:      Config{Guardrails: &falseVal},
			handle:   "@myagent",
			expected: false,
		},
		{
			name:     "@ayo namespace cannot disable guardrails",
			cfg:      Config{Guardrails: &falseVal},
			handle:   "@ayo",
			expected: true,
		},
		{
			name:     "@ayo.subagent namespace cannot disable guardrails",
			cfg:      Config{Guardrails: &falseVal},
			handle:   "@ayo.coding",
			expected: true,
		},
		{
			name:     "non-@ prefix still checks namespace",
			cfg:      Config{Guardrails: &falseVal},
			handle:   "ayo.research",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.GuardrailsEnabled(tt.handle)
			if got != tt.expected {
				t.Errorf("GuardrailsEnabled(%q) = %v, want %v", tt.handle, got, tt.expected)
			}
		})
	}
}

func TestLoadIncludesGuardrails(t *testing.T) {
	home := t.TempDir()
	cfg := config.Config{
		AgentsDir:    filepath.Join(home, "ayo", "agents"),
		DefaultModel: "gpt-4.1",
	}

	agentDir := filepath.Join(cfg.AgentsDir, "@testguard")
	mustWrite(t, filepath.Join(agentDir, "system.md"), "AGENT SYSTEM")
	writeAgentConfig(t, agentDir, Config{}) // Default: guardrails enabled

	ag, err := Load(cfg, "@testguard")
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	// Should contain guardrails prompt
	if !strings.Contains(ag.CombinedSystem, "<guardrails>") {
		t.Errorf("combined should contain <guardrails>, got:\n%s", ag.CombinedSystem)
	}
	if !strings.Contains(ag.CombinedSystem, "No malicious code") {
		t.Errorf("combined should contain guardrails content, got:\n%s", ag.CombinedSystem)
	}
	// Agent system should also be present
	if !strings.Contains(ag.CombinedSystem, "AGENT SYSTEM") {
		t.Errorf("combined should contain agent system, got:\n%s", ag.CombinedSystem)
	}
}

func TestLoadWithGuardrailsDisabled(t *testing.T) {
	home := t.TempDir()
	cfg := config.Config{
		AgentsDir:    filepath.Join(home, "ayo", "agents"),
		DefaultModel: "gpt-4.1",
	}

	falseVal := false
	agentDir := filepath.Join(cfg.AgentsDir, "@noguard")
	mustWrite(t, filepath.Join(agentDir, "system.md"), "AGENT SYSTEM")
	writeAgentConfig(t, agentDir, Config{Guardrails: &falseVal})

	ag, err := Load(cfg, "@noguard")
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	// Should NOT contain guardrails prompt
	if strings.Contains(ag.CombinedSystem, "<guardrails>") {
		t.Errorf("combined should NOT contain <guardrails> when disabled, got:\n%s", ag.CombinedSystem)
	}
	// Agent system should still be present
	if !strings.Contains(ag.CombinedSystem, "AGENT SYSTEM") {
		t.Errorf("combined should contain agent system, got:\n%s", ag.CombinedSystem)
	}
}
