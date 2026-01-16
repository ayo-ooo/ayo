package paths

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDataDir(t *testing.T) {
	dir := DataDir()
	if dir == "" {
		t.Error("DataDir returned empty string")
	}
	// In dev mode, ends with .ayo; in production, contains "ayo"
	if !strings.Contains(dir, "ayo") && !strings.HasSuffix(dir, ".ayo") {
		t.Errorf("DataDir should contain 'ayo' or end with '.ayo': got %s", dir)
	}
}

func TestConfigDir(t *testing.T) {
	dir := ConfigDir()
	if dir == "" {
		t.Error("ConfigDir returned empty string")
	}
	if !strings.Contains(dir, "ayo") {
		t.Errorf("ConfigDir should contain 'ayo': got %s", dir)
	}
}

func TestAgentsDir(t *testing.T) {
	dir := AgentsDir()
	if !strings.HasSuffix(dir, filepath.Join("ayo", "agents")) {
		t.Errorf("AgentsDir should end with ayo/agents: got %s", dir)
	}
}

func TestBuiltinAgentsDir(t *testing.T) {
	dir := BuiltinAgentsDir()
	// In dev mode: .ayo/agents, in production: ayo/agents
	if !strings.HasSuffix(dir, "agents") {
		t.Errorf("BuiltinAgentsDir should end with agents: got %s", dir)
	}
}

func TestSkillsDir(t *testing.T) {
	dir := SkillsDir()
	if !strings.HasSuffix(dir, filepath.Join("ayo", "skills")) {
		t.Errorf("SkillsDir should end with ayo/skills: got %s", dir)
	}
}

func TestBuiltinSkillsDir(t *testing.T) {
	dir := BuiltinSkillsDir()
	// In dev mode: .ayo/skills, in production: ayo/skills
	if !strings.HasSuffix(dir, "skills") {
		t.Errorf("BuiltinSkillsDir should end with skills: got %s", dir)
	}
}

func TestConfigFile(t *testing.T) {
	file := ConfigFile()
	if !strings.HasSuffix(file, "ayo.json") {
		t.Errorf("ConfigFile should end with ayo.json: got %s", file)
	}
}

func TestSystemPromptsDir(t *testing.T) {
	dir := SystemPromptsDir()
	if !strings.HasSuffix(dir, filepath.Join("ayo", "prompts")) {
		t.Errorf("SystemPromptsDir should end with ayo/prompts: got %s", dir)
	}
}

func TestVersionFile(t *testing.T) {
	file := VersionFile()
	if !strings.HasSuffix(file, ".builtin-version") {
		t.Errorf("VersionFile should end with .builtin-version: got %s", file)
	}
}

func TestIsDevMode(t *testing.T) {
	// When running tests from the repo, we should be in dev mode
	if !IsDevMode() {
		t.Log("Not in dev mode - running from installed binary or outside repo")
	}

	if IsDevMode() {
		root := DevRoot()
		if root == "" {
			t.Error("IsDevMode() is true but DevRoot() is empty")
		}
		// Verify go.mod exists at dev root
		goModPath := filepath.Join(root, "go.mod")
		if _, err := os.Stat(goModPath); err != nil {
			t.Errorf("Dev root %s should contain go.mod", root)
		}
	}
}

func TestDevModeDataDir(t *testing.T) {
	if !IsDevMode() {
		t.Skip("not in dev mode")
	}

	dataDir := DataDir()
	devRoot := DevRoot()

	// DataDir should be {devRoot}/.local/share/ayo
	expected := filepath.Join(devRoot, ".local", "share", "ayo")
	if dataDir != expected {
		t.Errorf("Dev DataDir: expected %s, got %s", expected, dataDir)
	}
}

func TestDevModeConfigDir(t *testing.T) {
	if !IsDevMode() {
		t.Skip("not in dev mode")
	}

	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows")
	}

	devRoot := DevRoot()
	configDir := ConfigDir()
	expected := filepath.Join(devRoot, ".config", "ayo")

	// ConfigDir should be project-local in dev mode
	if configDir != expected {
		t.Errorf("ConfigDir should be project-local in dev mode: expected %s, got %s", expected, configDir)
	}
}

func TestConfigDirStructure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows")
	}

	configDir := ConfigDir()

	// User dirs should be under ConfigDir
	if !strings.HasPrefix(AgentsDir(), configDir) {
		t.Errorf("AgentsDir should be under ConfigDir: got %s", AgentsDir())
	}
	if !strings.HasPrefix(SkillsDir(), configDir) {
		t.Errorf("SkillsDir should be under ConfigDir: got %s", SkillsDir())
	}
	if !strings.HasPrefix(SystemPromptsDir(), configDir) {
		t.Errorf("SystemPromptsDir should be under ConfigDir: got %s", SystemPromptsDir())
	}
}

func TestBuiltinDirsUnderDataDir(t *testing.T) {
	dataDir := DataDir()

	// Builtin dirs should be under DataDir
	if !strings.HasPrefix(BuiltinAgentsDir(), dataDir) {
		t.Errorf("BuiltinAgentsDir should be under DataDir: got %s (DataDir: %s)", BuiltinAgentsDir(), dataDir)
	}
	if !strings.HasPrefix(BuiltinSkillsDir(), dataDir) {
		t.Errorf("BuiltinSkillsDir should be under DataDir: got %s (DataDir: %s)", BuiltinSkillsDir(), dataDir)
	}
	if !strings.HasPrefix(VersionFile(), dataDir) {
		t.Errorf("VersionFile should be under DataDir: got %s (DataDir: %s)", VersionFile(), dataDir)
	}
}

func TestWindowsPaths(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("skipping Windows path test on non-Windows")
	}

	if IsDevMode() {
		t.Skip("skipping production path test in dev mode")
	}

	dataDir := DataDir()
	configDir := ConfigDir()

	// On Windows production, DataDir and ConfigDir should be the same
	if dataDir != configDir {
		t.Errorf("Windows DataDir and ConfigDir should match: data=%s, config=%s", dataDir, configDir)
	}

	if !strings.Contains(dataDir, "ayo") {
		t.Errorf("Windows DataDir should contain 'ayo': got %s", dataDir)
	}
}
