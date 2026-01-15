package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/run"
	"github.com/alexcabrera/ayo/internal/ui"
)

// runInteractiveChat handles the interactive chat session loop.
func runInteractiveChat(ctx context.Context, runner *run.Runner, ag agent.Agent, debug bool) error {
	uiHandler := ui.New(debug)
	reader := bufio.NewReader(os.Stdin)

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

		inputChan := make(chan string, 1)
		errChan := make(chan error, 1)

		go func() {
			input, err := reader.ReadString('\n')
			if err != nil {
				errChan <- err
				return
			}
			inputChan <- input
		}()

		var input string
		select {
		case <-sigChan:
			fmt.Println()
			return nil
		case err := <-errChan:
			if err == io.EOF {
				fmt.Println()
				return nil
			}
			return err
		case input = <-inputChan:
		}

		input = strings.TrimSpace(input)

		if input == "" {
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

		resp, err := runner.Chat(reqCtx, ag, input)
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
