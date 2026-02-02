package anim

import (
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func TestDefaultSettings(t *testing.T) {
	s := DefaultSettings()

	if s.Size != 15 {
		t.Errorf("Size = %d, want 15", s.Size)
	}
	if s.Label != "Working" {
		t.Errorf("Label = %q, want %q", s.Label, "Working")
	}
	if s.Interval != 80*time.Millisecond {
		t.Errorf("Interval = %v, want %v", s.Interval, 80*time.Millisecond)
	}
	if !s.CycleColors {
		t.Error("CycleColors should be true")
	}
}

func TestNew_DefaultValues(t *testing.T) {
	// Test with zero values to trigger defaults
	m := New(Settings{})

	if m.settings.Interval != 80*time.Millisecond {
		t.Errorf("Interval = %v, want %v", m.settings.Interval, 80*time.Millisecond)
	}
	if m.settings.Size != 15 {
		t.Errorf("Size = %d, want 15", m.settings.Size)
	}
	if !m.active {
		t.Error("Model should be active by default")
	}
}

func TestNew_CustomSettings(t *testing.T) {
	s := Settings{
		Size:     20,
		Label:    "Custom",
		Interval: 100 * time.Millisecond,
	}
	m := New(s)

	if m.settings.Size != 20 {
		t.Errorf("Size = %d, want 20", m.settings.Size)
	}
	if m.settings.Label != "Custom" {
		t.Errorf("Label = %q, want %q", m.settings.Label, "Custom")
	}
}

func TestModel_ID(t *testing.T) {
	m := New(DefaultSettings())
	id := m.ID()

	if id == "" {
		t.Error("ID should not be empty")
	}
}

func TestModel_IsActive(t *testing.T) {
	m := New(DefaultSettings())

	if !m.IsActive() {
		t.Error("Model should be active after creation")
	}

	m.Stop()
	if m.IsActive() {
		t.Error("Model should not be active after Stop()")
	}
}

func TestModel_Stop(t *testing.T) {
	m := New(DefaultSettings())
	m.Stop()

	if m.active {
		t.Error("active should be false after Stop()")
	}
}

func TestModel_Start(t *testing.T) {
	m := New(DefaultSettings())
	m.Stop()

	cmd := m.Start()

	if !m.active {
		t.Error("active should be true after Start()")
	}
	if cmd == nil {
		t.Error("Start() should return a command")
	}
}

func TestModel_SetLabel(t *testing.T) {
	m := New(DefaultSettings())
	m.SetLabel("New Label")

	if m.settings.Label != "New Label" {
		t.Errorf("Label = %q, want %q", m.settings.Label, "New Label")
	}
}

func TestModel_Init(t *testing.T) {
	m := New(DefaultSettings())
	cmd := m.Init()

	if cmd == nil {
		t.Error("Init() should return a command")
	}
}

func TestModel_View_Active(t *testing.T) {
	m := New(Settings{
		Label:      "Test",
		LabelColor: lipgloss.Color("#ffffff"),
		GradColorA: lipgloss.Color("#a78bfa"),
	})
	view := m.View()

	if view == "" {
		t.Error("View should not be empty when active")
	}
}

func TestModel_View_Inactive(t *testing.T) {
	m := New(DefaultSettings())
	m.Stop()
	view := m.View()

	if view != "" {
		t.Errorf("View = %q, want empty string when inactive", view)
	}
}

func TestModel_View_NoLabel(t *testing.T) {
	m := New(Settings{
		Label:      "",
		GradColorA: lipgloss.Color("#a78bfa"),
	})
	view := m.View()

	// Should only contain spinner, no label
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestModel_Update_OwnTickMsg(t *testing.T) {
	m := New(DefaultSettings())
	initialFrame := m.frame

	msg := TickMsg{ID: m.id}
	newM, cmd := m.Update(msg)

	if newM.frame == initialFrame {
		t.Error("frame should advance on own TickMsg")
	}
	if cmd == nil {
		t.Error("Update should return a tick command")
	}
}

func TestModel_Update_OtherTickMsg(t *testing.T) {
	m := New(DefaultSettings())
	initialFrame := m.frame

	msg := TickMsg{ID: "different-id"}
	newM, cmd := m.Update(msg)

	if newM.frame != initialFrame {
		t.Error("frame should not advance on other TickMsg")
	}
	if cmd != nil {
		t.Error("Update should return nil for other TickMsg")
	}
}

func TestModel_Update_InactiveIgnoresTick(t *testing.T) {
	m := New(DefaultSettings())
	m.Stop()
	initialFrame := m.frame

	msg := TickMsg{ID: m.id}
	newM, cmd := m.Update(msg)

	if newM.frame != initialFrame {
		t.Error("frame should not advance when inactive")
	}
	if cmd != nil {
		t.Error("Update should return nil when inactive")
	}
}

func TestModel_Update_OtherMsg(t *testing.T) {
	m := New(DefaultSettings())

	// Some other message type
	type OtherMsg struct{}
	_, cmd := m.Update(OtherMsg{})

	if cmd != nil {
		t.Error("Update should return nil for unknown message types")
	}
}

func TestTickMsg_Fields(t *testing.T) {
	msg := TickMsg{ID: "test-id"}
	if msg.ID != "test-id" {
		t.Errorf("ID = %q, want %q", msg.ID, "test-id")
	}
}

func TestSettings_Fields(t *testing.T) {
	s := Settings{
		Size:        20,
		Label:       "Loading",
		GradColorA:  lipgloss.Color("#111"),
		GradColorB:  lipgloss.Color("#222"),
		LabelColor:  lipgloss.Color("#333"),
		CycleColors: true,
		Interval:    50 * time.Millisecond,
	}

	if s.Size != 20 {
		t.Errorf("Size = %d, want 20", s.Size)
	}
	if s.Label != "Loading" {
		t.Errorf("Label = %q, want %q", s.Label, "Loading")
	}
	if s.Interval != 50*time.Millisecond {
		t.Errorf("Interval = %v, want %v", s.Interval, 50*time.Millisecond)
	}
}
