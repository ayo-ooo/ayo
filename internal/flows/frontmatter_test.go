package flows

import (
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantErr    error
		wantName   string
		wantDesc   string
		wantScript string
	}{
		{
			name: "valid simple flow",
			content: `#!/usr/bin/env bash
# ayo:flow
# name: test-flow
# description: A test flow

set -euo pipefail
echo "hello"`,
			wantName:   "test-flow",
			wantDesc:   "A test flow",
			wantScript: "set -euo pipefail\necho \"hello\"",
		},
		{
			name: "valid flow with all metadata",
			content: `#!/usr/bin/env bash
# ayo:flow
# name: full-flow
# description: A full featured flow
# version: 1.0.0
# author: test-author
# input: input.jsonschema
# output: output.jsonschema

echo "test"`,
			wantName:   "full-flow",
			wantDesc:   "A full featured flow",
			wantScript: "echo \"test\"",
		},
		{
			name: "no shebang",
			content: `# ayo:flow
# name: test
# description: test`,
			wantErr: ErrNoShebang,
		},
		{
			name: "wrong shebang",
			content: `#!/bin/bash
# ayo:flow
# name: test
# description: test`,
			wantErr: ErrNoShebang,
		},
		{
			name: "no flow marker",
			content: `#!/usr/bin/env bash
# name: test
# description: test
echo "hello"`,
			wantErr: ErrNoFlowMarker,
		},
		{
			name: "marker after non-comment",
			content: `#!/usr/bin/env bash
echo "before marker"
# ayo:flow
# name: test
# description: test`,
			wantErr: ErrNoFlowMarker,
		},
		{
			name: "empty script",
			content: `#!/usr/bin/env bash
# ayo:flow
# name: empty-flow
# description: An empty flow`,
			wantName:   "empty-flow",
			wantDesc:   "An empty flow",
			wantScript: "",
		},
		{
			name: "extra blank lines",
			content: `#!/usr/bin/env bash

# ayo:flow

# name: spaced-flow
# description: Flow with spaces


echo "test"`,
			wantName:   "spaced-flow",
			wantDesc:   "Flow with spaces",
			wantScript: "echo \"test\"",
		},
		{
			name: "metadata with colons in value",
			content: `#!/usr/bin/env bash
# ayo:flow
# name: colon-flow
# description: A flow with: colons in description

echo "test"`,
			wantName:   "colon-flow",
			wantDesc:   "A flow with: colons in description",
			wantScript: "echo \"test\"",
		},
		{
			name: "comment in script preserved",
			content: `#!/usr/bin/env bash
# ayo:flow
# name: commented-flow
# description: Flow with comments

# This is a script comment
echo "test"
# Another comment`,
			wantName:   "commented-flow",
			wantDesc:   "Flow with comments",
			wantScript: "# This is a script comment\necho \"test\"\n# Another comment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := ParseFrontmatter([]byte(tt.content))

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("ParseFrontmatter() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseFrontmatter() unexpected error: %v", err)
			}

			if raw.Frontmatter["name"] != tt.wantName {
				t.Errorf("name = %q, want %q", raw.Frontmatter["name"], tt.wantName)
			}
			if raw.Frontmatter["description"] != tt.wantDesc {
				t.Errorf("description = %q, want %q", raw.Frontmatter["description"], tt.wantDesc)
			}
			if raw.Script != tt.wantScript {
				t.Errorf("script = %q, want %q", raw.Script, tt.wantScript)
			}
		})
	}
}

func TestValidateFrontmatter(t *testing.T) {
	tests := []struct {
		name    string
		fm      map[string]string
		wantErr error
	}{
		{
			name: "valid",
			fm: map[string]string{
				"name":        "test",
				"description": "A test",
			},
			wantErr: nil,
		},
		{
			name: "missing name",
			fm: map[string]string{
				"description": "A test",
			},
			wantErr: ErrMissingName,
		},
		{
			name: "missing description",
			fm: map[string]string{
				"name": "test",
			},
			wantErr: ErrMissingDesc,
		},
		{
			name:    "empty",
			fm:      map[string]string{},
			wantErr: ErrMissingName,
		},
		{
			name: "empty name",
			fm: map[string]string{
				"name":        "",
				"description": "test",
			},
			wantErr: ErrMissingName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFrontmatter(tt.fm)
			if err != tt.wantErr {
				t.Errorf("ValidateFrontmatter() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidMetadataKey(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"name", true},
		{"description", true},
		{"input-schema", true},
		{"output_schema", true},
		{"version1", true},
		{"", false},
		{"Name", false},       // uppercase
		{"my key", false},     // space
		{"key:value", false},  // colon
		{"key.name", false},   // dot
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := isValidMetadataKey(tt.key); got != tt.want {
				t.Errorf("isValidMetadataKey(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}
