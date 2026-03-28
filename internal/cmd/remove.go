package cmd

import (
	"fmt"
	"os"

	"github.com/ayo-ooo/ayo/internal/registry"
	"github.com/spf13/cobra"
)

var removeDeleteBinary bool

var removeCmd = &cobra.Command{
	Use:     "remove <agent>",
	Aliases: []string{"rm", "unregister"},
	Short:   "Remove an agent from the registry",
	Long: `Remove an agent from the ayo registry.

This only removes the registry entry. The agent's binary and source
files are not deleted unless --delete-binary is specified.

Examples:
  ayo remove my-agent                Remove from registry only
  ayo remove my-agent --delete-binary  Also delete the binary`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := removeAgent(args[0]); err != nil {
			exitError(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().BoolVar(&removeDeleteBinary, "delete-binary", false, "Also delete the agent binary")
}

func removeAgent(name string) error {
	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("loading registry: %w", err)
	}

	entry := reg.Get(name)
	if entry == nil {
		return fmt.Errorf("agent '%s' not found in registry", name)
	}

	binaryPath := entry.BinaryPath

	if !reg.Remove(name) {
		return fmt.Errorf("failed to remove '%s'", name)
	}

	if err := reg.Save(); err != nil {
		return fmt.Errorf("saving registry: %w", err)
	}

	printSuccess(fmt.Sprintf("Removed '%s' from registry", name))

	if removeDeleteBinary && binaryPath != "" {
		if err := os.Remove(binaryPath); err != nil {
			if !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Warning: could not delete binary: %v\n", err)
			}
		} else {
			printSuccess(fmt.Sprintf("Deleted binary: %s", binaryPath))
		}
	}

	return nil
}
