package flows

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRun_SimpleFlow(t *testing.T) {
	tmpDir := t.TempDir()

	flowContent := `#!/usr/bin/env bash
# ayo:flow
# name: echo-flow
# description: Echo input back

INPUT="${1:-$(cat)}"
echo "$INPUT"
`
	flowPath := filepath.Join(tmpDir, "echo-flow.sh")
	if err := os.WriteFile(flowPath, []byte(flowContent), 0755); err != nil {
		t.Fatal(err)
	}

	flow, err := DiscoverOne(flowPath)
	if err != nil {
		t.Fatalf("DiscoverOne: %v", err)
	}

	result, err := Run(context.Background(), flow, RunOptions{
		Input: `{"test": "value"}`,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Status != RunStatusSuccess {
		t.Errorf("Status = %v, want %v", result.Status, RunStatusSuccess)
	}

	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}

	if !strings.Contains(result.Stdout, `{"test": "value"}`) {
		t.Errorf("Stdout = %q, want to contain input", result.Stdout)
	}

	if result.Duration <= 0 {
		t.Error("Duration should be positive")
	}
}

func TestRun_FailingFlow(t *testing.T) {
	tmpDir := t.TempDir()

	flowContent := `#!/usr/bin/env bash
# ayo:flow
# name: failing-flow
# description: Always fails

exit 42
`
	flowPath := filepath.Join(tmpDir, "failing-flow.sh")
	if err := os.WriteFile(flowPath, []byte(flowContent), 0755); err != nil {
		t.Fatal(err)
	}

	flow, err := DiscoverOne(flowPath)
	if err != nil {
		t.Fatalf("DiscoverOne: %v", err)
	}

	result, err := Run(context.Background(), flow, RunOptions{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Status != RunStatusFailed {
		t.Errorf("Status = %v, want %v", result.Status, RunStatusFailed)
	}

	if result.ExitCode != 42 {
		t.Errorf("ExitCode = %d, want 42", result.ExitCode)
	}
}

func TestRun_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}

	tmpDir := t.TempDir()

	flowContent := `#!/usr/bin/env bash
# ayo:flow
# name: slow-flow
# description: Takes too long

sleep 10
echo "done"
`
	flowPath := filepath.Join(tmpDir, "slow-flow.sh")
	if err := os.WriteFile(flowPath, []byte(flowContent), 0755); err != nil {
		t.Fatal(err)
	}

	flow, err := DiscoverOne(flowPath)
	if err != nil {
		t.Fatalf("DiscoverOne: %v", err)
	}

	result, err := Run(context.Background(), flow, RunOptions{
		Timeout: 500 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Status != RunStatusTimeout {
		t.Errorf("Status = %v, want %v", result.Status, RunStatusTimeout)
	}
}

func TestRun_EnvironmentVariables(t *testing.T) {
	tmpDir := t.TempDir()

	flowContent := `#!/usr/bin/env bash
# ayo:flow
# name: env-flow
# description: Check environment

echo "NAME=$AYO_FLOW_NAME"
echo "DIR=$AYO_FLOW_DIR"
echo "RUN_ID=$AYO_FLOW_RUN_ID"
`
	flowPath := filepath.Join(tmpDir, "env-flow.sh")
	if err := os.WriteFile(flowPath, []byte(flowContent), 0755); err != nil {
		t.Fatal(err)
	}

	flow, err := DiscoverOne(flowPath)
	if err != nil {
		t.Fatalf("DiscoverOne: %v", err)
	}

	result, err := Run(context.Background(), flow, RunOptions{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Status != RunStatusSuccess {
		t.Errorf("Status = %v, want %v", result.Status, RunStatusSuccess)
	}

	if !strings.Contains(result.Stdout, "NAME=env-flow") {
		t.Errorf("Stdout should contain flow name, got: %s", result.Stdout)
	}

	if !strings.Contains(result.Stdout, "DIR="+tmpDir) {
		t.Errorf("Stdout should contain flow dir, got: %s", result.Stdout)
	}

	if !strings.Contains(result.Stdout, "RUN_ID=") {
		t.Errorf("Stdout should contain run ID, got: %s", result.Stdout)
	}
}

func TestRun_ValidateOnly(t *testing.T) {
	tmpDir := t.TempDir()

	flowContent := `#!/usr/bin/env bash
# ayo:flow
# name: validate-flow
# description: Should not run

echo "This should not appear"
`
	flowPath := filepath.Join(tmpDir, "validate-flow.sh")
	if err := os.WriteFile(flowPath, []byte(flowContent), 0755); err != nil {
		t.Fatal(err)
	}

	flow, err := DiscoverOne(flowPath)
	if err != nil {
		t.Fatalf("DiscoverOne: %v", err)
	}

	result, err := Run(context.Background(), flow, RunOptions{
		Input:    `{"test": true}`,
		Validate: true,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Status != RunStatusSuccess {
		t.Errorf("Status = %v, want %v", result.Status, RunStatusSuccess)
	}

	if result.Stdout != "" {
		t.Errorf("Stdout should be empty in validate mode, got: %s", result.Stdout)
	}

	if result.InputUsed != `{"test": true}` {
		t.Errorf("InputUsed = %q, want %q", result.InputUsed, `{"test": true}`)
	}
}

func TestRun_InputFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create input file
	inputPath := filepath.Join(tmpDir, "input.json")
	if err := os.WriteFile(inputPath, []byte(`{"from": "file"}`), 0644); err != nil {
		t.Fatal(err)
	}

	flowContent := `#!/usr/bin/env bash
# ayo:flow
# name: file-input-flow
# description: Read from file

INPUT="${1:-$(cat)}"
echo "$INPUT"
`
	flowPath := filepath.Join(tmpDir, "file-input-flow.sh")
	if err := os.WriteFile(flowPath, []byte(flowContent), 0755); err != nil {
		t.Fatal(err)
	}

	flow, err := DiscoverOne(flowPath)
	if err != nil {
		t.Fatalf("DiscoverOne: %v", err)
	}

	result, err := Run(context.Background(), flow, RunOptions{
		InputFile: inputPath,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Status != RunStatusSuccess {
		t.Errorf("Status = %v, want %v", result.Status, RunStatusSuccess)
	}

	if !strings.Contains(result.Stdout, `{"from": "file"}`) {
		t.Errorf("Stdout = %q, want to contain file input", result.Stdout)
	}
}

func TestRun_DefaultInput(t *testing.T) {
	tmpDir := t.TempDir()

	flowContent := `#!/usr/bin/env bash
# ayo:flow
# name: default-input-flow
# description: Test default input

INPUT="${1:-$(cat)}"
echo "GOT: $INPUT"
`
	flowPath := filepath.Join(tmpDir, "default-input-flow.sh")
	if err := os.WriteFile(flowPath, []byte(flowContent), 0755); err != nil {
		t.Fatal(err)
	}

	flow, err := DiscoverOne(flowPath)
	if err != nil {
		t.Fatalf("DiscoverOne: %v", err)
	}

	result, err := Run(context.Background(), flow, RunOptions{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Status != RunStatusSuccess {
		t.Errorf("Status = %v, want %v", result.Status, RunStatusSuccess)
	}

	if !strings.Contains(result.Stdout, "GOT: {}") {
		t.Errorf("Stdout = %q, want to contain default empty object", result.Stdout)
	}
}

func TestResolveInput(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file
	inputPath := filepath.Join(tmpDir, "input.json")
	if err := os.WriteFile(inputPath, []byte(`{"file": true}`), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		opts    RunOptions
		want    string
		wantErr bool
	}{
		{
			name:    "explicit input",
			opts:    RunOptions{Input: `{"explicit": true}`},
			want:    `{"explicit": true}`,
			wantErr: false,
		},
		{
			name:    "input file",
			opts:    RunOptions{InputFile: inputPath},
			want:    `{"file": true}`,
			wantErr: false,
		},
		{
			name:    "default empty",
			opts:    RunOptions{},
			want:    "{}",
			wantErr: false,
		},
		{
			name:    "explicit takes precedence over file",
			opts:    RunOptions{Input: `{"explicit": true}`, InputFile: inputPath},
			want:    `{"explicit": true}`,
			wantErr: false,
		},
		{
			name:    "missing file",
			opts:    RunOptions{InputFile: "/nonexistent"},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveInput(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("resolveInput() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateRunID(t *testing.T) {
	id1 := generateRunID()
	id2 := generateRunID()

	if id1 == "" {
		t.Error("Run ID should not be empty")
	}

	if id1 == id2 {
		t.Error("Run IDs should be unique")
	}

	// ULID is 26 characters
	if len(id1) != 26 {
		t.Errorf("Run ID length = %d, want 26", len(id1))
	}
}
