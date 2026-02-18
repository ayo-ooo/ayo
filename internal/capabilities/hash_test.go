package capabilities

import (
	"testing"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/skills"
	"github.com/alexcabrera/ayo/internal/squads"
)

func TestComputeAgentHash_Consistent(t *testing.T) {
	ag := agent.Agent{
		System: "You are a helpful assistant.",
		Config: agent.Config{
			Description: "A helpful agent",
		},
		Skills: []skills.Metadata{
			{Name: "skill-a"},
			{Name: "skill-b"},
		},
	}

	hash1 := ComputeAgentHash(ag)
	hash2 := ComputeAgentHash(ag)

	if hash1 != hash2 {
		t.Errorf("hash should be consistent: %s != %s", hash1, hash2)
	}

	if hash1 == "" {
		t.Error("hash should not be empty")
	}

	// SHA256 produces 64 hex characters
	if len(hash1) != 64 {
		t.Errorf("hash should be 64 characters, got %d", len(hash1))
	}
}

func TestComputeAgentHash_DifferentContent(t *testing.T) {
	ag1 := agent.Agent{
		System: "You are agent 1.",
		Config: agent.Config{
			Description: "Agent 1",
		},
	}

	ag2 := agent.Agent{
		System: "You are agent 2.",
		Config: agent.Config{
			Description: "Agent 2",
		},
	}

	hash1 := ComputeAgentHash(ag1)
	hash2 := ComputeAgentHash(ag2)

	if hash1 == hash2 {
		t.Error("different agents should have different hashes")
	}
}

func TestComputeAgentHash_SkillOrderIndependent(t *testing.T) {
	ag1 := agent.Agent{
		System: "Test agent.",
		Skills: []skills.Metadata{
			{Name: "alpha"},
			{Name: "beta"},
			{Name: "gamma"},
		},
	}

	ag2 := agent.Agent{
		System: "Test agent.",
		Skills: []skills.Metadata{
			{Name: "gamma"},
			{Name: "alpha"},
			{Name: "beta"},
		},
	}

	hash1 := ComputeAgentHash(ag1)
	hash2 := ComputeAgentHash(ag2)

	if hash1 != hash2 {
		t.Error("skill order should not affect hash")
	}
}

func TestComputeAgentHash_DifferentSkills(t *testing.T) {
	ag1 := agent.Agent{
		System: "Test agent.",
		Skills: []skills.Metadata{
			{Name: "skill-a"},
		},
	}

	ag2 := agent.Agent{
		System: "Test agent.",
		Skills: []skills.Metadata{
			{Name: "skill-b"},
		},
	}

	hash1 := ComputeAgentHash(ag1)
	hash2 := ComputeAgentHash(ag2)

	if hash1 == hash2 {
		t.Error("different skills should produce different hashes")
	}
}

func TestComputeAgentHash_EmptyAgent(t *testing.T) {
	ag := agent.Agent{}

	hash := ComputeAgentHash(ag)

	if hash == "" {
		t.Error("hash should not be empty even for empty agent")
	}
}

func TestComputeSquadHash_Consistent(t *testing.T) {
	constitution := &squads.Constitution{
		Raw:       "# Mission\nBuild great things.",
		SquadName: "test-squad",
	}

	hash1 := ComputeSquadHash(constitution)
	hash2 := ComputeSquadHash(constitution)

	if hash1 != hash2 {
		t.Errorf("hash should be consistent: %s != %s", hash1, hash2)
	}

	if hash1 == "" {
		t.Error("hash should not be empty")
	}

	// SHA256 produces 64 hex characters
	if len(hash1) != 64 {
		t.Errorf("hash should be 64 characters, got %d", len(hash1))
	}
}

func TestComputeSquadHash_DifferentContent(t *testing.T) {
	constitution1 := &squads.Constitution{
		Raw: "# Mission 1",
	}

	constitution2 := &squads.Constitution{
		Raw: "# Mission 2",
	}

	hash1 := ComputeSquadHash(constitution1)
	hash2 := ComputeSquadHash(constitution2)

	if hash1 == hash2 {
		t.Error("different constitutions should have different hashes")
	}
}

func TestComputeSquadHash_NilConstitution(t *testing.T) {
	hash := ComputeSquadHash(nil)

	if hash != "" {
		t.Errorf("nil constitution should return empty hash, got %s", hash)
	}
}

func TestComputeSquadHashFromContent(t *testing.T) {
	content := "# Mission\nBuild great things."

	hash1 := ComputeSquadHashFromContent(content)
	hash2 := ComputeSquadHashFromContent(content)

	if hash1 != hash2 {
		t.Errorf("hash should be consistent: %s != %s", hash1, hash2)
	}

	// Should match ComputeSquadHash with same content
	constitution := &squads.Constitution{Raw: content}
	hash3 := ComputeSquadHash(constitution)

	if hash1 != hash3 {
		t.Errorf("ComputeSquadHashFromContent should match ComputeSquadHash: %s != %s", hash1, hash3)
	}
}

func TestComputeSquadHashFromContent_Empty(t *testing.T) {
	hash := ComputeSquadHashFromContent("")

	if hash == "" {
		t.Error("hash should not be empty even for empty content")
	}
}

func TestComputeAgentHash_DescriptionChanges(t *testing.T) {
	ag1 := agent.Agent{
		System: "Test agent.",
		Config: agent.Config{
			Description: "Description 1",
		},
	}

	ag2 := agent.Agent{
		System: "Test agent.",
		Config: agent.Config{
			Description: "Description 2",
		},
	}

	hash1 := ComputeAgentHash(ag1)
	hash2 := ComputeAgentHash(ag2)

	if hash1 == hash2 {
		t.Error("different descriptions should produce different hashes")
	}
}
