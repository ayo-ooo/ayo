package chat

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alexcabrera/ayo/internal/run"
)

// EventAggregator receives streaming events from a channel and forwards them
// to the Bubble Tea program. This decouples streaming from the TUI's message loop,
// preventing tick chain disruption.
type EventAggregator struct {
	events  <-chan run.StreamEvent
	program *tea.Program
}

// NewEventAggregator creates an aggregator that forwards events to the TUI.
func NewEventAggregator(events <-chan run.StreamEvent, program *tea.Program) *EventAggregator {
	return &EventAggregator{
		events:  events,
		program: program,
	}
}

// Subscribe runs a loop that forwards events to the TUI.
// This should be run in a goroutine. It exits when the context is canceled
// or when the event channel is closed.
func (a *EventAggregator) Subscribe(ctx context.Context) {
	for {
		select {
		case event, ok := <-a.events:
			if !ok {
				return
			}
			a.program.Send(event)
		case <-ctx.Done():
			return
		}
	}
}
