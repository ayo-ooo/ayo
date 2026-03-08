// Package squads provides squad management for agent team coordination.
package squads

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexcabrera/ayo/internal/debug"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/pelletier/go-toml/v2"
)

// TeamProject represents a team project loaded from a directory with team.toml
// This is the new team project format for the build system.
type TeamProject struct {
	// Name is the team name from team.toml
	Name string

	// Config is the team configuration
	Config *TeamConfig

	// Directory is the project directory
	Directory string

	// Constitution is the team's SQUAD.md constitution
	Constitution *Constitution

	// Schemas contains input/output JSON schemas for validation
	Schemas *SquadSchemas

	// Agents contains the list of agent paths
	Agents []string
}

// LoadTeamProject loads a team project from a directory.
// It looks for team.toml and loads the team configuration.
func LoadTeamProject(teamDir string) (*TeamProject, error) {
	// Load team configuration
	config, err := LoadTeamConfigFromTOML(teamDir)
	if err != nil {
		return nil, fmt.Errorf("load team config: %w", err)
	}
	if config == nil {
		return nil, fmt.Errorf("no team.toml found in %s", teamDir)
	}

	// Load constitution
	constitution, err := LoadTeamConstitution(teamDir)
	if err != nil {
		return nil, fmt.Errorf("load team constitution: %w", err)
	}

	// Load schemas
	schemas, err := LoadTeamSchemas(teamDir)
	if err != nil {
		return nil, fmt.Errorf("load team schemas: %w", err)
	}

	// Build agent list
	var agents []string
	for agentName, agentConfig := range config.Agents {
		agents = append(agents, agentName)
		// Ensure agent path is relative to team directory
		if !filepath.IsAbs(agentConfig.Path) {
			agentConfig.Path = filepath.Join(teamDir, agentConfig.Path)
		}
		config.Agents[agentName] = agentConfig
	}

	debug.Log("loaded team project", "name", config.Team.Name, "dir", teamDir, "agents", len(agents))

	return &TeamProject{
		Name:         config.Team.Name,
		Config:       config,
		Directory:    teamDir,
		Constitution: constitution,
		Schemas:      schemas,
		Agents:       agents,
	}, nil
}

// TryLoadTeamFromCurrentDir tries to load a team project from the current directory.
// Returns nil if no team.toml is found.
func TryLoadTeamFromCurrentDir() (*TeamProject, error) {
	// Check if team.toml exists in current directory
	if !TeamConfigExists(".") {
		return nil, nil
	}

	return LoadTeamProject(".")
}

// TryLoadTeamFromDir tries to load a team project from a specific directory.
// Returns nil if no team.toml is found.
func TryLoadTeamFromDir(teamDir string) (*TeamProject, error) {
	// Check if team.toml exists
	if !TeamConfigExists(teamDir) {
		return nil, nil
	}

	return LoadTeamProject(teamDir)
}

// ListTeamAgents returns the list of agents in a team project.
func (tp *TeamProject) ListTeamAgents() []string {
	return tp.Agents
}

// GetAgentPath returns the path to an agent's directory.
func (tp *TeamProject) GetAgentPath(agentName string) (string, bool) {
	if agentConfig, exists := tp.Config.Agents[agentName]; exists {
		return agentConfig.Path, true
	}
	return "", false
}

// GetWorkspacePath returns the path to the team's workspace.
func (tp *TeamProject) GetWorkspacePath() string {
	if tp.Config.Workspace.SharedPath != "" {
		return filepath.Join(tp.Directory, tp.Config.Workspace.SharedPath)
	}
	return paths.TeamWorkspaceDir(tp.Directory)
}

// GetOutputPath returns the path to the team's output directory.
func (tp *TeamProject) GetOutputPath() string {
	if tp.Config.Workspace.OutputPath != "" {
		return filepath.Join(tp.Directory, tp.Config.Workspace.OutputPath)
	}
	return filepath.Join(tp.GetWorkspacePath(), "results")
}

// EnsureTeamDirs creates all directories needed for a team project.
func EnsureTeamDirs(teamDir string) error {
	dirs := []string{
		paths.TeamWorkspaceDir(teamDir),
		paths.TeamAgentsDir(teamDir),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create team directory %s: %w", dir, err)
		}
	}
	return nil
}

// CreateDefaultTeamProject creates a default team project structure.
func CreateDefaultTeamProject(teamName, teamDir string, agentNames []string) error {
	// Create team.toml
	teamConfig := &TeamConfig{
		Team: struct {
			Name        string `toml:"name"`
			Description string `toml:"description"`
			Coordination string `toml:"coordination"`
		}{
			Name:        teamName,
			Description: "Team description",
			Coordination: "sequential",
		},
		Agents: make(map[string]struct {
			Path string `toml:"path"`
		}),
		Workspace: struct {
			SharedPath string `toml:"shared_path"`
			OutputPath string `toml:"output_path"`
		}{
			SharedPath: "workspace",
			OutputPath: "workspace/results",
		},
		Coordination: struct {
			Strategy      string `toml:"strategy"`
			MaxIterations int    `toml:"max_iterations"`
		}{
			Strategy:      "round-robin",
			MaxIterations: 5,
		},
	}

	// Add agents
	for _, agentName := range agentNames {
		teamConfig.Agents[agentName] = struct {
			Path string `toml:"path"`
		}{
			Path: filepath.Join("agents", agentName),
		}
	}

	// Create directories
	if err := EnsureTeamDirs(teamDir); err != nil {
		return err
	}

	// Create team.toml
	if err := saveTeamConfig(teamDir, teamConfig); err != nil {
		return err
	}

	// Create default constitution
	if err := CreateDefaultTeamConstitution(teamName, teamDir, agentNames); err != nil {
		return err
	}

	debug.Log("created default team project", "name", teamName, "dir", teamDir)
	return nil
}

// saveTeamConfig saves team configuration to team.toml
func saveTeamConfig(teamDir string, config *TeamConfig) error {
	teamPath := paths.TeamConfigPath(teamDir)
	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal team.toml: %w", err)
	}

	if err := os.WriteFile(teamPath, data, 0644); err != nil {
		return fmt.Errorf("write team.toml: %w", err)
	}

	debug.Log("saved team.toml", "dir", teamDir)
	return nil
}

// CreateDefaultTeamConstitution creates a default SQUAD.md template for a new team.
func CreateDefaultTeamConstitution(teamName, teamDir string, agentNames []string) error {
	var agentSection strings.Builder
	for _, agent := range agentNames {
		agentSection.WriteString(fmt.Sprintf(`### %s
**Role**: [Define this agent's role]
**Responsibilities**:
- [Responsibility 1]
- [Responsibility 2]

`, agent))
	}

	if len(agentNames) == 0 {
		agentSection.WriteString(`### @agent
**Role**: [Define this agent's role]
**Responsibilities**:
- [Responsibility 1]
- [Responsibility 2]

`)
	}

	template := fmt.Sprintf(`# Team: %s

## Mission

[Describe what this team is trying to accomplish in 1-2 paragraphs.]

## Context

[Background information all agents need: project constraints, technical decisions,
external dependencies, deadlines, or any shared knowledge.]

## Agents

%s
## Coordination

[How agents should work together: handoff protocols, communication patterns,
dependency chains, blocking rules.]

## Guidelines

[Specific rules or preferences for this team: coding style, testing requirements,
commit conventions, review process.]
`, teamName, agentSection.String())

	return SaveTeamConstitution(teamDir, template)
}