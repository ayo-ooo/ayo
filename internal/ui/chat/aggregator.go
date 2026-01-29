package chat

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alexcabrera/ayo/internal/run"
)

// debug logging
func debugLogAgg(format string, args ...interface{}) {
	f, err := os.OpenFile("/tmp/ayo_stream.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%d: [AGG] ", time.Now().UnixMilli())
	fmt.Fprintf(f, format, args...)
	fmt.Fprintln(f)
}

// EventAggregator receives streaming events from a channel and forwards them
// to the Bubble Tea program. This decouples streaming from the TUI's message loop,
// preventing tick chain disruption.
type EventAggregator struct {
	events  <-chan run.StreamEvent
	program *tea.Program
}

// NewEventAggregator creates an aggregator that forwards events to the TUI.
func NewEventAggregator(events <-chan run.StreamEvent, program *tea.Program) *EventAggregator {
	debugLogAgg("NewEventAggregator created")
	return &EventAggregator{
		events:  events,
		program: program,
	}
}

// Subscribe runs a loop that forwards events to the TUI.
// This should be run in a goroutine. It exits when the context is canceled
// or when the event channel is closed.
func (a *EventAggregator) Subscribe(ctx context.Context) {
	debugLogAgg("Subscribe started")
	for {
		select {
		case event, ok := <-a.events:
			if !ok {
				debugLogAgg("Channel closed")
				return
			}
			debugLogAgg("Received event type=%d, sending to program", event.Type)
			a.program.Send(event)
			debugLogAgg("Sent event to program")
		case <-ctx.Done():
			debugLogAgg("Context done")
			return
		}
	}
}
