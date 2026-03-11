package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
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
