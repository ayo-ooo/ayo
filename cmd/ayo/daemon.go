package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/daemon"
	"github.com/alexcabrera/ayo/internal/version"
)

func newDaemonCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Daemon management commands",
		Long:  "Commands for managing the ayo daemon process.",
	}

	cmd.AddCommand(newDaemonStartCmd(cfgPath))
	cmd.AddCommand(newDaemonStopCmd(cfgPath))

	return cmd
}

func newDaemonStartCmd(cfgPath *string) *cobra.Command {
	var foreground bool

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the ayo daemon",
		Long: `Start the ayo daemon process.

The daemon manages:
- Sandbox pool for isolated command execution
- LLM connection pooling
- Memory index caching

By default, the daemon runs in the background. Use --foreground to run
in the current terminal for debugging.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Check if already running
			if daemon.IsDaemonRunning() {
				client := daemon.NewClient()
				if err := client.Connect(ctx); err == nil {
					defer client.Close()
					if err := client.Ping(ctx); err == nil {
						fmt.Println("Daemon is already running")
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
				return runDaemonForeground(ctx, server, cfg.SocketPath)
			}

			// Background mode: start and detach
			return daemon.StartDaemonBackground()
		},
	}

	cmd.Flags().BoolVarP(&foreground, "foreground", "f", false, "Run in foreground (for debugging)")

	return cmd
}

func newDaemonStopCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the ayo daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if !daemon.IsDaemonRunning() {
				fmt.Println("Daemon is not running")
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

			fmt.Println("Daemon stopped")
			return nil
		},
	}

	return cmd
}

func runDaemonForeground(ctx context.Context, server *daemon.Server, socketPath string) error {
	// Styles for output
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	fmt.Println()
	fmt.Println(headerStyle.Render("  Ayo Daemon"))
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

	fmt.Println(infoStyle.Render("  Daemon started. Press Ctrl+C to stop."))
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
