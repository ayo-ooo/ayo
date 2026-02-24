package plugins

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSandboxConfigRegistry(t *testing.T) {
	// Create a fresh registry for testing
	reg := &SandboxConfigRegistry{
		configs: make(map[string]*SandboxConfig),
	}

	t.Run("register and get", func(t *testing.T) {
		config := &SandboxConfig{
			Name:        "test-config",
			Description: "Test config",
			PluginName:  "test-plugin",
		}

		if err := reg.Register(config); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, ok := reg.Get("test-config")
		if !ok {
			t.Fatal("expected to find config")
		}
		if got.Name != "test-config" {
			t.Errorf("expected name 'test-config', got %q", got.Name)
		}
	})

	t.Run("duplicate registration fails", func(t *testing.T) {
		config := &SandboxConfig{
			Name:       "test-config",
			PluginName: "another-plugin",
		}

		err := reg.Register(config)
		if err == nil {
			t.Error("expected error for duplicate registration")
		}
	})

	t.Run("list returns all configs", func(t *testing.T) {
		// Add another config
		config2 := &SandboxConfig{
			Name:       "test-config-2",
			PluginName: "test-plugin",
		}
		if err := reg.Register(config2); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		configs := reg.List()
		if len(configs) < 2 {
			t.Errorf("expected at least 2 configs, got %d", len(configs))
		}
	})

	t.Run("has returns correct value", func(t *testing.T) {
		if !reg.Has("test-config") {
			t.Error("expected Has to return true for existing config")
		}
		if reg.Has("nonexistent") {
			t.Error("expected Has to return false for nonexistent config")
		}
	})

	t.Run("clear removes all configs", func(t *testing.T) {
		reg.Clear()
		if len(reg.List()) != 0 {
			t.Error("expected empty list after clear")
		}
	})
}

func TestSandboxConfigLoadError(t *testing.T) {
	err := &SandboxConfigLoadError{
		PluginName: "test-plugin",
		ConfigName: "test-config",
		Err:        os.ErrNotExist,
	}

	expected := "failed to load sandbox config test-config from plugin test-plugin: file does not exist"
	if err.Error() != expected {
		t.Errorf("unexpected error message:\ngot: %s\nwant: %s", err.Error(), expected)
	}

	if err.Unwrap() != os.ErrNotExist {
		t.Error("Unwrap should return wrapped error")
	}
}

func TestLoadSandboxConfig(t *testing.T) {
	t.Run("loads config from sandbox.json", func(t *testing.T) {
		dir := t.TempDir()

		// Create sandbox config directory
		configDir := filepath.Join(dir, "sandboxes", "gpu-test")
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			t.Fatal(err)
		}

		// Create sandbox.json
		config := SandboxConfig{
			Name:        "gpu-test",
			Description: "GPU-enabled sandbox",
			BaseImage:   "nvidia/cuda:12.0",
			Env:         map[string]string{"CUDA_VISIBLE_DEVICES": "0"},
			ProviderRequirements: &ProviderRequirements{
				GPU:       true,
				MinMemory: "8G",
			},
		}
		data, _ := json.MarshalIndent(config, "", "  ")
		if err := os.WriteFile(filepath.Join(configDir, "sandbox.json"), data, 0o644); err != nil {
			t.Fatal(err)
		}

		plugin := &InstalledPlugin{
			Name: "test-plugin",
			Path: dir,
		}
		def := SandboxConfigDef{
			Name:        "gpu-test",
			Description: "GPU test config",
		}

		loaded, err := loadSandboxConfig(plugin, def)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if loaded.Name != "gpu-test" {
			t.Errorf("expected name 'gpu-test', got %q", loaded.Name)
		}
		if loaded.BaseImage != "nvidia/cuda:12.0" {
			t.Errorf("expected base image 'nvidia/cuda:12.0', got %q", loaded.BaseImage)
		}
		if loaded.PluginName != "test-plugin" {
			t.Errorf("expected plugin name 'test-plugin', got %q", loaded.PluginName)
		}
		if loaded.ProviderRequirements == nil || !loaded.ProviderRequirements.GPU {
			t.Error("expected GPU requirement to be set")
		}
	})

	t.Run("loads minimal config without sandbox.json", func(t *testing.T) {
		dir := t.TempDir()

		// Create sandbox config directory without sandbox.json
		configDir := filepath.Join(dir, "sandboxes", "minimal")
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			t.Fatal(err)
		}

		plugin := &InstalledPlugin{
			Name: "test-plugin",
			Path: dir,
		}
		def := SandboxConfigDef{
			Name:        "minimal",
			Description: "Minimal config",
		}

		loaded, err := loadSandboxConfig(plugin, def)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if loaded.Name != "minimal" {
			t.Errorf("expected name 'minimal', got %q", loaded.Name)
		}
		if loaded.Description != "Minimal config" {
			t.Errorf("expected description from def, got %q", loaded.Description)
		}
	})

	t.Run("uses custom path", func(t *testing.T) {
		dir := t.TempDir()

		// Create config in custom path
		configDir := filepath.Join(dir, "custom", "path")
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			t.Fatal(err)
		}

		plugin := &InstalledPlugin{
			Name: "test-plugin",
			Path: dir,
		}
		def := SandboxConfigDef{
			Name: "custom-config",
			Path: "custom/path",
		}

		loaded, err := loadSandboxConfig(plugin, def)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if loaded.Name != "custom-config" {
			t.Errorf("expected name 'custom-config', got %q", loaded.Name)
		}
		if loaded.ConfigPath != configDir {
			t.Errorf("expected config path %q, got %q", configDir, loaded.ConfigPath)
		}
	})

	t.Run("fails for missing directory", func(t *testing.T) {
		dir := t.TempDir()

		plugin := &InstalledPlugin{
			Name: "test-plugin",
			Path: dir,
		}
		def := SandboxConfigDef{
			Name: "missing",
		}

		_, err := loadSandboxConfig(plugin, def)
		if err == nil {
			t.Error("expected error for missing directory")
		}
	})

	t.Run("fails for invalid JSON", func(t *testing.T) {
		dir := t.TempDir()

		configDir := filepath.Join(dir, "sandboxes", "invalid")
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			t.Fatal(err)
		}

		// Write invalid JSON
		if err := os.WriteFile(filepath.Join(configDir, "sandbox.json"), []byte("not json"), 0o644); err != nil {
			t.Fatal(err)
		}

		plugin := &InstalledPlugin{
			Name: "test-plugin",
			Path: dir,
		}
		def := SandboxConfigDef{
			Name: "invalid",
		}

		_, err := loadSandboxConfig(plugin, def)
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("fails for empty name", func(t *testing.T) {
		plugin := &InstalledPlugin{
			Name: "test-plugin",
			Path: t.TempDir(),
		}
		def := SandboxConfigDef{
			Name: "",
		}

		_, err := loadSandboxConfig(plugin, def)
		if err == nil {
			t.Error("expected error for empty name")
		}
	})
}

func TestSandboxConfigDefValidation(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		m := &Manifest{
			Name:        "test",
			Version:     "1.0.0",
			Description: "Test plugin",
			SandboxConfigs: []SandboxConfigDef{
				{Name: "config-1"},
				{Name: "config-2"},
			},
		}

		if err := m.validateSandboxConfigs(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		m := &Manifest{
			Name:        "test",
			Version:     "1.0.0",
			Description: "Test plugin",
			SandboxConfigs: []SandboxConfigDef{
				{Name: ""},
			},
		}

		err := m.validateSandboxConfigs()
		if err == nil {
			t.Error("expected error for missing name")
		}
	})

	t.Run("duplicate names", func(t *testing.T) {
		m := &Manifest{
			Name:        "test",
			Version:     "1.0.0",
			Description: "Test plugin",
			SandboxConfigs: []SandboxConfigDef{
				{Name: "dup"},
				{Name: "dup"},
			},
		}

		err := m.validateSandboxConfigs()
		if err == nil {
			t.Error("expected error for duplicate names")
		}
	})
}
