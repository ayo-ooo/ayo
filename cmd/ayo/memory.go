package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/huh"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/db"
	"github.com/alexcabrera/ayo/internal/embedding"
	"github.com/alexcabrera/ayo/internal/memory"
	"github.com/alexcabrera/ayo/internal/ollama"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/smallmodel"
	"github.com/alexcabrera/ayo/internal/ui"
)

func newMemoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "memory",
		Aliases: []string{"memories"},
		Short:   "Manage agent memories",
		Long: `Manage persistent facts, preferences, and patterns learned across sessions.

Memories help agents provide personalized responses by remembering user
preferences, project facts, and behavioral patterns.

Categories:
  preference   User preferences (tools, styles, communication)
  fact         Facts about user or project
  correction   User corrections to agent behavior
  pattern      Observed behavioral patterns`,
	}

	cmd.AddCommand(newMemoryListCmd())
	cmd.AddCommand(newMemorySearchCmd())
	cmd.AddCommand(newMemoryShowCmd())
	cmd.AddCommand(newMemoryStoreCmd())
	cmd.AddCommand(newMemoryForgetCmd())
	cmd.AddCommand(newMemoryStatsCmd())
	cmd.AddCommand(newMemoryClearCmd())

	return cmd
}

func newMemoryListCmd() *cobra.Command {
	var agentFilter string
	var categoryFilter string
	var limit int64
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List memories",
		RunE: func(cmd *cobra.Command, args []string) error {
			dbConn, queries, err := db.ConnectWithQueries(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer dbConn.Close()

			svc := memory.NewService(queries, nil)

			memories, err := svc.List(cmd.Context(), agentFilter, limit, 0)
			if err != nil {
				return fmt.Errorf("failed to list memories: %w", err)
			}

			// Filter by category if specified
			if categoryFilter != "" {
				var filtered []memory.Memory
				for _, m := range memories {
					if string(m.Category) == categoryFilter {
						filtered = append(filtered, m)
					}
				}
				memories = filtered
			}

			if jsonOutput {
				return writeJSON(memoriesToJSON(memories))
			}

			if len(memories) == 0 {
				fmt.Println("No memories found")
				return nil
			}

			// Styles
			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
			idStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			categoryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141"))
			contentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
			agentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
			timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

			fmt.Println()
			fmt.Println(headerStyle.Render("  Memories"))
			fmt.Println(headerStyle.Render("  " + strings.Repeat("-", 60)))
			fmt.Println()

			for _, m := range memories {
				// Truncate content if too long
				content := m.Content
				if len(content) > 50 {
					content = content[:47] + "..."
				}

				idShort := m.ID
				if len(idShort) > 8 {
					idShort = idShort[:8]
				}

				agent := m.AgentHandle
				if agent == "" {
					agent = "global"
				}

				fmt.Printf("  %s  %s  %s\n",
					idStyle.Render(idShort),
					categoryStyle.Render(fmt.Sprintf("%-11s", m.Category)),
					contentStyle.Render(content),
				)
				fmt.Printf("     %s  %s\n",
					agentStyle.Render(agent),
					timeStyle.Render(m.CreatedAt.Format("2006-01-02 15:04")),
				)
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&agentFilter, "agent", "a", "", "Filter by agent handle")
	cmd.Flags().StringVarP(&categoryFilter, "category", "c", "", "Filter by category")
	cmd.Flags().Int64VarP(&limit, "limit", "n", 50, "Maximum number of memories to show")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

func newMemorySearchCmd() *cobra.Command {
	var agentFilter string
	var threshold float64
	var limit int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search memories semantically",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			dbConn, queries, err := db.ConnectWithQueries(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer dbConn.Close()

			// Try to create embedder for search
			embedder, err := createEmbedder()
			if err != nil {
				return fmt.Errorf("failed to create embedder (run 'ayo setup' to install): %w", err)
			}
			defer embedder.Close()

			svc := memory.NewService(queries, embedder)

			results, err := svc.Search(cmd.Context(), query, memory.SearchOptions{
				AgentHandle: agentFilter,
				Threshold:   float32(threshold),
				Limit:       limit,
			})
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			if jsonOutput {
				return writeJSON(searchResultsToJSON(results))
			}

			if len(results) == 0 {
				fmt.Println("No memories found matching the query")
				return nil
			}

			// Styles
			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
			idStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			scoreStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
			contentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
			categoryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141"))

			fmt.Println()
			fmt.Println(headerStyle.Render(fmt.Sprintf("  Search Results for: %s", query)))
			fmt.Println(headerStyle.Render("  " + strings.Repeat("-", 60)))
			fmt.Println()

			for _, r := range results {
				content := r.Memory.Content
				if len(content) > 60 {
					content = content[:57] + "..."
				}

				idShort := r.Memory.ID
				if len(idShort) > 8 {
					idShort = idShort[:8]
				}

				fmt.Printf("  %s  %s  %s\n",
					idStyle.Render(idShort),
					scoreStyle.Render(fmt.Sprintf("%.2f", r.Similarity)),
					contentStyle.Render(content),
				)
				fmt.Printf("     %s\n",
					categoryStyle.Render(string(r.Memory.Category)),
				)
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&agentFilter, "agent", "a", "", "Filter by agent handle")
	cmd.Flags().Float64VarP(&threshold, "threshold", "t", 0.3, "Minimum similarity threshold (0-1)")
	cmd.Flags().IntVarP(&limit, "limit", "n", 10, "Maximum number of results")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

func newMemoryShowCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show [id]",
		Short: "Show memory details",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbConn, queries, err := db.ConnectWithQueries(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer dbConn.Close()

			svc := memory.NewService(queries, nil)

			var mem memory.Memory

			if len(args) == 0 {
				// Interactive picker
				mem, err = pickMemory(cmd.Context(), svc, "Select memory to view")
				if err != nil {
					return err
				}
			} else {
				mem, err = svc.GetByPrefix(cmd.Context(), args[0])
				if err != nil {
					return fmt.Errorf("memory not found: %w", err)
				}
			}

			if jsonOutput {
				return writeJSON(memoryToJSON(mem))
			}

			// Styles
			labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("245"))
			valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
			contentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Italic(true)

			fmt.Println()
			fmt.Printf("%s %s\n", labelStyle.Render("ID:"), valueStyle.Render(mem.ID))
			fmt.Printf("%s %s\n", labelStyle.Render("Category:"), valueStyle.Render(string(mem.Category)))
			fmt.Printf("%s %s\n", labelStyle.Render("Status:"), valueStyle.Render(string(mem.Status)))
			fmt.Println()
			fmt.Printf("%s\n", labelStyle.Render("Content:"))
			fmt.Printf("  %s\n", contentStyle.Render(mem.Content))
			fmt.Println()

			if mem.AgentHandle != "" {
				fmt.Printf("%s %s\n", labelStyle.Render("Agent:"), valueStyle.Render(mem.AgentHandle))
			}
			if mem.PathScope != "" {
				fmt.Printf("%s %s\n", labelStyle.Render("Path:"), valueStyle.Render(mem.PathScope))
			}

			fmt.Printf("%s %.2f\n", labelStyle.Render("Confidence:"), mem.Confidence)
			fmt.Printf("%s %d\n", labelStyle.Render("Access Count:"), mem.AccessCount)
			fmt.Printf("%s %s\n", labelStyle.Render("Created:"), valueStyle.Render(mem.CreatedAt.Format("2006-01-02 15:04:05")))
			fmt.Printf("%s %s\n", labelStyle.Render("Updated:"), valueStyle.Render(mem.UpdatedAt.Format("2006-01-02 15:04:05")))

			if mem.SupersedesID != "" {
				fmt.Printf("%s %s\n", labelStyle.Render("Supersedes:"), valueStyle.Render(mem.SupersedesID))
			}
			if mem.SupersededByID != "" {
				fmt.Printf("%s %s\n", labelStyle.Render("Superseded By:"), valueStyle.Render(mem.SupersededByID))
			}

			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

func newMemoryStoreCmd() *cobra.Command {
	var agentHandle string
	var category string
	var pathScope string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "store <content>",
		Short: "Store a new memory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			content := args[0]
			ctx := cmd.Context()

			// Don't show spinner for JSON output (used by agents)
			var spinner *ui.Spinner
			if !jsonOutput {
				spinner = ui.NewSpinnerWithType("storing memory...", ui.SpinnerMemory)
				spinner.Start()
			}

			dbConn, queries, err := db.ConnectWithQueries(ctx, paths.DatabasePath())
			if err != nil {
				if spinner != nil {
					spinner.StopWithError("failed to connect to database")
				}
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer dbConn.Close()

			// Try to create embedder for storing with embeddings
			embedder, err := createEmbedder()
			if err != nil {
				// Fall back to storing without embeddings
				embedder = nil
			}
			if embedder != nil {
				defer embedder.Close()
			}

			svc := memory.NewService(queries, embedder)

			// Determine category: use provided or auto-categorize
			var cat memory.Category
			categoryProvided := cmd.Flags().Changed("category")

			if categoryProvided {
				// User explicitly provided a category
				switch category {
				case "preference":
					cat = memory.CategoryPreference
				case "fact":
					cat = memory.CategoryFact
				case "correction":
					cat = memory.CategoryCorrection
				case "pattern":
					cat = memory.CategoryPattern
				default:
					cat = memory.CategoryFact
				}
			} else {
				// Auto-categorize using small model
				if spinner != nil {
					spinner.Stop()
					spinner = ui.NewSpinnerWithType("categorizing...", ui.SpinnerMemory)
					spinner.Start()
				}

				smallSvc := smallmodel.NewService(smallmodel.Config{})
				if smallSvc.IsAvailable(ctx) {
					result, err := smallSvc.CategorizeMemory(ctx, content)
					if err == nil {
						switch result.Category {
						case "preference":
							cat = memory.CategoryPreference
						case "correction":
							cat = memory.CategoryCorrection
						case "pattern":
							cat = memory.CategoryPattern
						default:
							cat = memory.CategoryFact
						}
					} else {
						// Fallback to fact on error
						cat = memory.CategoryFact
					}
				} else {
					// Ollama not available, default to fact
					cat = memory.CategoryFact
				}

				if spinner != nil {
					spinner.Stop()
					spinner = ui.NewSpinnerWithType("storing memory...", ui.SpinnerMemory)
					spinner.Start()
				}
			}

			mem, err := svc.Create(ctx, memory.Memory{
				Content:     content,
				Category:    cat,
				AgentHandle: agentHandle,
				PathScope:   pathScope,
			})
			if err != nil {
				if spinner != nil {
					spinner.StopWithError("failed to store memory")
				}
				if jsonOutput {
					return writeJSON(map[string]interface{}{
						"success": false,
						"error":   err.Error(),
					})
				}
				return fmt.Errorf("failed to store memory: %w", err)
			}

			if spinner != nil {
				spinner.StopWithMessage("memory stored")
			}

			if jsonOutput {
				return writeJSON(map[string]interface{}{
					"success": true,
					"memory":  memoryToJSON(mem),
				})
			}

			// Show category in output so user knows what was chosen
			fmt.Printf("Stored as %s: %s\n", mem.Category, mem.ID[:8])
			return nil
		},
	}

	cmd.Flags().StringVarP(&agentHandle, "agent", "a", "", "Agent handle for scoping")
	cmd.Flags().StringVarP(&category, "category", "c", "fact", "Memory category (preference, fact, correction, pattern) - auto-detected if not specified")
	cmd.Flags().StringVarP(&pathScope, "path", "p", "", "Path scope for this memory")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

func newMemoryForgetCmd() *cobra.Command {
	var force bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "forget [id]",
		Short: "Forget a memory (soft delete)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbConn, queries, err := db.ConnectWithQueries(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer dbConn.Close()

			svc := memory.NewService(queries, nil)

			var mem memory.Memory

			if len(args) == 0 {
				// Interactive picker
				mem, err = pickMemory(cmd.Context(), svc, "Select memory to forget")
				if err != nil {
					return err
				}
			} else {
				mem, err = svc.GetByPrefix(cmd.Context(), args[0])
				if err != nil {
					if jsonOutput {
						return writeJSON(map[string]interface{}{
							"success": false,
							"error":   "memory not found",
						})
					}
					return fmt.Errorf("memory not found: %w", err)
				}
			}

			// Skip confirmation in JSON mode (assume --force)
			if !force && !jsonOutput {
				fmt.Printf("Memory to forget: %s\n", mem.Content)
				fmt.Print("Are you sure? [y/N] ")
				var confirm string
				fmt.Scanln(&confirm)
				if strings.ToLower(confirm) != "y" {
					fmt.Println("Cancelled")
					return nil
				}
			}

			err = svc.Forget(cmd.Context(), mem.ID)
			if err != nil {
				if jsonOutput {
					return writeJSON(map[string]interface{}{
						"success": false,
						"error":   err.Error(),
					})
				}
				return fmt.Errorf("failed to forget memory: %w", err)
			}

			if jsonOutput {
				return writeJSON(map[string]interface{}{
					"success":   true,
					"forgotten": true,
					"memory_id": mem.ID,
				})
			}

			fmt.Println("Memory forgotten")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

func newMemoryStatsCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show memory statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			dbConn, queries, err := db.ConnectWithQueries(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer dbConn.Close()

			svc := memory.NewService(queries, nil)

			total, err := svc.Count(cmd.Context(), "")
			if err != nil {
				return fmt.Errorf("failed to count memories: %w", err)
			}

			if jsonOutput {
				return writeJSON(map[string]interface{}{
					"total_active": total,
				})
			}

			// Styles
			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
			labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
			valueStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))

			fmt.Println()
			fmt.Println(headerStyle.Render("  Memory Statistics"))
			fmt.Println(headerStyle.Render("  " + strings.Repeat("-", 30)))
			fmt.Println()
			fmt.Printf("  %s %s\n", labelStyle.Render("Total Active Memories:"), valueStyle.Render(fmt.Sprintf("%d", total)))
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

func newMemoryClearCmd() *cobra.Command {
	var agentFilter string
	var force bool

	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear all memories",
		RunE: func(cmd *cobra.Command, args []string) error {
			dbConn, queries, err := db.ConnectWithQueries(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer dbConn.Close()

			svc := memory.NewService(queries, nil)

			count, err := svc.Count(cmd.Context(), agentFilter)
			if err != nil {
				return fmt.Errorf("failed to count memories: %w", err)
			}

			if count == 0 {
				fmt.Println("No memories to clear")
				return nil
			}

			target := "all memories"
			if agentFilter != "" {
				target = fmt.Sprintf("memories for %s", agentFilter)
			}

			if !force {
				fmt.Printf("This will forget %d %s.\n", count, target)
				fmt.Print("Type 'yes' to confirm: ")
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "yes" {
					fmt.Println("Cancelled")
					return nil
				}
			}

			err = svc.Clear(cmd.Context(), agentFilter)
			if err != nil {
				return fmt.Errorf("failed to clear memories: %w", err)
			}

			fmt.Printf("Cleared %d %s\n", count, target)
			return nil
		},
	}

	cmd.Flags().StringVarP(&agentFilter, "agent", "a", "", "Only clear memories for this agent")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	return cmd
}

func createEmbedder() (embedding.Embedder, error) {
	client := ollama.NewClient()
	if !client.IsAvailable(context.Background()) {
		return nil, fmt.Errorf("Ollama not available at %s", client.Host())
	}

	return embedding.NewOllamaEmbedder(embedding.OllamaConfig{}), nil
}

// writeJSON writes JSON to stdout.
func writeJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// memoryToJSON converts a memory to a JSON-friendly map.
func memoryToJSON(m memory.Memory) map[string]interface{} {
	result := map[string]interface{}{
		"id":           m.ID,
		"content":      m.Content,
		"category":     string(m.Category),
		"status":       string(m.Status),
		"confidence":   m.Confidence,
		"access_count": m.AccessCount,
		"created_at":   m.CreatedAt.Format(time.RFC3339),
		"updated_at":   m.UpdatedAt.Format(time.RFC3339),
	}
	if m.AgentHandle != "" {
		result["agent_handle"] = m.AgentHandle
	}
	if m.PathScope != "" {
		result["path_scope"] = m.PathScope
	}
	if m.SupersedesID != "" {
		result["supersedes_id"] = m.SupersedesID
	}
	if m.SupersededByID != "" {
		result["superseded_by_id"] = m.SupersededByID
	}
	return result
}

// memoriesToJSON converts a slice of memories to JSON-friendly format.
func memoriesToJSON(memories []memory.Memory) []map[string]interface{} {
	result := make([]map[string]interface{}, len(memories))
	for i, m := range memories {
		result[i] = memoryToJSON(m)
	}
	return result
}

// searchResultsToJSON converts search results to JSON-friendly format.
func searchResultsToJSON(results []memory.SearchResult) []map[string]interface{} {
	output := make([]map[string]interface{}, len(results))
	for i, r := range results {
		output[i] = map[string]interface{}{
			"memory":     memoryToJSON(r.Memory),
			"similarity": r.Similarity,
		}
	}
	return output
}

// pickMemory shows an interactive picker for selecting a memory.
func pickMemory(ctx context.Context, svc *memory.Service, title string) (memory.Memory, error) {
	memories, err := svc.List(ctx, "", 50, 0)
	if err != nil {
		return memory.Memory{}, fmt.Errorf("failed to list memories: %w", err)
	}

	if len(memories) == 0 {
		return memory.Memory{}, fmt.Errorf("no memories found")
	}

	// Build options for the select
	options := make([]huh.Option[string], len(memories))
	for i, m := range memories {
		// Truncate content for display
		content := m.Content
		if len(content) > 50 {
			content = content[:47] + "..."
		}
		label := fmt.Sprintf("[%s] %s", m.Category, content)
		options[i] = huh.NewOption(label, m.ID)
	}

	var selectedID string
	err = huh.NewSelect[string]().
		Title(title).
		Options(options...).
		Value(&selectedID).
		WithTheme(huh.ThemeCharm()).
		Run()

	if err != nil {
		return memory.Memory{}, err
	}

	// Find the selected memory
	for _, m := range memories {
		if m.ID == selectedID {
			return m, nil
		}
	}

	return memory.Memory{}, fmt.Errorf("memory not found")
}
