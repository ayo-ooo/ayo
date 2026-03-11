package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexcabrera/ayo/internal/build"
	"github.com/alexcabrera/ayo/internal/build/types"
)

func TestRunPackage(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test agent directory
	agentDir := filepath.Join(tmpDir, "test-agent")
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config.toml
	configContent := `[agent]
name = "test-agent"
description = "Test agent for packaging"
version = "1.0.0"
model = "gpt-4"

[cli]
mode = "freeform"
description = "Test CLI"
`
	configPath := filepath.Join(agentDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a mock binary in dist/
	distDir := filepath.Join(agentDir, "dist")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		t.Fatal(err)
	}

	binaryPath := filepath.Join(distDir, "test-agent-linux-amd64")
	binaryContent := []byte("mock binary content")
	if err := os.WriteFile(binaryPath, binaryContent, 0755); err != nil {
		t.Fatal(err)
	}

	// Run package
	err := runPackage(agentDir, "auto", "1.0.0")
	if err != nil {
		t.Fatalf("runPackage failed: %v", err)
	}

	// Check that releases directory exists
	releasesDir := filepath.Join(agentDir, "releases")
	if _, err := os.Stat(releasesDir); os.IsNotExist(err) {
		t.Fatalf("releases directory not created")
	}

	// Check that archive was created
	archivePath := filepath.Join(releasesDir, "test-agent-1.0.0-linux-amd64.tar.gz")
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Fatalf("archive not created: %s", archivePath)
	}

	// Check archive content
	file, err := os.Open(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	found := false
	for {
		header, err := tr.Next()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			t.Fatal(err)
		}
		if strings.Contains(header.Name, "test-agent-linux-amd64") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("binary not found in archive")
	}

	// Check checksums file
	checksumsPath := filepath.Join(releasesDir, "test-agent-1.0.0.sha256")
	content, err := os.ReadFile(checksumsPath)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "test-agent-1.0.0-linux-amd64.tar.gz") {
		t.Errorf("checksums file doesn't contain archive name")
	}
}

func TestFindBuiltBinaries(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with dist/ directory
	distDir := filepath.Join(tmpDir, "dist")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create mock binaries
	binaries := []string{"agent1-linux-amd64", "agent1-darwin-arm64", "agent1-windows-amd64.exe"}
	for _, bin := range binaries {
		path := filepath.Join(distDir, bin)
		if err := os.WriteFile(path, []byte("binary"), 0755); err != nil {
			t.Fatal(err)
		}
	}

	found, err := findBuiltBinaries(tmpDir, "agent1")
	if err != nil {
		t.Fatal(err)
	}

	if len(found) != 3 {
		t.Errorf("expected 3 binaries, got %d", len(found))
	}
}

func TestCreateTarGz(t *testing.T) {
	tmpDir := t.TempDir()

	sourcePath := filepath.Join(tmpDir, "source.txt")
	sourceContent := []byte("test content")
	if err := os.WriteFile(sourcePath, sourceContent, 0644); err != nil {
		t.Fatal(err)
	}

	destPath := filepath.Join(tmpDir, "archive.tar.gz")

	if err := createTarGz(sourcePath, destPath); err != nil {
		t.Fatalf("createTarGz failed: %v", err)
	}

	// Verify archive
	file, err := os.Open(destPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	header, err := tr.Next()
	if err != nil {
		t.Fatal(err)
	}

	if header.Name != "source.txt" {
		t.Errorf("unexpected file name in archive: %s", header.Name)
	}
}

func TestCreateZip(t *testing.T) {
	tmpDir := t.TempDir()

	sourcePath := filepath.Join(tmpDir, "source.txt")
	sourceContent := []byte("test content")
	if err := os.WriteFile(sourcePath, sourceContent, 0644); err != nil {
		t.Fatal(err)
	}

	destPath := filepath.Join(tmpDir, "archive.zip")

	if err := createZip(sourcePath, destPath); err != nil {
		t.Fatalf("createZip failed: %v", err)
	}

	// Verify archive
	archive, err := zip.OpenReader(destPath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()

	if len(archive.File) != 1 {
		t.Errorf("expected 1 file in zip, got %d", len(archive.File))
	}

	if archive.File[0].Name != "source.txt" {
		t.Errorf("unexpected file name in zip: %s", archive.File[0].Name)
	}
}

func TestComputeFileHash(t *testing.T) {
	tmpDir := t.TempDir()

	filePath := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatal(err)
	}

	hash1, err := computeFileHash(filePath)
	if err != nil {
		t.Fatal(err)
	}

	hash2, err := computeFileHash(filePath)
	if err != nil {
		t.Fatal(err)
	}

	if hash1 != hash2 {
		t.Errorf("hashes should be consistent")
	}

	if len(hash1) != 64 { // SHA-256 produces 64 hex characters
		t.Errorf("expected 64 character hash, got %d", len(hash1))
	}
}

func TestWriteChecksums(t *testing.T) {
	tmpDir := t.TempDir()

	path := filepath.Join(tmpDir, "checksums.sha256")
	checksums := map[string]string{
		"file1.tar.gz": "abc123",
		"file2.tar.gz": "def456",
	}

	if err := writeChecksums(path, checksums); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "abc123  file1.tar.gz") {
		t.Errorf("checksums file missing entry for file1")
	}

	if !strings.Contains(contentStr, "def456  file2.tar.gz") {
		t.Errorf("checksums file missing entry for file2")
	}
}

func TestRunPackage_NoBinaries(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test agent directory without binaries
	agentDir := filepath.Join(tmpDir, "test-agent")
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config.toml
	configContent := `[agent]
name = "test-agent"
description = "Test agent for packaging"
version = "1.0.0"
model = "gpt-4"

[cli]
mode = "freeform"
description = "Test CLI"
`
	configPath := filepath.Join(agentDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	err := runPackage(agentDir, "auto", "1.0.0")
	if err == nil {
		t.Errorf("expected error when no binaries found")
	}

	if !strings.Contains(err.Error(), "no built binaries") {
		t.Errorf("expected 'no built binaries' error, got: %v", err)
	}
}

func TestBumpVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		bumpType string
		expected string
		wantErr  bool
	}{
		{
			name:     "bump patch",
			version:  "1.0.0",
			bumpType: "patch",
			expected: "1.0.1",
			wantErr:  false,
		},
		{
			name:     "bump minor",
			version:  "1.0.0",
			bumpType: "minor",
			expected: "1.1.0",
			wantErr:  false,
		},
		{
			name:     "bump major",
			version:  "1.0.0",
			bumpType: "major",
			expected: "2.0.0",
			wantErr:  false,
		},
		{
			name:     "invalid version",
			version:  "1.0",
			bumpType: "patch",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid bump type",
			version:  "1.0.0",
			bumpType: "invalid",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bumpVersion(tt.version, tt.bumpType)
			if (err != nil) != tt.wantErr {
				t.Errorf("bumpVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("bumpVersion() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestUpdateChangelog(t *testing.T) {
	tmpDir := t.TempDir()

	path := filepath.Join(tmpDir, "CHANGELOG.md")

	// Test creating new CHANGELOG
	err := updateChangelog(path, "1.0.0", "My Agent", "Test agent")
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "# Changelog") {
		t.Errorf("CHANGELOG missing header")
	}

	if !strings.Contains(contentStr, "## [1.0.0]") {
		t.Errorf("CHANGELOG missing version section")
	}

	// Test updating existing CHANGELOG
	err = updateChangelog(path, "1.1.0", "My Agent", "Test agent")
	if err != nil {
		t.Fatal(err)
	}

	content, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	contentStr = string(content)
	if !strings.Contains(contentStr, "## [1.1.0]") {
		t.Errorf("CHANGELOG missing new version section")
	}

	if !strings.Contains(contentStr, "## [1.0.0]") {
		t.Errorf("CHANGELOG missing previous version section")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()

	config := &types.Config{
		Agent: types.AgentConfig{
			Name:        "test",
			Description: "test",
			Version:     "1.0.0",
			Model:       "gpt-4",
		},
		CLI: types.CLIConfig{
			Mode:        "freeform",
			Description: "test",
		},
	}

	err := saveConfig(tmpDir, config)
	if err != nil {
		t.Fatal(err)
	}

	loadedConfig, _, err := build.LoadConfigFromDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if loadedConfig.Agent.Version != "1.0.0" {
		t.Errorf("version not saved correctly: %s", loadedConfig.Agent.Version)
	}
}
