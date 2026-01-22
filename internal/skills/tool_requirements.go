package skills

// ToolSkillRequirement defines skills that must be attached when a tool is enabled.
type ToolSkillRequirement struct {
	// ToolName is the name of the tool (e.g., "agent_call").
	ToolName string
	// RequiredSkills is the list of skill names required by this tool.
	RequiredSkills []string
	// Reason explains why these skills are required.
	Reason string
}

// toolRequirements is the registry of tool-to-skill requirements.
var toolRequirements = []ToolSkillRequirement{
	{
		ToolName:       "agent_call",
		RequiredSkills: []string{"agent-discovery"},
		Reason:         "agent_call requires knowledge of available agents for delegation",
	},
	{
		ToolName:       "plan",
		RequiredSkills: []string{"planning"},
		Reason:         "plan tool requires the planning skill for effective task decomposition",
	},
}

// GetToolSkillRequirements returns all registered tool-skill requirements.
func GetToolSkillRequirements() []ToolSkillRequirement {
	return toolRequirements
}

// GetRequiredSkillsForTools returns skills required by the given tools.
// Returns a deduplicated list of skill names.
func GetRequiredSkillsForTools(tools []string) []string {
	toolSet := make(map[string]struct{}, len(tools))
	for _, t := range tools {
		toolSet[t] = struct{}{}
	}

	skillSet := make(map[string]struct{})
	for _, req := range toolRequirements {
		if _, ok := toolSet[req.ToolName]; ok {
			for _, s := range req.RequiredSkills {
				skillSet[s] = struct{}{}
			}
		}
	}

	skills := make([]string, 0, len(skillSet))
	for s := range skillSet {
		skills = append(skills, s)
	}
	return skills
}

// GetToolRequirementInfo returns requirement info for display purposes.
// Maps tool name to (required skills, reason).
type ToolRequirementInfo struct {
	ToolName       string
	RequiredSkills []string
	Reason         string
}

// GetToolRequirementsForTools returns requirement info for the given tools.
func GetToolRequirementsForTools(tools []string) []ToolRequirementInfo {
	toolSet := make(map[string]struct{}, len(tools))
	for _, t := range tools {
		toolSet[t] = struct{}{}
	}

	var result []ToolRequirementInfo
	for _, req := range toolRequirements {
		if _, ok := toolSet[req.ToolName]; ok {
			result = append(result, ToolRequirementInfo{
				ToolName:       req.ToolName,
				RequiredSkills: req.RequiredSkills,
				Reason:         req.Reason,
			})
		}
	}
	return result
}
