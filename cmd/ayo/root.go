package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/builtin"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/embedding"
	"github.com/alexcabrera/ayo/internal/memory"
	"github.com/alexcabrera/ayo/internal/ollama"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/pipe"
	"github.com/alexcabrera/ayo/internal/run"
	"github.com/alexcabrera/ayo/internal/session"
	"github.com/alexcabrera/ayo/internal/smallmodel"
	"github.com/alexcabrera/ayo/internal/ui"
)

func newRootCmd() *cobra.Command {
	var cfgPath string
	var attachments []string
	var debug bool

	cmd := &cobra.Command{
		Use:           "ayo [@agent] [prompt]",
		Short:         "Run AI agents",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ArbitraryArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Auto-install built-in agents and skills if needed (version-based)
			return builtin.Install()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConfig(&cfgPath, func(cfg config.Config) error {
				if len(args) == 0 {
					// No args: show help
					return cmd.Help()
				}

				// Determine agent handle and remaining args
				var handle string
				var promptArgs []string

				if strings.HasPrefix(args[0], "@") {
					// First arg is an agent handle
					handle = agent.NormalizeHandle(args[0])
					promptArgs = args[1:]
				} else {
					// First arg is not an agent handle: use default agent with all args as prompt
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
					if debug {
						fmt.Fprintf(os.Stderr, "Warning: session persistence unavailable: %v\n", err)
					}
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
					} else if debug {
						fmt.Fprintf(os.Stderr, "Warning: Ollama not available at %s, memory features disabled\n", cfg.OllamaHost)
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

				runner, err := run.NewRunner(cfg, debug, run.RunnerOptions{
					Services:         services,
					MemoryService:    memSvc,
					FormationService: formSvc,
					SmallModel:       smallModelSvc,
					MemoryQueue:      memQueue,
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

					result, err := runner.TextWithSession(ctx, ag, prompt, attachments)
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
				return runInteractiveChat(cmd.Context(), runner, ag, debug)
			})
		},
	}

	cmd.PersistentFlags().StringVar(&cfgPath, "config", defaultConfigPath(), "path to config file")
	cmd.Flags().StringSliceVarP(&attachments, "attachment", "a", nil, "file attachments")
	cmd.Flags().BoolVar(&debug, "debug", false, "show debug output including raw tool payloads")

	// Subcommands
	cmd.AddCommand(newSetupCmd(&cfgPath))
	cmd.AddCommand(newAgentsCmd(&cfgPath))
	cmd.AddCommand(newSkillsCmd(&cfgPath))
	cmd.AddCommand(newChainCmd(&cfgPath))
	cmd.AddCommand(newSessionsCmd(&cfgPath))
	cmd.AddCommand(newMemoryCmd())
	cmd.AddCommand(newDoctorCmd(&cfgPath))

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
