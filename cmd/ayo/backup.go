package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/sync"
)

func newBackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Manage backups of sandbox state",
		Long: `Manage backups of sandbox state, config, and data.

Backups include:
- Sandbox state (agent homes, shared files)
- Config (~/.config/ayo/)
- Data (~/.local/share/ayo/ except sandbox and backups)

Examples:
  ayo backup                      Create timestamped backup
  ayo backup --name pre-upgrade   Create named backup
  ayo backup list                 List all backups
  ayo backup show my-backup       Show backup details
  ayo backup restore my-backup    Restore from backup
  ayo backup export my-backup .   Export to portable archive
  ayo backup import backup.tar.gz Import from archive
  ayo backup prune                Clean old auto-backups`,
	}

	cmd.AddCommand(newBackupCreateCmd())
	cmd.AddCommand(newBackupListCmd())
	cmd.AddCommand(newBackupShowCmd())
	cmd.AddCommand(newBackupRestoreCmd())
	cmd.AddCommand(newBackupExportCmd())
	cmd.AddCommand(newBackupImportCmd())
	cmd.AddCommand(newBackupPruneCmd())

	return cmd
}

func newBackupCreateCmd() *cobra.Command {
	var name string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new backup",
		RunE: func(cmd *cobra.Command, args []string) error {
			backup, err := sync.CreateBackup(name)
			if err != nil {
				return fmt.Errorf("create backup: %w", err)
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(backup)
			}

			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Println(successStyle.Render(fmt.Sprintf("Created backup: %s", backup.Name)))
			fmt.Printf("  Path: %s\n", backup.Path)
			fmt.Printf("  Size: %s\n", humanize.Bytes(uint64(backup.Size)))
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "backup name (default: timestamp)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	return cmd
}

func newBackupListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all backups",
		RunE: func(cmd *cobra.Command, args []string) error {
			backups, err := sync.ListBackups()
			if err != nil {
				return fmt.Errorf("list backups: %w", err)
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(backups)
			}

			if len(backups) == 0 {
				dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
				fmt.Println(dimStyle.Render("No backups found. Use 'ayo backup create' to create one."))
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tTYPE\tSIZE\tCREATED")
			for _, b := range backups {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					b.Name,
					b.Type,
					humanize.Bytes(uint64(b.Size)),
					humanize.Time(b.CreatedAt),
				)
			}
			return w.Flush()
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	return cmd
}

func newBackupShowCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show backup details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			backup, err := sync.GetBackup(name)
			if err != nil {
				return err
			}

			manifest, err := sync.GetManifest(backup.Path)
			if err != nil {
				return fmt.Errorf("read manifest: %w", err)
			}

			if jsonOutput {
				result := map[string]any{
					"backup":   backup,
					"manifest": manifest,
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			fmt.Printf("Backup: %s\n", backup.Name)
			fmt.Printf("  Path:    %s\n", backup.Path)
			fmt.Printf("  Type:    %s\n", backup.Type)
			fmt.Printf("  Size:    %s\n", humanize.Bytes(uint64(backup.Size)))
			fmt.Printf("  Created: %s\n", backup.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Println()
			fmt.Printf("Manifest:\n")
			fmt.Printf("  Version: %s\n", manifest.Version)
			fmt.Printf("  Machine: %s\n", manifest.Machine)
			fmt.Printf("  Files:   %d\n", manifest.Files)
			fmt.Printf("  Size:    %s (uncompressed)\n", humanize.Bytes(uint64(manifest.Size)))
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	return cmd
}

func newBackupRestoreCmd() *cobra.Command {
	var noSafety bool

	cmd := &cobra.Command{
		Use:   "restore <name>",
		Short: "Restore from a backup",
		Long: `Restore sandbox state, config, and data from a backup.

By default, creates a safety backup before restoring.
Use --no-safety to skip creating a safety backup.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if err := sync.RestoreBackup(name, !noSafety); err != nil {
				return fmt.Errorf("restore backup: %w", err)
			}

			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Println(successStyle.Render(fmt.Sprintf("Restored from backup: %s", name)))
			return nil
		},
	}

	cmd.Flags().BoolVar(&noSafety, "no-safety", false, "skip creating a safety backup before restore")

	return cmd
}

func newBackupExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <name> <path>",
		Short: "Export backup to portable archive",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			destPath := args[1]

			if err := sync.ExportBackup(name, destPath); err != nil {
				return fmt.Errorf("export backup: %w", err)
			}

			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Println(successStyle.Render(fmt.Sprintf("Exported to: %s", destPath)))
			return nil
		},
	}

	return cmd
}

func newBackupImportCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "import <path>",
		Short: "Import backup from external archive",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			srcPath := args[0]

			backup, err := sync.ImportBackup(srcPath)
			if err != nil {
				return fmt.Errorf("import backup: %w", err)
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(backup)
			}

			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Println(successStyle.Render(fmt.Sprintf("Imported backup: %s", backup.Name)))
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	return cmd
}

func newBackupPruneCmd() *cobra.Command {
	var keepCount int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Clean old auto-backups",
		RunE: func(cmd *cobra.Command, args []string) error {
			removed, err := sync.PruneAutoBackups(keepCount)
			if err != nil {
				return fmt.Errorf("prune backups: %w", err)
			}

			if jsonOutput {
				result := map[string]int{"removed": removed}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			if removed == 0 {
				dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
				fmt.Println(dimStyle.Render("No backups to prune."))
			} else {
				successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
				fmt.Println(successStyle.Render(fmt.Sprintf("Pruned %d auto-backup(s)", removed)))
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&keepCount, "keep", 3, "number of auto-backups to keep")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	return cmd
}
