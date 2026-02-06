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
	"github.com/alexcabrera/ayo/internal/cli"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/embedding"
	"github.com/alexcabrera/ayo/internal/memory"
	"github.com/alexcabrera/ayo/internal/ollama"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/run"
	"github.com/alexcabrera/ayo/internal/session"
	"github.com/alexcabrera/ayo/internal/session/jsonl"
	"github.com/alexcabrera/ayo/internal/smallmodel"
	"github.com/alexcabrera/ayo/internal/ui"
)

// Ensure cli package is used
var _ = cli.Output{}

func newSessionsCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sessions",
		Aliases: []string{"session"},
		Short:   "Manage conversation sessions",
		Long: `Manage conversation sessions stored in the local database.

Sessions persist conversation history, allowing you to resume previous
conversations and review past interactions.

Storage: ~/.local/share/ayo/ayo.db`,
	}

	cmd.AddCommand(newSessionsListCmd())
	cmd.AddCommand(newSessionsShowCmd())
	cmd.AddCommand(newSessionsDeleteCmd())
	cmd.AddCommand(newSessionsContinueCmd(cfgPath))
	cmd.AddCommand(newSessionsMigrateCmd())
	cmd.AddCommand(newSessionsReindexCmd())

	return cmd
}

func newSessionsListCmd() *cobra.Command {
	var agentFilter string
	var sourceFilter string
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
			switch {
			case agentFilter != "":
				sessions, err = services.Sessions.ListByAgent(cmd.Context(), agentFilter, limit)
			case sourceFilter != "":
				sessions, err = services.Sessions.ListBySource(cmd.Context(), sourceFilter, limit)
			default:
				sessions, err = services.Sessions.List(cmd.Context(), limit)
			}
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			if len(sessions) == 0 {
				if globalOutput.JSON {
					globalOutput.PrintData([]struct{}{}, "")
					return nil
				}
				if !globalOutput.Quiet {
					fmt.Println("No sessions found")
				}
				return nil
			}

			// JSON output
			if globalOutput.JSON {
				type sessionJSON struct {
					ID           string `json:"id"`
					Agent        string `json:"agent"`
					Title        string `json:"title"`
					MessageCount int64  `json:"message_count"`
					Source       string `json:"source,omitempty"`
					UpdatedAt    int64  `json:"updated_at"`
				}
				var out []sessionJSON
				for _, s := range sessions {
					out = append(out, sessionJSON{
						ID:           s.ID,
						Agent:        s.AgentHandle,
						Title:        s.Title,
						MessageCount: s.MessageCount,
						Source:       s.Source,
						UpdatedAt:    s.UpdatedAt,
					})
				}
				globalOutput.PrintData(out, "")
				return nil
			}

			// Quiet mode: just list IDs
			if globalOutput.Quiet {
				for _, s := range sessions {
					fmt.Println(s.ID)
				}
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

				// Show source indicator for non-ayo sessions
				sourceIndicator := ""
				if s.Source != "" && s.Source != session.SourceAyo {
					sourceIndicator = countStyle.Render(fmt.Sprintf(" [%s]", s.Source))
				}

				// Print each session
				fmt.Printf("  %s  %s\n",
					idStyle.Render(s.ID[:8]),
					agentStyle.Render(s.AgentHandle),
				)
				fmt.Printf("    %s  %s%s  %s\n",
					titleStyle.Render(title),
					countStyle.Render(fmt.Sprintf("(%d msgs)", s.MessageCount)),
					sourceIndicator,
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
	cmd.Flags().StringVarP(&sourceFilter, "source", "s", "", "filter by source (ayo, crush, crush-via-ayo)")
	cmd.Flags().Int64VarP(&limit, "limit", "n", 20, "maximum number of sessions to show")

	return cmd
}

func newSessionsShowCmd() *cobra.Command {
	var latest bool

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
				// No session specified
				sessions, err := services.Sessions.List(cmd.Context(), 10)
				if err != nil {
					return fmt.Errorf("failed to list sessions: %w", err)
				}

				if len(sessions) == 0 {
					fmt.Println("No sessions found.")
					return nil
				}

				if latest {
					// Automatically pick the most recent session
					sess = sessions[0]
				} else {
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

	cmd.Flags().BoolVarP(&latest, "latest", "l", false, "show the most recent session without prompting")

	return cmd
}

func newSessionsDeleteCmd() *cobra.Command {
	var force bool
	var latest bool

	cmd := &cobra.Command{
		Use:   "delete [session-id]",
		Short: "Delete a session",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			services, err := session.Connect(cmd.Context(), paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer services.Close()

			var sess session.Session

			if len(args) == 0 {
				if !latest {
					return fmt.Errorf("session ID required (or use --latest)")
				}
				// Get the most recent session
				sessions, err := services.Sessions.List(cmd.Context(), 1)
				if err != nil {
					return fmt.Errorf("failed to list sessions: %w", err)
				}
				if len(sessions) == 0 {
					return fmt.Errorf("no sessions found")
				}
				sess = sessions[0]
			} else {
				// Find session by ID or prefix
				sess, err = findSession(cmd, services, args[0])
				if err != nil {
					return err
				}
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
	cmd.Flags().BoolVarP(&latest, "latest", "l", false, "delete the most recent session")

	return cmd
}

func newSessionsContinueCmd(cfgPath *string) *cobra.Command {
	var debug bool
	var latest bool

	cmd := &cobra.Command{
		Use:     "continue [session-id]",
		Aliases: []string{"resume"},
		Short:   "Continue a previous conversation session",
		Long: `Continue an interactive chat from a previous session.

If no session ID is provided:
  - With --latest: automatically continues the most recent session
  - Otherwise: shows a list of recent sessions to choose from

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
				// No session specified
				sessions, err := services.Sessions.List(cmd.Context(), 10)
				if err != nil {
					return fmt.Errorf("failed to list sessions: %w", err)
				}

				if len(sessions) == 0 {
					fmt.Println("No sessions found. Start a new chat with: ayo @agent")
					return nil
				}

				if latest {
					// Automatically pick the most recent session
					sess = sessions[0]
				} else {
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
			var memQueue *memory.Queue
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
			
			// Create async memory queue
			memQueue = memory.NewQueue(memSvc, memory.QueueConfig{
				BufferSize: 100,
				OnStatus: func(msg ui.AsyncStatusMsg) {
					switch msg.Status {
					case ui.AsyncStatusInProgress:
						fmt.Fprintf(os.Stderr, "  ◇ %s\n", msg.Message)
					case ui.AsyncStatusCompleted:
						fmt.Fprintf(os.Stderr, "  ◆ %s\n", msg.Message)
					case ui.AsyncStatusFailed:
						fmt.Fprintf(os.Stderr, "  × %s\n", msg.Message)
					}
				},
			})
			memQueue.Start()
			defer memQueue.Stop(5 * time.Second)

			// Create runner with services
			runner, err := run.NewRunner(cfg, debug, run.RunnerOptions{
				Services:         services,
				MemoryService:    memSvc,
				FormationService: formSvc,
				SmallModel:       smallModelSvc,
				MemoryQueue:      memQueue,
				SandboxProvider:  selectSandboxProvider(),
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
	cmd.Flags().BoolVarP(&latest, "latest", "l", false, "continue the most recent session without prompting")

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

func newSessionsMigrateCmd() *cobra.Command {
	var overwrite bool

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate sessions from SQLite to JSONL files",
		Long: `Migrate existing sessions from the SQLite database to JSONL files.

This converts session data to the new file-based storage format.
Each session is stored as a separate JSONL file in:
~/.local/share/ayo/sessions/@agent/YYYY-MM/session-id.jsonl

By default, existing files are skipped. Use --overwrite to replace them.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			services, err := session.Connect(ctx, paths.DatabasePath())
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer services.Close()

			structure := jsonl.NewStructure("")

			fmt.Println("Migrating sessions from SQLite to JSONL files...")

			result, err := jsonl.MigrateFromSQLite(ctx, services.Queries(), structure, overwrite)
			if err != nil {
				return fmt.Errorf("migration failed: %w", err)
			}

			// Print results
			fmt.Printf("\nMigration complete:\n")
			fmt.Printf("  Sessions migrated: %d\n", result.SessionsMigrated)
			fmt.Printf("  Sessions skipped:  %d\n", result.SessionsSkipped)
			fmt.Printf("  Messages migrated: %d\n", result.MessagesMigrated)

			if len(result.Errors) > 0 {
				fmt.Printf("\nErrors encountered (%d):\n", len(result.Errors))
				for _, e := range result.Errors {
					fmt.Printf("  - %s\n", e)
				}
			}

			fmt.Printf("\nSession files stored in: %s\n", structure.Root)

			return nil
		},
	}

	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "overwrite existing session files")

	return cmd
}

func newSessionsReindexCmd() *cobra.Command {
	var fullRebuild bool

	cmd := &cobra.Command{
		Use:   "reindex",
		Short: "Rebuild the session search index",
		Long: `Rebuild the SQLite search index from JSONL session files.

The index is used for fast session listing and searching. It is
derived from session file headers and can be fully rebuilt.

By default, performs an incremental sync (adds new files, removes orphans).
Use --full to completely rebuild the index from scratch.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			structure := jsonl.NewStructure("")

			if !structure.Exists() {
				fmt.Println("No session files found. Nothing to index.")
				return nil
			}

			idx, err := jsonl.OpenIndex(structure)
			if err != nil {
				return fmt.Errorf("open index: %w", err)
			}
			defer idx.Close()

			if fullRebuild {
				fmt.Println("Rebuilding session index from scratch...")
				result, err := idx.Rebuild()
				if err != nil {
					return fmt.Errorf("rebuild: %w", err)
				}

				fmt.Printf("\nIndex rebuilt:\n")
				fmt.Printf("  Sessions indexed: %d\n", result.Indexed)
				if len(result.Errors) > 0 {
					fmt.Printf("  Errors: %d\n", len(result.Errors))
					for _, e := range result.Errors {
						fmt.Printf("    - %s\n", e)
					}
				}
			} else {
				fmt.Println("Syncing session index...")
				result, err := idx.Sync()
				if err != nil {
					return fmt.Errorf("sync: %w", err)
				}

				fmt.Printf("\nIndex synced:\n")
				fmt.Printf("  Sessions added:   %d\n", result.Added)
				fmt.Printf("  Sessions removed: %d\n", result.Removed)
				if len(result.Errors) > 0 {
					fmt.Printf("  Errors: %d\n", len(result.Errors))
					for _, e := range result.Errors {
						fmt.Printf("    - %s\n", e)
					}
				}
			}

			count, _ := idx.Count()
			fmt.Printf("\nTotal indexed sessions: %d\n", count)
			fmt.Printf("Index path: %s\n", structure.IndexDB)

			return nil
		},
	}

	cmd.Flags().BoolVar(&fullRebuild, "full", false, "fully rebuild index (not incremental)")

	return cmd
}
