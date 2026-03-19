package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateProject(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := "test-agent"

	// Change to temp dir to test relative path creation
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	err := createProject(projectName)
	if err != nil {
		t.Fatalf("createProject() error = %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(projectName); os.IsNotExist(err) {
		t.Error("createProject() did not create directory")
	}

	// Verify config.toml was created
	configPath := filepath.Join(projectName, "config.toml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config.toml: %v", err)
	}

	if !bytes.Contains(data, []byte("name = \"test-agent\"")) {
		t.Error("config.toml should contain agent name")
	}

	// Verify system.md was created
	systemPath := filepath.Join(projectName, "system.md")
	data, err = os.ReadFile(systemPath)
	if err != nil {
		t.Fatalf("Failed to read system.md: %v", err)
	}

	if !bytes.Contains(data, []byte("Agent Instructions")) {
		t.Error("system.md should contain template content")
	}

	// Verify .gitignore was created
	gitignorePath := filepath.Join(projectName, ".gitignore")
	data, err = os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("Failed to read .gitignore: %v", err)
	}

	if !bytes.Contains(data, []byte("test-agent")) {
		t.Error(".gitignore should contain agent name")
	}
}

func TestCreateProject_DirectoryExists(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := "existing-agent"

	// Create directory first
	if err := os.MkdirAll(filepath.Join(tmpDir, projectName), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Change to temp dir
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	err := createProject(projectName)
	if err == nil {
		t.Error("createProject() should return error when directory exists")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Error should mention directory exists, got: %v", err)
	}
}

func TestValidateProject_InvalidPath(t *testing.T) {
	err := validateProject("/nonexistent/path")
	if err == nil {
		t.Error("validateProject() should return error for non-existent path")
	}
}

func TestValidateProject_PathIsFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	err := validateProject(tmpFile)
	if err == nil {
		t.Error("validateProject() should return error when path is a file")
	}

	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("Error should mention not a directory, got: %v", err)
	}
}

func TestValidateProject_MissingConfig(t *testing.T) {
	dir := t.TempDir()

	err := validateProject(dir)
	if err == nil {
		t.Error("validateProject() should return error when config.toml is missing")
	}
}

func TestValidateProject_MissingSystem(t *testing.T) {
	dir := t.TempDir()

	// Create config.toml only
	configContent := `[agent]
name = "test"
version = "1.0.0"

[model]

[defaults]`
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config.toml: %v", err)
	}

	err := validateProject(dir)
	if err == nil {
		t.Error("validateProject() should return error when system.md is missing")
	}
}

func TestPrintError(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	printError("test error message")

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "test error message") {
		t.Errorf("printError() output should contain message, got: %s", output)
	}
}

func TestPrintSuccess(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printSuccess("test success message")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "test success message") {
		t.Errorf("printSuccess() output should contain message, got: %s", output)
	}
}

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}
