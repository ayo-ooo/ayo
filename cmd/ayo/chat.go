package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/charmbracelet/x/input"
	"github.com/charmbracelet/x/term"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/run"
	"github.com/alexcabrera/ayo/internal/ui"
)

// runInteractiveChat handles the interactive chat session loop.
func runInteractiveChat(ctx context.Context, runner *run.Runner, ag agent.Agent, debug bool) error {
	uiHandler := ui.New(debug)

	// Set up signal handling for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	defer signal.Stop(sigChan)

	uiHandler.PrintChatHeaderWithSkills(ag.Handle, len(ag.Skills))

	for {
		select {
		case <-sigChan:
			fmt.Println()
			return nil
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		uiHandler.PrintUserPrompt()

		// Read input with Ctrl+H detection
		inputResult, err := readInputWithCtrlH(ctx, sigChan)
		if err != nil {
			return err
		}

		if inputResult.interrupted {
			fmt.Println()
			return nil
		}

		if inputResult.ctrlH {
			// Show history viewer
			messages, err := runner.GetSessionMessages(ctx, ag.Handle)
			if err != nil {
				uiHandler.PrintError(fmt.Sprintf("Failed to load history: %v", err))
				continue
			}
			if len(messages) == 0 {
				fmt.Println()
				fmt.Println("No conversation history yet.")
				fmt.Println()
				continue
			}

			result, err := ui.RunHistoryViewer(messages, ag.Handle, "Session History")
			if err != nil {
				uiHandler.PrintError(fmt.Sprintf("Failed to show history: %v", err))
				continue
			}

			if result == ui.HistoryViewerQuit {
				return nil
			}

			uiHandler.PrintChatHeaderWithSkills(ag.Handle, len(ag.Skills))
			continue
		}

		userInput := strings.TrimSpace(inputResult.text)
		if userInput == "" {
			continue
		}

		// Create a cancellable context for this request
		reqCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)

		// Listen for Ctrl+C during execution to cancel the request
		done := make(chan struct{})
		go func() {
			select {
			case <-sigChan:
				fmt.Println("\nInterrupted")
				cancel()
			case <-done:
			}
		}()

		resp, err := runner.Chat(reqCtx, ag, userInput)
		close(done)
		cancel()

		if err != nil {
			// Don't print error for context cancellation (user interrupted)
			if reqCtx.Err() != context.Canceled {
				uiHandler.PrintError(err.Error())
			}
			fmt.Println()
			continue
		}

		uiHandler.PrintResult(resp)
		fmt.Println()
	}
}

// inputResult holds the result of reading user input.
type inputResult struct {
	text        string
	ctrlH       bool
	interrupted bool
}

// readInputWithCtrlH reads user input character by character, detecting Ctrl+H.
func readInputWithCtrlH(ctx context.Context, sigChan <-chan os.Signal) (inputResult, error) {
	// Enter raw mode
	oldState, err := term.MakeRaw(os.Stdin.Fd())
	if err != nil {
		return inputResult{}, fmt.Errorf("failed to enter raw mode: %w", err)
	}
	defer term.Restore(os.Stdin.Fd(), oldState)

	// Create input reader
	drv, err := input.NewReader(os.Stdin, os.Getenv("TERM"), 0)
	if err != nil {
		return inputResult{}, fmt.Errorf("failed to create input reader: %w", err)
	}
	defer drv.Close()

	var buf strings.Builder

	for {
		select {
		case <-ctx.Done():
			return inputResult{interrupted: true}, nil
		case <-sigChan:
			return inputResult{interrupted: true}, nil
		default:
		}

		evs, err := drv.ReadEvents()
		if err != nil {
			return inputResult{}, err
		}

		for _, ev := range evs {
			switch ev := ev.(type) {
			case input.KeyPressEvent:
				keyStr := ev.String()

				switch keyStr {
				case "ctrl+c":
					return inputResult{interrupted: true}, nil

				case "ctrl+h":
					// Clear the line visually and return ctrl+h result
					fmt.Print("\r\033[K") // Clear line
					return inputResult{ctrlH: true}, nil

				case "enter":
					fmt.Println() // Move to next line
					return inputResult{text: buf.String()}, nil

				case "backspace":
					// Handle backspace
					if buf.Len() > 0 {
						s := buf.String()
						buf.Reset()
						buf.WriteString(s[:len(s)-1])
						fmt.Print("\b \b") // Erase character visually
					}

				case "space":
					buf.WriteRune(' ')
					fmt.Print(" ")

				case "tab":
					buf.WriteRune('\t')
					fmt.Print("\t")

				default:
					// Regular character input
					if len(keyStr) == 1 && keyStr[0] >= 32 && keyStr[0] < 127 {
						buf.WriteString(keyStr)
						fmt.Print(keyStr)
					} else if len(ev.Text) > 0 {
						buf.WriteString(ev.Text)
						fmt.Print(ev.Text)
					}
				}
			}
		}
	}
}
