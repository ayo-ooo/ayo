package squads

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateSquadDir(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	squadDir := filepath.Join(tmpDir, "test-squad")
	if err := os.MkdirAll(squadDir, 0755); err != nil {
		t.Fatalf("create squad dir: %v", err)
	}

	// Create SQUAD.md with frontmatter
	squadMD := `---
lead: "@architect"
agents: ["@frontend", "@backend"]
planners:
  near_term: ayo-todos
  long_term: ayo-tickets
input_accepts: "@planner"
---
# Mission

Build the auth system...
`
	squadMDPath := filepath.Join(squadDir, "SQUAD.md")
	if err := os.WriteFile(squadMDPath, []byte(squadMD), 0644); err != nil {
		t.Fatalf("write SQUAD.md: %v", err)
	}

	// Run migration
	result := migrateSquadDir("test-squad", squadDir, MigrateOptions{})

	// Check result
	if result.Error != nil {
		t.Fatalf("migration failed: %v", result.Error)
	}
	if !result.Migrated {
		t.Fatal("expected migration to occur")
	}
	if result.Skipped {
		t.Fatal("unexpected skip")
	}

	// Verify ayo.json was created
	ayoJSONPath := filepath.Join(squadDir, "ayo.json")
	data, err := os.ReadFile(ayoJSONPath)
	if err != nil {
		t.Fatalf("read ayo.json: %v", err)
	}

	var cfg AyoConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parse ayo.json: %v", err)
	}

	if cfg.Squad.Lead != "@architect" {
		t.Errorf("expected lead @architect, got %s", cfg.Squad.Lead)
	}
	if len(cfg.Squad.Agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(cfg.Squad.Agents))
	}
	if cfg.Squad.InputAccepts != "@planner" {
		t.Errorf("expected input_accepts @planner, got %s", cfg.Squad.InputAccepts)
	}
	if cfg.Squad.Planners == nil {
		t.Fatal("expected planners to be set")
	}
	if cfg.Squad.Planners.NearTerm != "ayo-todos" {
		t.Errorf("expected near_term ayo-todos, got %s", cfg.Squad.Planners.NearTerm)
	}

	// Verify SQUAD.md was stripped
	strippedData, err := os.ReadFile(squadMDPath)
	if err != nil {
		t.Fatalf("read stripped SQUAD.md: %v", err)
	}

	strippedContent := string(strippedData)
	if strings.Contains(strippedContent, "---") {
		t.Error("SQUAD.md still contains frontmatter delimiters")
	}
	if !strings.Contains(strippedContent, "# Mission") {
		t.Error("SQUAD.md body was not preserved")
	}
	if !strings.Contains(strippedContent, "Build the auth system") {
		t.Error("SQUAD.md body content was not preserved")
	}
}

func TestMigrateSquadDir_AlreadyMigrated(t *testing.T) {
	tmpDir := t.TempDir()
	squadDir := filepath.Join(tmpDir, "migrated-squad")
	if err := os.MkdirAll(squadDir, 0755); err != nil {
		t.Fatalf("create squad dir: %v", err)
	}

	// Create existing ayo.json
	ayoJSONPath := filepath.Join(squadDir, "ayo.json")
	if err := os.WriteFile(ayoJSONPath, []byte(`{"version":"1"}`), 0644); err != nil {
		t.Fatalf("write ayo.json: %v", err)
	}

	// Create SQUAD.md with frontmatter
	squadMDPath := filepath.Join(squadDir, "SQUAD.md")
	if err := os.WriteFile(squadMDPath, []byte("---\nlead: \"@old\"\n---\n# Test"), 0644); err != nil {
		t.Fatalf("write SQUAD.md: %v", err)
	}

	// Run migration
	result := migrateSquadDir("migrated-squad", squadDir, MigrateOptions{})

	// Should be skipped
	if !result.Skipped {
		t.Error("expected migration to be skipped")
	}
	if result.Migrated {
		t.Error("unexpected migration")
	}
}

func TestMigrateSquadDir_Force(t *testing.T) {
	tmpDir := t.TempDir()
	squadDir := filepath.Join(tmpDir, "force-squad")
	if err := os.MkdirAll(squadDir, 0755); err != nil {
		t.Fatalf("create squad dir: %v", err)
	}

	// Create existing ayo.json
	ayoJSONPath := filepath.Join(squadDir, "ayo.json")
	if err := os.WriteFile(ayoJSONPath, []byte(`{"version":"1"}`), 0644); err != nil {
		t.Fatalf("write ayo.json: %v", err)
	}

	// Create SQUAD.md with frontmatter
	squadMDPath := filepath.Join(squadDir, "SQUAD.md")
	if err := os.WriteFile(squadMDPath, []byte("---\nlead: \"@new\"\n---\n# Test"), 0644); err != nil {
		t.Fatalf("write SQUAD.md: %v", err)
	}

	// Run migration with force
	result := migrateSquadDir("force-squad", squadDir, MigrateOptions{Force: true})

	// Should not be skipped
	if result.Skipped {
		t.Error("expected migration not to be skipped with force")
	}
	if !result.Migrated {
		t.Error("expected migration to occur")
	}

	// Verify ayo.json was overwritten
	data, _ := os.ReadFile(ayoJSONPath)
	var cfg AyoConfig
	json.Unmarshal(data, &cfg)
	if cfg.Squad == nil || cfg.Squad.Lead != "@new" {
		t.Error("ayo.json was not overwritten")
	}
}

func TestMigrateSquadDir_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	squadDir := filepath.Join(tmpDir, "dryrun-squad")
	if err := os.MkdirAll(squadDir, 0755); err != nil {
		t.Fatalf("create squad dir: %v", err)
	}

	// Create SQUAD.md with frontmatter
	originalContent := "---\nlead: \"@architect\"\n---\n# Test"
	squadMDPath := filepath.Join(squadDir, "SQUAD.md")
	if err := os.WriteFile(squadMDPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("write SQUAD.md: %v", err)
	}

	// Run dry run
	result := migrateSquadDir("dryrun-squad", squadDir, MigrateOptions{DryRun: true})

	// Should not be migrated
	if result.Migrated {
		t.Error("expected no migration in dry run")
	}
	if result.Skipped {
		t.Error("unexpected skip in dry run")
	}

	// Verify ayo.json was NOT created
	ayoJSONPath := filepath.Join(squadDir, "ayo.json")
	if _, err := os.Stat(ayoJSONPath); err == nil {
		t.Error("ayo.json should not be created in dry run")
	}

	// Verify SQUAD.md was NOT modified
	data, _ := os.ReadFile(squadMDPath)
	if string(data) != originalContent {
		t.Error("SQUAD.md should not be modified in dry run")
	}
}

func TestMigrateSquadDir_NoFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	squadDir := filepath.Join(tmpDir, "nofm-squad")
	if err := os.MkdirAll(squadDir, 0755); err != nil {
		t.Fatalf("create squad dir: %v", err)
	}

	// Create SQUAD.md without frontmatter
	squadMDPath := filepath.Join(squadDir, "SQUAD.md")
	if err := os.WriteFile(squadMDPath, []byte("# Just markdown\n\nNo frontmatter here."), 0644); err != nil {
		t.Fatalf("write SQUAD.md: %v", err)
	}

	// Run migration
	result := migrateSquadDir("nofm-squad", squadDir, MigrateOptions{})

	// Should be skipped (no frontmatter to migrate)
	if !result.Skipped {
		t.Error("expected migration to be skipped")
	}
}

func TestNeedsMigrationDir(t *testing.T) {
	tmpDir := t.TempDir()
	squadDir := filepath.Join(tmpDir, "needs-migration")
	if err := os.MkdirAll(squadDir, 0755); err != nil {
		t.Fatalf("create squad dir: %v", err)
	}

	// Create SQUAD.md with frontmatter but no ayo.json
	squadMDPath := filepath.Join(squadDir, "SQUAD.md")
	if err := os.WriteFile(squadMDPath, []byte("---\nlead: \"@test\"\n---\n# Test"), 0644); err != nil {
		t.Fatalf("write SQUAD.md: %v", err)
	}

	if !needsMigrationDir(squadDir) {
		t.Error("expected squad to need migration")
	}

	// Create ayo.json
	ayoJSONPath := filepath.Join(squadDir, "ayo.json")
	if err := os.WriteFile(ayoJSONPath, []byte(`{"version":"1"}`), 0644); err != nil {
		t.Fatalf("write ayo.json: %v", err)
	}

	if needsMigrationDir(squadDir) {
		t.Error("expected squad to not need migration after ayo.json exists")
	}
}

func TestDeprecationWarningDir(t *testing.T) {
	tmpDir := t.TempDir()
	squadDir := filepath.Join(tmpDir, "deprecated-squad")
	if err := os.MkdirAll(squadDir, 0755); err != nil {
		t.Fatalf("create squad dir: %v", err)
	}

	squadMDPath := filepath.Join(squadDir, "SQUAD.md")
	if err := os.WriteFile(squadMDPath, []byte("---\nlead: \"@test\"\n---\n# Test"), 0644); err != nil {
		t.Fatalf("write SQUAD.md: %v", err)
	}

	warning := deprecationWarningDir("deprecated-squad", squadDir)
	if warning == "" {
		t.Error("expected deprecation warning")
	}
	if !strings.Contains(warning, "deprecated") {
		t.Error("warning should mention deprecated")
	}
	if !strings.Contains(warning, "ayo migrate squad") {
		t.Error("warning should mention migrate command")
	}
}
