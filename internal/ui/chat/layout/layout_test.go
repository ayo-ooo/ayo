package layout

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// MockFocusable is a test implementation of Focusable
type MockFocusable struct {
	focused bool
}

func (m *MockFocusable) Focus() tea.Cmd {
	m.focused = true
	return nil
}

func (m *MockFocusable) Blur() tea.Cmd {
	m.focused = false
	return nil
}

func (m *MockFocusable) IsFocused() bool {
	return m.focused
}

func TestFocusable_Interface(t *testing.T) {
	var f Focusable = &MockFocusable{}

	if f.IsFocused() {
		t.Error("should not be focused initially")
	}

	f.Focus()
	if !f.IsFocused() {
		t.Error("should be focused after Focus()")
	}

	f.Blur()
	if f.IsFocused() {
		t.Error("should not be focused after Blur()")
	}
}

// MockSizeable is a test implementation of Sizeable
type MockSizeable struct {
	width, height int
}

func (m *MockSizeable) SetSize(w, h int) tea.Cmd {
	m.width = w
	m.height = h
	return nil
}

func (m *MockSizeable) GetSize() (int, int) {
	return m.width, m.height
}

func TestSizeable_Interface(t *testing.T) {
	var s Sizeable = &MockSizeable{}

	w, h := s.GetSize()
	if w != 0 || h != 0 {
		t.Errorf("initial size = (%d, %d), want (0, 0)", w, h)
	}

	s.SetSize(80, 24)
	w, h = s.GetSize()
	if w != 80 || h != 24 {
		t.Errorf("size = (%d, %d), want (80, 24)", w, h)
	}
}

// MockPositional is a test implementation of Positional
type MockPositional struct {
	x, y int
}

func (m *MockPositional) SetPosition(x, y int) tea.Cmd {
	m.x = x
	m.y = y
	return nil
}

func TestPositional_Interface(t *testing.T) {
	m := &MockPositional{}
	var p Positional = m

	p.SetPosition(10, 20)
	if m.x != 10 || m.y != 20 {
		t.Errorf("position = (%d, %d), want (10, 20)", m.x, m.y)
	}
}

// MockHelp is a test implementation of Help
type MockHelp struct {
	bindings []key.Binding
}

func (m *MockHelp) Bindings() []key.Binding {
	return m.bindings
}

func TestHelp_Interface(t *testing.T) {
	bindings := []key.Binding{
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	}

	var h Help = &MockHelp{bindings: bindings}

	got := h.Bindings()
	if len(got) != 2 {
		t.Errorf("len(Bindings()) = %d, want 2", len(got))
	}
}

// MockModel is a test implementation of Model
type MockModel struct {
	value string
}

func (m MockModel) Init() tea.Cmd {
	return nil
}

func (m MockModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m MockModel) View() string {
	return m.value
}

func TestModel_Interface(t *testing.T) {
	var m Model = MockModel{value: "test"}

	if m.Init() != nil {
		t.Error("Init() should return nil")
	}

	updated, cmd := m.Update(nil)
	if cmd != nil {
		t.Error("Update() should return nil cmd")
	}
	if updated.View() != "test" {
		t.Errorf("View() = %q, want %q", updated.View(), "test")
	}
}

// Test that interfaces can be composed
type ComposableComponent struct {
	MockFocusable
	MockSizeable
}

func TestComposedInterfaces(t *testing.T) {
	c := &ComposableComponent{}

	// Should satisfy both interfaces
	var _ Focusable = c
	var _ Sizeable = c

	c.Focus()
	c.SetSize(100, 50)

	if !c.IsFocused() {
		t.Error("should be focused")
	}

	w, h := c.GetSize()
	if w != 100 || h != 50 {
		t.Errorf("size = (%d, %d), want (100, 50)", w, h)
	}
}
