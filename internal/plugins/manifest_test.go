package plugins

import (
	"encoding/json"
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
		wantErr  bool
	}{
		{
			ref:      "https://github.com/owner/ayo-plugins-test.git",
			wantURL:  "https://github.com/owner/ayo-plugins-test.git",
			wantName: "test",
		},
		{
			ref:      "https://github.com/owner/ayo-plugins-test",
			wantURL:  "https://github.com/owner/ayo-plugins-test",
			wantName: "test",
		},
		{
			ref:      "https://gitlab.com/org/ayo-plugins-mytools.git",
			wantURL:  "https://gitlab.com/org/ayo-plugins-mytools.git",
			wantName: "mytools",
		},
		{
			ref:      "git@github.com:owner/ayo-plugins-test.git",
			wantURL:  "git@github.com:owner/ayo-plugins-test.git",
			wantName: "test",
		},
		{
			ref:      "https://github.com/owner/custom-repo-name",
			wantURL:  "https://github.com/owner/custom-repo-name",
			wantName: "custom-repo-name",
		},
		{
			ref:     "owner/test",
			wantErr: true,
		},
		{
			ref:     "test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			gotURL, gotName, err := ParsePluginURL(tt.ref)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParsePluginURL(%q) expected error, got nil", tt.ref)
				}
				return
			}
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

func TestDependenciesUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantDeps []BinaryDep
		wantErr  bool
	}{
		{
			name: "simple string array",
			json: `{"binaries": ["foo", "bar"]}`,
			wantDeps: []BinaryDep{
				{Name: "foo"},
				{Name: "bar"},
			},
		},
		{
			name: "object with install hint",
			json: `{"binaries": [{"name": "crush", "install_hint": "Install with: brew install crush"}]}`,
			wantDeps: []BinaryDep{
				{Name: "crush", InstallHint: "Install with: brew install crush"},
			},
		},
		{
			name: "object with install command",
			json: `{"binaries": [{"name": "crush", "install_cmd": "go install github.com/charmbracelet/crush@latest"}]}`,
			wantDeps: []BinaryDep{
				{Name: "crush", InstallCmd: "go install github.com/charmbracelet/crush@latest"},
			},
		},
		{
			name: "mixed string and object",
			json: `{"binaries": ["simple", {"name": "complex", "install_hint": "See docs", "install_url": "https://example.com"}]}`,
			wantDeps: []BinaryDep{
				{Name: "simple"},
				{Name: "complex", InstallHint: "See docs", InstallURL: "https://example.com"},
			},
		},
		{
			name: "empty binaries",
			json: `{"binaries": []}`,
			wantDeps: nil,
		},
		{
			name: "no binaries field",
			json: `{"plugins": ["other-plugin"]}`,
			wantDeps: nil,
		},
		{
			name:    "object missing name",
			json:    `{"binaries": [{"install_hint": "no name!"}]}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var deps Dependencies
			err := json.Unmarshal([]byte(tt.json), &deps)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(deps.Binaries) != len(tt.wantDeps) {
				t.Fatalf("got %d deps, want %d", len(deps.Binaries), len(tt.wantDeps))
			}

			for i, got := range deps.Binaries {
				want := tt.wantDeps[i]
				if got.Name != want.Name {
					t.Errorf("deps[%d].Name = %q, want %q", i, got.Name, want.Name)
				}
				if got.InstallHint != want.InstallHint {
					t.Errorf("deps[%d].InstallHint = %q, want %q", i, got.InstallHint, want.InstallHint)
				}
				if got.InstallURL != want.InstallURL {
					t.Errorf("deps[%d].InstallURL = %q, want %q", i, got.InstallURL, want.InstallURL)
				}
				if got.InstallCmd != want.InstallCmd {
					t.Errorf("deps[%d].InstallCmd = %q, want %q", i, got.InstallCmd, want.InstallCmd)
				}
			}
		})
	}
}

func TestDependenciesMarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		deps Dependencies
		want string
	}{
		{
			name: "simple deps as strings",
			deps: Dependencies{
				Binaries: []BinaryDep{{Name: "foo"}, {Name: "bar"}},
			},
			want: `{"binaries":["foo","bar"]}`,
		},
		{
			name: "full dep as object",
			deps: Dependencies{
				Binaries: []BinaryDep{{Name: "crush", InstallHint: "See docs"}},
			},
			want: `{"binaries":[{"name":"crush","install_hint":"See docs"}]}`,
		},
		{
			name: "mixed simple and full",
			deps: Dependencies{
				Binaries: []BinaryDep{
					{Name: "simple"},
					{Name: "complex", InstallCmd: "brew install complex"},
				},
			},
			want: `{"binaries":["simple",{"name":"complex","install_cmd":"brew install complex"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.deps)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestDependenciesGetBinaryNames(t *testing.T) {
	deps := &Dependencies{
		Binaries: []BinaryDep{
			{Name: "foo"},
			{Name: "bar", InstallHint: "some hint"},
			{Name: "baz"},
		},
	}

	names := deps.GetBinaryNames()
	if len(names) != 3 {
		t.Fatalf("got %d names, want 3", len(names))
	}

	want := []string{"foo", "bar", "baz"}
	for i, got := range names {
		if got != want[i] {
			t.Errorf("names[%d] = %q, want %q", i, got, want[i])
		}
	}

	// Test nil dependencies
	var nilDeps *Dependencies
	if names := nilDeps.GetBinaryNames(); names != nil {
		t.Errorf("nil deps should return nil, got %v", names)
	}
}

func TestLoadManifestWithDependencies(t *testing.T) {
	dir := t.TempDir()

	manifest := `{
		"name": "test-plugin",
		"version": "1.0.0",
		"description": "A test plugin with dependencies",
		"dependencies": {
			"binaries": [
				"simple-dep",
				{
					"name": "complex-dep",
					"install_hint": "Run: brew install complex-dep",
					"install_cmd": "brew install complex-dep"
				}
			],
			"plugins": ["other-plugin"]
		}
	}`

	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	m, err := LoadManifest(dir)
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}

	if m.Dependencies == nil {
		t.Fatal("Dependencies is nil")
	}

	if len(m.Dependencies.Binaries) != 2 {
		t.Fatalf("got %d binaries, want 2", len(m.Dependencies.Binaries))
	}

	// Check simple dep
	if m.Dependencies.Binaries[0].Name != "simple-dep" {
		t.Errorf("binaries[0].Name = %q, want %q", m.Dependencies.Binaries[0].Name, "simple-dep")
	}
	if m.Dependencies.Binaries[0].InstallHint != "" {
		t.Errorf("binaries[0].InstallHint should be empty")
	}

	// Check complex dep
	if m.Dependencies.Binaries[1].Name != "complex-dep" {
		t.Errorf("binaries[1].Name = %q, want %q", m.Dependencies.Binaries[1].Name, "complex-dep")
	}
	if m.Dependencies.Binaries[1].InstallHint != "Run: brew install complex-dep" {
		t.Errorf("binaries[1].InstallHint = %q, want %q", m.Dependencies.Binaries[1].InstallHint, "Run: brew install complex-dep")
	}
	if m.Dependencies.Binaries[1].InstallCmd != "brew install complex-dep" {
		t.Errorf("binaries[1].InstallCmd = %q, want %q", m.Dependencies.Binaries[1].InstallCmd, "brew install complex-dep")
	}

	// Check plugins
	if len(m.Dependencies.Plugins) != 1 || m.Dependencies.Plugins[0] != "other-plugin" {
		t.Errorf("Plugins = %v, want [other-plugin]", m.Dependencies.Plugins)
	}
}
