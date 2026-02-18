package capabilities

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/squads"
)

// ComputeAgentHash computes a content hash for an agent.
// The hash includes the agent's description, system prompt, and skill names.
// This is used for lazy invalidation of embeddings - if the hash changes,
// the agent's capabilities need to be re-inferred.
func ComputeAgentHash(ag agent.Agent) string {
	h := sha256.New()

	// Include description
	h.Write([]byte(ag.Config.Description))
	h.Write([]byte("\n---\n"))

	// Include system prompt
	h.Write([]byte(ag.System))
	h.Write([]byte("\n---\n"))

	// Include skill names (sorted for consistency)
	skillNames := make([]string, len(ag.Skills))
	for i, skill := range ag.Skills {
		skillNames[i] = skill.Name
	}
	sort.Strings(skillNames)
	h.Write([]byte(strings.Join(skillNames, "|")))

	return hex.EncodeToString(h.Sum(nil))
}

// ComputeSquadHash computes a content hash for a squad.
// The hash is based on the squad's constitution content.
// This is used for lazy invalidation of embeddings - if the hash changes,
// the squad's embeddings need to be regenerated.
func ComputeSquadHash(constitution *squads.Constitution) string {
	if constitution == nil {
		return ""
	}

	h := sha256.New()
	h.Write([]byte(constitution.Raw))
	return hex.EncodeToString(h.Sum(nil))
}

// ComputeSquadHashFromContent computes a hash directly from constitution content.
// This is useful when you have the raw content but haven't parsed it yet.
func ComputeSquadHashFromContent(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))
}
