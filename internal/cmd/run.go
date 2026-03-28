package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/ayo-ooo/ayo/internal/registry"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <agent> [args...]",
	Short: "Run a registered agent by name",
	Long: `Run a registered agent by name, passing through all arguments.

This is a convenience wrapper — it looks up the agent's binary path
in the registry and executes it with the given arguments.

Examples:
  ayo run my-agent '{"text": "hello"}'
  ayo run translator --text "hello world"
  echo '{"text":"hello"}' | ayo run summarize`,
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runRegisteredAgent(args[0], args[1:]); err != nil {
			exitError(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runRegisteredAgent(name string, args []string) error {
	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("loading registry: %w", err)
	}

	entry := reg.Get(name)
	if entry == nil {
		return fmt.Errorf("agent '%s' not found in registry\n\nRun 'ayo list' to see registered agents", name)
	}

	if entry.BinaryPath == "" {
		return fmt.Errorf("agent '%s' has no binary path — rebuild with 'ayo runthat --register'", name)
	}

	if _, err := os.Stat(entry.BinaryPath); err != nil {
		return fmt.Errorf("binary not found at %s — rebuild with 'ayo runthat --register'", entry.BinaryPath)
	}

	// Execute the agent binary, passing through all args and stdio
	execCmd := exec.Command(entry.BinaryPath, args...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("running agent: %w", err)
	}

	return nil
}
