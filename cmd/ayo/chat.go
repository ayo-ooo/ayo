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
	// The runner's StreamWriter will send streaming events through the channel.
	// This function just triggers the chat and returns the final response.
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

	// Create the tea.Program and model with the new channel-based architecture
	// This sets up:
	// 1. An event channel for streaming events
	// 2. An EventAggregator that forwards events to the TUI via program.Send()
	// 3. A ChannelWriter that the runner will use to write events
	program, _, channelWriter := chat.RunWithChannel(ctx, ag, sessionID, sendFn)

	// Set the stream writer on the runner so streaming events go through the channel
	runner.SetStreamWriter(channelWriter)

	// Run the TUI
	finalModel, err := program.Run()
	if err != nil {
		return err
	}

	m := finalModel.(chat.Model)
	scrollback := m.ScrollbackContent()

	// Dump scrollback to terminal after exiting alt-screen
	if scrollback != "" {
		fmt.Print(scrollback)
	}

	return nil
}
