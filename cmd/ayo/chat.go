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
	// The runner's StreamHandler (set below) will send streaming events to the TUI.
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

	// Create the tea.Program and model, then set up the TUIStreamHandler
	program, _ := chat.RunWithProgram(ctx, ag, sessionID, sendFn)

	// Create a TUIStreamHandler with memory service for initial memory loading
	handlerOpts := []chat.TUIStreamHandlerOption{}
	if memSvc := runner.MemoryService(); memSvc != nil {
		handlerOpts = append(handlerOpts, chat.WithMemoryService(memSvc))
	}
	handler := chat.NewTUIStreamHandler(program, handlerOpts...)

	// Set the handler on the runner so streaming events go to the TUI
	runner.SetStreamHandler(handler)

	// Send initial memories to the TUI if memory is enabled for this agent
	if ag.Config.Memory.Enabled && ag.Config.Memory.Retrieval.AutoInject {
		// Use a general query to get relevant memories for the session start
		// The agent handle helps scope the search
		_ = handler.SendInitialMemories(
			ctx,
			"session context", // Generic query for initial memories
			ag.Handle,
			ag.Config.Memory.Retrieval.Threshold,
			ag.Config.Memory.Retrieval.MaxMemories,
		)
	}

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
