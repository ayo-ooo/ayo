package server

import (
	"encoding/json"
	"net/http"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/config"
)

// AgentInfo represents an agent for API responses.
type AgentInfo struct {
	Handle      string   `json:"handle"`
	Description string   `json:"description,omitempty"`
	Model       string   `json:"model,omitempty"`
	BuiltIn     bool     `json:"built_in"`
	Tools       []string `json:"tools,omitempty"`
	Skills      []string `json:"skills,omitempty"`
	HasMemory   bool     `json:"has_memory"`
}

// AgentHandler handles agent-related API requests.
type AgentHandler struct {
	config config.Config
}

// NewAgentHandler creates a new agent handler.
func NewAgentHandler(cfg config.Config) *AgentHandler {
	return &AgentHandler{config: cfg}
}

// ListAgents returns all available agents.
func (h *AgentHandler) ListAgents(w http.ResponseWriter, r *http.Request) {
	handles, err := agent.ListHandles(h.config)
	if err != nil {
		http.Error(w, "failed to list agents: "+err.Error(), http.StatusInternalServerError)
		return
	}

	agents := make([]AgentInfo, 0, len(handles))
	for _, handle := range handles {
		ag, err := agent.Load(h.config, handle)
		if err != nil {
			// Skip agents that fail to load
			continue
		}
		agents = append(agents, agentToInfo(ag))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

// GetAgent returns details for a specific agent.
func (h *AgentHandler) GetAgent(w http.ResponseWriter, r *http.Request) {
	handle := r.PathValue("handle")
	if handle == "" {
		http.Error(w, "handle required", http.StatusBadRequest)
		return
	}

	// Normalize handle (add @ if missing)
	if handle[0] != '@' {
		handle = "@" + handle
	}

	ag, err := agent.Load(h.config, handle)
	if err != nil {
		http.Error(w, "agent not found: "+handle, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agentToInfo(ag))
}

// agentToInfo converts an agent to API response format.
func agentToInfo(ag agent.Agent) AgentInfo {
	skills := make([]string, 0, len(ag.Skills))
	for _, s := range ag.Skills {
		skills = append(skills, s.Name)
	}

	return AgentInfo{
		Handle:      ag.Handle,
		Description: ag.Config.Description,
		Model:       ag.Model,
		BuiltIn:     ag.BuiltIn,
		Tools:       ag.Config.AllowedTools,
		Skills:      skills,
		HasMemory:   ag.Config.Memory.Enabled,
	}
}
