package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/build"
	"github.com/alexcabrera/ayo/internal/build/types"
)

// ModuleRoot is set at build time via ldflags
var ModuleRoot string

func init() {
	// Set default module root if not provided via ldflags
	if ModuleRoot == "" {
		if dir, err := findModuleRoot(); err == nil {
			ModuleRoot = dir
		}
	}
}

func newBuildCmd() *cobra.Command {
	var output string
	var targetOS, targetArch string

	cmd := &cobra.Command{
		Use:   "build [directory]",
		Short: "Build an agent or team executable",
		Long: `Build a standalone executable from a config.toml or team.toml file.

The build process:
1. Reads and validates the configuration
2. Generates a main.go stub
3. Embeds configuration and resources
4. Compiles to a standalone binary

The resulting binary can be distributed and run independently.

Examples:
  ayo build ./myagent
  ayo build ./myteam
  ayo build . --output ./bin/myagent
  ayo build . --target-os linux --target-arch amd64`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]
			return runBuild(dir, output, targetOS, targetArch)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output binary path (default: <agent-name> or <team-name>)")
	cmd.Flags().StringVar(&targetOS, "target-os", runtime.GOOS, "Target OS (default: current OS)")
	cmd.Flags().StringVar(&targetArch, "target-arch", runtime.GOARCH, "Target architecture (default: current arch)")

	return cmd
}

func runBuild(dir, output, targetOS, targetArch string) error {
	// Ensure we have the module root
	if ModuleRoot == "" {
		return fmt.Errorf("module root not found - please build ayo from its source directory or use -ldflags to set ModuleRoot")
	}

	// Resolve directory to absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve directory: %w", err)
	}

	// Load configuration
	config, configPath, err := build.LoadConfigFromDir(absDir)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Determine output name
	if output == "" {
		output = config.Agent.Name
		if targetOS == "windows" {
			output += ".exe"
		}
	}

	// Resolve output to absolute path
	outputPath, err := filepath.Abs(output)
	if err != nil {
		return fmt.Errorf("resolve output path: %w", err)
	}

	// Prepare resources in agent directory for embedding
	// We'll build in-place in the agent directory to avoid module resolution issues
	agentBuildDir := filepath.Join(absDir, ".build")
	if err := os.MkdirAll(agentBuildDir, 0755); err != nil {
		return fmt.Errorf("create build dir: %w", err)
	}
	// defer os.RemoveAll(agentBuildDir)  // Commented out for debugging

	// Create go.mod in .build directory with replace directive to ayo module
	// This allows building standalone while accessing ayo's internal packages
	goModContent := fmt.Sprintf(`module agent

go 1.25.5

require github.com/alexcabrera/ayo v0.0.0

replace github.com/alexcabrera/ayo => %s
`, ModuleRoot)
	if err := os.WriteFile(filepath.Join(agentBuildDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("create go.mod: %w", err)
	}

	// Create go.sum
	_ = os.WriteFile(filepath.Join(agentBuildDir, "go.sum"), []byte(""), 0644)

	// Generate main.go stub in agent's .build directory
	mainGoPath := filepath.Join(agentBuildDir, "main.go")
	if err := generateMainStub(mainGoPath, config, configPath); err != nil {
		return fmt.Errorf("generate main stub: %w", err)
	}

	// Run go mod tidy to resolve all dependencies (now that main.go exists)
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = agentBuildDir
	if outputBytes, err := tidyCmd.CombinedOutput(); err != nil {
		// Non-fatal - continue with what we have
		fmt.Fprintf(os.Stderr, "Warning: go mod tidy failed: %v\n%s\n", err, outputBytes)
	}

	// Copy config file to .build for embedding
	configDest := filepath.Join(agentBuildDir, "config.toml")
	if err := copyFile(configPath, configDest); err != nil {
		return fmt.Errorf("copy config: %w", err)
	}

	// Copy system prompt if exists
	promptsDir := filepath.Join(absDir, "prompts")
	promptDest := filepath.Join(agentBuildDir, "prompts")
	if err := os.MkdirAll(promptDest, 0755); err != nil {
		return fmt.Errorf("create prompts dir: %w", err)
	}
	systemPromptPath := filepath.Join(promptsDir, "system.md")
	if _, err := os.Stat(systemPromptPath); err == nil {
		if err := copyFile(systemPromptPath, filepath.Join(promptDest, "system.md")); err != nil {
			return fmt.Errorf("copy system prompt: %w", err)
		}
	} else {
		// Create empty system.md if not exists
		if err := os.WriteFile(filepath.Join(promptDest, "system.md"), []byte(""), 0644); err != nil {
			return fmt.Errorf("create empty system prompt: %w", err)
		}
	}

	// Copy skills directory if exists
	skillsDir := filepath.Join(absDir, "skills")
	if _, err := os.Stat(skillsDir); err == nil {
		skillsDest := filepath.Join(agentBuildDir, "skills")
		if err := copyDir(skillsDir, skillsDest); err != nil {
			return fmt.Errorf("copy skills: %w", err)
		}
	} else {
		// Create empty skills directory with placeholder file (embed requires at least one file)
		skillsDir := filepath.Join(agentBuildDir, "skills")
		if err := os.MkdirAll(skillsDir, 0755); err != nil {
			return fmt.Errorf("create skills dir: %w", err)
		}
		// Create a placeholder SKILL.md file so embed doesn't fail
		_ = os.WriteFile(filepath.Join(skillsDir, "placeholder.md"), []byte("# Placeholder Skill\n\nThis is a placeholder to satisfy embed requirements.\n"), 0644)
	}

	// Copy tools directory if exists
	toolsDir := filepath.Join(absDir, "tools")
	if _, err := os.Stat(toolsDir); err == nil {
		toolsDest := filepath.Join(agentBuildDir, "tools")
		if err := copyDir(toolsDir, toolsDest); err != nil {
			return fmt.Errorf("copy tools: %w", err)
		}
	} else {
		// Create empty tools directory with placeholder file (embed requires at least one file)
		toolsDir := filepath.Join(agentBuildDir, "tools")
		if err := os.MkdirAll(toolsDir, 0755); err != nil {
			return fmt.Errorf("create tools dir: %w", err)
		}
		// Create a placeholder file so embed doesn't fail
		_ = os.WriteFile(filepath.Join(toolsDir, "placeholder"), []byte("# Placeholder tool file\n"), 0644)
	}

	// Run go build from agent's .build directory
	// Use -modfile to specify the go.mod we created
	buildCmd := exec.Command("go", "build", "-modfile", filepath.Join(agentBuildDir, "go.mod"), "-o", outputPath, mainGoPath)
	buildCmd.Dir = agentBuildDir
	buildCmd.Env = append(os.Environ(),
		fmt.Sprintf("GOOS=%s", targetOS),
		fmt.Sprintf("GOARCH=%s", targetArch),
		"CGO_ENABLED=0",
	)

	// Capture output
	outputBytes, err := buildCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go build failed: %w\nOutput: %s", err, outputBytes)
	}

	// Make executable on Unix systems
	if targetOS != "windows" {
		if err := os.Chmod(outputPath, 0755); err != nil {
			return fmt.Errorf("make executable: %w", err)
		}
	}

	fmt.Printf("Successfully built: %s\n", outputPath)
	return nil
}

// generateMainStub creates the main.go file for the built executable
func generateMainStub(path string, config *types.Config, configPath string) error {
	// Generate complete main.go with embedded resources
	stub := fmt.Sprintf(`package main

import (
	"embed"
	"os"

	"github.com/alexcabrera/ayo/pkg/build/runtime"
)

//go:embed config.toml
var configToml []byte

//go:embed prompts/system.md
var systemPrompt []byte

//go:embed skills/*
var skills embed.FS

//go:embed tools/*
var tools embed.FS

func main() {
	if err := runtime.Execute(configToml, systemPrompt, skills, tools); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
`)

	return os.WriteFile(path, []byte(stub), 0644)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// copyDir recursively copies a directory from src to dst
func copyDir(src, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectories
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy files
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// findModuleRoot finds the root of the ayo Go module by searching for go.mod
func findModuleRoot() (string, error) {
	// Strategy: Try multiple starting points to find the ayo module
	// 1. Search from executable location
	// 2. Search from current working directory
	// 3. Search from user's home directory

	startPaths := []string{}

	// Try executable location
	if execPath, err := os.Executable(); err == nil {
		if absPath, err := filepath.Abs(execPath); err == nil {
			startPaths = append(startPaths, filepath.Dir(absPath))
		}
	}

	// Try current working directory
	if cwd, err := os.Getwd(); err == nil {
		startPaths = append(startPaths, cwd)
	}

	// Try to find from each starting point
	for _, startPath := range startPaths {
		if result, err := searchForModuleFrom(startPath); err == nil {
			return result, nil
		}
	}

	return "", fmt.Errorf("could not find ayo module root (tried searching from executable and working directory)")
}

// searchForModuleFrom searches upward from a starting directory for the ayo module
func searchForModuleFrom(startDir string) (string, error) {
	dir := startDir
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if info, err := os.Stat(goModPath); err == nil && !info.IsDir() {
			// Found go.mod, check if it's for ayo module
			data, err := os.ReadFile(goModPath)
			if err != nil {
				return "", fmt.Errorf("read go.mod: %w", err)
			}
			if strings.Contains(string(data), "module github.com/alexcabrera/ayo") {
				return dir, nil
			}
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding ayo module
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("ayo module not found from %s", startDir)
}

