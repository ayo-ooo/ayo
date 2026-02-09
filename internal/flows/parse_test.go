package flows

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseYAMLFlow(t *testing.T) {
	// Create a temporary flow file
	tmpDir := t.TempDir()
	flowPath := filepath.Join(tmpDir, "test.yaml")

	flowContent := `
version: 1
name: test-flow
description: A test flow
steps:
  - id: step1
    type: shell
    run: echo "hello"
  - id: step2
    type: agent
    agent: "@summarizer"
    prompt: "Summarize the output"
    depends_on: [step1]
`

	if err := os.WriteFile(flowPath, []byte(flowContent), 0644); err != nil {
		t.Fatal(err)
	}

	flow, err := ParseYAMLFlow(flowPath)
	if err != nil {
		t.Fatalf("ParseYAMLFlow failed: %v", err)
	}

	if flow.Version != 1 {
		t.Errorf("expected version 1, got %d", flow.Version)
	}
	if flow.Name != "test-flow" {
		t.Errorf("expected name 'test-flow', got %q", flow.Name)
	}
	if len(flow.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(flow.Steps))
	}

	// Check first step
	if flow.Steps[0].ID != "step1" {
		t.Errorf("expected step1 id, got %q", flow.Steps[0].ID)
	}
	if flow.Steps[0].Type != FlowStepTypeShell {
		t.Errorf("expected shell type, got %q", flow.Steps[0].Type)
	}

	// Check second step
	if flow.Steps[1].ID != "step2" {
		t.Errorf("expected step2 id, got %q", flow.Steps[1].ID)
	}
	if flow.Steps[1].Type != FlowStepTypeAgent {
		t.Errorf("expected agent type, got %q", flow.Steps[1].Type)
	}
	if len(flow.Steps[1].DependsOn) != 1 || flow.Steps[1].DependsOn[0] != "step1" {
		t.Errorf("expected depends_on [step1], got %v", flow.Steps[1].DependsOn)
	}
}

func TestParseYAMLFlowBytes(t *testing.T) {
	data := []byte(`
version: 1
name: inline-flow
steps:
  - id: only
    type: shell
    run: echo test
`)

	flow, err := ParseYAMLFlowBytes(data)
	if err != nil {
		t.Fatalf("ParseYAMLFlowBytes failed: %v", err)
	}

	if flow.Name != "inline-flow" {
		t.Errorf("expected name 'inline-flow', got %q", flow.Name)
	}
}

func TestParseYAMLFlow_InvalidYAML(t *testing.T) {
	data := []byte(`
version: 1
name: broken
steps:
  - id: bad
    type: shell
    run: echo
  invalid: [yaml: structure
`)

	_, err := ParseYAMLFlowBytes(data)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestParseYAMLFlow_NotFound(t *testing.T) {
	_, err := ParseYAMLFlow("/nonexistent/path/flow.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestIsYAMLFlowFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"flow.yaml", true},
		{"flow.yml", true},
		{"flow.json", false},
		{"flow.sh", false},
		{"flowYAML", false},
		{".yaml", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsYAMLFlowFile(tt.path); got != tt.expected {
				t.Errorf("IsYAMLFlowFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestDiscoverYAMLFlows(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid flow
	validFlow := `
version: 1
name: valid-flow
steps:
  - id: step1
    type: shell
    run: echo hello
`
	if err := os.WriteFile(filepath.Join(tmpDir, "valid.yaml"), []byte(validFlow), 0644); err != nil {
		t.Fatal(err)
	}

	// Create another valid flow
	anotherFlow := `
version: 1
name: another-flow
steps:
  - id: step1
    type: agent
    agent: "@test"
    prompt: "Do something"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "another.yml"), []byte(anotherFlow), 0644); err != nil {
		t.Fatal(err)
	}

	// Create invalid flow (should be skipped)
	invalidFlow := `
version: 1
name: invalid-flow
steps: []
`
	if err := os.WriteFile(filepath.Join(tmpDir, "invalid.yaml"), []byte(invalidFlow), 0644); err != nil {
		t.Fatal(err)
	}

	// Create non-YAML file (should be skipped)
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("not a flow"), 0644); err != nil {
		t.Fatal(err)
	}

	flows, err := DiscoverYAMLFlows(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverYAMLFlows failed: %v", err)
	}

	if len(flows) != 2 {
		t.Errorf("expected 2 flows, got %d", len(flows))
	}

	names := make(map[string]bool)
	for _, f := range flows {
		names[f.Name] = true
	}
	if !names["valid-flow"] {
		t.Error("missing valid-flow")
	}
	if !names["another-flow"] {
		t.Error("missing another-flow")
	}
}

func TestDiscoverYAMLFlows_NonexistentDir(t *testing.T) {
	flows, err := DiscoverYAMLFlows("/nonexistent/path")
	if err != nil {
		t.Fatalf("expected nil error for nonexistent dir, got: %v", err)
	}
	if flows != nil {
		t.Error("expected nil flows for nonexistent dir")
	}
}
