package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/db"
	"github.com/alexcabrera/ayo/internal/embedding"
	"github.com/alexcabrera/ayo/internal/memory"
	"github.com/alexcabrera/ayo/internal/paths"
)

func newMemoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Manage agent memories",
	}

	cmd.AddCommand(newMemoryListCmd())
	cmd.AddCommand(newMemorySearchCmd())
	cmd.AddCommand(newMemoryShowCmd())
	cmd.AddCommand(newMemoryForgetCmd())
	cmd.AddCommand(newMemoryStatsCmd())
	cmd.AddCommand(newMemoryClearCmd())

	return cmd
}

func getMemoryService(ctx interface{ Context() interface{} }) (*memory.Service, *db.Queries, *sql.DB, error) {
	dbConn, queries, err := db.ConnectWithQueries(ctx.(interface{ Context() interface{} }).Context().(interface {
		Done() <-chan struct{}
		Err() error
		Deadline() (time.Time, bool)
		Value(interface{}) interface{}
	}), paths.DatabasePath())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Use a nil embedder for CLI operations (no embedding generation needed for list/show/etc.)
	svc := memory.NewService(queries, nil)
	return svc, queries, dbConn, nil
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

			if len(memories) == 0 {
				fmt.Println("No memories found")
				return nil
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
	cmd.Flags().Float64VarP(&threshold, "threshold", "t", 0.5, "Minimum similarity threshold (0-1)")
	cmd.Flags().IntVarP(&limit, "limit", "n", 10, "Maximum number of results")

	return cmd
}

func newMemoryShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show memory details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			dbConn, queries, err := db.ConnectWithQueries(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer dbConn.Close()

			svc := memory.NewService(queries, nil)

			mem, err := svc.GetByPrefix(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("memory not found: %w", err)
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

func newMemoryForgetCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "forget <id>",
		Short: "Forget a memory (soft delete)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			dbConn, queries, err := db.ConnectWithQueries(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer dbConn.Close()

			svc := memory.NewService(queries, nil)

			// Show the memory first
			mem, err := svc.GetByPrefix(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("memory not found: %w", err)
			}

			if !force {
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
				return fmt.Errorf("failed to forget memory: %w", err)
			}

			fmt.Println("Memory forgotten")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	return cmd
}

func newMemoryStatsCmd() *cobra.Command {
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
	if !embedding.IsModelAvailable() {
		return nil, fmt.Errorf("embedding model not available")
	}

	return embedding.NewLocalEmbedder(embedding.LocalConfig{})
}
