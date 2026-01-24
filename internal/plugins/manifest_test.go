package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifest(t *testing.T) {
	// Create temp directory
	dir := t.TempDir()

	// Create valid manifest
	manifest := `{
		"name": "test-plugin",
		"version": "1.0.0",
		"description": "A test plugin",
		"agents": ["@test-agent"],
		"skills": ["test-skill"],
		"tools": ["test-tool"]
	}`

	// Create required directories
	os.MkdirAll(filepath.Join(dir, "agents", "@test-agent"), 0o755)
	os.MkdirAll(filepath.Join(dir, "skills", "test-skill"), 0o755)
	os.MkdirAll(filepath.Join(dir, "tools", "test-tool"), 0o755)

	// Write manifest
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	// Load manifest
	m, err := LoadManifest(dir)
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}

	// Verify fields
	if m.Name != "test-plugin" {
		t.Errorf("Name = %q, want %q", m.Name, "test-plugin")
	}
	if m.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", m.Version, "1.0.0")
	}
	if m.Description != "A test plugin" {
		t.Errorf("Description = %q, want %q", m.Description, "A test plugin")
	}
	if len(m.Agents) != 1 || m.Agents[0] != "@test-agent" {
		t.Errorf("Agents = %v, want [@test-agent]", m.Agents)
	}
}

func TestLoadManifestMissingFile(t *testing.T) {
	dir := t.TempDir()

	_, err := LoadManifest(dir)
	if err != ErrManifestNotFound {
		t.Errorf("Expected ErrManifestNotFound, got %v", err)
	}
}

func TestLoadManifestInvalidJSON(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadManifest(dir)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestManifestValidation(t *testing.T) {
	tests := []struct {
		name    string
		m       Manifest
		wantErr error
	}{
		{
			name:    "missing name",
			m:       Manifest{Version: "1.0.0", Description: "test"},
			wantErr: ErrMissingName,
		},
		{
			name:    "missing version",
			m:       Manifest{Name: "test", Description: "test"},
			wantErr: ErrMissingVersion,
		},
		{
			name:    "missing description",
			m:       Manifest{Name: "test", Version: "1.0.0"},
			wantErr: ErrMissingDescription,
		},
		{
			name:    "invalid name format",
			m:       Manifest{Name: "Test Plugin!", Version: "1.0.0", Description: "test"},
			wantErr: ErrInvalidName,
		},
		{
			name:    "invalid version format",
			m:       Manifest{Name: "test", Version: "v1.0", Description: "test"},
			wantErr: ErrInvalidVersion,
		},
		{
			name: "valid manifest",
			m:    Manifest{Name: "test", Version: "1.0.0", Description: "test"},
		},
		{
			name: "valid manifest with hyphen",
			m:    Manifest{Name: "test-plugin", Version: "1.0.0-beta", Description: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.m.Validate()
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Expected error %v, got nil", tt.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestParsePluginURL(t *testing.T) {
	tests := []struct {
		ref      string
		wantURL  string
		wantName string
	}{
		{
			ref:      "https://github.com/owner/ayo-plugins-test.git",
			wantURL:  "https://github.com/owner/ayo-plugins-test.git",
			wantName: "test",
		},
		{
			ref:      "owner/test",
			wantURL:  "https://github.com/owner/ayo-plugins-test.git",
			wantName: "test",
		},
		{
			ref:      "owner/ayo-plugins-test",
			wantURL:  "https://github.com/owner/ayo-plugins-test.git",
			wantName: "test",
		},
		{
			ref:      "test",
			wantURL:  "https://github.com/test/ayo-plugins-test.git",
			wantName: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			gotURL, gotName, err := ParsePluginURL(tt.ref)
			if err != nil {
				t.Fatalf("ParsePluginURL(%q) error: %v", tt.ref, err)
			}
			if gotURL != tt.wantURL {
				t.Errorf("URL = %q, want %q", gotURL, tt.wantURL)
			}
			if gotName != tt.wantName {
				t.Errorf("Name = %q, want %q", gotName, tt.wantName)
			}
		})
	}
}

func TestExtractNameFromRepo(t *testing.T) {
	tests := []struct {
		repo string
		want string
	}{
		{"ayo-plugins-test", "test"},
		{"ayo-plugins-test.git", "test"},
		{"my-repo", "my-repo"},
	}

	for _, tt := range tests {
		t.Run(tt.repo, func(t *testing.T) {
			got := ExtractNameFromRepo(tt.repo)
			if got != tt.want {
				t.Errorf("ExtractNameFromRepo(%q) = %q, want %q", tt.repo, got, tt.want)
			}
		})
	}
}
