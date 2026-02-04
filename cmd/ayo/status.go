package main

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/daemon"
	"github.com/alexcabrera/ayo/internal/version"
)

func newStatusCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show ayo daemon and service status",
		Long:  "Display the current status of the ayo daemon, sandbox pool, and related services.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Styles
			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
			okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
			errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
			labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(20)
			valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

			status := func(name string, ok bool, value string) {
				indicator := okStyle.Render("●")
				if !ok {
					indicator = errStyle.Render("●")
				}
				fmt.Printf("  %s %s %s\n", indicator, labelStyle.Render(name), valueStyle.Render(value))
			}

			warn := func(name string, value string) {
				indicator := warnStyle.Render("●")
				fmt.Printf("  %s %s %s\n", indicator, labelStyle.Render(name), valueStyle.Render(value))
			}

			fmt.Println()
			fmt.Println(headerStyle.Render("  Ayo Status"))
			fmt.Println()

			// CLI version
			fmt.Printf("  %s %s\n", labelStyle.Render("CLI Version:"), valueStyle.Render(version.Version))
			fmt.Println()

			// Daemon status
			fmt.Println(headerStyle.Render("  Daemon"))

			daemonRunning := daemon.IsDaemonRunning()
			if !daemonRunning {
				status("Status:", false, "not running")
				fmt.Println()
				fmt.Println("  Run 'ayo daemon start' to start the daemon")
				fmt.Println()
				return nil
			}

			// Try to connect and get detailed status
			client := daemon.NewClient()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			if err := client.Connect(ctx); err != nil {
				warn("Status:", "running but cannot connect")
				fmt.Println()
				return nil
			}
			defer client.Close()

			daemonStatus, err := client.Status(ctx)
			if err != nil {
				warn("Status:", "running but cannot get status")
				fmt.Println()
				return nil
			}

			status("Status:", true, "running")
			fmt.Printf("  %s %s\n", labelStyle.Render("Version:"), valueStyle.Render(daemonStatus.Version))
			fmt.Printf("  %s %s\n", labelStyle.Render("PID:"), valueStyle.Render(fmt.Sprintf("%d", daemonStatus.PID)))
			fmt.Printf("  %s %s\n", labelStyle.Render("Uptime:"), valueStyle.Render(formatUptime(time.Duration(daemonStatus.Uptime)*time.Second)))
			fmt.Printf("  %s %s\n", labelStyle.Render("Memory:"), valueStyle.Render(formatBytes(daemonStatus.MemoryUsage)))
			fmt.Println()

			// Sandbox pool status
			fmt.Println(headerStyle.Render("  Sandbox Pool"))
			fmt.Printf("  %s %s\n", labelStyle.Render("Total:"), valueStyle.Render(fmt.Sprintf("%d", daemonStatus.Sandboxes.Total)))
			fmt.Printf("  %s %s\n", labelStyle.Render("Idle:"), valueStyle.Render(fmt.Sprintf("%d", daemonStatus.Sandboxes.Idle)))
			fmt.Printf("  %s %s\n", labelStyle.Render("In Use:"), valueStyle.Render(fmt.Sprintf("%d", daemonStatus.Sandboxes.InUse)))
			fmt.Println()

			return nil
		},
	}

	return cmd
}

func formatUptime(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
