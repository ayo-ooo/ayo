package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/alexcabrera/ayo/internal/audit"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "View file modification audit logs",
	Long:  "View and export audit logs of file modifications made by agents.",
}

var auditListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent audit entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		agent, _ := cmd.Flags().GetString("agent")
		session, _ := cmd.Flags().GetString("session")
		action, _ := cmd.Flags().GetString("action")
		limit, _ := cmd.Flags().GetInt("limit")
		sinceStr, _ := cmd.Flags().GetString("since")

		filter := audit.Filter{
			Agent:   agent,
			Session: session,
			Action:  action,
			Limit:   limit,
		}

		if sinceStr != "" {
			since, err := time.Parse(time.RFC3339, sinceStr)
			if err != nil {
				// Try relative duration
				dur, err := time.ParseDuration(sinceStr)
				if err != nil {
					return fmt.Errorf("invalid since format: %s", sinceStr)
				}
				filter.Since = time.Now().Add(-dur)
			} else {
				filter.Since = since
			}
		}

		logger, err := audit.NewFileLogger()
		if err != nil {
			return fmt.Errorf("open audit log: %w", err)
		}
		defer logger.Close()

		entries, err := logger.Query(filter)
		if err != nil {
			return fmt.Errorf("query audit log: %w", err)
		}

		if globalOutput.JSON {
			return json.NewEncoder(os.Stdout).Encode(entries)
		}

		if len(entries) == 0 {
			fmt.Println("No audit entries found.")
			return nil
		}

		// Styles
		headerStyle := lipgloss.NewStyle().Bold(true)
		timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		agentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
		actionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))

		for _, entry := range entries {
			ts := timeStyle.Render(entry.Timestamp.Format("2006-01-02 15:04:05"))
			agent := agentStyle.Render(entry.Agent)
			action := actionStyle.Render(entry.Action)
			fmt.Printf("%s  %s  %s  %s\n", ts, agent, action, entry.Path)
			if !globalOutput.Quiet {
				fmt.Printf("         %s: %s\n", headerStyle.Render("Approval"), entry.Approval)
				if entry.Size > 0 {
					fmt.Printf("         %s: %d bytes\n", headerStyle.Render("Size"), entry.Size)
				}
				fmt.Println()
			}
		}

		return nil
	},
}

var auditExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export audit entries to file",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")
		agent, _ := cmd.Flags().GetString("agent")
		session, _ := cmd.Flags().GetString("session")

		filter := audit.Filter{
			Agent:   agent,
			Session: session,
		}

		logger, err := audit.NewFileLogger()
		if err != nil {
			return fmt.Errorf("open audit log: %w", err)
		}
		defer logger.Close()

		entries, err := logger.Query(filter)
		if err != nil {
			return fmt.Errorf("query audit log: %w", err)
		}

		switch format {
		case "json":
			return json.NewEncoder(os.Stdout).Encode(entries)
		case "csv":
			return exportCSV(entries)
		default:
			return fmt.Errorf("unknown format: %s (use 'json' or 'csv')", format)
		}
	},
}

func exportCSV(entries []audit.Entry) error {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	// Header
	w.Write([]string{"timestamp", "agent", "session", "action", "path", "approval", "size", "hash"})

	// Data
	for _, e := range entries {
		w.Write([]string{
			e.Timestamp.Format(time.RFC3339),
			e.Agent,
			e.Session,
			e.Action,
			e.Path,
			e.Approval,
			fmt.Sprintf("%d", e.Size),
			e.ContentHash,
		})
	}

	return w.Error()
}

func init() {
	// List command flags
	auditListCmd.Flags().String("agent", "", "Filter by agent (e.g., @ayo)")
	auditListCmd.Flags().String("session", "", "Filter by session ID")
	auditListCmd.Flags().String("action", "", "Filter by action (create, update, delete)")
	auditListCmd.Flags().Int("limit", 50, "Maximum entries to show")
	auditListCmd.Flags().String("since", "", "Show entries since (RFC3339 or duration like '1h')")

	// Export command flags
	auditExportCmd.Flags().String("format", "json", "Export format (json or csv)")
	auditExportCmd.Flags().String("agent", "", "Filter by agent")
	auditExportCmd.Flags().String("session", "", "Filter by session ID")

	auditCmd.AddCommand(auditListCmd)
	auditCmd.AddCommand(auditExportCmd)
}
