//go:build ignore

package main

import (
	"context"
	"fmt"
	"os"
	"time"
	
	tea "github.com/charmbracelet/bubbletea"
	
	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/run"
	"github.com/alexcabrera/ayo/internal/session"
	"github.com/alexcabrera/ayo/internal/ui/chat"
)

func main() {
	fmt.Fprintln(os.Stderr, "DEBUG TEST: Starting")
	ctx := context.Background()
	
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}
	
	ag, err := agent.Load(cfg, "@ayo")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Agent error: %v\n", err)
		os.Exit(1)
	}
	
	services, _ := session.Connect(ctx, paths.DatabasePath())
	if services != nil {
		defer services.Close()
	}
	
	runner, err := run.NewRunner(cfg, false, run.RunnerOptions{
		Services: services,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Runner error: %v\n", err)
		os.Exit(1)
	}
	
	sessionID := runner.GetSessionID(ag.Handle)
	
	sendFn := func(ctx context.Context, message string) (string, error) {
		fmt.Fprintln(os.Stderr, "DEBUG: sendFn called with message:", message)
		result, err := runner.Chat(ctx, ag, message)
		fmt.Fprintf(os.Stderr, "DEBUG: runner.Chat returned: err=%v, result=%q\n", err, result)
		return result, err
	}
	
	program, model := chat.RunWithProgram(ctx, ag, sessionID, sendFn)
	_ = model
	
	// Create a debug stream handler that logs
	debugHandler := &debugStreamHandler{
		inner: chat.NewTUIStreamHandler(program),
	}
	runner.SetStreamHandler(debugHandler)
	
	fmt.Fprintln(os.Stderr, "DEBUG TEST: Starting program.Run()")
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

type debugStreamHandler struct {
	inner *chat.TUIStreamHandler
}

func (h *debugStreamHandler) OnTextDelta(id, text string) error {
	fmt.Fprintf(os.Stderr, "DEBUG OnTextDelta: %q\n", text)
	return h.inner.OnTextDelta(id, text)
}

func (h *debugStreamHandler) OnTextEnd(id string) error {
	fmt.Fprintln(os.Stderr, "DEBUG OnTextEnd")
	return h.inner.OnTextEnd(id)
}

func (h *debugStreamHandler) OnToolCall(tc interface{}) error {
	fmt.Fprintf(os.Stderr, "DEBUG OnToolCall: %v\n", tc)
	return nil
}

func (h *debugStreamHandler) OnToolResult(result interface{}, duration time.Duration) error {
	fmt.Fprintf(os.Stderr, "DEBUG OnToolResult: duration=%v\n", duration)
	return nil
}

func (h *debugStreamHandler) OnReasoningStart(id string) error {
	fmt.Fprintln(os.Stderr, "DEBUG OnReasoningStart")
	return h.inner.OnReasoningStart(id)
}

func (h *debugStreamHandler) OnReasoningDelta(id, text string) error {
	fmt.Fprintf(os.Stderr, "DEBUG OnReasoningDelta: %q\n", text)
	return h.inner.OnReasoningDelta(id, text)
}

func (h *debugStreamHandler) OnReasoningEnd(id string, duration time.Duration) error {
	fmt.Fprintf(os.Stderr, "DEBUG OnReasoningEnd: duration=%v\n", duration)
	return h.inner.OnReasoningEnd(id, duration)
}

func (h *debugStreamHandler) OnError(err error) error {
	fmt.Fprintf(os.Stderr, "DEBUG OnError: %v\n", err)
	return h.inner.OnError(err)
}

// Implement the Cmd helper if needed
func (h *debugStreamHandler) Cmd() tea.Cmd {
	return nil
}
