package main

import (
	"context"
	"encoding/json"
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
	"github.com/alexcabrera/ayo/internal/doctor"
	"github.com/alexcabrera/ayo/internal/ollama"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/providers"
	"github.com/alexcabrera/ayo/internal/sandbox"
	"github.com/alexcabrera/ayo/internal/version"
)

func newDoctorCmd(cfgPath *string) *cobra.Command {
	var verbose bool
	var fix bool

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
			checker := doctor.NewChecker()

			// JSON output mode
			if globalOutput.JSON {
				return runDoctorJSON(ctx, cfg, checker)
			}

			// Styles
			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
			okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
			errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
			labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(30)

			check := func(name string, ok bool, msg string) {
				status := okStyle.Render("✓")
				if !ok {
					status = errStyle.Render("✗")
				}
				fmt.Printf("  %s %s %s\n", status, labelStyle.Render(name), msg)
			}

			warn := func(name string, msg string) {
				status := warnStyle.Render("⚠")
				fmt.Printf("  %s %s %s\n", status, labelStyle.Render(name), msg)
			}

			fmt.Println()
			fmt.Println(headerStyle.Render("  Ayo Doctor"))
			fmt.Println(headerStyle.Render("  " + strings.Repeat("─", 50)))
			fmt.Println()

			// System Requirements
			fmt.Println(headerStyle.Render("  System Requirements"))
			checker.CheckSystemRequirements(ctx)
			fmt.Printf("  %s %s %s\n", okStyle.Render("✓"), labelStyle.Render("Ayo Version:"), version.Version)
			fmt.Printf("  %s %s %s\n", okStyle.Render("✓"), labelStyle.Render("Go Version:"), runtime.Version())
			fmt.Printf("  %s %s %s/%s\n", okStyle.Render("✓"), labelStyle.Render("Platform:"), runtime.GOOS, runtime.GOARCH)
			fmt.Println()

			// Daemon
			fmt.Println(headerStyle.Render("  Daemon"))
			checker.CheckDaemon(ctx)
			printCheckerResults(checker, "Daemon", check, warn)
			fmt.Println()

			// Check paths
			fmt.Println(headerStyle.Render("  Paths"))
			checker.CheckPaths(ctx)
			printCheckerResults(checker, "Paths", check, warn)
			fmt.Println()

			// API Keys
			fmt.Println(headerStyle.Render("  API Keys"))
			checker.CheckAPIKeys(ctx)
			printCheckerResults(checker, "API Keys", check, warn)
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
				warn("Binary:", "not found in PATH")
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
				}
			} else {
				warn("Service:", "not running - memory features disabled")
			}
			fmt.Println()

			// Check database
			fmt.Println(headerStyle.Render("  Database"))
			dbPath := paths.DatabasePath()
			if fileExists(dbPath) {
				dbConn, queries, err := db.ConnectWithQueries(ctx, dbPath)
				if err != nil {
					check("Connection:", false, err.Error())
				} else {
					check("Connection:", true, "OK")
					defer dbConn.Close()

					sessions, _ := queries.CountSessions(ctx)
					check("Sessions:", true, fmt.Sprintf("%d", sessions))

					memories, _ := queries.ListMemories(ctx, db.ListMemoriesParams{
						Lim: 1000,
					})
					check("Active Memories:", true, fmt.Sprintf("%d", len(memories)))
				}
			} else {
				warn("Database:", "not created yet - run 'ayo setup'")
			}
			fmt.Println()

			// Squads
			fmt.Println(headerStyle.Render("  Squads"))
			checker.CheckSquads(ctx)
			printCheckerResults(checker, "Squads", check, warn)
			fmt.Println()

			// Check sandbox
			fmt.Println(headerStyle.Render("  Sandbox"))
			sandboxProvider := selectSandboxProvider()
			if sandboxProvider == nil {
				warn("Provider:", "none available")
			} else {
				check("Provider:", true, sandboxProvider.Name())

				if verbose {
					fmt.Println("  Testing sandbox execution...")
					testCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
					defer cancel()

					sb, createErr := sandboxProvider.Create(testCtx, providers.SandboxCreateOptions{
						Name: "ayo-doctor-test",
					})
					if createErr != nil {
						check("Create:", false, createErr.Error())
					} else {
						check("Create:", true, sb.ID)

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

			// Summary
			summary := checker.Summary()
			fmt.Println(headerStyle.Render("  " + strings.Repeat("─", 50)))
			fmt.Printf("  Summary: %s passed, %s warnings, %s errors\n",
				okStyle.Render(fmt.Sprintf("%d", summary.Passed)),
				warnStyle.Render(fmt.Sprintf("%d", summary.Warnings)),
				errStyle.Render(fmt.Sprintf("%d", summary.Errors)))

			if summary.Warnings > 0 || summary.Errors > 0 {
				fmt.Println()
				fmt.Println("  Run 'ayo doctor --fix' to attempt automatic fixes.")
			}
			fmt.Println()

			// --fix mode
			if fix {
				return runDoctorFixes(ctx, summary)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")
	cmd.Flags().BoolVar(&fix, "fix", false, "Attempt automatic fixes")

	return cmd
}

func printCheckerResults(checker *doctor.Checker, category string, check func(string, bool, string), warn func(string, string)) {
	summary := checker.Summary()
	for _, r := range summary.Results {
		if r.Category != category {
			continue
		}
		switch r.Status {
		case doctor.StatusPass:
			check(r.Name+":", true, r.Message)
		case doctor.StatusWarn:
			warn(r.Name+":", r.Message)
			if r.Fix != "" {
				fmt.Printf("    Fix: %s\n", r.Fix)
			}
		case doctor.StatusFail:
			check(r.Name+":", false, r.Message)
			if r.Fix != "" {
				fmt.Printf("    Fix: %s\n", r.Fix)
			}
		}
	}
}

func runDoctorJSON(ctx context.Context, cfg config.Config, checker *doctor.Checker) error {
	checker.CheckSystemRequirements(ctx)
	checker.CheckDaemon(ctx)
	checker.CheckPaths(ctx)
	checker.CheckAPIKeys(ctx)
	checker.CheckSquads(ctx)

	summary := checker.Summary()
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(summary)
}

func runDoctorFixes(ctx context.Context, summary doctor.Summary) error {
	fmt.Println("  Attempting fixes...")
	fmt.Println()

	fixed := 0
	cantFix := 0

	for _, r := range summary.Results {
		if r.Status == doctor.StatusPass {
			continue
		}
		if !r.Fixable {
			if r.Fix != "" {
				fmt.Printf("  ✗ %s (requires manual action)\n", r.Name)
				fmt.Printf("    %s\n", r.Fix)
				cantFix++
			}
			continue
		}

		// Attempt common fixes
		switch {
		case strings.Contains(r.Name, "Daemon"):
			fmt.Print("  Attempting to start daemon...")
			cmd := exec.CommandContext(ctx, "ayo", "daemon", "start")
			if err := cmd.Run(); err != nil {
				fmt.Println(" failed")
			} else {
				fmt.Println(" done")
				fixed++
			}
		case strings.Contains(r.Name, "Directory"):
			fmt.Printf("  Creating %s...", r.Name)
			if strings.Contains(r.Message, paths.ConfigDir()) {
				if err := os.MkdirAll(paths.ConfigDir(), 0755); err != nil {
					fmt.Println(" failed")
				} else {
					fmt.Println(" done")
					fixed++
				}
			} else if strings.Contains(r.Message, paths.DataDir()) {
				if err := os.MkdirAll(paths.DataDir(), 0755); err != nil {
					fmt.Println(" failed")
				} else {
					fmt.Println(" done")
					fixed++
				}
			}
		}
	}

	fmt.Println()
	if fixed > 0 {
		fmt.Printf("  Fixed %d issue(s)\n", fixed)
	}
	if cantFix > 0 {
		fmt.Printf("  %d issue(s) require manual action\n", cantFix)
	}

	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
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
