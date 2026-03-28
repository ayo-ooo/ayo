package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ayo-ooo/ayo/internal/project"
	"github.com/ayo-ooo/ayo/internal/registry"
	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register <path>",
	Short: "Register an agent with the ayo registry",
	Long: `Register an agent project directory or compiled binary with the ayo registry.

If the path is a directory containing config.toml, ayo registers it as a project.
If the path is a binary, ayo invokes it with --ayo-describe to extract metadata.

Examples:
  ayo register ./my-agent          Register a project directory
  ayo register ./my-agent-binary   Register a compiled binary`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := registerAgent(args[0]); err != nil {
			exitError(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
}

func registerAgent(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}

	var entry registry.Entry

	if info.IsDir() {
		// It's a project directory — parse it
		proj, err := project.ParseProject(absPath)
		if err != nil {
			return fmt.Errorf("parsing project: %w", err)
		}

		agentType := "conversational"
		if proj.Input != nil {
			agentType = "tool"
		}

		entry = registry.Entry{
			Name:        proj.Config.Name,
			Version:     proj.Config.Version,
			Description: proj.Config.Description,
			SourcePath:  absPath,
			Type:        agentType,
		}

		// Check if a compiled binary exists alongside
		binaryPath := filepath.Join(filepath.Dir(absPath), proj.Config.Name)
		if _, err := os.Stat(binaryPath); err == nil {
			entry.BinaryPath = binaryPath
		}
	} else {
		// It's a binary — invoke --ayo-describe
		entry, err = describeFromBinary(absPath)
		if err != nil {
			return fmt.Errorf("reading agent metadata: %w", err)
		}
		entry.BinaryPath = absPath
	}

	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("loading registry: %w", err)
	}

	reg.Register(entry)

	if err := reg.Save(); err != nil {
		return fmt.Errorf("saving registry: %w", err)
	}

	printSuccess(fmt.Sprintf("Registered '%s' (%s, %s)", entry.Name, entry.Type, entry.Version))
	return nil
}

// describeFromBinary invokes a binary with --ayo-describe and parses its metadata.
func describeFromBinary(binaryPath string) (registry.Entry, error) {
	cmd := exec.Command(binaryPath, "--ayo-describe")
	output, err := cmd.Output()
	if err != nil {
		return registry.Entry{}, fmt.Errorf("running %s --ayo-describe: %w (is this an ayo agent?)", filepath.Base(binaryPath), err)
	}

	var meta struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		Description string `json:"description"`
		Type        string `json:"type"`
	}
	if err := json.Unmarshal(output, &meta); err != nil {
		return registry.Entry{}, fmt.Errorf("parsing metadata: %w", err)
	}

	return registry.Entry{
		Name:        meta.Name,
		Version:     meta.Version,
		Description: meta.Description,
		Type:        meta.Type,
		BinaryPath:  binaryPath,
	}, nil
}

// RegisterFromBuild registers an agent after a successful build.
// Used by runthat --register.
func RegisterFromBuild(proj *project.Project, binaryPath string) error {
	agentType := "conversational"
	if proj.Input != nil {
		agentType = "tool"
	}

	absBinary, err := filepath.Abs(binaryPath)
	if err != nil {
		return fmt.Errorf("resolving binary path: %w", err)
	}

	absSource, err := filepath.Abs(proj.Path)
	if err != nil {
		return fmt.Errorf("resolving source path: %w", err)
	}

	entry := registry.Entry{
		Name:        proj.Config.Name,
		Version:     proj.Config.Version,
		Description: proj.Config.Description,
		SourcePath:  absSource,
		BinaryPath:  absBinary,
		Type:        agentType,
	}

	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("loading registry: %w", err)
	}

	reg.Register(entry)

	if err := reg.Save(); err != nil {
		return fmt.Errorf("saving registry: %w", err)
	}

	return nil
}
