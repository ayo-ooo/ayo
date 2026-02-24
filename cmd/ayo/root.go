package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/builtin"
	"github.com/alexcabrera/ayo/internal/cli"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/daemon"
	"github.com/alexcabrera/ayo/internal/debug"
	"github.com/alexcabrera/ayo/internal/embedding"
	"github.com/alexcabrera/ayo/internal/memory"
	"github.com/alexcabrera/ayo/internal/ollama"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/pipe"
	"github.com/alexcabrera/ayo/internal/planners"
	"github.com/alexcabrera/ayo/internal/run"
	"github.com/alexcabrera/ayo/internal/session"
	"github.com/alexcabrera/ayo/internal/share"
	"github.com/alexcabrera/ayo/internal/smallmodel"
	"github.com/alexcabrera/ayo/internal/squads"
	"github.com/alexcabrera/ayo/internal/ui"
)

// Global output flags accessible to all subcommands
var globalOutput cli.Output

// Global no-jodas flag (auto-approve file modifications)
var globalNoJodas bool

func newRootCmd() *cobra.Command {
	var cfgPath string
	var attachments []string
	var debugFlag bool
	var modelOverride string
	var sessionID string
	var continueSession bool
	var outputDir string

	cmd := &cobra.Command{
		Use:   "ayo [@agent | #squad] [prompt]",
		Short: "AI agents that live on your machine",
		Long: `ayo - AI agents that live on your machine

Run AI agents in isolated sandboxes with tool access and Unix pipe integration.

Usage:
  ayo [prompt]              Chat with @ayo (default agent)
  ayo @agent [prompt]       Chat with specific agent
  ayo #squad [prompt]       Send task to squad (multi-agent coordination)

Common Commands:
  agent                     Manage agents (create, list, show, delete)
  squad                     Manage squads (create, list, destroy)
  trigger                   Manage triggers (schedule, watch, list)
  service                   Control background service (start, stop, status)
  doctor                    Check system health

Flags:
  -y, --no-jodas            Auto-approve file modifications
  -q, --quiet               Suppress non-essential output
      --json                Output in JSON format
  -a, --attach FILE         Attach file to prompt
  -c, --continue            Continue most recent session
  -s, --session ID          Continue specific session

Examples:
  ayo "explain this code"              Chat with @ayo
  ayo @reviewer "review my changes"    Chat with @reviewer agent
  ayo #frontend "build auth feature"   Send to frontend squad
  ayo -a main.go "fix the bug"         Attach file to prompt
  ayo -c "also add tests"              Continue last session`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ArbitraryArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// Only complete first argument (handle)
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completeHandles(toComplete)
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Enable debug logging if --debug flag is set
			if debugFlag {
				debug.SetEnabled(true)
				debug.Log("debug mode enabled")
			}

			// Load stored credentials into environment
			if err := config.InjectCredentials(); err != nil {
				// Non-fatal: just log in debug mode
				debug.Log("failed to load credentials", "error", err)
			}

			// Auto-install built-in agents and skills if needed (version-based)
			return builtin.Install()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConfig(&cfgPath, func(cfg config.Config) error {
				// Check for first-run (no providers configured)
				if !config.HasAnyProvider() {
					warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
					fmt.Fprintln(os.Stderr, warnStyle.Render("No API providers configured. Run 'ayo setup' to configure."))
					fmt.Fprintln(os.Stderr)
				}

				// Apply model override from flag
				if modelOverride != "" {
					cfg.DefaultModel = modelOverride
				}

				// Determine agent handle and remaining args
				var handle string
				var promptArgs []string

				if len(args) == 0 {
					// No args: use default agent, no prompt (interactive mode)
					handle = agent.DefaultAgent
					promptArgs = nil
				} else if strings.HasPrefix(args[0], "@") {
					// First arg is an agent handle
					handle = agent.NormalizeHandle(args[0])
					promptArgs = args[1:]
				} else if squads.IsSquadHandle(args[0]) {
					// First arg is a squad handle (#squad-name)
					squadHandle := squads.NormalizeHandle(args[0])
					if !squads.ValidateHandle(squadHandle) {
						return fmt.Errorf("invalid squad handle: %s", args[0])
					}
					prompt := strings.Join(args[1:], " ")
					return invokeSquad(cmd.Context(), squadHandle, prompt)
				} else {
					// First arg is not an agent handle
					// Check if it looks like a potential subcommand typo
					if looksLikeSubcommand(args[0]) {
						return fmt.Errorf("unknown command %q\n\nTo send a prompt to an agent, use quotes:\n  ayo \"your prompt here\"\n\nFor available commands, run: ayo --help", args[0])
					}
					// Use default agent with all args as prompt
					handle = agent.DefaultAgent
					promptArgs = args
				}

				ag, err := agent.Load(cfg, handle)
				if err != nil {
					return err
				}

				// Initialize session services
				services, err := session.Connect(cmd.Context(), paths.DatabasePath())
				if err != nil {
					// Log warning but continue without persistence
					debug.Log("session persistence unavailable", "error", err)
					services = nil
				}
				if services != nil {
					defer services.Close()
				}

				// Create memory services if database available
				var memSvc *memory.Service
				var formSvc *memory.FormationService
				var smallModelSvc *smallmodel.Service
				var memQueue *memory.Queue
				if services != nil {
					// Create Ollama-based embedder and small model service
					var embedder embedding.Embedder
					ollamaClient := ollama.NewClient(ollama.WithHost(cfg.OllamaHost))
					if ollamaClient.IsAvailable(cmd.Context()) {
						embedder = embedding.NewOllamaEmbedder(embedding.OllamaConfig{
							Host:  cfg.OllamaHost,
							Model: cfg.Embedding.Model,
						})
						smallModelSvc = smallmodel.NewService(smallmodel.Config{
							Host:  cfg.OllamaHost,
							Model: cfg.SmallModel,
						})
					} else {
						debug.Log("Ollama not available, memory features disabled", "host", cfg.OllamaHost)
					}
					memSvc = memory.NewService(services.Queries(), embedder)
					if embedder != nil {
						defer embedder.Close()
					}
					formSvc = memory.NewFormationService(memSvc)
					
					// Create async memory queue
					memQueue = memory.NewQueue(memSvc, memory.QueueConfig{
						BufferSize: 100,
						OnStatus: func(msg ui.AsyncStatusMsg) {
							// For now, just print to stderr - will be wired to TUI later
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
					
					// Register callback for memory formation feedback
					formSvc.OnFormation(func(result memory.FormationResult) {
						var msg string
						switch result.EventType() {
						case memory.FormationEventCreated:
							msg = "  ◆ Remembered"
						case memory.FormationEventSkipped:
							msg = "  ◇ Already remembered"
						case memory.FormationEventSuperseded:
							msg = "  ◆ Memory updated"
						case memory.FormationEventFailed:
							msg = "  × Failed to remember"
						default:
							return
						}
						fmt.Fprintln(os.Stderr, msg)
					})
				}

				// Create share service for request_access tool
				shareSvc := share.NewService()
				if err := shareSvc.Load(); err != nil {
					debug.Log("Failed to load share service", "error", err)
				}

				// Create planner manager for per-sandbox planners
				plannerMgr := planners.NewSandboxPlannerManager(nil, cfg)

				runner, err := run.NewRunner(cfg, debugFlag, run.RunnerOptions{
					Services:         services,
					MemoryService:    memSvc,
					FormationService: formSvc,
					SmallModel:       smallModelSvc,
					MemoryQueue:      memQueue,
					SandboxProvider:  selectSandboxProvider(),
					ShareService:     shareSvc,
					PlannerManager:   plannerMgr,
				})
				if err != nil {
					return err
				}

				// Non-interactive mode: prompt provided as positional args or stdin
				if len(promptArgs) > 0 || pipe.IsStdinPiped() {
					var prompt string

					if pipe.IsStdinPiped() {
						// Read from stdin
						stdinData, err := pipe.ReadStdin()
						if err != nil {
							return fmt.Errorf("read stdin: %w", err)
						}
						stdinData = strings.TrimSpace(stdinData)

						if ag.HasInputSchema() {
							// Agent has input schema: stdin must be valid JSON matching schema
							if err := ag.ValidateInput(stdinData); err != nil {
								return printInputValidationError(err)
							}
							prompt = stdinData
						} else {
							// Agent has no input schema: build preamble with context
							prompt = buildFreeformPreamble(stdinData)
						}

						// If there are also positional args, append them
						if len(promptArgs) > 0 {
							prompt = prompt + "\n\n" + strings.Join(promptArgs, " ")
						}
					} else {
						// No stdin, use positional args
						prompt = strings.Join(promptArgs, " ")

						// Validate input against schema if agent has one
						if err := ag.ValidateInput(prompt); err != nil {
							return printInputValidationError(err)
						}
					}

					ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
					defer cancel()

					var result run.TextResult
					
					// Determine session ID for continuation
					effectiveSessionID := sessionID
					if continueSession && effectiveSessionID == "" {
						// --continue without -s: use the latest session
						latestSess, err := getLatestSession(cmd.Context(), services)
						if err != nil {
							return fmt.Errorf("no sessions found to continue: %w", err)
						}
						effectiveSessionID = latestSess.ID
					}
					
					if effectiveSessionID != "" {
						// Continue existing session
						result, err = runner.ContinueSessionWithPrompt(ctx, ag, effectiveSessionID, prompt, attachments)
					} else {
						// Start new session
						result, err = runner.TextWithSession(ctx, ag, prompt, attachments)
					}
					if err != nil {
						return err
					}

					// Wait for any pending memory formations to complete
					runner.WaitForFormations(2 * time.Second)

					// Output to stdout (for piping)
					fmt.Println(result.Response)

					// Print session ID to stderr (visible even when piped)
					if result.SessionID != "" && !pipe.IsStdoutPiped() {
						sessionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
						fmt.Fprintln(os.Stderr, sessionStyle.Render(fmt.Sprintf("\nSession: %s", result.SessionID)))
					}
					return nil
				}

				// Interactive mode
				return runInteractiveChat(cmd.Context(), runner, ag, debugFlag)
			})
		},
	}

	cmd.PersistentFlags().StringVar(&cfgPath, "config", defaultConfigPath(), "path to config file")
	cmd.PersistentFlags().BoolVar(&globalOutput.JSON, "json", false, "output in JSON format")
	cmd.PersistentFlags().BoolVarP(&globalOutput.Quiet, "quiet", "q", false, "minimal output, suppress informational messages")
	cmd.PersistentFlags().BoolVarP(&globalNoJodas, "no-jodas", "y", false, "auto-approve all file modifications (WARNING: use with caution)")
	cmd.Flags().StringSliceVarP(&attachments, "attachment", "a", nil, "file attachments")
	cmd.Flags().BoolVar(&debugFlag, "debug", false, "show debug output including raw tool payloads")
	cmd.Flags().StringVarP(&modelOverride, "model", "m", "", "model to use (overrides config default)")
	cmd.Flags().BoolVarP(&continueSession, "continue", "c", false, "continue the most recent session")
	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "continue a specific session by ID")
	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "target directory for work products (default: current directory)")

	// Make outputDir accessible globally for squad operations
	_ = outputDir // Used by squad commands via flag lookup

	// Subcommands
	cmd.AddCommand(newSetupCmd(&cfgPath))
	cmd.AddCommand(newAgentsCmd(&cfgPath))
	cmd.AddCommand(newSkillsCmd(&cfgPath))
	cmd.AddCommand(newFlowsCmd(&cfgPath))
	cmd.AddCommand(newChainCmd(&cfgPath))
	cmd.AddCommand(newSessionsCmd(&cfgPath))
	cmd.AddCommand(newMemoryCmd())
	cmd.AddCommand(newDoctorCmd(&cfgPath))
	cmd.AddCommand(newPluginsCmd(&cfgPath))
	cmd.AddCommand(newSandboxCmd(&cfgPath))
	cmd.AddCommand(newShareCmd())
	cmd.AddCommand(newBackupCmd())
	cmd.AddCommand(newSyncCmd())
	cmd.AddCommand(newTriggerCmd())
	cmd.AddCommand(newTicketCmd())
	cmd.AddCommand(newSquadCmd())
	cmd.AddCommand(newPlannerCmd(&cfgPath))
	cmd.AddCommand(newIndexCmd(&cfgPath))
	cmd.AddCommand(auditCmd)

	// Hidden backwards-compat alias: `ayo daemon` -> `ayo sandbox service`
	cmd.AddCommand(newDaemonAliasCmd(&cfgPath))

	return cmd
}

func defaultConfigPath() string {
	return paths.ConfigFile()
}

func loadConfig(cfgPath string) (config.Config, error) {
	return config.Load(cfgPath)
}

func withConfig(cfgPath *string, fn func(config.Config) error) error {
	cfg, err := loadConfig(*cfgPath)
	if err != nil {
		return err
	}
	return fn(cfg)
}

// printInputValidationError prints a formatted input validation error to stderr.
// Returns a simple error to signal failure without duplicating the message.
func printInputValidationError(err error) error {
	// Check if it's our custom InputValidationError
	var validationErr *agent.InputValidationError
	if errors.As(err, &validationErr) {
		// Style the error
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
		codeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, errorStyle.Render("  ERROR: Agent requires structured JSON input"))
		fmt.Fprintln(os.Stderr)

		// The error message from InputValidationError already has the schema info
		errMsg := validationErr.Error()
		for _, line := range strings.Split(errMsg, "\n") {
			if strings.HasPrefix(line, "Expected format:") || strings.HasPrefix(line, "Your input") || strings.HasPrefix(line, "This agent") || strings.HasPrefix(line, "Validation error:") {
				fmt.Fprintln(os.Stderr, "  "+headerStyle.Render(line))
			} else {
				fmt.Fprintln(os.Stderr, "  "+codeStyle.Render(line))
			}
		}
		fmt.Fprintln(os.Stderr)
		return errors.New("input validation failed")
	}
	return err
}

// formatInputValidationError creates a detailed error message for input validation failures.
// buildFreeformPreamble creates a preamble for agents without input schemas
// when receiving piped input from another agent.
func buildFreeformPreamble(jsonInput string) string {
	ctx := pipe.GetChainContext()

	var preamble strings.Builder
	preamble.WriteString("You received structured output from a previous agent in a chain.\n\n")

	if ctx != nil {
		if ctx.Source != "" {
			preamble.WriteString(fmt.Sprintf("Source agent: %s\n", ctx.Source))
		}
		if ctx.SourceDescription != "" {
			preamble.WriteString(fmt.Sprintf("Description: %s\n", ctx.SourceDescription))
		}
		preamble.WriteString(fmt.Sprintf("Chain depth: %d\n", ctx.Depth))
		preamble.WriteString("\n")
	}

	preamble.WriteString("The output is provided below as JSON:\n\n")
	preamble.WriteString("```json\n")
	preamble.WriteString(jsonInput)
	preamble.WriteString("\n```")

	return preamble.String()
}

// getLatestSession returns the most recent session from the database.
func getLatestSession(ctx context.Context, services *session.Services) (session.Session, error) {
	if services == nil {
		return session.Session{}, errors.New("session storage not available")
	}
	sessions, err := services.Sessions.List(ctx, 1)
	if err != nil {
		return session.Session{}, err
	}
	if len(sessions) == 0 {
		return session.Session{}, errors.New("no sessions found")
	}
	return sessions[0], nil
}

// looksLikeSubcommand checks if a string looks like it could be a mistyped subcommand.
// Returns true if the string is a single lowercase word that resembles command syntax.
func looksLikeSubcommand(s string) bool {
	// Must be non-empty
	if s == "" {
		return false
	}

	// If it contains spaces, it's probably a prompt
	if strings.Contains(s, " ") {
		return false
	}

	// If it starts with special chars like quotes, brackets, etc., it's probably data
	if strings.ContainsAny(string(s[0]), `"'[{(<>`) {
		return false
	}

	// If it's all lowercase letters, hyphens, or underscores (command-like pattern)
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || r == '-' || r == '_') {
			return false
		}
	}

	return true
}

// InvocationType represents the type of invocation parsed from CLI args.
type InvocationType string

const (
	InvocationTypeAgent InvocationType = "agent"
	InvocationTypeSquad InvocationType = "squad"
)

// ParsedInvocation holds the result of parsing CLI arguments.
type ParsedInvocation struct {
	Type       InvocationType
	Handle     string
	PromptArgs []string
}

// ParseInvocation parses CLI arguments to determine the invocation type.
// Returns the type (agent or squad), the handle, and remaining args as prompt.
func ParseInvocation(args []string) ParsedInvocation {
	if len(args) == 0 {
		return ParsedInvocation{
			Type:       InvocationTypeAgent,
			Handle:     agent.DefaultAgent,
			PromptArgs: nil,
		}
	}

	first := args[0]

	// Check for agent handle (@agent)
	if strings.HasPrefix(first, "@") {
		return ParsedInvocation{
			Type:       InvocationTypeAgent,
			Handle:     agent.NormalizeHandle(first),
			PromptArgs: args[1:],
		}
	}

	// Check for squad handle (#squad)
	if squads.IsSquadHandle(first) {
		return ParsedInvocation{
			Type:       InvocationTypeSquad,
			Handle:     squads.NormalizeHandle(first),
			PromptArgs: args[1:],
		}
	}

	// Default to @ayo agent with all args as prompt
	return ParsedInvocation{
		Type:       InvocationTypeAgent,
		Handle:     agent.DefaultAgent,
		PromptArgs: args,
	}
}

// knownSubcommands returns a list of all registered subcommand names.
var knownSubcommands = []string{
	"setup", "agents", "skills", "flows", "chain", "sessions", "memory",
	"doctor", "plugins", "serve", "sandbox", "share", "backup", "sync",
	"triggers", "ticket", "tickets", "squad", "squads", "service", "daemon",
}

// completeHandles returns completion suggestions for agent (@) and squad (#) handles.
func completeHandles(toComplete string) ([]string, cobra.ShellCompDirective) {
	var completions []string

	// Load config for agent listing
	cfg, err := config.Load(paths.ConfigFile())
	if err != nil {
		cfg = config.Config{}
	}

	// If completing a # prefix, suggest squad names
	if prefix, found := strings.CutPrefix(toComplete, "#"); found {
		squadNames, err := paths.ListSquads()
		if err == nil {
			for _, name := range squadNames {
				if strings.HasPrefix(name, prefix) {
					completions = append(completions, "#"+name)
				}
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}

	// If completing a @ prefix, suggest agent handles
	if strings.HasPrefix(toComplete, "@") {
		prefix := toComplete
		agents, err := agent.ListHandles(cfg)
		if err == nil {
			for _, handle := range agents {
				if strings.HasPrefix(handle, prefix) {
					completions = append(completions, handle)
				}
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}

	// No prefix yet - suggest both @ agents and # squads
	agents, err := agent.ListHandles(cfg)
	if err == nil {
		for _, handle := range agents {
			if strings.HasPrefix(handle, toComplete) || strings.HasPrefix(handle, "@"+toComplete) {
				completions = append(completions, handle)
			}
		}
	}

	squadNames, err := paths.ListSquads()
	if err == nil {
		for _, name := range squadNames {
			if strings.HasPrefix(name, toComplete) || strings.HasPrefix("#"+name, toComplete) {
				completions = append(completions, "#"+name)
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// invokeSquad dispatches work to a squad synchronously via the daemon.
// It connects to the daemon, sends the prompt, waits for a result, and prints output.
func invokeSquad(ctx context.Context, handle, prompt string) error {
	client, err := daemon.ConnectOrStart(ctx)
	if err != nil {
		return fmt.Errorf("connect to daemon: %w", err)
	}
	defer client.Close()

	name := squads.StripPrefix(handle)
	result, err := client.SquadDispatch(ctx, daemon.SquadDispatchParams{
		Name:           name,
		Prompt:         prompt,
		StartIfStopped: true,
	})
	if err != nil {
		return fmt.Errorf("dispatch to squad: %w", err)
	}

	if result.Error != "" {
		return fmt.Errorf("squad error: %s", result.Error)
	}

	if result.Raw != "" {
		fmt.Println(result.Raw)
	} else if result.Output != nil {
		// Structured output - encode as JSON
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result.Output); err != nil {
			return fmt.Errorf("encode output: %w", err)
		}
	}

	return nil
}
