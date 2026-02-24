package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSquadDefValidation(t *testing.T) {
	tests := []struct {
		name    string
		squads  []SquadDef
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty squads is valid",
			squads:  []SquadDef{},
			wantErr: false,
		},
		{
			name: "valid squad",
			squads: []SquadDef{
				{
					Name:        "dev-team",
					Description: "Development team",
					Agents:      []string{"@developer", "@reviewer"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			squads: []SquadDef{
				{
					Description: "No name squad",
				},
			},
			wantErr: true,
			errMsg:  "squad name is required",
		},
		{
			name: "duplicate names",
			squads: []SquadDef{
				{Name: "team-a"},
				{Name: "team-a"},
			},
			wantErr: true,
			errMsg:  "duplicate squad name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manifest{
				Name:        "test-plugin",
				Version:     "1.0.0",
				Description: "Test plugin",
				Squads:      tt.squads,
			}

			err := m.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !containsString(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %v", tt.errMsg, err)
				}
			}
		})
	}
}

func TestSquadRegistry(t *testing.T) {
	reg := &SquadRegistry{
		squads: make(map[string]*PluginSquad),
	}

	// Test Register
	squad1 := &PluginSquad{
		Name:        "test-squad",
		Description: "Test squad",
		PluginName:  "test-plugin",
	}

	if err := reg.Register(squad1); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Test Get
	got, ok := reg.Get("test-squad")
	if !ok {
		t.Fatal("Get() returned not found")
	}
	if got.Name != "test-squad" {
		t.Errorf("Get() name = %v, want %v", got.Name, "test-squad")
	}

	// Test duplicate registration
	squad2 := &PluginSquad{
		Name:       "test-squad",
		PluginName: "another-plugin",
	}
	if err := reg.Register(squad2); err == nil {
		t.Error("expected error for duplicate registration")
	}

	// Test Has
	if !reg.Has("test-squad") {
		t.Error("Has() returned false for existing squad")
	}
	if reg.Has("nonexistent") {
		t.Error("Has() returned true for nonexistent squad")
	}

	// Test List
	list := reg.List()
	if len(list) != 1 {
		t.Errorf("List() len = %v, want 1", len(list))
	}

	// Test Clear
	reg.Clear()
	if reg.Has("test-squad") {
		t.Error("Clear() didn't remove squads")
	}
}

func TestLoadPluginSquad(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	squadDir := filepath.Join(tmpDir, "squads", "dev-team")
	if err := os.MkdirAll(squadDir, 0755); err != nil {
		t.Fatalf("failed to create squad dir: %v", err)
	}

	// Create SQUAD.md
	squadMD := `---
lead: "@architect"
agents: ["@developer", "@tester"]
---
# Dev Team

This is the development team.
`
	if err := os.WriteFile(filepath.Join(squadDir, "SQUAD.md"), []byte(squadMD), 0644); err != nil {
		t.Fatalf("failed to write SQUAD.md: %v", err)
	}

	// Create ayo.json
	ayoJSON := `{
  "agents": ["@developer", "@tester"],
  "workspace_mount": "/host/code"
}
`
	if err := os.WriteFile(filepath.Join(squadDir, "ayo.json"), []byte(ayoJSON), 0644); err != nil {
		t.Fatalf("failed to write ayo.json: %v", err)
	}

	plugin := &InstalledPlugin{
		Name: "test-plugin",
		Path: tmpDir,
	}

	def := SquadDef{
		Name:        "dev-team",
		Description: "Development team",
		Agents:      []string{"@developer", "@tester"},
	}

	squad, err := loadPluginSquad(plugin, def)
	if err != nil {
		t.Fatalf("loadPluginSquad() error = %v", err)
	}

	if squad.Name != "dev-team" {
		t.Errorf("Name = %v, want dev-team", squad.Name)
	}
	if squad.PluginName != "test-plugin" {
		t.Errorf("PluginName = %v, want test-plugin", squad.PluginName)
	}
	if !squad.HasConstitution() {
		t.Error("HasConstitution() = false, want true")
	}
	if !squad.HasConfig() {
		t.Error("HasConfig() = false, want true")
	}

	// Test ReadConstitution
	content, err := squad.ReadConstitution()
	if err != nil {
		t.Fatalf("ReadConstitution() error = %v", err)
	}
	if !containsString(string(content), "Dev Team") {
		t.Error("ReadConstitution() content doesn't contain expected text")
	}

	// Test ReadConfig
	config, err := squad.ReadConfig()
	if err != nil {
		t.Fatalf("ReadConfig() error = %v", err)
	}
	if config == nil {
		t.Fatal("ReadConfig() returned nil")
	}
	if _, ok := config["agents"]; !ok {
		t.Error("ReadConfig() doesn't contain agents key")
	}
}

func TestLoadPluginSquadNoConstitution(t *testing.T) {
	// Create temp directory structure without SQUAD.md
	tmpDir := t.TempDir()
	squadDir := filepath.Join(tmpDir, "squads", "no-constitution")
	if err := os.MkdirAll(squadDir, 0755); err != nil {
		t.Fatalf("failed to create squad dir: %v", err)
	}

	plugin := &InstalledPlugin{
		Name: "test-plugin",
		Path: tmpDir,
	}

	def := SquadDef{
		Name: "no-constitution",
	}

	_, err := loadPluginSquad(plugin, def)
	if err == nil {
		t.Error("expected error for missing SQUAD.md")
	}
	if !containsString(err.Error(), "SQUAD.md not found") {
		t.Errorf("expected error about SQUAD.md, got: %v", err)
	}
}

func TestLoadPluginSquadCustomPath(t *testing.T) {
	// Create temp directory structure with custom path
	tmpDir := t.TempDir()
	squadDir := filepath.Join(tmpDir, "custom", "path", "my-team")
	if err := os.MkdirAll(squadDir, 0755); err != nil {
		t.Fatalf("failed to create squad dir: %v", err)
	}

	// Create SQUAD.md
	if err := os.WriteFile(filepath.Join(squadDir, "SQUAD.md"), []byte("# My Team"), 0644); err != nil {
		t.Fatalf("failed to write SQUAD.md: %v", err)
	}

	plugin := &InstalledPlugin{
		Name: "test-plugin",
		Path: tmpDir,
	}

	def := SquadDef{
		Name: "my-team",
		Path: "custom/path/my-team", // Custom path instead of default squads/my-team
	}

	squad, err := loadPluginSquad(plugin, def)
	if err != nil {
		t.Fatalf("loadPluginSquad() error = %v", err)
	}

	if squad.SquadPath != squadDir {
		t.Errorf("SquadPath = %v, want %v", squad.SquadPath, squadDir)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
