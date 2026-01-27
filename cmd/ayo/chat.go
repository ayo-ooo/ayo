package main

import (
	"context"
	"fmt"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/run"
	"github.com/alexcabrera/ayo/internal/session"
	"github.com/alexcabrera/ayo/internal/ui/chat"
)

// runInteractiveChat handles the interactive chat session loop using the alt-screen TUI.
func runInteractiveChat(ctx context.Context, runner *run.Runner, ag agent.Agent, debug bool) error {
	// Get session ID for display
	sessionID := runner.GetSessionID(ag.Handle)

	// Create a send function that wraps runner.Chat
	// Note: runner.Chat streams to stdout which goes to alt-screen, and returns empty string.
	// We need to retrieve the actual response from session messages after the call.
	sendFn := func(ctx context.Context, message string) (string, error) {
		_, err := runner.Chat(ctx, ag, message)
		if err != nil {
			return "", err
		}

		// Retrieve the last assistant message from the session
		messages, err := runner.GetSessionMessages(ctx, ag.Handle)
		if err != nil {
			return "", err
		}

		// Find the last assistant message
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == "assistant" {
				// Extract text content from parts
				for _, part := range messages[i].Parts {
					if textPart, ok := part.(session.TextContent); ok {
						return textPart.Text, nil
					}
				}
			}
		}

		return "", nil
	}

	// Run the alt-screen chat TUI
	result, scrollback, err := chat.Run(ctx, ag, sessionID, sendFn)
	if err != nil {
		return err
	}

	// Dump scrollback to terminal after exiting alt-screen
	if scrollback != "" {
		fmt.Print(scrollback)
	}

	if result == chat.ResultError {
		return fmt.Errorf("chat ended with error")
	}

	return nil
}
