package squads

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/debug"
	"github.com/alexcabrera/ayo/internal/paths"
)

// SquadContext represents preserved context for a squad.
type SquadContext struct {
	// Session info
	LastSessionID   string    `json:"last_session_id,omitempty"`
	LastSessionTime time.Time `json:"last_session_time,omitempty"`
	SessionCount    int       `json:"session_count"`

	// Agent memories (agent handle -> memory content)
	AgentMemories map[string]string `json:"agent_memories,omitempty"`

	// General notes
	Notes []string `json:"notes,omitempty"`

	// Metadata
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// SaveContext saves the squad context to /.context/session.json.
func SaveContext(squadName string, ctx *SquadContext) error {
	contextDir := paths.SquadContextDir(squadName)
	if err := os.MkdirAll(contextDir, 0755); err != nil {
		return err
	}

	ctx.Modified = time.Now()
	if ctx.Created.IsZero() {
		ctx.Created = ctx.Modified
	}

	sessionPath := filepath.Join(contextDir, "session.json")
	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(sessionPath, data, 0644); err != nil {
		return err
	}

	debug.Log("saved squad context", "squad", squadName)
	return nil
}

// LoadContext loads the squad context from /.context/session.json.
func LoadContext(squadName string) (*SquadContext, error) {
	contextDir := paths.SquadContextDir(squadName)
	sessionPath := filepath.Join(contextDir, "session.json")

	data, err := os.ReadFile(sessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty context
			return &SquadContext{
				AgentMemories: make(map[string]string),
				Created:       time.Now(),
			}, nil
		}
		return nil, err
	}

	var ctx SquadContext
	if err := json.Unmarshal(data, &ctx); err != nil {
		return nil, err
	}

	if ctx.AgentMemories == nil {
		ctx.AgentMemories = make(map[string]string)
	}

	debug.Log("loaded squad context", "squad", squadName, "sessions", ctx.SessionCount)
	return &ctx, nil
}

// SaveAgentMemory saves an agent's memory to the context.
func SaveAgentMemory(squadName, agentHandle, memory string) error {
	ctx, err := LoadContext(squadName)
	if err != nil {
		return err
	}

	ctx.AgentMemories[agentHandle] = memory
	return SaveContext(squadName, ctx)
}

// LoadAgentMemory loads an agent's memory from the context.
func LoadAgentMemory(squadName, agentHandle string) (string, error) {
	ctx, err := LoadContext(squadName)
	if err != nil {
		return "", err
	}

	return ctx.AgentMemories[agentHandle], nil
}

// AddNote adds a note to the squad context.
func AddNote(squadName, note string) error {
	ctx, err := LoadContext(squadName)
	if err != nil {
		return err
	}

	ctx.Notes = append(ctx.Notes, note)
	return SaveContext(squadName, ctx)
}

// RecordSession records a new session in the context.
func RecordSession(squadName, sessionID string) error {
	ctx, err := LoadContext(squadName)
	if err != nil {
		return err
	}

	ctx.LastSessionID = sessionID
	ctx.LastSessionTime = time.Now()
	ctx.SessionCount++

	return SaveContext(squadName, ctx)
}

// ClearContext removes all context data for a squad.
func ClearContext(squadName string) error {
	contextDir := paths.SquadContextDir(squadName)
	return os.RemoveAll(contextDir)
}

// Constitution represents the parsed SQUAD.md file.
type Constitution struct {
	// Raw is the raw markdown content (without frontmatter).
	Raw string

	// SquadName is the squad this constitution belongs to.
	SquadName string

	// Frontmatter contains parsed YAML frontmatter fields.
	Frontmatter ConstitutionFrontmatter
}

// ConstitutionFrontmatter represents the YAML frontmatter in SQUAD.md.
type ConstitutionFrontmatter struct {
	// Name is the squad name (optional, usually derived from directory).
	Name string `yaml:"name"`

	// Planners configures which planners this squad should use.
	// Falls back to global config if not specified.
	Planners config.PlannersConfig `yaml:"planners"`

	// Lead specifies which agent is the squad lead.
	// Defaults to @ayo if not specified.
	Lead string `yaml:"lead"`

	// InputAccepts specifies which agent receives input directly.
	// Defaults to Lead if not specified.
	InputAccepts string `yaml:"input_accepts"`

	// Agents lists agent handles in this squad (optional).
	// If not specified, agents are parsed from ### @agent sections.
	Agents []string `yaml:"agents"`
}

// WithDefaults returns a copy with default values applied.
func (f ConstitutionFrontmatter) WithDefaults() ConstitutionFrontmatter {
	result := f
	if result.Lead == "" {
		result.Lead = "@ayo"
	}
	if result.InputAccepts == "" {
		result.InputAccepts = result.Lead
	}
	result.Planners = result.Planners.WithDefaults()
	return result
}

// GetInputAcceptsAgent returns the agent that should receive direct input.
// Returns the normalized agent handle (with @ prefix).
func (f ConstitutionFrontmatter) GetInputAcceptsAgent() string {
	if f.InputAccepts == "" {
		if f.Lead == "" {
			return "@ayo"
		}
		return f.Lead
	}
	// Ensure @ prefix
	if len(f.InputAccepts) > 0 && f.InputAccepts[0] != '@' {
		return "@" + f.InputAccepts
	}
	return f.InputAccepts
}

// GetAgents returns the list of agent handles in this constitution.
// If agents are specified in frontmatter, those are returned.
// Otherwise, agent handles are parsed from ### @agent sections.
func (c *Constitution) GetAgents() []string {
	// If frontmatter specifies agents, use those
	if len(c.Frontmatter.Agents) > 0 {
		agents := make([]string, len(c.Frontmatter.Agents))
		for i, a := range c.Frontmatter.Agents {
			if len(a) > 0 && a[0] != '@' {
				agents[i] = "@" + a
			} else {
				agents[i] = a
			}
		}
		return agents
	}

	// Parse from markdown sections like "### @backend"
	return parseAgentSections(c.Raw)
}

// parseAgentSections extracts agent handles from ### @agent sections in markdown.
func parseAgentSections(content string) []string {
	var agents []string
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		// Look for ### @something
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "### @") {
			// Extract the agent handle
			rest := strings.TrimPrefix(trimmed, "### ")
			parts := strings.Fields(rest)
			if len(parts) > 0 && strings.HasPrefix(parts[0], "@") {
				agents = append(agents, parts[0])
			}
		}
	}
	return agents
}

// LoadConstitution loads the SQUAD.md file for a squad.
// Returns nil if the file doesn't exist.
func LoadConstitution(squadName string) (*Constitution, error) {
	path := paths.SquadConstitutionPath(squadName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read SQUAD.md: %w", err)
	}

	frontmatter, body, err := parseFrontmatter(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse SQUAD.md frontmatter: %w", err)
	}

	return &Constitution{
		Raw:         body,
		SquadName:   squadName,
		Frontmatter: frontmatter,
	}, nil
}

// parseFrontmatter extracts YAML frontmatter from markdown content.
// Frontmatter must be delimited by "---" at the start and end.
// Returns the parsed frontmatter, the remaining body, and any error.
func parseFrontmatter(content string) (ConstitutionFrontmatter, string, error) {
	var fm ConstitutionFrontmatter

	// Check for frontmatter delimiter
	if !strings.HasPrefix(content, "---") {
		return fm, content, nil
	}

	// Find the closing delimiter
	// Skip the first "---" and find the next one
	rest := content[3:]
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	}

	closeIdx := strings.Index(rest, "\n---")
	if closeIdx == -1 {
		// No closing delimiter, treat entire content as body
		return fm, content, nil
	}

	// Extract frontmatter YAML
	frontmatterYAML := rest[:closeIdx]

	// Extract body (skip the closing "---" and optional newline)
	body := rest[closeIdx+4:]
	if len(body) > 0 && body[0] == '\n' {
		body = body[1:]
	}

	// Parse YAML
	if err := yaml.Unmarshal([]byte(frontmatterYAML), &fm); err != nil {
		return fm, "", fmt.Errorf("invalid YAML: %w", err)
	}

	return fm, body, nil
}

// SaveConstitution saves the SQUAD.md file for a squad.
func SaveConstitution(squadName, content string) error {
	path := paths.SquadConstitutionPath(squadName)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write SQUAD.md: %w", err)
	}
	debug.Log("saved squad constitution", "squad", squadName)
	return nil
}

// SaveTeamConstitution saves the SQUAD.md file for a team project.
func SaveTeamConstitution(teamDir, content string) error {
	path := filepath.Join(teamDir, "SQUAD.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write SQUAD.md: %w", err)
	}
	debug.Log("saved team constitution", "team", teamDir)
	return nil
}

// LoadTeamConstitution loads the SQUAD.md file for a team project.
func LoadTeamConstitution(teamDir string) (*Constitution, error) {
	path := filepath.Join(teamDir, "SQUAD.md")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read SQUAD.md: %w", err)
	}

	frontmatter, body, err := parseFrontmatter(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse SQUAD.md frontmatter: %w", err)
	}

	return &Constitution{
		Raw:         body,
		SquadName:   filepath.Base(teamDir),
		Frontmatter: frontmatter,
	}, nil
}

// CreateDefaultConstitution creates a default SQUAD.md template for a new squad.
func CreateDefaultConstitution(squadName string, agents []string) error {
	var agentSection strings.Builder
	for _, agent := range agents {
		agentSection.WriteString(fmt.Sprintf(`### %s
**Role**: [Define this agent's role]
**Responsibilities**:
- [Responsibility 1]
- [Responsibility 2]

`, agent))
	}

	if len(agents) == 0 {
		agentSection.WriteString(`### @agent
**Role**: [Define this agent's role]
**Responsibilities**:
- [Responsibility 1]
- [Responsibility 2]

`)
	}

	template := fmt.Sprintf(`# Squad: %s

## Mission

[Describe what this squad is trying to accomplish in 1-2 paragraphs.]

## Context

[Background information all agents need: project constraints, technical decisions,
external dependencies, deadlines, or any shared knowledge.]

## Agents

%s
## Coordination

[How agents should work together: handoff protocols, communication patterns,
dependency chains, blocking rules.]

## Guidelines

[Specific rules or preferences for this squad: coding style, testing requirements,
commit conventions, review process.]
`, squadName, agentSection.String())

	return SaveConstitution(squadName, template)
}

// FormatForSystemPrompt formats the constitution for injection into an agent's system prompt.
func (c *Constitution) FormatForSystemPrompt() string {
	if c == nil || c.Raw == "" {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<squad_context>\n")
	sb.WriteString(strings.TrimSpace(c.Raw))
	sb.WriteString("\n</squad_context>")
	return sb.String()
}

// InjectConstitution adds squad constitution context to a system prompt.
// The constitution is inserted after the environment context but before the agent's persona.
func InjectConstitution(systemPrompt string, constitution *Constitution) string {
	if constitution == nil || constitution.Raw == "" {
		return systemPrompt
	}

	formatted := constitution.FormatForSystemPrompt()
	return systemPrompt + "\n\n" + formatted
}
