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

	"github.com/alexcabrera/ayo/internal/cli"
	"github.com/alexcabrera/ayo/internal/db"
	"github.com/alexcabrera/ayo/internal/embedding"
	"github.com/alexcabrera/ayo/internal/memory"
	"github.com/alexcabrera/ayo/internal/memory/zettelkasten"
	"github.com/alexcabrera/ayo/internal/ollama"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/smallmodel"
	"github.com/alexcabrera/ayo/internal/ui"
)

// Ensure cli package is used
var _ = cli.Output{}

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
	cmd.AddCommand(newMemoryReindexCmd())
	cmd.AddCommand(newMemoryTopicsCmd())
	cmd.AddCommand(newMemoryLinkCmd())
	cmd.AddCommand(newMemoryMergeCmd())
	cmd.AddCommand(newMemoryMigrateCmd())
	cmd.AddCommand(newMemoryExportCmd())
	cmd.AddCommand(newMemoryImportCmd())

	return cmd
}

func newMemoryListCmd() *cobra.Command {
	var agentFilter string
	var categoryFilter string
	var limit int64

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

			if globalOutput.JSON {
				return writeJSON(memoriesToJSON(memories))
			}

			if len(memories) == 0 {
				if !globalOutput.Quiet {
					fmt.Println("No memories found")
				}
				return nil
			}

			// Quiet mode: just list IDs
			if globalOutput.Quiet {
				for _, m := range memories {
					fmt.Println(m.ID)
				}
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

	return cmd
}

func newMemorySearchCmd() *cobra.Command {
	var agentFilter string
	var threshold float64
	var limit int

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

			if globalOutput.JSON {
				return writeJSON(searchResultsToJSON(results))
			}

			if len(results) == 0 {
				if !globalOutput.Quiet {
					fmt.Println("No memories found matching the query")
				}
				return nil
			}

			// Quiet mode: just list IDs
			if globalOutput.Quiet {
				for _, r := range results {
					fmt.Println(r.Memory.ID)
				}
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

	return cmd
}

func newMemoryShowCmd() *cobra.Command {
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

			if globalOutput.JSON {
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
					return writeJSON(map[string]any{
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
				return writeJSON(map[string]any{
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
						return writeJSON(map[string]any{
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
					return writeJSON(map[string]any{
						"success": false,
						"error":   err.Error(),
					})
				}
				return fmt.Errorf("failed to forget memory: %w", err)
			}

			if jsonOutput {
				return writeJSON(map[string]any{
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
				return writeJSON(map[string]any{
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

func newMemoryReindexCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "reindex",
		Short: "Rebuild the memory search index from source files",
		Long: `Rebuild the .index.sqlite file from memory markdown files.

This command scans all memory files in the zettelkasten directory and
rebuilds the SQLite search index. The index is derived and can always
be rebuilt from the source files.

Use this if:
- The index becomes corrupted
- You manually edited memory files
- After restoring memory files from backup`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Show spinner unless JSON output
			var spinner *ui.Spinner
			if !jsonOutput {
				spinner = ui.NewSpinnerWithType("rebuilding memory index...", ui.SpinnerMemory)
				spinner.Start()
			}

			// Get memory directory
			memDir := zettelkasten.DefaultMemoryDir()

			// Create structure and index
			structure := zettelkasten.NewStructure(memDir)
			if !structure.Exists() {
				if spinner != nil {
					spinner.StopWithMessage("no memory directory found")
				}
				if jsonOutput {
					return writeJSON(map[string]any{
						"success": false,
						"error":   "no memory directory found",
					})
				}
				fmt.Println("No memory directory found. Nothing to reindex.")
				return nil
			}

			// Count files first
			files, err := structure.ListAllMemories()
			if err != nil {
				if spinner != nil {
					spinner.StopWithError("failed to list memory files")
				}
				return fmt.Errorf("failed to list memory files: %w", err)
			}

			if len(files) == 0 {
				if spinner != nil {
					spinner.StopWithMessage("no memory files found")
				}
				if jsonOutput {
					return writeJSON(map[string]any{
						"success": true,
						"count":   0,
						"message": "no memory files found",
					})
				}
				fmt.Println("No memory files found. Index is empty.")
				return nil
			}

			// Open index and rebuild
			idx := zettelkasten.NewIndex(structure)
			if err := idx.Open(ctx); err != nil {
				if spinner != nil {
					spinner.StopWithError("failed to open index")
				}
				return fmt.Errorf("failed to open index: %w", err)
			}
			defer idx.Close()

			if err := idx.Rebuild(ctx); err != nil {
				if spinner != nil {
					spinner.StopWithError("failed to rebuild index")
				}
				return fmt.Errorf("failed to rebuild index: %w", err)
			}

			// Get final count
			count, _ := idx.Count(ctx)

			if spinner != nil {
				spinner.StopWithMessage(fmt.Sprintf("reindexed %d memories", count))
			}

			if jsonOutput {
				return writeJSON(map[string]any{
					"success": true,
					"count":   count,
				})
			}

			fmt.Printf("Successfully reindexed %d memories\n", count)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

func newMemoryTopicsCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "topics",
		Short: "List all memory topics",
		Long: `List all topics that have been assigned to memories.

Topics help organize memories by subject matter. Memories can have
multiple topics, and topics are automatically created when storing
memories with the --topics flag or via the zettelkasten file format.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get memory directory
			memDir := zettelkasten.DefaultMemoryDir()

			// Create structure
			structure := zettelkasten.NewStructure(memDir)
			if !structure.Exists() {
				if jsonOutput {
					return writeJSON(map[string]any{
						"topics": []string{},
					})
				}
				fmt.Println("No topics found")
				return nil
			}

			// List topics
			topics, err := structure.ListTopics()
			if err != nil {
				return fmt.Errorf("failed to list topics: %w", err)
			}

			if jsonOutput {
				return writeJSON(map[string]any{
					"topics": topics,
				})
			}

			if len(topics) == 0 {
				fmt.Println("No topics found")
				return nil
			}

			// Styles
			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
			topicStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141"))

			fmt.Println()
			fmt.Println(headerStyle.Render("  Memory Topics"))
			fmt.Println(headerStyle.Render("  " + strings.Repeat("-", 30)))
			fmt.Println()

			for _, topic := range topics {
				fmt.Printf("  %s\n", topicStyle.Render(topic))
			}
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

func newMemoryLinkCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "link <id1> <id2>",
		Short: "Create a bidirectional link between two memories",
		Long: `Create a bidirectional link between two memories.

Links establish relationships between related memories, forming a
knowledge graph. Links are symmetric: if A links to B, B links to A.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id1, id2 := args[0], args[1]

			// Get memory directory
			memDir := zettelkasten.DefaultMemoryDir()

			// Create and initialize provider
			provider := zettelkasten.NewProvider()
			if err := provider.Init(cmd.Context(), map[string]any{"root": memDir}); err != nil {
				return fmt.Errorf("failed to initialize memory provider: %w", err)
			}
			defer provider.Close()

			// Create the link
			if err := provider.Link(cmd.Context(), id1, id2); err != nil {
				if jsonOutput {
					return writeJSON(map[string]any{
						"error": err.Error(),
					})
				}
				return fmt.Errorf("failed to link memories: %w", err)
			}

			if jsonOutput {
				return writeJSON(map[string]any{
					"linked": []string{id1, id2},
				})
			}

			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			fmt.Printf("%s Linked %s <-> %s\n", successStyle.Render("✓"), id1, id2)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

func newMemoryMergeCmd() *cobra.Command {
	var dryRun bool
	var jsonOutput bool
	var threshold float64
	var unclearThreshold float64

	cmd := &cobra.Command{
		Use:   "merge",
		Short: "Find and merge similar memories",
		Long: `Find and merge similar memories automatically.

Similar memories are merged into one, with the newer memory becoming
the keeper and older memories marked as superseded. Memories that are
similar but not identical are flagged for clarification.

Thresholds:
  --threshold       Similarity above which memories are merged (default: 0.90)
  --unclear         Similarity range for flagging (default: 0.75)
                    Memories with similarity between unclear and threshold
                    are flagged for user clarification.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get memory directory
			memDir := zettelkasten.DefaultMemoryDir()

			// Create and initialize provider
			provider := zettelkasten.NewProvider()
			if err := provider.Init(cmd.Context(), map[string]any{"root": memDir}); err != nil {
				return fmt.Errorf("failed to initialize memory provider: %w", err)
			}
			defer provider.Close()

			// Create structure and index
			structure := zettelkasten.NewStructure(memDir)
			if !structure.Exists() {
				if jsonOutput {
					return writeJSON(map[string]any{
						"merged":  0,
						"flagged": 0,
						"linked":  0,
					})
				}
				fmt.Println("No memories found")
				return nil
			}

			idx := zettelkasten.NewIndex(structure)
			if err := idx.Open(cmd.Context()); err != nil {
				return fmt.Errorf("failed to open index: %w", err)
			}
			defer idx.Close()

			// Create merger config
			config := zettelkasten.DefaultMergeConfig()
			config.DryRun = dryRun
			if threshold > 0 {
				config.SimilarityThreshold = float32(threshold)
			}
			if unclearThreshold > 0 {
				config.UnclearThreshold = float32(unclearThreshold)
			}

			merger := zettelkasten.NewMerger(provider, idx, nil, config)

			// Find candidates
			candidates, err := merger.FindCandidates(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to find candidates: %w", err)
			}

			if len(candidates) == 0 {
				if jsonOutput {
					return writeJSON(map[string]any{
						"merged":     0,
						"flagged":    0,
						"linked":     0,
						"candidates": []any{},
					})
				}
				fmt.Println("No merge candidates found")
				return nil
			}

			// Execute merge
			result, err := merger.Execute(cmd.Context(), candidates)
			if err != nil {
				return fmt.Errorf("failed to execute merge: %w", err)
			}

			if jsonOutput {
				candidatesJSON := make([]map[string]any, len(result.Candidates))
				for i, c := range result.Candidates {
					candidatesJSON[i] = map[string]any{
						"memory_a":   c.MemoryA.Frontmatter.ID,
						"memory_b":   c.MemoryB.Frontmatter.ID,
						"similarity": c.Similarity,
						"action":     string(c.Action),
						"reason":     c.Reason,
					}
				}
				errStrings := make([]string, len(result.Errors))
				for i, e := range result.Errors {
					errStrings[i] = e.Error()
				}
				return writeJSON(map[string]any{
					"merged":     result.Merged,
					"flagged":    result.FlaggedAsUnclear,
					"linked":     result.Linked,
					"candidates": candidatesJSON,
					"errors":     errStrings,
					"dry_run":    dryRun,
				})
			}

			// Print results
			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

			if dryRun {
				fmt.Println(warnStyle.Render("Dry run - no changes made"))
				fmt.Println()
			}

			fmt.Printf("  Merged:  %d\n", result.Merged)
			fmt.Printf("  Flagged: %d\n", result.FlaggedAsUnclear)
			fmt.Printf("  Linked:  %d\n", result.Linked)

			if len(result.Errors) > 0 {
				fmt.Println()
				fmt.Println(warnStyle.Render("Errors:"))
				for _, e := range result.Errors {
					fmt.Printf("  - %s\n", e.Error())
				}
			}

			if result.Merged > 0 || result.FlaggedAsUnclear > 0 || result.Linked > 0 {
				fmt.Println()
				fmt.Println(successStyle.Render("✓ Memory merge complete"))
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate merge without making changes")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().Float64Var(&threshold, "threshold", 0, "Similarity threshold for auto-merge (default: 0.90)")
	cmd.Flags().Float64Var(&unclearThreshold, "unclear", 0, "Similarity threshold for flagging unclear (default: 0.75)")

	return cmd
}

func newMemoryMigrateCmd() *cobra.Command {
	var overwrite bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate memories from SQLite to Zettelkasten files",
		Long: `Migrate existing memories from the SQLite database to Zettelkasten markdown files.

This is a one-way migration that:
- Reads all memories from the SQLite database
- Creates corresponding markdown files in the zettelkasten directory
- Preserves all metadata including timestamps, categories, and relationships

By default, existing files are skipped. Use --overwrite to replace them.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Connect to SQLite database
			dbConn, queries, err := db.ConnectWithQueries(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer dbConn.Close()

			// Get memory directory
			memDir := zettelkasten.DefaultMemoryDir()

			// Create and initialize provider
			provider := zettelkasten.NewProvider()
			if err := provider.Init(cmd.Context(), map[string]any{"root": memDir}); err != nil {
				return fmt.Errorf("failed to initialize memory provider: %w", err)
			}
			defer provider.Close()

			// Perform migration
			result, err := zettelkasten.MigrateFromSQLite(cmd.Context(), queries, provider, overwrite)
			if err != nil {
				return fmt.Errorf("migration failed: %w", err)
			}

			if jsonOutput {
				return writeJSON(map[string]any{
					"migrated": result.Migrated,
					"skipped":  result.Skipped,
					"failed":   result.Failed,
					"errors":   result.Errors,
				})
			}

			// Print results
			successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

			fmt.Printf("  Migrated: %d\n", result.Migrated)
			fmt.Printf("  Skipped:  %d\n", result.Skipped)
			fmt.Printf("  Failed:   %d\n", result.Failed)

			if len(result.Errors) > 0 {
				fmt.Println()
				fmt.Println(warnStyle.Render("Errors:"))
				for _, e := range result.Errors {
					fmt.Printf("  - %s\n", e)
				}
			}

			if result.Migrated > 0 {
				fmt.Println()
				fmt.Println(successStyle.Render("✓ Migration complete"))
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing files")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

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
func writeJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// memoryToJSON converts a memory to a JSON-friendly map.
func memoryToJSON(m memory.Memory) map[string]any {
	result := map[string]any{
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
func memoriesToJSON(memories []memory.Memory) []map[string]any {
	result := make([]map[string]any, len(memories))
	for i, m := range memories {
		result[i] = memoryToJSON(m)
	}
	return result
}

// searchResultsToJSON converts search results to JSON-friendly format.
func searchResultsToJSON(results []memory.SearchResult) []map[string]any {
	output := make([]map[string]any, len(results))
	for i, r := range results {
		output[i] = map[string]any{
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

func newMemoryExportCmd() *cobra.Command {
	var agentFilter string
	var includeEmbeddings bool
	var sinceStr string

	cmd := &cobra.Command{
		Use:   "export <file>",
		Short: "Export memories to a JSON file",
		Long: `Export memories to a JSON file for backup or transfer.

Examples:
  ayo memory export memories.json
  ayo memory export --agent @ayo agent-memories.json
  ayo memory export --include-embeddings full-backup.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			dbConn, queries, err := db.ConnectWithQueries(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}
			defer dbConn.Close()

			svc := memory.NewService(queries, nil)

			opts := memory.ExportOptions{
				AgentHandle:       agentFilter,
				IncludeEmbeddings: includeEmbeddings,
			}

			if sinceStr != "" {
				since, err := time.Parse(time.RFC3339, sinceStr)
				if err != nil {
					return fmt.Errorf("invalid since time: %w", err)
				}
				opts.Since = since
			}

			data, err := svc.Export(cmd.Context(), opts)
			if err != nil {
				return fmt.Errorf("export memories: %w", err)
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(data)
			}

			file, err := os.Create(filePath)
			if err != nil {
				return fmt.Errorf("create file: %w", err)
			}
			defer file.Close()

			enc := json.NewEncoder(file)
			enc.SetIndent("", "  ")
			if err := enc.Encode(data); err != nil {
				return fmt.Errorf("write file: %w", err)
			}

			fmt.Printf("Exported %d memories to %s\n", len(data.Memories), filePath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&agentFilter, "agent", "a", "", "Filter by agent")
	cmd.Flags().BoolVar(&includeEmbeddings, "include-embeddings", false, "Include embedding vectors")
	cmd.Flags().StringVar(&sinceStr, "since", "", "Export memories created after (RFC3339 format)")

	return cmd
}

func newMemoryImportCmd() *cobra.Command {
	var merge bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import memories from a JSON file",
		Long: `Import memories from a JSON file.

Examples:
  ayo memory import memories.json
  ayo memory import --merge memories.json  # Don't overwrite existing
  ayo memory import --dry-run memories.json  # Show what would be imported`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			file, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("open file: %w", err)
			}
			defer file.Close()

			var data memory.ExportData
			if err := json.NewDecoder(file).Decode(&data); err != nil {
				return fmt.Errorf("parse file: %w", err)
			}

			dbConn, queries, err := db.ConnectWithQueries(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}
			defer dbConn.Close()

			svc := memory.NewService(queries, nil)

			opts := memory.ImportOptions{
				Merge:  merge,
				DryRun: dryRun,
			}

			result, err := svc.Import(cmd.Context(), &data, opts)
			if err != nil {
				return fmt.Errorf("import memories: %w", err)
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(result)
			}

			if dryRun {
				fmt.Printf("Would import %d memories, skip %d\n", result.Imported, result.Skipped)
			} else {
				fmt.Printf("Imported %d memories, skipped %d\n", result.Imported, result.Skipped)
			}
			if len(result.Errors) > 0 {
				fmt.Printf("Errors: %d\n", len(result.Errors))
				for _, e := range result.Errors {
					fmt.Printf("  %s\n", e)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&merge, "merge", false, "Don't overwrite existing memories")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be imported without making changes")

	return cmd
}
