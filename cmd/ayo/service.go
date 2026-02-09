package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/daemon"
	"github.com/alexcabrera/ayo/internal/version"
)

func newSandboxServiceCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Manage the sandbox service",
		Long:  "Commands for managing the sandbox background service (start, stop, status).",
	}

	cmd.AddCommand(newServiceStartCmd(cfgPath))
	cmd.AddCommand(newServiceStopCmd(cfgPath))
	cmd.AddCommand(newServiceStatusCmd(cfgPath))

	return cmd
}

// newDaemonAliasCmd creates a hidden "daemon" command that redirects to "sandbox service"
// for backwards compatibility.
func newDaemonAliasCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "daemon",
		Hidden: true,
		Short:  "Alias for 'sandbox service' (deprecated)",
	}

	cmd.AddCommand(newServiceStartCmd(cfgPath))
	cmd.AddCommand(newServiceStopCmd(cfgPath))
	cmd.AddCommand(newServiceStatusCmd(cfgPath))

	return cmd
}

func newServiceStartCmd(cfgPath *string) *cobra.Command {
	var foreground bool

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the ayo service",
		Long: `Start the ayo background service.

The service manages:
- Sandbox pool for isolated command execution
- LLM connection pooling
- Memory index caching

By default, the service runs in the background. Use --foreground to run
in the current terminal for debugging.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Check if already running
			if daemon.IsDaemonRunning() {
				client := daemon.NewClient()
				if err := client.Connect(ctx); err == nil {
					defer client.Close()
					if err := client.Ping(ctx); err == nil {
						fmt.Println("Service is already running")
						return nil
					}
				}
			}

			cfg := daemon.DefaultServerConfig()

			server, err := daemon.NewServer(cfg)
			if err != nil {
				return fmt.Errorf("create server: %w", err)
			}

			if foreground {
				return runServiceForeground(ctx, server, cfg.SocketPath)
			}

			// Background mode: start and detach with spinner
			return startServiceBackground(ctx)
		},
	}

	cmd.Flags().BoolVarP(&foreground, "foreground", "f", false, "Run in foreground (for debugging)")

	return cmd
}

func newServiceStopCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the ayo service",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if !daemon.IsDaemonRunning() {
				fmt.Println("Service is not running")
				return nil
			}

			client := daemon.NewClient()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connect to daemon: %w", err)
			}
			defer client.Close()

			if err := client.Shutdown(ctx, true); err != nil {
				return fmt.Errorf("shutdown daemon: %w", err)
			}

			fmt.Println("Service stopped")
			return nil
		},
	}

	return cmd
}

func newServiceStatusCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show service status",
		Long:  "Display the current status of the ayo service, sandbox pool, and related resources.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Check for JSON output early
			if globalOutput.JSON {
				return serviceStatusJSON(ctx)
			}

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
			fmt.Println(headerStyle.Render("  Ayo Service Status"))
			fmt.Println()

			// CLI version
			fmt.Printf("  %s %s\n", labelStyle.Render("CLI Version:"), valueStyle.Render(version.Version))
			fmt.Println()

			// Service status
			fmt.Println(headerStyle.Render("  Service"))

			serviceRunning := daemon.IsDaemonRunning()
			if !serviceRunning {
				status("Status:", false, "not running")
				fmt.Println()
				fmt.Println("  Run 'ayo sandbox service start' to start the service")
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

func serviceStatusJSON(ctx context.Context) error {
	type statusOutput struct {
		CLIVersion     string `json:"cli_version"`
		ServiceRunning bool   `json:"service_running"`
		ServiceVersion string `json:"service_version,omitempty"`
		ServicePID     int    `json:"service_pid,omitempty"`
		Uptime         int64  `json:"uptime_seconds,omitempty"`
		MemoryBytes    int64  `json:"memory_bytes,omitempty"`
		SandboxTotal   int    `json:"sandbox_total,omitempty"`
		SandboxIdle    int    `json:"sandbox_idle,omitempty"`
		SandboxInUse   int    `json:"sandbox_in_use,omitempty"`
	}

	out := statusOutput{
		CLIVersion:     version.Version,
		ServiceRunning: daemon.IsDaemonRunning(),
	}

	if out.ServiceRunning {
		client := daemon.NewClient()
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := client.Connect(ctx); err == nil {
			defer client.Close()
			if daemonStatus, err := client.Status(ctx); err == nil {
				out.ServiceVersion = daemonStatus.Version
				out.ServicePID = daemonStatus.PID
				out.Uptime = daemonStatus.Uptime
				out.MemoryBytes = daemonStatus.MemoryUsage
				out.SandboxTotal = daemonStatus.Sandboxes.Total
				out.SandboxIdle = daemonStatus.Sandboxes.Idle
				out.SandboxInUse = daemonStatus.Sandboxes.InUse
			}
		}
	}

	return json.NewEncoder(os.Stdout).Encode(out)
}

func startServiceBackground(ctx context.Context) error {
	okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	checkmark := okStyle.Render("✓")

	var startErr error

	// Start the daemon process with spinner
	spinErr := spinner.New().
		Title("Starting service...").
		Type(spinner.Dots).
		Style(lipgloss.NewStyle().Foreground(lipgloss.Color("212"))).
		ActionWithErr(func(_ context.Context) error {
			startErr = daemon.StartDaemonBackground()
			return startErr
		}).
		Run()

	if spinErr != nil {
		return spinErr
	}
	if startErr != nil {
		return fmt.Errorf("start service: %w", startErr)
	}

	// Wait for service to become ready with spinner
	var ready bool
	spinErr = spinner.New().
		Title("Waiting for service to become ready...").
		Type(spinner.Dots).
		Style(lipgloss.NewStyle().Foreground(lipgloss.Color("212"))).
		ActionWithErr(func(_ context.Context) error {
			// Poll for the service to start accepting connections
			// Use longer timeout since initial sandbox creation can take a while
			timeout := time.After(30 * time.Second)
			ticker := time.NewTicker(200 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-timeout:
					return fmt.Errorf("timed out waiting for service")
				case <-ticker.C:
					if daemon.IsDaemonRunning() {
						client := daemon.NewClient()
						ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
						if err := client.Connect(ctx); err == nil {
							if err := client.Ping(ctx); err == nil {
								client.Close()
								cancel()
								ready = true
								return nil
							}
							client.Close()
						}
						cancel()
					}
				}
			}
		}).
		Run()

	if spinErr != nil {
		return spinErr
	}

	if ready {
		fmt.Printf("%s Service started\n", checkmark)
	}

	return nil
}

func runServiceForeground(ctx context.Context, server *daemon.Server, socketPath string) error {
	// Styles for output
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	fmt.Println()
	fmt.Println(headerStyle.Render("  Ayo Service"))
	fmt.Printf("  %s\n", infoStyle.Render(fmt.Sprintf("Version: %s", version.Version)))
	fmt.Printf("  %s\n", infoStyle.Render(fmt.Sprintf("Socket: %s", socketPath)))
	fmt.Println()

	// Handle shutdown signals
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Start server
	if err := server.Start(ctx, socketPath); err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	fmt.Println(infoStyle.Render("  Service started. Press Ctrl+C to stop."))
	fmt.Println()

	// Wait for shutdown
	<-ctx.Done()

	// Stop server
	stopCtx := context.Background()
	if err := server.Stop(stopCtx); err != nil {
		return fmt.Errorf("stop server: %w", err)
	}

	return nil
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
