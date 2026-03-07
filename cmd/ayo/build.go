package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/build"
	"github.com/alexcabrera/ayo/internal/build/types"
)

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
  ayo build ./myteam --team
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

	// Create temporary build directory
	tmpDir, err := os.MkdirTemp("", "ayo-build-*")
	if err != nil {
		return fmt.Errorf("create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Generate main.go stub
	mainGoPath := filepath.Join(tmpDir, "main.go")
	if err := generateMainStub(mainGoPath, config, configPath); err != nil {
		return fmt.Errorf("generate main stub: %w", err)
	}

	// Copy config file to embed
	configDest := filepath.Join(tmpDir, "config.toml")
	if err := copyFile(configPath, configDest); err != nil {
		return fmt.Errorf("copy config: %w", err)
	}

	// Run go build
	buildCmd := exec.Command("go", "build", "-o", outputPath, mainGoPath)
	buildCmd.Dir = tmpDir
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
	// This is a simplified stub - in reality, this would be more sophisticated
	// and would need to properly initialize the Fantasy framework, load tools, etc.

	stub := fmt.Sprintf(`package main

import (
	_ "embed"
	"fmt"
)

//go:embed config.toml
var configToml []byte

func main() {
	// Placeholder: this would initialize the agent from embedded config
	// and execute with the CLI arguments

	fmt.Printf("Agent: %%s\n", "%s")
	fmt.Printf("Description: %%s\n", "%s")
	fmt.Printf("Config size: %%d bytes\n", len(configToml))

	// TODO: Implement actual agent execution
	// - Parse embedded config
	// - Set up Fantasy agent
	// - Parse CLI flags according to config
	// - Execute agent with parsed input
	// - Validate output against schema
}
`, config.Agent.Name, config.Agent.Description)

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
