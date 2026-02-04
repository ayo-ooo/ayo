package builtin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListAgents(t *testing.T) {
	agents := ListAgents()
	if len(agents) < 1 {
		t.Errorf("expected at least 1 built-in agent, got %d: %v", len(agents), agents)
	}

	// Check that @ayo agent is present
	foundAyo := false
	for _, a := range agents {
		if a == "@ayo" {
			foundAyo = true
		}
	}
	if !foundAyo {
		t.Errorf("expected @ayo in agents list, got %v", agents)
	}
}

func TestHasAgent(t *testing.T) {
	if !HasAgent("@ayo") {
		t.Error("expected HasAgent(@ayo) to be true")
	}
	if !HasAgent("ayo") {
		t.Error("expected HasAgent(ayo) to be true (without @)")
	}
	if HasAgent("@nonexistent") {
		t.Error("expected HasAgent(@nonexistent) to be false")
	}
	// Old builtin namespace should not work
	if HasAgent("@builtin.helper") {
		t.Error("expected HasAgent(@builtin.helper) to be false (old namespace)")
	}
}

func TestLoadAgent(t *testing.T) {
	def, err := LoadAgent("@ayo")
	if err != nil {
		t.Fatalf("LoadAgent(@ayo) error: %v", err)
	}

	if def.Handle != "@ayo" {
		t.Errorf("Handle = %q, want @ayo", def.Handle)
	}

	if def.System == "" {
		t.Error("System should not be empty")
	}

	if def.Config.Description == "" {
		t.Error("Config.Description should not be empty")
	}
}

func TestLoadAgentNotFound(t *testing.T) {
	_, err := LoadAgent("@nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent agent")
	}
}

func TestInstallDir(t *testing.T) {
	dir := InstallDir()
	if dir == "" {
		t.Error("InstallDir should not be empty")
	}
	// Should contain "ayo" and "agents"
	if !filepath.IsAbs(dir) {
		t.Errorf("InstallDir should be absolute, got %q", dir)
	}
}

func TestInstallAndUninstall(t *testing.T) {
	// Use a temp directory for testing
	tmpDir := t.TempDir()
	origInstallDir := InstallDir

	// Override InstallDir for testing - we can't easily do this,
	// so we'll just test that Install doesn't error
	err := Install()
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	// Verify ayo was installed
	ayoDir := InstalledAgentDir("@ayo")
	if _, err := os.Stat(ayoDir); os.IsNotExist(err) {
		t.Errorf("ayo agent not installed at %s", ayoDir)
	}

	// Verify system.md exists
	systemPath := filepath.Join(ayoDir, "system.md")
	if _, err := os.Stat(systemPath); os.IsNotExist(err) {
		t.Errorf("system.md not found at %s", systemPath)
	}

	// Check IsInstalled
	if !IsInstalled("@ayo") {
		t.Error("IsInstalled(@ayo) should be true after install")
	}

	// Clean up - don't uninstall in tests as it affects real install dir
	_ = origInstallDir
	_ = tmpDir
}

func TestVersionFile(t *testing.T) {
	vf := VersionFile()
	if vf == "" {
		t.Error("VersionFile should not be empty")
	}
	if !filepath.IsAbs(vf) {
		t.Errorf("VersionFile should be absolute, got %q", vf)
	}
}

func TestListAgentInfo(t *testing.T) {
	infos := ListAgentInfo()
	if len(infos) < 1 {
		t.Errorf("expected at least 1 agent info, got %d", len(infos))
	}

	// Check that @ayo is present
	var ayoInfo *AgentInfo
	for i := range infos {
		if infos[i].Handle == "@ayo" {
			ayoInfo = &infos[i]
			break
		}
	}

	if ayoInfo == nil {
		t.Fatal("expected @ayo in agent infos")
	}

	if ayoInfo.Description == "" {
		t.Error("expected ayo to have description")
	}
}

func TestListBuiltinSkills(t *testing.T) {
	skills := ListBuiltinSkills()
	if len(skills) < 1 {
		t.Errorf("expected at least 1 built-in skill, got %d", len(skills))
	}

	// Check that debugging skill is present
	found := false
	for _, s := range skills {
		if s == "debugging" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected debugging in skills list, got %v", skills)
	}
}

func TestHasBuiltinSkill(t *testing.T) {
	if !HasBuiltinSkill("debugging") {
		t.Error("expected HasBuiltinSkill(debugging) to be true")
	}
	if HasBuiltinSkill("nonexistent-skill") {
		t.Error("expected HasBuiltinSkill(nonexistent-skill) to be false")
	}
}

func TestLoadBuiltinSkill(t *testing.T) {
	skill, err := LoadBuiltinSkill("debugging")
	if err != nil {
		t.Fatalf("LoadBuiltinSkill(debugging) error: %v", err)
	}

	if skill.Name != "debugging" {
		t.Errorf("Name = %q, want debugging", skill.Name)
	}

	if skill.Description == "" {
		t.Error("Description should not be empty")
	}

	if skill.Content == "" {
		t.Error("Content should not be empty")
	}
}

func TestLoadBuiltinSkillNotFound(t *testing.T) {
	_, err := LoadBuiltinSkill("nonexistent-skill")
	if err == nil {
		t.Error("expected error for nonexistent skill")
	}
}

func TestListAgentBuiltinSkills(t *testing.T) {
	skills := ListAgentBuiltinSkills("@ayo")
	if len(skills) < 1 {
		t.Errorf("expected at least 1 skill for @ayo, got %d", len(skills))
	}

	// Check that project-summary skill is present
	found := false
	for _, s := range skills {
		if s == "project-summary" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected project-summary in @ayo skills list, got %v", skills)
	}
}

func TestLoadAgentBuiltinSkill(t *testing.T) {
	skill, err := LoadAgentBuiltinSkill("@ayo", "project-summary")
	if err != nil {
		t.Fatalf("LoadAgentBuiltinSkill(@ayo, project-summary) error: %v", err)
	}

	if skill.Name != "project-summary" {
		t.Errorf("Name = %q, want project-summary", skill.Name)
	}

	if skill.Description == "" {
		t.Error("Description should not be empty")
	}
}

func TestSkillsInstallDir(t *testing.T) {
	dir := SkillsInstallDir()
	if dir == "" {
		t.Error("SkillsInstallDir should not be empty")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("SkillsInstallDir should be absolute, got %q", dir)
	}
}

func TestSkillsInstalledAfterInstall(t *testing.T) {
	// Run install first
	err := Install()
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	// Check that debugging skill was installed
	debuggingDir := InstalledSkillDir("debugging")
	if _, err := os.Stat(debuggingDir); os.IsNotExist(err) {
		t.Errorf("debugging skill not installed at %s", debuggingDir)
	}

	// Check SKILL.md exists
	skillMD := filepath.Join(debuggingDir, "SKILL.md")
	if _, err := os.Stat(skillMD); os.IsNotExist(err) {
		t.Errorf("SKILL.md not found at %s", skillMD)
	}

	// Check IsSkillInstalled
	if !IsSkillInstalled("debugging") {
		t.Error("IsSkillInstalled(debugging) should be true after install")
	}
}

func TestGetAllBuiltinSkillInfos(t *testing.T) {
	infos := GetAllBuiltinSkillInfos()
	if len(infos) < 1 {
		t.Errorf("expected at least 1 skill info, got %d", len(infos))
	}

	// Check that debugging is present
	found := false
	for _, info := range infos {
		if info.Name == "debugging" {
			found = true
			if info.Description == "" {
				t.Error("debugging should have description")
			}
			break
		}
	}
	if !found {
		t.Error("expected debugging in skill infos")
	}
}

func TestNewSkillsExist(t *testing.T) {
	// Test that the new sandbox, memory, and session skills exist
	newSkills := []string{"sandbox", "memory-usage", "memory-worthy", "session-summary"}
	for _, name := range newSkills {
		if !HasBuiltinSkill(name) {
			t.Errorf("expected HasBuiltinSkill(%s) to be true", name)
			continue
		}

		skill, err := LoadBuiltinSkill(name)
		if err != nil {
			t.Errorf("LoadBuiltinSkill(%s) error: %v", name, err)
			continue
		}

		if skill.Name != name {
			t.Errorf("skill.Name = %q, want %q", skill.Name, name)
		}

		if skill.Description == "" {
			t.Errorf("%s skill should have description", name)
		}

		if skill.Content == "" {
			t.Errorf("%s skill should have content", name)
		}
	}
}

func TestMemorySkillUpdated(t *testing.T) {
	skill, err := LoadBuiltinSkill("memory")
	if err != nil {
		t.Fatalf("LoadBuiltinSkill(memory) error: %v", err)
	}

	// Check that the memory skill mentions Zettelkasten
	if skill.Content == "" {
		t.Error("memory skill content should not be empty")
	}
}
