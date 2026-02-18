package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/capabilities"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/daemon"
	"github.com/alexcabrera/ayo/internal/embedding"
	"github.com/alexcabrera/ayo/internal/ollama"
	"github.com/alexcabrera/ayo/internal/squads"
)

func newIndexCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Manage the entity index for agent/squad selection",
		Long:  "Commands for managing the embedding index used for semantic agent and squad selection.",
	}

	cmd.AddCommand(newIndexStatusCmd(cfgPath))
	cmd.AddCommand(newIndexRebuildCmd(cfgPath))
	cmd.AddCommand(newIndexSearchCmd(cfgPath))

	return cmd
}

func newIndexStatusCmd(cfgPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show index statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(*cfgPath)
			if err != nil {
				cfg = config.Config{}
			}

			// Styles
			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
			labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(25)
			valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))

			fmt.Println()
			fmt.Println(headerStyle.Render("  Entity Index Status"))
			fmt.Println(headerStyle.Render("  " + strings.Repeat("-", 40)))
			fmt.Println()

			// Count agents
			handles, err := agent.ListHandles(cfg)
			if err != nil {
				handles = []string{}
			}
			fmt.Printf("  %s %s\n", labelStyle.Render("Agents Found:"), valueStyle.Render(fmt.Sprintf("%d", len(handles))))

			// Count squads via daemon
			ctx := cmd.Context()
			squadCount := 0
			client, err := daemon.ConnectOrStart(ctx)
			if err == nil {
				defer client.Close()
				result, err := client.SquadList(ctx)
				if err == nil {
					squadCount = len(result.Squads)
				}
			}
			fmt.Printf("  %s %s\n", labelStyle.Render("Squads Found:"), valueStyle.Render(fmt.Sprintf("%d", squadCount)))

			// Check embedding provider
			embedModel := cfg.Embedding.Model
			if embedModel == "" {
				embedModel = "nomic-embed-text"
			}
			fmt.Printf("  %s %s\n", labelStyle.Render("Embedding Model:"), valueStyle.Render(embedModel))

			fmt.Println()
			fmt.Printf("  %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Run 'ayo index rebuild' to regenerate embeddings"))
			fmt.Println()

			return nil
		},
	}
}

func newIndexRebuildCmd(cfgPath *string) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "rebuild",
		Short: "Rebuild the entity index",
		Long:  "Regenerate embeddings for all agents and squads. This may take some time depending on the number of entities.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(*cfgPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			ctx := cmd.Context()

			// Create embedder
			embedder, err := createIndexEmbedder()
			if err != nil {
				return fmt.Errorf("create embedder: %w", err)
			}
			defer embedder.Close()

			idx := capabilities.NewLazyEntityIndex(embedder)

			// Styles
			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
			infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))

			fmt.Println()
			fmt.Println(headerStyle.Render("  Rebuilding Entity Index"))
			fmt.Println()

			// Index agents
			handles, err := agent.ListHandles(cfg)
			if err != nil {
				handles = []string{}
			}
			fmt.Printf("  %s\n", infoStyle.Render(fmt.Sprintf("Indexing %d agents...", len(handles))))

			agentStart := time.Now()
			for _, handle := range handles {
				ag, err := agent.Load(cfg, handle)
				if err != nil {
					fmt.Printf("  Warning: failed to load %s: %v\n", handle, err)
					continue
				}
				_, err = idx.GetAgentEmbedding(ctx, ag)
				if err != nil {
					fmt.Printf("  Warning: failed to embed %s: %v\n", handle, err)
					continue
				}
			}
			fmt.Printf("  %s\n", successStyle.Render(fmt.Sprintf("  Indexed %d agents in %v", idx.AgentCount(), time.Since(agentStart).Round(time.Millisecond))))

			// Index squads via daemon
			client, err := daemon.ConnectOrStart(ctx)
			if err != nil {
				fmt.Printf("  Warning: failed to connect to daemon: %v\n", err)
			} else {
				defer client.Close()
				result, err := client.SquadList(ctx)
				if err != nil {
					fmt.Printf("  Warning: failed to list squads: %v\n", err)
				} else {
					fmt.Printf("  %s\n", infoStyle.Render(fmt.Sprintf("Indexing %d squads...", len(result.Squads))))
					squadStart := time.Now()
					for _, squadInfo := range result.Squads {
						constitution, err := squads.LoadConstitution(squadInfo.Name)
						if err != nil {
							fmt.Printf("  Warning: failed to load constitution for %s: %v\n", squadInfo.Name, err)
							continue
						}
						_, err = idx.GetSquadEmbedding(ctx, constitution, squadInfo.Name)
						if err != nil {
							fmt.Printf("  Warning: failed to embed %s: %v\n", squadInfo.Name, err)
							continue
						}
					}
					fmt.Printf("  %s\n", successStyle.Render(fmt.Sprintf("  Indexed %d squads in %v", idx.SquadCount(), time.Since(squadStart).Round(time.Millisecond))))
				}
			}

			fmt.Println()
			fmt.Printf("  %s\n", successStyle.Render(fmt.Sprintf("Total: %d entities indexed", idx.TotalCount())))
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force rebuild even if index is up to date")

	return cmd
}

func newIndexSearchCmd(cfgPath *string) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search for agents and squads by capability",
		Long:  "Test semantic search without dispatching. Shows ranked results for the given query.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(*cfgPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			ctx := cmd.Context()
			query := args[0]

			// Create embedder
			embedder, err := createIndexEmbedder()
			if err != nil {
				return fmt.Errorf("create embedder: %w", err)
			}
			defer embedder.Close()

			// Build index first
			idx, err := buildEntityIndex(ctx, cfg, embedder)
			if err != nil {
				return err
			}

			// Create searcher and search
			searcher := capabilities.NewUnifiedSearcher(idx, embedder)
			results, err := searcher.Search(ctx, query, limit)
			if err != nil {
				return fmt.Errorf("search: %w", err)
			}

			if len(results) == 0 {
				fmt.Println("No results found.")
				return nil
			}

			// Output
			if globalOutput.JSON {
				globalOutput.Print(results, "")
				return nil
			}

			// Styles
			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
			typeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
			handleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
			scoreStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
			descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

			fmt.Println()
			fmt.Println(headerStyle.Render(fmt.Sprintf("  Search Results for: %q", query)))
			fmt.Println()

			for i, r := range results {
				typeLabel := typeStyle.Render(fmt.Sprintf("[%s]", r.Type))
				handle := handleStyle.Render(r.Handle)
				score := scoreStyle.Render(fmt.Sprintf("(%.2f)", r.Score))

				fmt.Printf("  %d. %s %s %s\n", i+1, typeLabel, handle, score)

				if r.Description != "" {
					desc := r.Description
					if len(desc) > 80 {
						desc = desc[:77] + "..."
					}
					fmt.Printf("     %s\n", descStyle.Render(desc))
				}
			}
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 10, "Maximum number of results")

	return cmd
}

// createIndexEmbedder creates an embedder for indexing.
func createIndexEmbedder() (embedding.Embedder, error) {
	client := ollama.NewClient()
	if !client.IsAvailable(context.Background()) {
		return nil, fmt.Errorf("Ollama not available at %s", client.Host())
	}
	return embedding.NewOllamaEmbedder(embedding.OllamaConfig{}), nil
}

// buildEntityIndex creates and populates the entity index.
func buildEntityIndex(ctx context.Context, cfg config.Config, embedder embedding.Embedder) (*capabilities.LazyEntityIndex, error) {
	idx := capabilities.NewLazyEntityIndex(embedder)

	// Index agents
	handles, err := agent.ListHandles(cfg)
	if err == nil {
		for _, handle := range handles {
			ag, err := agent.Load(cfg, handle)
			if err != nil {
				continue // Skip failed loads
			}
			idx.GetAgentEmbedding(ctx, ag)
		}
	}

	// Index squads via daemon
	client, err := daemon.ConnectOrStart(ctx)
	if err == nil {
		defer client.Close()
		result, err := client.SquadList(ctx)
		if err == nil {
			for _, squadInfo := range result.Squads {
				constitution, err := squads.LoadConstitution(squadInfo.Name)
				if err != nil {
					continue
				}
				idx.GetSquadEmbedding(ctx, constitution, squadInfo.Name)
			}
		}
	}

	return idx, nil
}
