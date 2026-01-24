package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/db"
	"github.com/alexcabrera/ayo/internal/ollama"
	"github.com/alexcabrera/ayo/internal/paths"
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

					smallModel := cfg.SmallModel
					if smallModel == "" {
						smallModel = "ministral-3:3b"
					}
					// Strip provider prefix (e.g., "ollama/") if present
					if strings.Contains(smallModel, "/") {
						parts := strings.SplitN(smallModel, "/", 2)
						smallModel = parts[len(parts)-1]
					}
					hasSmall := false
					for _, m := range models {
						if strings.HasPrefix(m.Name, smallModel) || strings.Contains(m.Name, smallModel) {
							hasSmall = true
							break
						}
					}
					check("Small Model:", hasSmall, smallModel)
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
						Limit: 1000,
					})
					check("Active Memories:", true, fmt.Sprintf("%d", len(memories)))
				}
			} else {
				warn("Database:", "not created yet - run 'ayo setup'")
			}
			fmt.Println()

			// Check config
			fmt.Println(headerStyle.Render("  Configuration"))
			if configExists {
				if cfg.DefaultModel != "" {
					check("Default Model:", true, cfg.DefaultModel)
				} else {
					warn("Default Model:", "not set")
				}
			} else {
				warn("Config:", "not found - using defaults")
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
