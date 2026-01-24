package plugins

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRegistry(t *testing.T) {
	// Use a temp directory for registry
	dir := t.TempDir()
	origDataDir := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", dir)
	defer os.Setenv("XDG_DATA_HOME", origDataDir)

	// Create ayo subdirectory
	os.MkdirAll(filepath.Join(dir, "ayo"), 0o755)

	// Load empty registry
	reg, err := LoadRegistry()
	if err != nil {
		t.Fatalf("LoadRegistry failed: %v", err)
	}

	if len(reg.Plugins) != 0 {
		t.Errorf("New registry should be empty, got %d plugins", len(reg.Plugins))
	}

	// Add a plugin
	plugin := &InstalledPlugin{
		Name:        "test",
		Version:     "1.0.0",
		GitURL:      "https://github.com/test/ayo-plugins-test",
		GitCommit:   "abc123",
		InstalledAt: time.Now(),
		Path:        "/path/to/test",
		Agents:      []string{"@test"},
	}

	if err := reg.Add(plugin); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Verify plugin was added
	if !reg.Has("test") {
		t.Error("Registry should have 'test' plugin")
	}

	// Get plugin
	got, err := reg.Get("test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Name != "test" {
		t.Errorf("Name = %q, want %q", got.Name, "test")
	}

	// Try adding duplicate
	if err := reg.Add(plugin); err == nil {
		t.Error("Adding duplicate should fail")
	}

	// Update plugin
	plugin.Version = "2.0.0"
	if err := reg.Update(plugin); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, _ = reg.Get("test")
	if got.Version != "2.0.0" {
		t.Errorf("Version = %q, want %q", got.Version, "2.0.0")
	}

	// List plugins
	list := reg.List()
	if len(list) != 1 {
		t.Errorf("List() returned %d plugins, want 1", len(list))
	}

	// Remove plugin
	if err := reg.Remove("test"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if reg.Has("test") {
		t.Error("Registry should not have 'test' after removal")
	}

	// Try removing non-existent
	if err := reg.Remove("nonexistent"); err == nil {
		t.Error("Removing non-existent should fail")
	}
}

func TestInstalledPluginRenames(t *testing.T) {
	p := &InstalledPlugin{
		Name:   "test",
		Agents: []string{"@crush"},
		Renames: map[string]string{
			"@crush": "@my-crush",
		},
	}

	// Test GetResolvedAgentHandle
	if got := p.GetResolvedAgentHandle("@crush"); got != "@my-crush" {
		t.Errorf("GetResolvedAgentHandle(@crush) = %q, want @my-crush", got)
	}

	// Test non-renamed handle
	if got := p.GetResolvedAgentHandle("@other"); got != "@other" {
		t.Errorf("GetResolvedAgentHandle(@other) = %q, want @other", got)
	}

	// Test GetOriginalAgentHandle
	if got := p.GetOriginalAgentHandle("@my-crush"); got != "@crush" {
		t.Errorf("GetOriginalAgentHandle(@my-crush) = %q, want @crush", got)
	}

	// Test non-renamed handle
	if got := p.GetOriginalAgentHandle("@other"); got != "@other" {
		t.Errorf("GetOriginalAgentHandle(@other) = %q, want @other", got)
	}
}

func TestRegistryListEnabled(t *testing.T) {
	reg := &Registry{
		Version: CurrentRegistryVersion,
		Plugins: map[string]*InstalledPlugin{
			"enabled": {
				Name:     "enabled",
				Disabled: false,
			},
			"disabled": {
				Name:     "disabled",
				Disabled: true,
			},
		},
	}

	enabled := reg.ListEnabled()
	if len(enabled) != 1 {
		t.Errorf("ListEnabled() returned %d plugins, want 1", len(enabled))
	}
	if enabled[0].Name != "enabled" {
		t.Errorf("ListEnabled()[0].Name = %q, want %q", enabled[0].Name, "enabled")
	}
}
