package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/db"
	"github.com/alexcabrera/ayo/internal/ollama"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/sandbox"
	"github.com/alexcabrera/ayo/internal/version"
)

func newDoctorCmd(cfgPath *string) *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check system health and dependencies",
		Long:  "Diagnose the ayo installation, checking Ollama, models, database, and configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(*cfgPath)
			if err != nil {
				cfg = config.Config{}
			}

			ctx := cmd.Context()

			// Styles
			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
			okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
			errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
			labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(30)

			check := func(name string, ok bool, msg string) {
				status := okStyle.Render("OK")
				if !ok {
					status = errStyle.Render("FAIL")
				}
				fmt.Printf("  %s %s %s\n", labelStyle.Render(name), status, msg)
			}

			warn := func(name string, msg string) {
				status := warnStyle.Render("WARN")
				fmt.Printf("  %s %s %s\n", labelStyle.Render(name), status, msg)
			}

			fmt.Println()
			fmt.Println(headerStyle.Render("  Ayo Doctor"))
			fmt.Println(headerStyle.Render("  " + strings.Repeat("-", 50)))
			fmt.Println()

			// Version info
			fmt.Println(headerStyle.Render("  System"))
			fmt.Printf("  %s %s\n", labelStyle.Render("Ayo Version:"), version.Version)
			fmt.Printf("  %s %s\n", labelStyle.Render("Go Version:"), runtime.Version())
			fmt.Printf("  %s %s/%s\n", labelStyle.Render("Platform:"), runtime.GOOS, runtime.GOARCH)
			fmt.Println()

			// Check paths
			fmt.Println(headerStyle.Render("  Paths"))
			configExists := fileExists(paths.ConfigFile())
			if configExists {
				check("Config File:", true, paths.ConfigFile())
			} else {
				warn("Config File:", paths.ConfigFile()+" (using defaults)")
			}

			dbExists := fileExists(paths.DatabasePath())
			check("Database:", dbExists, paths.DatabasePath())

			builtinExists := dirExists(paths.UserDataDir())
			check("Data Directory:", builtinExists, paths.UserDataDir())
			fmt.Println()

			// Check Ollama
			fmt.Println(headerStyle.Render("  Ollama"))
			ollamaHost := cfg.OllamaHost
			if ollamaHost == "" {
				ollamaHost = "http://localhost:11434"
			}

			// Check if ollama binary exists
			ollamaPath, pathErr := exec.LookPath("ollama")
			if pathErr != nil {
				check("Binary:", false, "not found in PATH")
			} else {
				check("Binary:", true, ollamaPath)
			}

			// Check if Ollama is running
			client := ollama.NewClient(ollama.WithHost(ollamaHost))
			isAvailable := client.IsAvailable(ctx)
			check("Service:", isAvailable, ollamaHost)

			if isAvailable {
				// Check version
				ver, verErr := client.GetVersion(ctx)
				if verErr != nil {
					warn("Version:", "could not get version")
				} else {
					check("Version:", true, ver)
				}

				// Check models
				models, modelsErr := client.ListModels(ctx)
				if modelsErr != nil {
					warn("Models:", "could not list models")
				} else {
					check("Models:", len(models) > 0, fmt.Sprintf("%d installed", len(models)))

					if verbose {
						for _, m := range models {
							fmt.Printf("       - %s\n", m.Name)
						}
					}

					// Check for required models
					embModel := cfg.Embedding.Model
					if embModel == "" {
						embModel = "nomic-embed-text"
					}
					hasEmbedding := false
					for _, m := range models {
						if strings.HasPrefix(m.Name, embModel) || strings.Contains(m.Name, embModel) {
							hasEmbedding = true
							break
						}
					}
					check("Embedding Model:", hasEmbedding, embModel)

					// Small model check - only validate in Ollama if it's an Ollama model
					small := cfg.GetSmallModel()
					smallModel := small.String()
					
					// Check if this is an Ollama model (has ollama provider or ollama/ prefix)
					isOllamaModel := small.Provider == "ollama" || strings.HasPrefix(smallModel, "ollama/")
					displayModel := smallModel
					if isOllamaModel {
						// Strip ollama/ prefix for display and lookup
						displayModel = strings.TrimPrefix(smallModel, "ollama/")
					}
					
					if isOllamaModel {
						// Validate Ollama model exists
						hasSmall := false
						for _, m := range models {
							if strings.HasPrefix(m.Name, displayModel) || strings.Contains(m.Name, displayModel) {
								hasSmall = true
								break
							}
						}
						check("Small Model:", hasSmall, displayModel+" (ollama)")
					} else {
						// Non-Ollama model - just report it's configured
						check("Small Model:", true, smallModel+" (cloud)")
					}
				}
			} else {
				warn("Service:", "not running - memory features will be disabled")
			}
			fmt.Println()

			// Check database
			fmt.Println(headerStyle.Render("  Database"))
			if dbExists {
				dbConn, queries, err := db.ConnectWithQueries(ctx, paths.DatabasePath())
				if err != nil {
					check("Connection:", false, err.Error())
				} else {
					check("Connection:", true, "OK")
					defer dbConn.Close()

					// Count sessions and memories
					sessions, _ := queries.CountSessions(ctx)
					check("Sessions:", true, fmt.Sprintf("%d", sessions))

					// Count active memories
					memories, _ := queries.ListMemories(ctx, db.ListMemoriesParams{
						Lim: 1000,
					})
					check("Active Memories:", true, fmt.Sprintf("%d", len(memories)))
				}
			} else {
				warn("Database:", "not created yet - run 'ayo setup'")
			}
			fmt.Println()

			// Check config
			fmt.Println(headerStyle.Render("  Configuration"))
			largeModel := cfg.GetLargeModel()
			smallModelCfg := cfg.GetSmallModel()
			
			if !largeModel.IsEmpty() {
				check("Large Model:", true, largeModel.String())
			} else {
				warn("Large Model:", "not configured")
			}
			
			if !smallModelCfg.IsEmpty() {
				check("Small Model:", true, smallModelCfg.String())
			} else {
				warn("Small Model:", "not configured")
			}
			fmt.Println()

			// Check sandbox
			fmt.Println(headerStyle.Render("  Sandbox"))
			sandboxProvider := selectSandboxProvider()
			if sandboxProvider == nil {
				warn("Provider:", "none available")
			} else {
				check("Provider:", true, sandboxProvider.Name())

				// Test sandbox if verbose
				if verbose {
					fmt.Println("  Testing sandbox execution...")
					testCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
					defer cancel()

					// Create a test sandbox
					sb, createErr := sandboxProvider.Create(testCtx, providers.SandboxCreateOptions{
						Name: "ayo-doctor-test",
					})
					if createErr != nil {
						check("Create:", false, createErr.Error())
					} else {
						check("Create:", true, sb.ID)

						// Execute a simple command
						result, execErr := sandboxProvider.Exec(testCtx, sb.ID, providers.ExecOptions{
							Command: "echo",
							Args:    []string{"hello from sandbox"},
						})
						if execErr != nil {
							check("Exec:", false, execErr.Error())
						} else {
							output := strings.TrimSpace(result.Stdout)
							check("Exec:", result.ExitCode == 0, fmt.Sprintf("exit=%d output=%q", result.ExitCode, output))
						}

						// Cleanup
						if delErr := sandboxProvider.Delete(testCtx, sb.ID, true); delErr != nil {
							warn("Cleanup:", delErr.Error())
						} else {
							check("Cleanup:", true, "removed test sandbox")
						}
					}
				} else {
					fmt.Println("  Run with -v to test sandbox execution")
				}
			}
			fmt.Println()

			// Recommendations
			var recommendations []string
			if !isAvailable {
				recommendations = append(recommendations, "Start Ollama: ollama serve")
			}
			if pathErr != nil {
				recommendations = append(recommendations, "Install Ollama: https://ollama.ai")
			}


			if len(recommendations) > 0 {
				fmt.Println(headerStyle.Render("  Recommendations"))
				for _, r := range recommendations {
					fmt.Printf("  - %s\n", r)
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")

	return cmd
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// selectSandboxProvider returns the appropriate sandbox provider for the current platform.
func selectSandboxProvider() providers.SandboxProvider {
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		provider := sandbox.NewAppleProvider()
		if provider.IsAvailable() {
			return provider
		}
	}
	if runtime.GOOS == "linux" {
		provider := sandbox.NewLinuxProvider()
		if provider.IsAvailable() {
			return provider
		}
	}
	return nil
}
