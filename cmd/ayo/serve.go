package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/db"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/server"
	"github.com/alexcabrera/ayo/internal/server/tunnel"
	"github.com/alexcabrera/ayo/internal/session"
)

func newServeCmd(cfgPath *string) *cobra.Command {
	var host string
	var port int
	var useTunnel bool

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP API server",
		Long: `Start the HTTP API server for remote access to ayo agents.

The server provides:
- REST API for agent listing and chat
- SSE streaming for real-time responses
- Embedded web client at /
- QR code pairing for mobile clients

Examples:
  ayo serve                    # Start on random port
  ayo serve --port 8080        # Start on specific port
  ayo serve --host 0.0.0.0     # Allow external connections
  ayo serve --tunnel           # Create HTTPS tunnel via cloudflared`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(*cfgPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Connect to database
			dbConn, queries, err := db.ConnectWithQueries(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer dbConn.Close()

			// Initialize services
			services := session.NewServices(dbConn, queries)

			// Determine address
			addr := fmt.Sprintf("%s:%d", host, port)

			// Determine display host for QR code (used when not tunneling)
			displayHost := host
			if host == "0.0.0.0" {
				if hostname, err := os.Hostname(); err == nil {
					// Ensure .local suffix for mDNS resolution
					if !strings.HasSuffix(hostname, ".local") {
						hostname = hostname + ".local"
					}
					displayHost = hostname
				}
			}

			// Token will be set by server
			var token string

			// Tunnel URL (set if --tunnel is used)
			var tunnelURL string

			// Handle shutdown
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			// Track active tunnel for cleanup
			var activeTunnel tunnel.Tunnel

			go func() {
				<-sigCh
				fmt.Println("\nShutting down...")
				if activeTunnel != nil {
					activeTunnel.Stop()
				}
				cancel()
			}()

			// Create server with OnReady callback to print connection info
			srv := server.New(cfg, server.Options{
				Addr:        addr,
				Services:    services,
				AllowRemote: host != "127.0.0.1" && host != "localhost" || useTunnel,
				OnReady: func(actualAddr string) {
					// Get the actual port from the address
					_, actualPort, _ := net.SplitHostPort(actualAddr)

					// Start tunnel if requested
					if useTunnel && tunnelURL == "" {
						provider := tunnel.DefaultProvider()
						if provider == nil {
							fmt.Printf("Error: no tunnel provider available (install cloudflared: brew install cloudflared)\n")
							return
						}

						fmt.Printf("Starting %s tunnel...\n", provider.Name())
						var err error
						activeTunnel, err = provider.Start(ctx, actualAddr)
						if err != nil {
							fmt.Printf("Error: failed to start tunnel: %v\n", err)
							return
						}
						tunnelURL = activeTunnel.URL()
					}

					// Use tunnel URL if available, otherwise local URL
					var url string
					if tunnelURL != "" {
						url = tunnelURL
					} else {
						displayAddr := fmt.Sprintf("%s:%s", displayHost, actualPort)
						url = fmt.Sprintf("http://%s", displayAddr)
					}

					qrASCII, _, err := server.GenerateQRCodeWithURL(url, token)
					if err != nil {
						// Fallback to text-only if QR fails
						fmt.Printf("Starting ayo server...\n")
						fmt.Printf("Web client: %s/\n", url)
						fmt.Printf("API base: %s\n", url)
						fmt.Printf("Token: %s\n", token)
					} else {
						fmt.Print(server.FormatQRDisplay(qrASCII, url, token))
					}
					fmt.Printf("Press Ctrl+C to stop\n")
				},
			})

			// Get token after server is created
			token = srv.Token()

			// Check tunnel availability early
			if useTunnel {
				provider := tunnel.DefaultProvider()
				if provider == nil {
					return fmt.Errorf("no tunnel provider available (install cloudflared: brew install cloudflared)")
				}
			}

			// Start server
			if err := srv.Start(ctx); err != nil {
				return fmt.Errorf("server error: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&host, "host", "127.0.0.1", "Host to bind to")
	cmd.Flags().IntVarP(&port, "port", "p", 0, "Port to listen on (0 for random)")
	cmd.Flags().BoolVarP(&useTunnel, "tunnel", "t", false, "Create HTTPS tunnel via cloudflared")

	return cmd
}
