package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/squads"
)

func newMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate configurations to new formats",
		Long: `Migrate existing configurations to newer formats.

Available migrations:
  squad    Migrate SQUAD.md frontmatter to ayo.json
  squads   Migrate all squads at once`,
	}

	cmd.AddCommand(newMigrateSquadCmd())
	cmd.AddCommand(newMigrateSquadsCmd())

	return cmd
}

func newMigrateSquadCmd() *cobra.Command {
	var dryRun bool
	var force bool

	cmd := &cobra.Command{
		Use:   "squad <name>",
		Short: "Migrate a squad from SQUAD.md frontmatter to ayo.json",
		Long: `Migrate a squad's configuration from SQUAD.md frontmatter to ayo.json.

This migration:
1. Parses YAML frontmatter from SQUAD.md
2. Creates ayo.json with the configuration
3. Strips frontmatter from SQUAD.md (keeps markdown body)

The migration is idempotent - running it multiple times is safe.

Examples:
  ayo migrate squad auth-team
  ayo migrate squad auth-team --dry-run
  ayo migrate squad auth-team --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			squadName := args[0]

			opts := squads.MigrateOptions{
				DryRun: dryRun,
				Force:  force,
			}

			result := squads.MigrateSquadConfig(squadName, opts)

			if globalOutput.JSON {
				globalOutput.PrintData(result, "")
				return nil
			}

			if result.Error != nil {
				return fmt.Errorf("migration failed: %w", result.Error)
			}

			if result.Skipped {
				if !globalOutput.Quiet {
					fmt.Printf("Squad '%s' already migrated or has no frontmatter\n", squadName)
				}
				return nil
			}

			if dryRun {
				fmt.Printf("Would migrate squad '%s':\n", squadName)
				fmt.Printf("  Create: %s\n", result.AyoJSONPath)
				fmt.Printf("  Update: %s (strip frontmatter)\n", result.SquadMDPath)
				return nil
			}

			if result.Migrated {
				fmt.Printf("✓ Migrated squad '%s'\n", squadName)
				fmt.Printf("  Created: %s\n", result.AyoJSONPath)
				fmt.Printf("  Updated: %s\n", result.SquadMDPath)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would change without modifying files")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing ayo.json")

	return cmd
}

func newMigrateSquadsCmd() *cobra.Command {
	var dryRun bool
	var force bool

	cmd := &cobra.Command{
		Use:   "squads",
		Short: "Migrate all squads from SQUAD.md frontmatter to ayo.json",
		Long: `Migrate all squads' configuration from SQUAD.md frontmatter to ayo.json.

Examples:
  ayo migrate squads
  ayo migrate squads --dry-run
  ayo migrate squads --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := squads.MigrateOptions{
				DryRun: dryRun,
				Force:  force,
			}

			results, err := squads.MigrateAllSquads(opts)
			if err != nil {
				return err
			}

			if globalOutput.JSON {
				globalOutput.PrintData(results, "")
				return nil
			}

			if len(results) == 0 {
				if !globalOutput.Quiet {
					fmt.Println("No squads found")
				}
				return nil
			}

			var migrated, skipped, failed int
			for _, r := range results {
				if r.Error != nil {
					failed++
					if !globalOutput.Quiet {
						fmt.Printf("✗ %s: %v\n", r.SquadName, r.Error)
					}
				} else if r.Skipped {
					skipped++
					if !globalOutput.Quiet {
						fmt.Printf("- %s (skipped)\n", r.SquadName)
					}
				} else if r.Migrated {
					migrated++
					if !globalOutput.Quiet {
						fmt.Printf("✓ %s\n", r.SquadName)
					}
				} else if dryRun {
					migrated++
					if !globalOutput.Quiet {
						fmt.Printf("? %s (would migrate)\n", r.SquadName)
					}
				}
			}

			if !globalOutput.Quiet {
				fmt.Printf("\nMigration complete: %d migrated, %d skipped, %d failed\n",
					migrated, skipped, failed)
			}

			if failed > 0 {
				return fmt.Errorf("%d squad(s) failed to migrate", failed)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would change without modifying files")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing ayo.json")

	return cmd
}
