package plugins

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadPlanners_NoPlugins(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	SetTestDataDir(tmpDir)
	defer SetTestDataDir("")

	// Should return empty results with no errors
	loaded, errors := LoadPlanners()

	if len(loaded) != 0 {
		t.Errorf("expected no loaded planners, got %d", len(loaded))
	}
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

func TestLoadPlanners_PluginWithPlanners(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	SetTestDataDir(tmpDir)
	defer SetTestDataDir("")

	// Create plugin directory
	pluginDir := filepath.Join(tmpDir, "plugins", "test-plugin")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("create plugin dir: %v", err)
	}

	// Create manifest with planners
	manifest := `{
		"name": "test-plugin",
		"version": "1.0.0",
		"description": "Test plugin",
		"planners": [
			{
				"name": "test-planner",
				"type": "near",
				"description": "A test planner"
			}
		]
	}`
	manifestPath := filepath.Join(pluginDir, "manifest.json")
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	// Create registry with the plugin
	reg := &Registry{
		Version: CurrentRegistryVersion,
		Plugins: map[string]*InstalledPlugin{
			"test-plugin": {
				Name:        "test-plugin",
				Version:     "1.0.0",
				Path:        pluginDir,
				InstalledAt: time.Now(),
			},
		},
	}
	if err := reg.Save(); err != nil {
		t.Fatalf("save registry: %v", err)
	}

	// Load planners
	loaded, errors := LoadPlanners()

	// Should have loaded one planner (though it's a no-op since no entry point)
	if len(loaded) != 1 {
		t.Errorf("expected 1 loaded planner, got %d", len(loaded))
	}
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %v", errors)
	}

	if len(loaded) > 0 {
		if loaded[0].Name != "test-planner" {
			t.Errorf("expected planner name 'test-planner', got %s", loaded[0].Name)
		}
		if loaded[0].PluginName != "test-plugin" {
			t.Errorf("expected plugin name 'test-plugin', got %s", loaded[0].PluginName)
		}
	}
}

func TestLoadPlanners_InvalidPlannerType(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	SetTestDataDir(tmpDir)
	defer SetTestDataDir("")

	// Create plugin directory
	pluginDir := filepath.Join(tmpDir, "plugins", "test-plugin")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("create plugin dir: %v", err)
	}

	// Create manifest with invalid planner type
	manifest := `{
		"name": "test-plugin",
		"version": "1.0.0",
		"description": "Test plugin",
		"planners": [
			{
				"name": "bad-planner",
				"type": "invalid"
			}
		]
	}`
	manifestPath := filepath.Join(pluginDir, "manifest.json")
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	// Create registry with the plugin
	reg := &Registry{
		Version: CurrentRegistryVersion,
		Plugins: map[string]*InstalledPlugin{
			"test-plugin": {
				Name:        "test-plugin",
				Version:     "1.0.0",
				Path:        pluginDir,
				InstalledAt: time.Now(),
			},
		},
	}
	if err := reg.Save(); err != nil {
		t.Fatalf("save registry: %v", err)
	}

	// Load planners
	loaded, errors := LoadPlanners()

	// Should have error for invalid type
	if len(loaded) != 0 {
		t.Errorf("expected no loaded planners, got %d", len(loaded))
	}
	if len(errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(errors))
	}
}

func TestLoadPlanners_MissingEntryPoint(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	SetTestDataDir(tmpDir)
	defer SetTestDataDir("")

	// Create plugin directory
	pluginDir := filepath.Join(tmpDir, "plugins", "test-plugin")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("create plugin dir: %v", err)
	}

	// Create manifest with entry point that doesn't exist
	manifest := `{
		"name": "test-plugin",
		"version": "1.0.0",
		"description": "Test plugin",
		"planners": [
			{
				"name": "external-planner",
				"type": "long",
				"entry_point": "planners/missing.so"
			}
		]
	}`
	manifestPath := filepath.Join(pluginDir, "manifest.json")
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	// Create registry with the plugin
	reg := &Registry{
		Version: CurrentRegistryVersion,
		Plugins: map[string]*InstalledPlugin{
			"test-plugin": {
				Name:        "test-plugin",
				Version:     "1.0.0",
				Path:        pluginDir,
				InstalledAt: time.Now(),
			},
		},
	}
	if err := reg.Save(); err != nil {
		t.Fatalf("save registry: %v", err)
	}

	// Load planners
	loaded, errors := LoadPlanners()

	// Should have error for missing entry point
	if len(loaded) != 0 {
		t.Errorf("expected no loaded planners, got %d", len(loaded))
	}
	if len(errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(errors))
	}
}

func TestListPluginPlanners(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	SetTestDataDir(tmpDir)
	defer SetTestDataDir("")

	// Create plugin directories with manifests
	plugin1Dir := filepath.Join(tmpDir, "plugins", "plugin1")
	plugin2Dir := filepath.Join(tmpDir, "plugins", "plugin2")
	if err := os.MkdirAll(plugin1Dir, 0o755); err != nil {
		t.Fatalf("create plugin1 dir: %v", err)
	}
	if err := os.MkdirAll(plugin2Dir, 0o755); err != nil {
		t.Fatalf("create plugin2 dir: %v", err)
	}

	// Write manifests (with required description field)
	manifest1 := `{
		"name": "plugin1",
		"version": "1.0.0",
		"description": "Test plugin 1",
		"planners": [
			{"name": "planner-a", "type": "near"},
			{"name": "planner-b", "type": "long"}
		]
	}`
	manifest2 := `{
		"name": "plugin2",
		"version": "1.0.0",
		"description": "Test plugin 2",
		"planners": [
			{"name": "planner-c", "type": "near"}
		]
	}`
	if err := os.WriteFile(filepath.Join(plugin1Dir, "manifest.json"), []byte(manifest1), 0o644); err != nil {
		t.Fatalf("write manifest1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(plugin2Dir, "manifest.json"), []byte(manifest2), 0o644); err != nil {
		t.Fatalf("write manifest2: %v", err)
	}

	// Create registry
	reg := &Registry{
		Version: CurrentRegistryVersion,
		Plugins: map[string]*InstalledPlugin{
			"plugin1": {Name: "plugin1", Version: "1.0.0", Path: plugin1Dir, InstalledAt: time.Now()},
			"plugin2": {Name: "plugin2", Version: "1.0.0", Path: plugin2Dir, InstalledAt: time.Now()},
		},
	}
	if err := reg.Save(); err != nil {
		t.Fatalf("save registry: %v", err)
	}

	// List planners
	planners, err := ListPluginPlanners()
	if err != nil {
		t.Fatalf("ListPluginPlanners: %v", err)
	}

	if len(planners) != 3 {
		t.Errorf("expected 3 planners, got %d", len(planners))
	}
}

func TestGetPlannerPlugin(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	SetTestDataDir(tmpDir)
	defer SetTestDataDir("")

	// Create plugin directory
	pluginDir := filepath.Join(tmpDir, "plugins", "my-plugin")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("create plugin dir: %v", err)
	}

	// Create manifest (with required description field)
	manifest := `{
		"name": "my-plugin",
		"version": "1.0.0",
		"description": "My test plugin",
		"planners": [
			{"name": "my-planner", "type": "near"}
		]
	}`
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	// Create registry
	reg := &Registry{
		Version: CurrentRegistryVersion,
		Plugins: map[string]*InstalledPlugin{
			"my-plugin": {Name: "my-plugin", Version: "1.0.0", Path: pluginDir, InstalledAt: time.Now()},
		},
	}
	if err := reg.Save(); err != nil {
		t.Fatalf("save registry: %v", err)
	}

	// Test finding a planner
	plugin, err := GetPlannerPlugin("my-planner")
	if err != nil {
		t.Fatalf("GetPlannerPlugin: %v", err)
	}
	if plugin == nil {
		t.Fatal("expected to find plugin for my-planner")
	}
	if plugin.Name != "my-plugin" {
		t.Errorf("expected plugin name 'my-plugin', got %s", plugin.Name)
	}

	// Test not finding a planner
	plugin, err = GetPlannerPlugin("nonexistent")
	if err != nil {
		t.Fatalf("GetPlannerPlugin: %v", err)
	}
	if plugin != nil {
		t.Error("expected nil for nonexistent planner")
	}
}

func TestPlannerLoadError(t *testing.T) {
	err := &PlannerLoadError{
		PluginName:  "test-plugin",
		PlannerName: "test-planner",
		Err:         os.ErrNotExist,
	}

	expected := "failed to load planner test-planner from plugin test-plugin: file does not exist"
	if err.Error() != expected {
		t.Errorf("error message = %q, want %q", err.Error(), expected)
	}

	if err.Unwrap() != os.ErrNotExist {
		t.Errorf("Unwrap() should return original error")
	}
}

func TestLoadExternalPlanner_InvalidExtension(t *testing.T) {
	// Create a temp file that is not a .so file
	tmpDir := t.TempDir()
	badPath := filepath.Join(tmpDir, "planner.txt")
	if err := os.WriteFile(badPath, []byte("not a plugin"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := loadExternalPlanner(badPath)
	if err == nil {
		t.Error("expected error for non-.so file")
	}
}

func TestLoadExternalPlanner_InvalidSoFile(t *testing.T) {
	// Create a fake .so file that isn't a valid plugin
	tmpDir := t.TempDir()
	fakeSo := filepath.Join(tmpDir, "planner.so")
	if err := os.WriteFile(fakeSo, []byte("not a valid so file"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := loadExternalPlanner(fakeSo)
	if err == nil {
		t.Error("expected error for invalid .so file")
	}
}
