// Package ui provides terminal user interface components.
package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
)

// ApprovalRequest contains information about a file modification requiring approval.
type ApprovalRequest struct {
	Agent       string // Agent requesting the modification
	Action      string // "create", "update", "delete"
	Path        string // Host path
	Content     string // New content (for create/update)
	OldContent  string // Current content (for diff display in updates)
	Reason      string // Reason provided by agent
	SessionID   string // Session ID for caching
}

// ApprovalResponse contains the user's decision.
type ApprovalResponse struct {
	Approved      bool // Whether the request was approved
	AlwaysApprove bool // User selected "A" (always approve this session)
}

// ApprovalPrompter prompts users for file modification approval.
type ApprovalPrompter interface {
	Prompt(req ApprovalRequest) (ApprovalResponse, error)
}

// InteractiveApprovalPrompter implements ApprovalPrompter using stdin/stdout.
type InteractiveApprovalPrompter struct {
	in  *bufio.Reader
	out *os.File
}

// NewInteractiveApprovalPrompter creates a new interactive prompter.
func NewInteractiveApprovalPrompter() *InteractiveApprovalPrompter {
	return &InteractiveApprovalPrompter{
		in:  bufio.NewReader(os.Stdin),
		out: os.Stdout,
	}
}

// Prompt displays an approval prompt and waits for user input.
func (p *InteractiveApprovalPrompter) Prompt(req ApprovalRequest) (ApprovalResponse, error) {
	// Styles
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	reasonStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)
	optionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	boxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(0, 1)

	// Build content
	var content strings.Builder
	
	actionVerb := actionToVerb(req.Action)
	content.WriteString(titleStyle.Render(fmt.Sprintf("%s wants to %s:", req.Agent, actionVerb)))
	content.WriteString("\n")
	content.WriteString("  " + pathStyle.Render(req.Path))
	content.WriteString("\n")
	
	if req.Reason != "" {
		content.WriteString("\n")
		content.WriteString(reasonStyle.Render("Reason: " + req.Reason))
	}
	
	content.WriteString("\n\n")
	content.WriteString("─────────────────────────────────────────────────\n")
	
	options := []string{"[Y]es", "[N]o"}
	if req.Action == "update" && req.OldContent != "" {
		options = append(options, "[D]iff")
	}
	options = append(options, "[A]lways", "[?]Help")
	
	content.WriteString(optionStyle.Render(strings.Join(options, "  ")))
	
	// Print box
	fmt.Fprintln(p.out)
	fmt.Fprintln(p.out, boxStyle.Render(content.String()))
	
	// Input loop
	for {
		fmt.Fprint(p.out, "> ")
		
		input, err := p.in.ReadString('\n')
		if err != nil {
			return ApprovalResponse{}, err
		}
		
		input = strings.TrimSpace(strings.ToLower(input))
		
		switch input {
		case "y", "yes":
			return ApprovalResponse{Approved: true}, nil
		case "n", "no":
			return ApprovalResponse{Approved: false}, nil
		case "a", "always":
			return ApprovalResponse{Approved: true, AlwaysApprove: true}, nil
		case "d", "diff":
			if req.Action == "update" && req.OldContent != "" {
				p.showDiff(req.Path, req.OldContent, req.Content)
				// Re-show the prompt
				fmt.Fprintln(p.out, boxStyle.Render(content.String()))
			} else {
				fmt.Fprintln(p.out, "Diff not available for this action.")
			}
		case "?", "help":
			p.showHelp()
		default:
			fmt.Fprintln(p.out, "Invalid option. Press Y, N, D, A, or ? for help.")
		}
	}
}

func (p *InteractiveApprovalPrompter) showDiff(path, oldContent, newContent string) {
	edits := myers.ComputeEdits(span.URIFromPath(path), oldContent, newContent)
	unified := gotextdiff.ToUnified(path+" (current)", path+" (new)", oldContent, edits)
	diff := fmt.Sprint(unified)
	
	diffStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	addStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	delStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	
	fmt.Fprintln(p.out)
	for _, line := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(line, "+"):
			fmt.Fprintln(p.out, addStyle.Render(line))
		case strings.HasPrefix(line, "-"):
			fmt.Fprintln(p.out, delStyle.Render(line))
		default:
			fmt.Fprintln(p.out, diffStyle.Render(line))
		}
	}
	fmt.Fprintln(p.out)
}

func (p *InteractiveApprovalPrompter) showHelp() {
	help := `
File Modification Approval Help:

  Y  Approve this file modification
  N  Deny this request (agent will be notified)
  D  Show diff between current and new content
  A  Approve this and all future requests in this session
  ?  Show this help message

Press Enter after your choice.
`
	fmt.Fprintln(p.out, help)
}

func actionToVerb(action string) string {
	switch action {
	case "create":
		return "create"
	case "update":
		return "update"
	case "delete":
		return "delete"
	default:
		return "modify"
	}
}

// AutoApprovalPrompter always approves requests (for --no-jodas mode).
type AutoApprovalPrompter struct{}

// Prompt always returns approved.
func (p *AutoApprovalPrompter) Prompt(req ApprovalRequest) (ApprovalResponse, error) {
	return ApprovalResponse{Approved: true}, nil
}
