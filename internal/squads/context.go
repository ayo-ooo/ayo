package squads

import (
	"encoding/json"
	"os"
	"path/filepath"
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
