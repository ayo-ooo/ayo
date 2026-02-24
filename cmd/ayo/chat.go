package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/run"
	"github.com/alexcabrera/ayo/internal/ui/interactive"
	"github.com/charmbracelet/lipgloss"
)

// runInteractiveChat handles the interactive chat session loop using simple line-based input.
func runInteractiveChat(ctx context.Context, runner *run.Runner, ag agent.Agent, _ bool) error {
	// Create interactive writer
	writer := interactive.NewWriter()
	runner.SetStreamWriter(writer)

	// Get session ID for display
	sessionID := runner.GetSessionID(ag.Handle)

	// Styles
	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#67e8f9")).Bold(true)
	sessionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))

	// Print session info
	fmt.Printf("\n%s %s\n", promptStyle.Render(ag.Handle), sessionStyle.Render("("+sessionID[:8]+")"))
	fmt.Println(sessionStyle.Render("Type 'exit' or Ctrl+D to quit"))
	fmt.Println()

	// Simple input loop using bufio
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		// Print prompt
		fmt.Print(promptStyle.Render("> "))

		// Read input
		if !scanner.Scan() {
			// EOF (Ctrl+D) or error
			fmt.Println()
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Check for exit command
		if input == "exit" || input == "quit" || input == "/exit" || input == "/quit" {
			break
		}

		// Print agent header
		fmt.Printf("\n%s\n", promptStyle.Render("@"+ag.Handle+":"))

		// Send to agent
		_, err := runner.Chat(ctx, ag, input)
		if err != nil {
			fmt.Printf("\nError: %v\n", err)
		}
		fmt.Println()
	}

	return nil
}
