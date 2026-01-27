package main

import (
	"context"
	"fmt"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/run"
	"github.com/alexcabrera/ayo/internal/ui/chat"
)

// runInteractiveChat handles the interactive chat session loop using the alt-screen TUI.
func runInteractiveChat(ctx context.Context, runner *run.Runner, ag agent.Agent, debug bool) error {
	// Get session ID for display
	sessionID := runner.GetSessionID(ag.Handle)

	// Create a send function that wraps runner.Chat
	sendFn := func(ctx context.Context, message string) (string, error) {
		return runner.Chat(ctx, ag, message)
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
