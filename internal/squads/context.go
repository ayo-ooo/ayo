package squads

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	// Raw is the raw markdown content.
	Raw string

	// SquadName is the squad this constitution belongs to.
	SquadName string
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

	return &Constitution{
		Raw:       string(data),
		SquadName: squadName,
	}, nil
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
