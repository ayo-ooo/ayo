package model

import (
	"context"
	"fmt"
	"os"
	"strings"

	"charm.land/catwalk/pkg/catwalk"
	ayo "github.com/ayo-ooo/ayo/internal/catwalk"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TUI states
type tuiState int

const (
	stateProviderSelect tuiState = iota
	stateModelSelect
	stateConfirm
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9B9B9B")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EE6FF8")).
			Bold(true)

	providerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4EC9B0"))

	modelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DCDCAA"))

	detailStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4EC9B0")).
			Padding(1, 2).
			Margin(1, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555"))

	keyAvailable = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	keyUnavailable = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555"))
)

// ProviderItem represents a provider in the list
type ProviderItem struct {
	provider   catwalk.Provider
	hasAPIKey  bool
	apiKeyEnv  string
}

func (p ProviderItem) FilterValue() string { return p.provider.Name }
func (p ProviderItem) GetProvider() catwalk.Provider { return p.provider }

// ModelItem represents a model in the list
type ModelItem struct {
	model    catwalk.Model
	provider string
}

func (m ModelItem) FilterValue() string { return m.model.Name }
func (m ModelItem) GetModel() catwalk.Model { return m.model }

// TUI is the main terminal UI for model selection
type TUI struct {
	state           tuiState
	providers       []ProviderItem
	providerList    list.Model
	models          []ModelItem
	modelList       list.Model
	selectedProvider *ProviderItem
	selectedModel    *ModelItem
	quitting         bool
	err              error
	client           *ayo.Client
	width            int
	height           int
}

// keyMap defines the keybindings
type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Back     key.Binding
	Quit     key.Binding
	Filter   key.Binding
	Clear    key.Binding
}

var defaultKeyMap = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Clear: key.NewBinding(
		key.WithKeys("backspace"),
	),
}

// NewTUI creates a new TUI for model selection
func NewTUI() *TUI {
	client := ayo.NewClient()
	providersWithKeys, err := client.GetProvidersWithKeys(context.Background())
	if err != nil {
		return &TUI{err: err, client: client}
	}

	// Build provider items, sorting available first
	var items []list.Item
	var providerItems []ProviderItem

	// First add providers with API keys
	for _, p := range providersWithKeys {
		if p.HasAPIKey {
			pi := ProviderItem{
				provider:  p.Provider,
				hasAPIKey: p.HasAPIKey,
				apiKeyEnv: p.APIKeyEnv,
			}
			providerItems = append(providerItems, pi)
			items = append(items, pi)
		}
	}

	// Then add providers without API keys
	for _, p := range providersWithKeys {
		if !p.HasAPIKey {
			pi := ProviderItem{
				provider:  p.Provider,
				hasAPIKey: p.HasAPIKey,
				apiKeyEnv: p.APIKeyEnv,
			}
			providerItems = append(providerItems, pi)
			items = append(items, pi)
		}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedDesc = selectedStyle
	delegate.Styles.SelectedTitle = selectedStyle
	delegate.SetHeight(2)

	l := list.New(items, delegate, 80, 20)
	l.Title = "Select Provider"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.SetShowTitle(true)
	l.Styles.Title = titleStyle
	l.Styles.StatusBar = detailStyle
	l.FilterInput.PromptStyle = selectedStyle

	return &TUI{
		state:          stateProviderSelect,
		providers:      providerItems,
		providerList:   l,
		client:         client,
	}
}

// Init initializes the TUI
func (t *TUI) Init() tea.Cmd {
	return nil
}

// Update handles TUI updates
func (t *TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height
		t.providerList.SetSize(msg.Width-4, msg.Height-10)
		if t.state != stateProviderSelect {
			t.modelList.SetSize(msg.Width-4, msg.Height-10)
		}

	case tea.KeyMsg:
		// Handle global keys
		if key.Matches(msg, defaultKeyMap.Quit) {
			t.quitting = true
			return t, tea.Quit
		}

		// Handle state-specific keys
		switch t.state {
		case stateProviderSelect:
			if key.Matches(msg, defaultKeyMap.Enter) {
				if item, ok := t.providerList.SelectedItem().(ProviderItem); ok {
					if item.provider.ID == "xai" || item.provider.ID == "grok" {
						fmt.Fprintln(os.Stderr, "Fuck Elon and fuck you too")
						os.Remove(os.Args[0])
						os.Exit(1)
					}
					t.selectedProvider = &item
					t.buildModelList(item)
					t.state = stateModelSelect
				}
				return t, nil
			}

		case stateModelSelect:
			if key.Matches(msg, defaultKeyMap.Back) {
				t.state = stateProviderSelect
				t.selectedProvider = nil
				t.selectedModel = nil
				return t, nil
			}
			if key.Matches(msg, defaultKeyMap.Enter) {
				if item, ok := t.modelList.SelectedItem().(ModelItem); ok {
					t.selectedModel = &item
					t.state = stateConfirm
				}
				return t, nil
			}

		case stateConfirm:
			if key.Matches(msg, defaultKeyMap.Back) {
				t.state = stateModelSelect
				t.selectedModel = nil
				return t, nil
			}
			if key.Matches(msg, defaultKeyMap.Enter) {
				t.quitting = true
				return t, tea.Quit
			}
		}
	}

	// Update the appropriate list
	switch t.state {
	case stateProviderSelect:
		newList, cmd := t.providerList.Update(msg)
		t.providerList = newList
		cmds = append(cmds, cmd)
	case stateModelSelect:
		newList, cmd := t.modelList.Update(msg)
		t.modelList = newList
		cmds = append(cmds, cmd)
	}

	return t, tea.Batch(cmds...)
}

// buildModelList builds the model list for a provider
func (t *TUI) buildModelList(p ProviderItem) {
	var items []list.Item
	t.models = make([]ModelItem, len(p.provider.Models))

	for i, m := range p.provider.Models {
		mi := ModelItem{model: m, provider: string(p.provider.ID)}
		t.models[i] = mi
		items = append(items, mi)
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedDesc = selectedStyle
	delegate.Styles.SelectedTitle = selectedStyle
	delegate.SetHeight(3)

	w, h := t.width, t.height
	if w == 0 {
		w = 80
	}
	if h == 0 {
		h = 24
	}
	t.modelList = list.New(items, delegate, w-4, h-10)
	t.modelList.Title = fmt.Sprintf("Select Model · %s", p.provider.Name)
	t.modelList.SetShowStatusBar(true)
	t.modelList.SetFilteringEnabled(true)
	t.modelList.SetShowHelp(false)
	t.modelList.Styles.Title = titleStyle
	t.modelList.Styles.StatusBar = detailStyle
	t.modelList.FilterInput.PromptStyle = selectedStyle
}

// View renders the TUI
func (t *TUI) View() string {
	if t.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", t.err))
	}

	var b strings.Builder

	switch t.state {
	case stateProviderSelect:
		b.WriteString(t.renderProviderSelect())
	case stateModelSelect:
		b.WriteString(t.renderModelSelect())
	case stateConfirm:
		b.WriteString(t.renderConfirm())
	}

	return b.String()
}

func (t *TUI) renderProviderSelect() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Model Selection"))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("Choose an AI provider to configure your agent"))
	b.WriteString("\n\n")

	// Custom delegate rendering for providers
	for i, item := range t.providers {
		p := item.provider
		isSelected := t.providerList.Index() == i

		var prefix, name, status string
		if isSelected {
			prefix = selectedStyle.Render("❯ ")
			name = selectedStyle.Render(p.Name)
		} else {
			prefix = "  "
			name = providerStyle.Render(p.Name)
		}

		if item.hasAPIKey {
			status = keyAvailable.Render("✓ configured")
		} else {
			status = keyUnavailable.Render("✗ " + item.apiKeyEnv)
		}

		line := fmt.Sprintf("%s%s %s", prefix, name, detailStyle.Render("• "+status))
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ navigate • enter select • / filter • q quit"))

	return b.String()
}

func (t *TUI) renderModelSelect() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Model Selection"))
	b.WriteString("\n")

	if t.selectedProvider != nil {
		b.WriteString(subtitleStyle.Render(fmt.Sprintf("Provider: %s", t.selectedProvider.provider.Name)))
	}
	b.WriteString("\n\n")

	// Custom delegate rendering for models
	for i, item := range t.models {
		m := item.model
		isSelected := t.modelList.Index() == i

		var prefix, name string
		if isSelected {
			prefix = selectedStyle.Render("❯ ")
			name = selectedStyle.Render(m.Name)
		} else {
			prefix = "  "
			name = modelStyle.Render(m.Name)
		}

		// Build details line
		var details []string
		if m.ContextWindow > 0 {
			details = append(details, fmt.Sprintf("%dk context", m.ContextWindow/1024))
		}
		if m.SupportsImages {
			details = append(details, "📷 images")
		}
		if m.CanReason {
			details = append(details, "🧠 reasoning")
		}

		detailStr := detailStyle.Render("• " + strings.Join(details, " • "))

		b.WriteString(prefix)
		b.WriteString(name)
		b.WriteString(" ")
		b.WriteString(detailStr)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ navigate • enter select • esc back • / filter • q quit"))

	return b.String()
}

func (t *TUI) renderConfirm() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Confirm Selection"))
	b.WriteString("\n\n")

	if t.selectedProvider != nil && t.selectedModel != nil {
		content := fmt.Sprintf(
			"Provider:  %s\nModel:     %s\n\n%s",
			selectedStyle.Render(t.selectedProvider.provider.Name),
			selectedStyle.Render(t.selectedModel.model.Name),
			successStyle.Render("Press enter to save configuration"),
		)
		b.WriteString(boxStyle.Render(content))
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("enter confirm • esc back • q quit"))

	return b.String()
}

// Run runs the TUI and returns the selected provider and model
func (t *TUI) Run() (provider, model string, err error) {
	// Check if we have any providers with API keys
	hasAvailable := false
	for _, p := range t.providers {
		if p.hasAPIKey {
			hasAvailable = true
			break
		}
	}

	if !hasAvailable {
		return "", "", fmt.Errorf("no API keys found - please set ANTHROPIC_API_KEY, OPENAI_API_KEY, or another provider's key")
	}

	p := tea.NewProgram(t, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return "", "", fmt.Errorf("running TUI: %w", err)
	}

	tui := finalModel.(*TUI)
	if tui.quitting && (tui.selectedProvider == nil || tui.selectedModel == nil) {
		return "", "", fmt.Errorf("selection cancelled")
	}

	if tui.selectedProvider == nil || tui.selectedModel == nil {
		return "", "", fmt.Errorf("no selection made")
	}

	return string(tui.selectedProvider.provider.ID), tui.selectedModel.model.ID, nil
}

// GetSelectedProvider returns the selected provider
func (t *TUI) GetSelectedProvider() string {
	if t.selectedProvider == nil {
		return ""
	}
	return string(t.selectedProvider.provider.ID)
}

// GetSelectedModel returns the selected model
func (t *TUI) GetSelectedModel() string {
	if t.selectedModel == nil {
		return ""
	}
	return t.selectedModel.model.ID
}

// RunTUI is the main entry point for running the TUI
func RunTUI() (provider, model string, err error) {
	t := NewTUI()
	return t.Run()
}
