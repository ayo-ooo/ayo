package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Integration tests for complete ayo workflow
// These tests are skipped by default and require a full ayo build

func TestFullWorkflow(t *testing.T) {
	// Skip in normal test runs - requires full ayo binary
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build ayo
	ayoPath, err := buildAyo()
	if err != nil {
		t.Skipf("Failed to build ayo: %v", err)
	}

	tmpDir := t.TempDir()

	// Step 1: Create a new agent
	t.Run("CreateAgent", func(t *testing.T) {
		cmd := exec.Command(ayoPath, "fresh", "test-agent")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("fresh command failed: %v\nOutput: %s", err, output)
		}

		agentDir := filepath.Join(tmpDir, "test-agent")
		if _, err := os.Stat(agentDir); os.IsNotExist(err) {
			t.Fatalf("agent directory not created")
		}

		configPath := filepath.Join(agentDir, "config.toml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Fatalf("config.toml not created")
		}
	})

	agentDir := filepath.Join(tmpDir, "test-agent")

	// Step 2: Validate configuration
	t.Run("ValidateConfig", func(t *testing.T) {
		cmd := exec.Command(ayoPath, "checkit", agentDir)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("checkit command failed: %v\nOutput: %s", err, output)
		}
	})

	// Step 3: Build the agent
	t.Run("BuildAgent", func(t *testing.T) {
		cmd := exec.Command(ayoPath, "build", agentDir)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("build command failed: %v\nOutput: %s", err, output)
		}

		binPath := filepath.Join(agentDir, ".build", "bin", "test-agent")
		if _, err := os.Stat(binPath); os.IsNotExist(err) {
			t.Fatalf("binary not created at %s", binPath)
		}
	})

	// Step 4: Version bump
	t.Run("VersionBump", func(t *testing.T) {
		cmd := exec.Command(ayoPath, "release", agentDir, "--bump", "patch")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("release command failed: %v\nOutput: %s", err, output)
		}
	})

	// Step 5: Build for distribution
	t.Run("BuildAll", func(t *testing.T) {
		cmd := exec.Command(ayoPath, "build", agentDir, "--all")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("build --all command failed: %v\nOutput: %s", err, output)
		}

		distDir := filepath.Join(agentDir, "dist")
		if _, err := os.Stat(distDir); os.IsNotExist(err) {
			t.Fatalf("dist directory not created")
		}
	})

	// Step 6: Package
	t.Run("PackageAgent", func(t *testing.T) {
		cmd := exec.Command(ayoPath, "package", agentDir, "--version", "0.0.1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("package command failed: %v\nOutput: %s", err, output)
		}

		releasesDir := filepath.Join(agentDir, "releases")
		if _, err := os.Stat(releasesDir); os.IsNotExist(err) {
			t.Fatalf("releases directory not created")
		}

		// Check for checksum file
		checksumsPath := filepath.Join(releasesDir, "test-agent-0.0.1.sha256")
		if _, err := os.Stat(checksumsPath); os.IsNotExist(err) {
			t.Fatalf("checksums file not created")
		}
	})

	// Step 7: Clean
	t.Run("CleanBuild", func(t *testing.T) {
		cmd := exec.Command(ayoPath, "clean", agentDir)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("clean command failed: %v\nOutput: %s", err, output)
		}

		buildDir := filepath.Join(agentDir, ".build")
		if _, err := os.Stat(buildDir); !os.IsNotExist(err) {
			t.Logf("warning: .build directory still exists after clean")
		}
	})
}

func TestHelpSystem(t *testing.T) {
	ayoPath := t.TempDir() + "/ayo"
	// Use the pre-built ayo if available, or skip
	if _, err := os.Stat("/tmp/ayo-dev"); err == nil {
		ayoPath = "/tmp/ayo-dev"
	} else {
		t.Skip("No ayo binary available")
	}

	// Test root help
	t.Run("RootHelp", func(t *testing.T) {
		cmd := exec.Command(ayoPath, "--help")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("--help failed: %v", err)
		}

		if !bytes.Contains(output, []byte("Build system")) {
			t.Error("help output doesn't contain expected text")
		}
	})

	// Test build help
	t.Run("BuildHelp", func(t *testing.T) {
		cmd := exec.Command(ayoPath, "build", "--help")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("build --help failed: %v", err)
		}

		if !bytes.Contains(output, []byte("Build a standalone executable")) {
			t.Error("build help doesn't contain expected text")
		}
	})

	// Test package help
	t.Run("PackageHelp", func(t *testing.T) {
		cmd := exec.Command(ayoPath, "package", "--help")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("package --help failed: %v", err)
		}

		if !bytes.Contains(output, []byte("distributable archives")) {
			t.Error("package help doesn't contain expected text")
		}
	})

	// Test dev help
	t.Run("DevHelp", func(t *testing.T) {
		cmd := exec.Command(ayoPath, "dev", "--help")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("dev --help failed: %v", err)
		}

		if !bytes.Contains(output, []byte("automatic rebuilds")) {
			t.Error("dev help doesn't contain expected text")
		}
	})

	// Test release help
	t.Run("ReleaseHelp", func(t *testing.T) {
		cmd := exec.Command(ayoPath, "release", "--help")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("release --help failed: %v", err)
		}

		if !bytes.Contains(output, []byte("version")) {
			t.Error("release help doesn't contain expected text")
		}
	})
}

func TestExamplesBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ayoPath, err := buildAyo()
	if err != nil {
		t.Skipf("Failed to build ayo: %v", err)
	}

	// Test building qa-bot example
	t.Run("QABotExample", func(t *testing.T) {
		examplesDir := getExamplesDir()
		if examplesDir == "" {
			t.Skip("Examples directory not found")
		}

		qaBotDir := filepath.Join(examplesDir, "qa-bot")
		if _, err := os.Stat(qaBotDir); os.IsNotExist(err) {
			t.Skip("qa-bot example not found")
		}

		cmd := exec.Command(ayoPath, "build", qaBotDir)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("failed to build qa-bot: %v\nOutput: %s", err, output)
		}
	})

	// Test building file-processor example
	t.Run("FileProcessorExample", func(t *testing.T) {
		examplesDir := getExamplesDir()
		if examplesDir == "" {
			t.Skip("Examples directory not found")
		}

		fileProcessorDir := filepath.Join(examplesDir, "file-processor")
		if _, err := os.Stat(fileProcessorDir); os.IsNotExist(err) {
			t.Skip("file-processor example not found")
		}

		cmd := exec.Command(ayoPath, "build", fileProcessorDir)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("failed to build file-processor: %v\nOutput: %s", err, output)
		}
	})
}

func buildAyo() (string, error) {
	tmpDir := os.TempDir()
	ayoPath := filepath.Join(tmpDir, "ayo-test")

	cmd := exec.Command("go", "build", "-o", ayoPath, "./cmd/ayo")
	cmd.Dir = getModuleRoot()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("build failed: %v\nOutput: %s", err, output)
	}

	return ayoPath, nil
}

func getModuleRoot() string {
	// Try to find the module root
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Walk up to find go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return ""
}

func getExamplesDir() string {
	moduleRoot := getModuleRoot()
	if moduleRoot == "" {
		return ""
	}
	return filepath.Join(moduleRoot, "examples")
}
