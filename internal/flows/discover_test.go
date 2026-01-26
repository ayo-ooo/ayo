package flows

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscover(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	// Create simple flow
	simpleFlow := `#!/usr/bin/env bash
# ayo:flow
# name: simple-flow
# description: A simple flow

echo "hello"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "simple-flow.sh"), []byte(simpleFlow), 0755); err != nil {
		t.Fatal(err)
	}

	// Create packaged flow
	pkgDir := filepath.Join(tmpDir, "packaged-flow")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatal(err)
	}

	pkgFlow := `#!/usr/bin/env bash
# ayo:flow
# name: packaged-flow
# description: A packaged flow

echo "packaged"
`
	if err := os.WriteFile(filepath.Join(pkgDir, "flow.sh"), []byte(pkgFlow), 0755); err != nil {
		t.Fatal(err)
	}

	// Add schemas
	if err := os.WriteFile(filepath.Join(pkgDir, "input.jsonschema"), []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "output.jsonschema"), []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Create non-flow .sh file (no marker)
	nonFlow := `#!/usr/bin/env bash
echo "not a flow"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "not-a-flow.sh"), []byte(nonFlow), 0755); err != nil {
		t.Fatal(err)
	}

	// Discover flows
	flows, err := Discover([]string{tmpDir})
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}

	if len(flows) != 2 {
		t.Errorf("Discover() found %d flows, want 2", len(flows))
	}

	// Check simple flow
	var simpleFound, pkgFound bool
	for _, f := range flows {
		switch f.Name {
		case "simple-flow":
			simpleFound = true
			if f.HasInputSchema() || f.HasOutputSchema() {
				t.Error("simple-flow should not have schemas")
			}
		case "packaged-flow":
			pkgFound = true
			if !f.HasInputSchema() {
				t.Error("packaged-flow should have input schema")
			}
			if !f.HasOutputSchema() {
				t.Error("packaged-flow should have output schema")
			}
		}
	}

	if !simpleFound {
		t.Error("simple-flow not found")
	}
	if !pkgFound {
		t.Error("packaged-flow not found")
	}
}

func TestDiscover_Deduplication(t *testing.T) {
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	flow1 := `#!/usr/bin/env bash
# ayo:flow
# name: duplicate
# description: First version

echo "first"
`
	flow2 := `#!/usr/bin/env bash
# ayo:flow
# name: duplicate
# description: Second version

echo "second"
`

	if err := os.WriteFile(filepath.Join(tmpDir1, "duplicate.sh"), []byte(flow1), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir2, "duplicate.sh"), []byte(flow2), 0755); err != nil {
		t.Fatal(err)
	}

	// First directory takes precedence
	flows, err := Discover([]string{tmpDir1, tmpDir2})
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}

	if len(flows) != 1 {
		t.Errorf("Discover() found %d flows, want 1", len(flows))
	}

	if flows[0].Description != "First version" {
		t.Errorf("Expected first version, got %q", flows[0].Description)
	}
}

func TestDiscover_MissingDirectory(t *testing.T) {
	flows, err := Discover([]string{"/nonexistent/path"})
	if err != nil {
		t.Fatalf("Discover() should not error on missing dir: %v", err)
	}
	if len(flows) != 0 {
		t.Errorf("Discover() found %d flows, want 0", len(flows))
	}
}

func TestDiscover_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	flows, err := Discover([]string{tmpDir})
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}
	if len(flows) != 0 {
		t.Errorf("Discover() found %d flows, want 0", len(flows))
	}
}

func TestDiscoverOne_SimpleFlow(t *testing.T) {
	tmpDir := t.TempDir()

	flowContent := `#!/usr/bin/env bash
# ayo:flow
# name: test-flow
# description: A test flow
# version: 1.0.0
# author: test

echo "test"
`
	path := filepath.Join(tmpDir, "test-flow.sh")
	if err := os.WriteFile(path, []byte(flowContent), 0755); err != nil {
		t.Fatal(err)
	}

	flow, err := DiscoverOne(path)
	if err != nil {
		t.Fatalf("DiscoverOne() error: %v", err)
	}

	if flow.Name != "test-flow" {
		t.Errorf("Name = %q, want %q", flow.Name, "test-flow")
	}
	if flow.Metadata.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", flow.Metadata.Version, "1.0.0")
	}
	if flow.Metadata.Author != "test" {
		t.Errorf("Author = %q, want %q", flow.Metadata.Author, "test")
	}
}

func TestDiscoverOne_PackagedFlow(t *testing.T) {
	tmpDir := t.TempDir()
	pkgDir := filepath.Join(tmpDir, "my-flow")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatal(err)
	}

	flowContent := `#!/usr/bin/env bash
# ayo:flow
# name: my-flow
# description: My flow

echo "test"
`
	if err := os.WriteFile(filepath.Join(pkgDir, "flow.sh"), []byte(flowContent), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "input.jsonschema"), []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	flow, err := DiscoverOne(pkgDir)
	if err != nil {
		t.Fatalf("DiscoverOne() error: %v", err)
	}

	if flow.Name != "my-flow" {
		t.Errorf("Name = %q, want %q", flow.Name, "my-flow")
	}
	if !flow.HasInputSchema() {
		t.Error("Expected input schema")
	}
	if flow.HasOutputSchema() {
		t.Error("Did not expect output schema")
	}
}

func TestDiscoverOne_NotFound(t *testing.T) {
	_, err := DiscoverOne("/nonexistent/path")
	if err == nil {
		t.Error("DiscoverOne() should error on missing path")
	}
}

func TestDiscoverOne_InvalidFlow(t *testing.T) {
	tmpDir := t.TempDir()

	// No marker
	content := `#!/usr/bin/env bash
echo "not a flow"
`
	path := filepath.Join(tmpDir, "invalid.sh")
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}

	_, err := DiscoverOne(path)
	if err == nil {
		t.Error("DiscoverOne() should error on invalid flow")
	}
}

func TestSourceFromPath(t *testing.T) {
	tests := []struct {
		path string
		want FlowSource
	}{
		{"/home/user/.local/share/ayo/flows", FlowSourceBuiltin},
		{"/home/user/.config/ayo/flows", FlowSourceUser},
		{"/project/.ayo/flows", FlowSourceProject},
		{"/some/random/path", FlowSourceUser},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := sourceFromPath(tt.path)
			if got != tt.want {
				t.Errorf("sourceFromPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
