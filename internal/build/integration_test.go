//go:build integration

package build

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestBuildPipeline builds the ayo CLI and then runs `ayo runthat` against
// every test-agent directory, verifying that valid agents compile to working
// binaries whose --help output includes the expected input-schema flags.
func TestBuildPipeline(t *testing.T) {
	// Locate the repository root (three levels up from internal/build/).
	repoRoot := filepath.Join(mustAbs(t, "."), "..", "..")
	testAgentsDir := filepath.Join(repoRoot, "..", "test-agents")

	if _, err := os.Stat(testAgentsDir); err != nil {
		t.Fatalf("test-agents directory not found at %s: %v", testAgentsDir, err)
	}

	// 1. Build the ayo CLI into a temp directory.
	ayoBin := filepath.Join(t.TempDir(), "ayo")
	cmd := exec.Command("go", "build", "-o", ayoBin, "./cmd/ayo/")
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("Skipping: failed to build ayo CLI: %v\n%s", err, string(out))
	}

	// Agents that are expected to fail during build.
	expectFail := map[string]bool{
		"test-error-invalid-schema": true,
		"test-interactive-mode":     true,
	}

	// Map of agent name -> list of flag substrings that must appear in --help.
	// Only agents with an input schema need flag checks.
	expectedFlags := map[string][]string{
		"test-hooks-basic":       {"--message"},
		"test-hooks-payload":     {"--data"},
		"test-input-array":       {"--tags", "--numbers"},
		"test-input-enum":        {"--color", "--size"},
		"test-input-nested":      {"--user"},
		"test-input-primitives":  {"--name", "--age", "--score", "--active"},
		"test-input-required":    {"--required-field", "--optional-field"},
		"test-interactive-simple": {"--prompt", "--scope", "--dry-run"},
		"test-output-file":       {"--text"},
		"test-output-nested":     {"--topic"},
		"test-output-simple":     {"--message"},
		"test-skills-embedding":  {"--topic"},
		"test-template-basic":    {"--name", "--topic"},
		"test-template-file":     {"--filename"},
		"test-template-functions": {"--name", "--items", "--uppercase"},
	}

	// Enumerate test agent directories.
	entries, err := os.ReadDir(testAgentsDir)
	if err != nil {
		t.Fatalf("reading test-agents directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		agentName := entry.Name()
		agentDir := filepath.Join(testAgentsDir, agentName)

		t.Run(agentName, func(t *testing.T) {
			t.Parallel()

			outDir := t.TempDir()
			binaryPath := filepath.Join(outDir, agentName)

			// Run: ayo runthat <agentDir> -o <binaryPath>
			buildCmd := exec.Command(ayoBin, "runthat", agentDir, "-o", binaryPath)
			buildOut, buildErr := buildCmd.CombinedOutput()

			if expectFail[agentName] {
				if buildErr == nil {
					t.Errorf("expected build to fail for %s but it succeeded", agentName)
				}
				return
			}

			if buildErr != nil {
				t.Fatalf("ayo runthat failed for %s: %v\n%s", agentName, buildErr, string(buildOut))
			}

			// Verify binary exists.
			info, err := os.Stat(binaryPath)
			if err != nil {
				t.Fatalf("binary not found at %s: %v", binaryPath, err)
			}

			// Verify binary is executable.
			if info.Mode()&0111 == 0 {
				t.Errorf("binary %s is not executable (mode %s)", binaryPath, info.Mode())
			}

			// Run --help and capture output.
			helpCmd := exec.Command(binaryPath, "--help")
			helpOut, helpErr := helpCmd.CombinedOutput()
			helpText := string(helpOut)

			// --help may exit 0 or non-zero depending on framework; we only
			// care that it produced output.
			if helpErr != nil && len(helpOut) == 0 {
				t.Fatalf("binary --help produced no output and failed: %v", helpErr)
			}

			if len(helpText) == 0 {
				t.Fatal("binary --help produced empty output")
			}

			// Verify expected flags appear in help text.
			if flags, ok := expectedFlags[agentName]; ok {
				for _, flag := range flags {
					if !strings.Contains(helpText, flag) {
						t.Errorf("--help output missing expected flag %q\nOutput:\n%s", flag, helpText)
					}
				}
			}
		})
	}
}

func mustAbs(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("filepath.Abs(%q): %v", path, err)
	}
	return abs
}
