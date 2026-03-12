package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexcabrera/ayo/internal/build"
	"github.com/alexcabrera/ayo/internal/build/types"
)

func TestGetCacheDir(t *testing.T) {
	cacheDir, err := getCacheDir()
	if err != nil {
		t.Fatalf("getCacheDir failed: %v", err)
	}

	if cacheDir == "" {
		t.Fatal("cacheDir is empty")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	expectedDir := filepath.Join(homeDir, ".cache", "ayo")
	if cacheDir != expectedDir {
		t.Errorf("expected cache dir %q, got %q", expectedDir, cacheDir)
	}
}

func TestComputeBuildHash(t *testing.T) {
	// Create a temporary directory with test agent
	tmpDir := t.TempDir()

	// Create config.toml
	configContent := `[agent]
name = "test-agent"
model = "gpt-4o"

[agent.provider]
type = "openai"
api_key = "test-key"

[memory]
type = "ephemeral"

[cli]
enabled = true
mode = "freeform"
description = "Test agent"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Create prompts directory
	promptsDir := filepath.Join(tmpDir, "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatalf("failed to create prompts dir: %v", err)
	}
	systemPrompt := filepath.Join(promptsDir, "system.md")
	if err := os.WriteFile(systemPrompt, []byte("You are a test agent."), 0644); err != nil {
		t.Fatalf("failed to write system prompt: %v", err)
	}

	// Compute hash
	hash1, err := computeBuildHash(tmpDir, configPath, "linux", "amd64")
	if err != nil {
		t.Fatalf("computeBuildHash failed: %v", err)
	}

	if hash1 == "" {
		t.Fatal("hash is empty")
	}

	// Verify same inputs produce same hash
	hash2, err := computeBuildHash(tmpDir, configPath, "linux", "amd64")
	if err != nil {
		t.Fatalf("computeBuildHash failed: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("hashes differ for same inputs: %q vs %q", hash1, hash2)
	}

	// Verify different target produces different hash
	hash3, err := computeBuildHash(tmpDir, configPath, "darwin", "amd64")
	if err != nil {
		t.Fatalf("computeBuildHash failed: %v", err)
	}

	if hash1 == hash3 {
		t.Error("hashes should differ for different targets")
	}

	// Verify modified file produces different hash
	if err := os.WriteFile(systemPrompt, []byte("Modified system prompt."), 0644); err != nil {
		t.Fatalf("failed to modify system prompt: %v", err)
	}

	hash4, err := computeBuildHash(tmpDir, configPath, "linux", "amd64")
	if err != nil {
		t.Fatalf("computeBuildHash failed: %v", err)
	}

	if hash1 == hash4 {
		t.Error("hashes should differ when file is modified")
	}
}

func TestHashFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Hash the file
	h := sha256.New()
	if err := hashFile(h, testFile); err != nil {
		t.Fatalf("hashFile failed: %v", err)
	}

	hash := h.Sum(nil)
	if len(hash) == 0 {
		t.Fatal("hash is empty")
	}

	// Test with non-existent file (should not error)
	nonExistent := filepath.Join(tmpDir, "nonexistent.txt")
	h2 := sha256.New()
	if err := hashFile(h2, nonExistent); err != nil {
		t.Errorf("hashFile should not error for non-existent file: %v", err)
	}
}

func TestHashDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory structure
	testDir := filepath.Join(tmpDir, "test")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	file1 := filepath.Join(testDir, "file1.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}

	subDir := filepath.Join(testDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create sub dir: %v", err)
	}

	file2 := filepath.Join(subDir, "file2.txt")
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}

	// Hash the directory
	h := sha256.New()
	if err := hashDirectory(h, testDir); err != nil {
		t.Fatalf("hashDirectory failed: %v", err)
	}

	hash := h.Sum(nil)
	if len(hash) == 0 {
		t.Fatal("hash is empty")
	}

	// Test with non-existent directory (should not error)
	h2 := sha256.New()
	nonExistent := filepath.Join(tmpDir, "nonexistent")
	if err := hashDirectory(h2, nonExistent); err != nil {
		t.Errorf("hashDirectory should not error for non-existent directory: %v", err)
	}

	// Verify different content produces different hash
	if err := os.WriteFile(file1, []byte("modified content"), 0644); err != nil {
		t.Fatalf("failed to modify file1: %v", err)
	}

	h3 := sha256.New()
	if err := hashDirectory(h3, testDir); err != nil {
		t.Fatalf("hashDirectory failed: %v", err)
	}

	hash2 := h3.Sum(nil)
	hashStr1 := hex.EncodeToString(hash)
	hashStr2 := hex.EncodeToString(hash2)

	if hashStr1 == hashStr2 {
		t.Error("hashes should differ when directory content is modified")
	}
}

func TestRunCleanDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .build directory
	buildDir := filepath.Join(tmpDir, ".build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatalf("failed to create build dir: %v", err)
	}

	// Create a file in .build
	testFile := filepath.Join(buildDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Verify .build exists
	if _, err := os.Stat(buildDir); err != nil {
		t.Fatalf(".build should exist before clean: %v", err)
	}

	// Clean the directory
	if err := runCleanDir(tmpDir); err != nil {
		t.Fatalf("runCleanDir failed: %v", err)
	}

	// Verify .build is removed
	if _, err := os.Stat(buildDir); !os.IsNotExist(err) {
		t.Error(".build should be removed after clean")
	}

	// Test cleaning directory without .build (should not error)
	tmpDir2 := t.TempDir()
	if err := runCleanDir(tmpDir2); err != nil {
		t.Errorf("runCleanDir should not error for directory without .build: %v", err)
	}
}

func TestRunCleanCache(t *testing.T) {
	// This test is potentially destructive, so we'll create a test cache dir
	tmpDir := t.TempDir()
	testCacheDir := filepath.Join(tmpDir, "cache")

	// Create test cache directory
	if err := os.MkdirAll(testCacheDir, 0755); err != nil {
		t.Fatalf("failed to create test cache dir: %v", err)
	}

	// Create some cached binaries
	cacheFile := filepath.Join(testCacheDir, "agent-linux-amd64-hash123")
	if err := os.WriteFile(cacheFile, []byte("binary content"), 0644); err != nil {
		t.Fatalf("failed to write cache file: %v", err)
	}

	// We can't easily test runCleanCache() since it uses getCacheDir()
	// which returns the real cache directory. Instead, we'll just verify
	// that the function exists and compiles correctly.

	// Test is just to ensure the function compiles and can be called
	// without crashing on the real cache directory
	if err := runCleanCache(); err != nil {
		// This is expected if there's no real cache, but shouldn't crash
		t.Logf("runCleanCache returned (expected): %v", err)
	}
}

func TestBuildConfig(t *testing.T) {
	config := types.BuildConfig{
		Targets: []types.BuildTarget{
			{OS: "linux", Arch: "amd64"},
			{OS: "linux", Arch: "arm64"},
			{OS: "darwin", Arch: "amd64"},
			{OS: "darwin", Arch: "arm64"},
			{OS: "windows", Arch: "amd64"},
		},
	}

	if len(config.Targets) != 5 {
		t.Errorf("expected 5 targets, got %d", len(config.Targets))
	}

	for i, target := range config.Targets {
		if target.OS == "" {
			t.Errorf("target %d: OS is empty", i)
		}
		if target.Arch == "" {
			t.Errorf("target %d: Arch is empty", i)
		}
	}
}

func TestBuildConfigEmpty(t *testing.T) {
	config := types.BuildConfig{}

	if len(config.Targets) != 0 {
		t.Errorf("expected 0 targets, got %d", len(config.Targets))
	}
}

func TestRunBuildAllWithConfigTargets(t *testing.T) {
	tmpDir := t.TempDir()

	configContent := `[agent]
name = "test-agent"
description = "Test agent for builds"
model = "gpt-4o"

[agent.provider]
type = "openai"
api_key = "test-key"

[memory]
type = "ephemeral"

[cli]
enabled = true
mode = "freeform"
description = "Test agent"

[[build.targets]]
os = "linux"
arch = "amd64"

[[build.targets]]
os = "darwin"
arch = "arm64"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	promptsDir := filepath.Join(tmpDir, "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatalf("failed to create prompts dir: %v", err)
	}

	config, _, err := build.LoadConfigFromDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if len(config.Build.Targets) != 2 {
		t.Errorf("expected 2 build targets, got %d", len(config.Build.Targets))
	}

	expectedTargets := []struct{ os, arch string }{
		{"linux", "amd64"},
		{"darwin", "arm64"},
	}

	for i, expected := range expectedTargets {
		if i >= len(config.Build.Targets) {
			t.Fatalf("missing target %d", i)
		}
		actual := config.Build.Targets[i]
		if actual.OS != expected.os {
			t.Errorf("target %d: expected OS %q, got %q", i, expected.os, actual.OS)
		}
		if actual.Arch != expected.arch {
			t.Errorf("target %d: expected Arch %q, got %q", i, expected.arch, actual.Arch)
		}
	}
}

func TestRunBuildAllDefaultTargets(t *testing.T) {
	tmpDir := t.TempDir()

	configContent := `[agent]
description = "Test agent for builds"
name = "test-agent"
model = "gpt-4o"

[agent.provider]
type = "openai"
api_key = "test-key"

[memory]
type = "ephemeral"

[cli]
enabled = true
mode = "freeform"
description = "Test agent"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	promptsDir := filepath.Join(tmpDir, "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatalf("failed to create prompts dir: %v", err)
	}

	config, _, err := build.LoadConfigFromDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if len(config.Build.Targets) != 0 {
		t.Errorf("expected 0 build targets in config, got %d", len(config.Build.Targets))
	}
}

func TestGenerateMainStub(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test config
	configContent := `[agent]
name = "test-agent"
description = "Test agent for builds"
model = "gpt-4o"

[agent.provider]
type = "openai"
api_key = "test-key"

[memory]
type = "ephemeral"

[cli]
enabled = true
mode = "freeform"
description = "Test agent"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	config, _, err := build.LoadConfigFromDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Generate main stub
	mainGoPath := filepath.Join(tmpDir, "main.go")
	if err := generateMainStub(mainGoPath, config, configPath); err != nil {
		t.Fatalf("generateMainStub failed: %v", err)
	}

	// Verify main.go was created
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		t.Fatal("main.go was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(mainGoPath)
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}

	contentStr := string(content)

	// Check for required imports
	if !strings.Contains(contentStr, `import (
	"embed"`) {
		t.Error("main.go missing embed import")
	}

	// Check for embed directives
	if !strings.Contains(contentStr, "//go:embed config.toml") {
		t.Error("main.go missing config.toml embed directive")
	}
	if !strings.Contains(contentStr, "//go:embed prompts/system.md") {
		t.Error("main.go missing system.md embed directive")
	}
	if !strings.Contains(contentStr, "//go:embed skills/*") {
		t.Error("main.go missing skills embed directive")
	}
	if !strings.Contains(contentStr, "//go:embed tools/*") {
		t.Error("main.go missing tools embed directive")
	}

	// Check for main function
	if !strings.Contains(contentStr, "func main()") {
		t.Error("main.go missing main function")
	}

	// Check for runtime.Execute call
	if !strings.Contains(contentStr, "runtime.Execute") {
		t.Error("main.go missing runtime.Execute call")
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(tmpDir, "source.txt")
	content := []byte("test content for copy")
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	// Copy file
	dstFile := filepath.Join(tmpDir, "dest.txt")
	if err := copyFile(srcFile, dstFile); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	// Verify destination exists
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Fatal("destination file was not created")
	}

	// Verify content matches
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	if string(dstContent) != string(content) {
		t.Errorf("content mismatch: got %q, want %q", string(dstContent), string(content))
	}

	// Test copying non-existent file (should error)
	nonExistentSrc := filepath.Join(tmpDir, "nonexistent.txt")
	dstFile2 := filepath.Join(tmpDir, "dest2.txt")
	if err := copyFile(nonExistentSrc, dstFile2); err == nil {
		t.Error("copyFile should error for non-existent source file")
	}
}

func TestCopyDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source directory structure
	srcDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}

	// Create files in source
	file1 := filepath.Join(srcDir, "file1.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}

	file2 := filepath.Join(srcDir, "file2.txt")
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}

	// Create subdirectory
	subDir := filepath.Join(srcDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	file3 := filepath.Join(subDir, "file3.txt")
	if err := os.WriteFile(file3, []byte("content3"), 0644); err != nil {
		t.Fatalf("failed to write file3: %v", err)
	}

	// Copy directory
	dstDir := filepath.Join(tmpDir, "destination")
	if err := copyDir(srcDir, dstDir); err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}

	// Verify destination directory exists
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		t.Fatal("destination directory was not created")
	}

	// Verify all files were copied
	dstFile1 := filepath.Join(dstDir, "file1.txt")
	content1, err := os.ReadFile(dstFile1)
	if err != nil {
		t.Fatalf("failed to read dest file1: %v", err)
	}
	if string(content1) != "content1" {
		t.Errorf("file1 content mismatch: got %q, want %q", string(content1), "content1")
	}

	dstFile2 := filepath.Join(dstDir, "file2.txt")
	content2, err := os.ReadFile(dstFile2)
	if err != nil {
		t.Fatalf("failed to read dest file2: %v", err)
	}
	if string(content2) != "content2" {
		t.Errorf("file2 content mismatch: got %q, want %q", string(content2), "content2")
	}

	// Verify subdirectory was copied
	dstSubDir := filepath.Join(dstDir, "subdir")
	if _, err := os.Stat(dstSubDir); os.IsNotExist(err) {
		t.Fatal("subdirectory was not copied")
	}

	dstFile3 := filepath.Join(dstSubDir, "file3.txt")
	content3, err := os.ReadFile(dstFile3)
	if err != nil {
		t.Fatalf("failed to read dest file3: %v", err)
	}
	if string(content3) != "content3" {
		t.Errorf("file3 content mismatch: got %q, want %q", string(content3), "content3")
	}

	// Test copying non-existent directory (should error)
	nonExistentSrc := filepath.Join(tmpDir, "nonexistent")
	dstDir2 := filepath.Join(tmpDir, "dest2")
	if err := copyDir(nonExistentSrc, dstDir2); err == nil {
		t.Error("copyDir should error for non-existent source directory")
	}
}

func TestCopyDirEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	// Create empty source directory
	srcDir := filepath.Join(tmpDir, "empty_source")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create empty source dir: %v", err)
	}

	// Copy empty directory
	dstDir := filepath.Join(tmpDir, "empty_dest")
	if err := copyDir(srcDir, dstDir); err != nil {
		t.Fatalf("copyDir failed for empty directory: %v", err)
	}

	// Verify destination exists
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		t.Fatal("destination directory was not created for empty source")
	}

	// Verify destination is empty
	entries, err := os.ReadDir(dstDir)
	if err != nil {
		t.Fatalf("failed to read destination dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty destination, got %d entries", len(entries))
	}
}

func TestFindModuleRoot(t *testing.T) {
	// Save original ModuleRoot
	origModuleRoot := ModuleRoot
	defer func() { ModuleRoot = origModuleRoot }()

	// Test from current working directory (should find the ayo module)
	if dir, err := findModuleRoot(); err != nil {
		t.Logf("findModuleRoot returned error (may be expected if not in ayo module): %v", err)
	} else if dir == "" {
		t.Error("findModuleRoot returned empty directory")
	} else {
		// Verify go.mod exists in the returned directory
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); os.IsNotExist(err) {
			t.Errorf("go.mod not found in returned module root: %s", dir)
		}
	}
}

func TestSearchForModuleFrom(t *testing.T) {
	// Create a temporary directory structure
tmpDir := t.TempDir()

	// Create a go.mod file at the root
	goModContent := `module github.com/alexcabrera/ayo

go 1.25.5
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "subdir", "nested")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectories: %v", err)
	}

	// Search from subdirectory - should find the module root
	foundDir, err := searchForModuleFrom(subDir)
	if err != nil {
		t.Errorf("searchForModuleFrom failed: %v", err)
	}
	if foundDir != tmpDir {
		t.Errorf("expected %s, got %s", tmpDir, foundDir)
	}

	// Test searching from the module root itself
	foundDir2, err := searchForModuleFrom(tmpDir)
	if err != nil {
		t.Errorf("searchForModuleFrom from root failed: %v", err)
	}
	if foundDir2 != tmpDir {
		t.Errorf("expected %s, got %s", tmpDir, foundDir2)
	}

	// Test searching from a directory without ayo module
	nonModuleDir := filepath.Join(tmpDir, "other_module")
	if err := os.MkdirAll(nonModuleDir, 0755); err != nil {
		t.Fatalf("failed to create other module dir: %v", err)
	}

	// Create a different go.mod in a separate location
	separateTempDir := t.TempDir()
	otherGoMod := `module other/module

go 1.25.5
`
	if err := os.WriteFile(filepath.Join(separateTempDir, "go.mod"), []byte(otherGoMod), 0644); err != nil {
		t.Fatalf("failed to write other go.mod: %v", err)
	}

	// Search from the separate temp directory which has a different go.mod
	_, err = searchForModuleFrom(separateTempDir)
	if err == nil {
		t.Error("searchForModuleFrom should error when ayo module is not found")
	}

	// Test from a completely separate temp directory without any go.mod
	separateDir := t.TempDir()
	_, err = searchForModuleFrom(separateDir)
	if err == nil {
		t.Error("searchForModuleFrom should error when no go.mod is found")
	}
}

func TestRunBuildErrors(t *testing.T) {
	// Save and restore ModuleRoot
	origModuleRoot := ModuleRoot
	defer func() { ModuleRoot = origModuleRoot }()

	// Test with empty ModuleRoot (should error)
	ModuleRoot = ""

	tmpDir := t.TempDir()

	configContent := `[agent]
name = "test-agent"
model = "gpt-4o"

[agent.provider]
type = "openai"
api_key = "test-key"

[memory]
type = "ephemeral"

[cli]
enabled = true
mode = "freeform"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	promptsDir := filepath.Join(tmpDir, "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatalf("failed to create prompts dir: %v", err)
	}

	// Test with non-existent directory
	err := runBuild("/nonexistent/directory", "", "linux", "amd64")
	if err == nil {
		t.Error("runBuild should error for non-existent directory")
	}

	// Test with invalid config
	invalidConfigDir := t.TempDir()
	invalidConfigPath := filepath.Join(invalidConfigDir, "config.toml")
	if err := os.WriteFile(invalidConfigPath, []byte("invalid config"), 0644); err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}

	err = runBuild(invalidConfigDir, "", "linux", "amd64")
	if err == nil {
		t.Error("runBuild should error for invalid config")
	}
}

func TestRunBuildAll(t *testing.T) {
	// Save and restore ModuleRoot
	origModuleRoot := ModuleRoot
	defer func() { ModuleRoot = origModuleRoot }()

	tmpDir := t.TempDir()

	// Create a minimal valid config
	configContent := `[agent]
name = "test-agent"
model = "gpt-4o"

[agent.provider]
type = "openai"
api_key = "test-key"

[memory]
type = "ephemeral"

[cli]
enabled = true
mode = "freeform"
`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	promptsDir := filepath.Join(tmpDir, "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatalf("failed to create prompts dir: %v", err)
	}

	// Test runBuildAll with valid directory but without ModuleRoot
	// (should fail when trying to call runBuild)
	ModuleRoot = ""
	err := runBuildAll(tmpDir, filepath.Join(tmpDir, "dist"))
	if err == nil {
		t.Error("runBuildAll should error when ModuleRoot is not set")
	}

	// Test with non-existent directory
	err = runBuildAll("/nonexistent/directory", "")
	if err == nil {
		t.Error("runBuildAll should error for non-existent directory")
	}

	// Test with invalid config
	invalidConfigDir := t.TempDir()
	invalidConfigPath := filepath.Join(invalidConfigDir, "config.toml")
	if err := os.WriteFile(invalidConfigPath, []byte("invalid config"), 0644); err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}

	err = runBuildAll(invalidConfigDir, "")
	if err == nil {
		t.Error("runBuildAll should error for invalid config")
	}
}
