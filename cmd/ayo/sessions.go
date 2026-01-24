package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/embedding"
	"github.com/alexcabrera/ayo/internal/memory"
	"github.com/alexcabrera/ayo/internal/ollama"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/run"
	"github.com/alexcabrera/ayo/internal/session"
	"github.com/alexcabrera/ayo/internal/smallmodel"
	"github.com/alexcabrera/ayo/internal/ui"
)

func newSessionsCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sessions",
		Aliases: []string{"session"},
		Short:   "Manage conversation sessions",
	}

	cmd.AddCommand(newSessionsListCmd())
	cmd.AddCommand(newSessionsShowCmd())
	cmd.AddCommand(newSessionsDeleteCmd())
	cmd.AddCommand(newSessionsContinueCmd(cfgPath))

	return cmd
}

func newSessionsListCmd() *cobra.Command {
	var agentFilter string
	var limit int64

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List conversation sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			services, err := session.Connect(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer services.Close()

			var sessions []session.Session
			if agentFilter != "" {
				sessions, err = services.Sessions.ListByAgent(cmd.Context(), agentFilter, limit)
			} else {
				sessions, err = services.Sessions.List(cmd.Context(), limit)
			}
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			if len(sessions) == 0 {
				fmt.Println("No sessions found")
				return nil
			}

			// Styles
			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
			idStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			agentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141"))
			titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
			countStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
			timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

			fmt.Println()
			fmt.Println(headerStyle.Render("  Sessions"))
			fmt.Println(headerStyle.Render("  " + strings.Repeat("-", 60)))
			fmt.Println()

			for _, s := range sessions {
				// Truncate title if too long
				title := s.Title
				if len(title) > 40 {
					title = title[:37] + "..."
				}

				// Format time
				timeAgo := formatTimeAgo(s.UpdatedAt)

				// Print each session
				fmt.Printf("  %s  %s\n",
					idStyle.Render(s.ID[:8]),
					agentStyle.Render(s.AgentHandle),
				)
				fmt.Printf("    %s  %s  %s\n",
					titleStyle.Render(title),
					countStyle.Render(fmt.Sprintf("(%d msgs)", s.MessageCount)),
					timeStyle.Render(timeAgo),
				)
				fmt.Println()
			}

			fmt.Printf("  %s\n", countStyle.Render(fmt.Sprintf("%d sessions", len(sessions))))
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().StringVarP(&agentFilter, "agent", "a", "", "filter by agent handle")
	cmd.Flags().Int64VarP(&limit, "limit", "n", 20, "maximum number of sessions to show")

	return cmd
}

func newSessionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [session-id]",
		Short: "Show session details and conversation",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			services, err := session.Connect(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer services.Close()

			var sess session.Session

			if len(args) == 0 {
				// No session specified, show selector
				sessions, err := services.Sessions.List(cmd.Context(), 10)
				if err != nil {
					return fmt.Errorf("failed to list sessions: %w", err)
				}

				if len(sessions) == 0 {
					fmt.Println("No sessions found.")
					return nil
				}

				options := make([]huh.Option[string], len(sessions))
				for i, s := range sessions {
					timeAgo := formatTimeAgo(s.UpdatedAt)
					title := s.Title
					if len(title) > 30 {
						title = title[:27] + "..."
					}
					label := fmt.Sprintf("%s  %s  %s  %s",
						s.ID[:8],
						s.AgentHandle,
						title,
						timeAgo,
					)
					options[i] = huh.NewOption(label, s.ID)
				}

				var selectedID string
				err = huh.NewSelect[string]().
					Title("Select a session to show:").
					Options(options...).
					Value(&selectedID).
					Run()
				if err != nil {
					return err
				}

				sess, err = services.Sessions.Get(cmd.Context(), selectedID)
				if err != nil {
					return fmt.Errorf("failed to get session: %w", err)
				}
			} else {
				// Find session by query
				sess, err = findSession(cmd, services, args[0])
				if err != nil {
					return err
				}
			}

			// Get messages
			messages, err := services.Messages.List(cmd.Context(), sess.ID)
			if err != nil {
				return fmt.Errorf("failed to list messages: %w", err)
			}

			// Styles
			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
			labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
			valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

			// Session details
			fmt.Println()
			fmt.Println(headerStyle.Render("  Session Details"))
			fmt.Println(headerStyle.Render("  " + strings.Repeat("─", 60)))
			fmt.Println()
			fmt.Printf("  %s %s\n", labelStyle.Render("ID:"), valueStyle.Render(sess.ID))
			fmt.Printf("  %s %s\n", labelStyle.Render("Agent:"), valueStyle.Render(sess.AgentHandle))
			fmt.Printf("  %s %s\n", labelStyle.Render("Title:"), valueStyle.Render(sess.Title))
			fmt.Printf("  %s %s\n", labelStyle.Render("Messages:"), valueStyle.Render(fmt.Sprintf("%d", sess.MessageCount)))
			fmt.Printf("  %s %s\n", labelStyle.Render("Created:"), valueStyle.Render(formatTime(sess.CreatedAt)))
			fmt.Printf("  %s %s\n", labelStyle.Render("Updated:"), valueStyle.Render(formatTime(sess.UpdatedAt)))
			fmt.Println()

			// Conversation using common renderer
			if len(messages) > 0 {
				fmt.Println(headerStyle.Render("  Conversation"))
				fmt.Println(headerStyle.Render("  " + strings.Repeat("─", 60)))
				fmt.Println()
				output := ui.RenderHistory(messages, sess.AgentHandle)
				fmt.Println(output)
				fmt.Println()
			} else {
				fmt.Println(labelStyle.Render("  No messages in this session."))
				fmt.Println()
			}

			return nil
		},
	}

	return cmd
}

func newSessionsDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <session-id>",
		Short: "Delete a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionQuery := args[0]

			services, err := session.Connect(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer services.Close()

			// Find session by ID or prefix
			sess, err := findSession(cmd, services, sessionQuery)
			if err != nil {
				return err
			}

			// Confirm deletion unless --force
			if !force {
				var confirm bool
				err := huh.NewConfirm().
					Title(fmt.Sprintf("Delete session \"%s\"?", sess.Title)).
					Description(fmt.Sprintf("ID: %s (%d messages)", sess.ID[:8], sess.MessageCount)).
					Affirmative("Delete").
					Negative("Cancel").
					Value(&confirm).
					Run()
				if err != nil {
					return err
				}
				if !confirm {
					fmt.Println("Cancelled")
					return nil
				}
			}

			if err := services.Sessions.Delete(cmd.Context(), sess.ID); err != nil {
				return fmt.Errorf("failed to delete session: %w", err)
			}

			fmt.Printf("Deleted session %s\n", sess.ID[:8])
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "delete without confirmation")

	return cmd
}

func newSessionsContinueCmd(cfgPath *string) *cobra.Command {
	var debug bool

	cmd := &cobra.Command{
		Use:     "continue [session-id]",
		Aliases: []string{"resume"},
		Short:   "Continue a previous conversation session",
		Long: `Continue an interactive chat from a previous session.

If no session ID is provided, shows a list of recent sessions to choose from.
Supports session ID prefix matching and title search.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(*cfgPath)
			if err != nil {
				return err
			}

			services, err := session.Connect(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer services.Close()

			var sess session.Session

			if len(args) == 0 {
				// No session specified, show recent sessions
				sessions, err := services.Sessions.List(cmd.Context(), 10)
				if err != nil {
					return fmt.Errorf("failed to list sessions: %w", err)
				}

				if len(sessions) == 0 {
					fmt.Println("No sessions found. Start a new chat with: ayo @agent")
					return nil
				}

				// Show selector
				options := make([]huh.Option[string], len(sessions))
				for i, s := range sessions {
					timeAgo := formatTimeAgo(s.UpdatedAt)
					title := s.Title
					if len(title) > 30 {
						title = title[:27] + "..."
					}
					label := fmt.Sprintf("%s  %s  %s  %s",
						s.ID[:8],
						s.AgentHandle,
						title,
						timeAgo,
					)
					options[i] = huh.NewOption(label, s.ID)
				}

				var selectedID string
				err = huh.NewSelect[string]().
					Title("Select a session to continue:").
					Options(options...).
					Value(&selectedID).
					Run()
				if err != nil {
					return err
				}

				sess, err = services.Sessions.Get(cmd.Context(), selectedID)
				if err != nil {
					return fmt.Errorf("failed to get session: %w", err)
				}
			} else {
				// Find session by query
				sess, err = findSession(cmd, services, args[0])
				if err != nil {
					return err
				}
			}

			// Load the agent
			ag, err := agent.Load(cfg, sess.AgentHandle)
			if err != nil {
				return fmt.Errorf("failed to load agent %s: %w", sess.AgentHandle, err)
			}

			// Load session messages
			messages, err := services.Messages.List(cmd.Context(), sess.ID)
			if err != nil {
				return fmt.Errorf("failed to load messages: %w", err)
			}

			// Create memory services with Ollama if available
			var embedder embedding.Embedder
			var smallModelSvc *smallmodel.Service
			ollamaClient := ollama.NewClient(ollama.WithHost(cfg.OllamaHost))
			if ollamaClient.IsAvailable(cmd.Context()) {
				embedder = embedding.NewOllamaEmbedder(embedding.OllamaConfig{
					Host:  cfg.OllamaHost,
					Model: cfg.Embedding.Model,
				})
				defer embedder.Close()
				smallModelSvc = smallmodel.NewService(smallmodel.Config{
					Host:  cfg.OllamaHost,
					Model: cfg.SmallModel,
				})
			} else if debug {
				fmt.Fprintf(os.Stderr, "Warning: Ollama not available at %s, memory features disabled\n", cfg.OllamaHost)
			}
			memSvc := memory.NewService(services.Queries(), embedder)
			formSvc := memory.NewFormationService(memSvc)

			// Create runner with services
			runner, err := run.NewRunner(cfg, debug, run.RunnerOptions{
				Services:         services,
				MemoryService:    memSvc,
				FormationService: formSvc,
				SmallModel:       smallModelSvc,
			})
			if err != nil {
				return err
			}

			// Resume the session
			if err := runner.ResumeSession(cmd.Context(), ag, sess.ID, messages); err != nil {
				return fmt.Errorf("failed to resume session: %w", err)
			}

			// Show preview of last 3 messages if there's history
			if len(messages) > 0 {
				preview := ui.RenderHistoryPreview(messages, sess.AgentHandle, 3)
				if preview != "" {
					fmt.Println()
					fmt.Println(preview)
					fmt.Println()
				}
			}

			// Run interactive chat
			return runInteractiveChat(cmd.Context(), runner, ag, debug)
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "show debug output")

	return cmd
}

// findSession finds a session by exact ID, prefix match, or title search.
// If multiple matches are found, prompts user to select one.
func findSession(cmd *cobra.Command, services *session.Services, query string) (session.Session, error) {
	ctx := cmd.Context()

	// Try exact match first
	sess, err := services.Sessions.Get(ctx, query)
	if err == nil {
		return sess, nil
	}

	// Try prefix match
	matches, err := services.Sessions.GetByPrefix(ctx, query)
	if err != nil {
		return session.Session{}, fmt.Errorf("failed to search sessions: %w", err)
	}

	if len(matches) == 0 {
		// Try title search
		matches, err = services.Sessions.Search(ctx, query, 10)
		if err != nil {
			return session.Session{}, fmt.Errorf("failed to search sessions: %w", err)
		}
	}

	if len(matches) == 0 {
		return session.Session{}, fmt.Errorf("no session found matching %q", query)
	}

	if len(matches) == 1 {
		return matches[0], nil
	}

	// Multiple matches: show selector
	options := make([]huh.Option[string], len(matches))
	for i, s := range matches {
		label := fmt.Sprintf("%s - %s (%s)", s.ID[:8], s.Title, s.AgentHandle)
		if len(label) > 60 {
			label = label[:57] + "..."
		}
		options[i] = huh.NewOption(label, s.ID)
	}

	var selectedID string
	err = huh.NewSelect[string]().
		Title("Multiple sessions found. Select one:").
		Options(options...).
		Value(&selectedID).
		Run()
	if err != nil {
		return session.Session{}, err
	}

	return services.Sessions.Get(ctx, selectedID)
}

func formatTimeAgo(unixTime int64) string {
	t := time.Unix(unixTime, 0)
	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	case d < 24*time.Hour:
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("Jan 2, 2006")
	}
}

func formatTime(unixTime int64) string {
	t := time.Unix(unixTime, 0)
	return t.Format("Jan 2, 2006 3:04 PM")
}
