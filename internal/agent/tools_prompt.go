package agent

import (
	"strings"

	"github.com/alexcabrera/ayo/internal/builtin"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/plugins"
)

// isSearchToolAvailable checks if a search tool is configured and available.
func isSearchToolAvailable() bool {
	// Check if there's a default_tools.search configured
	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return false
	}

	if cfg.DefaultTools == nil {
		return false
	}

	searchTool := cfg.DefaultTools["search"]
	if searchTool == "" {
		return false
	}

	// Check if the tool exists in plugins
	registry, err := plugins.LoadRegistry()
	if err != nil {
		return false
	}

	for _, plugin := range registry.ListEnabled() {
		for _, tool := range plugin.Tools {
			if tool == searchTool {
				return true
			}
		}
	}

	return false
}

// BuildToolsPrompt returns system prompt instructions for tool usage
func BuildToolsPrompt(allowedTools []string) string {
	// Default to bash if no tools specified
	if len(allowedTools) == 0 {
		allowedTools = []string{"bash", "agent_call"}
	}

	hasBash := false
	hasAgentCall := false
	hasMemory := false
	hasSearch := false
	for _, t := range allowedTools {
		switch t {
		case "bash":
			hasBash = true
		case "agent_call":
			hasAgentCall = true
		case "memory":
			hasMemory = true
		case "search":
			hasSearch = true
		}
	}

	// Check if search tool is actually available (not just in allowed list)
	searchAvailable := hasSearch && isSearchToolAvailable()

	if !hasBash && !hasAgentCall && !hasMemory && !searchAvailable {
		return ""
	}

	var b strings.Builder
	b.WriteString("<tools>\n\n")

	if hasBash {
		b.WriteString("<bash>\n")
		b.WriteString("You have a bash tool for executing shell commands on the local system.\n\n")

		b.WriteString("**ONLY** use bash when the user's request requires local system interaction (files, processes, installations, etc.).\n")
		b.WriteString("**DO NOT** use bash for questions that can be answered from knowledge, web searches, or other tools.\n")
		b.WriteString("**DO NOT** make gratuitous tool calls - no `ls`, `true`, `echo`, or any command unless it directly serves the user's request.\n\n")

		b.WriteString("**NEVER** call `date`, `timedatectl`, or any time-checking command - the current datetime is already in your system context.\n\n")

		b.WriteString("When bash IS needed:\n")
		b.WriteString("- Don't ask permission, just run the command\n")
		b.WriteString("- Don't explain what you'll do, just do it\n")
		b.WriteString("- Don't say \"you can run...\" - run it yourself\n")
		b.WriteString("- Report results, not instructions\n\n")

		b.WriteString("Examples:\n")
		b.WriteString("- \"what's in this directory\" → run `ls -la`, show results\n")
		b.WriteString("- \"install express\" → run `npm install express`, confirm done\n")
		b.WriteString("- \"how much disk space\" → run `df -h`, report numbers\n\n")

		b.WriteString("Required parameters:\n")
		b.WriteString("- `command`: The shell command\n")
		b.WriteString("- `description`: What you're doing (shown in UI)\n\n")

		b.WriteString("Optional: `timeout_seconds`, `working_dir`\n")
		b.WriteString("</bash>\n\n")
	}

	if hasAgentCall {
		// Only include agent_call section if there are builtin agents to call
		agentInfos := builtin.ListAgentInfo()
		pluginAgents := plugins.ListPluginAgents()

		if len(agentInfos) == 0 && len(pluginAgents) == 0 {
			// No agents available, skip agent_call section
		} else {
			b.WriteString("<agent_call>\n")
			b.WriteString("You have access to specialized agents via the agent_call tool.\n\n")

			if len(agentInfos) > 0 {
				b.WriteString("Available builtin agents:\n\n")

				// Dynamically list builtin agents with their delegation hints
				for _, info := range agentInfos {
					b.WriteString("### ")
					b.WriteString(info.Handle)
					b.WriteString("\n")
					if info.Description != "" {
						b.WriteString(info.Description)
						b.WriteString("\n")
					}
					if info.DelegateHint != "" {
						b.WriteString("**When to use**: ")
						b.WriteString(info.DelegateHint)
						b.WriteString("\n")
					}
					b.WriteString("\n")
				}
			}

			if len(pluginAgents) > 0 {
				b.WriteString("Available plugin agents:\n\n")

				for _, info := range pluginAgents {
					b.WriteString("### ")
					b.WriteString(info.Handle)
					b.WriteString("\n")
					b.WriteString("(from plugin: ")
					b.WriteString(info.PluginName)
					b.WriteString(")\n\n")
				}
			}

			b.WriteString("The called agent runs as a subprocess and returns its complete response.\n")
			b.WriteString("You can then incorporate or process that response as needed.\n\n")

			b.WriteString("Required parameters:\n")
			b.WriteString("- `agent`: The agent handle (e.g., '@ayo')\n")
			b.WriteString("- `prompt`: The prompt/question to send to the agent\n\n")

			b.WriteString("Optional: `timeout_seconds` (default 120, max 300)\n")
			b.WriteString("</agent_call>\n\n")
		}
	}

	if hasMemory {
		b.WriteString("<memory>\n")
		b.WriteString("You have a memory tool for managing persistent memories across sessions.\n\n")

		b.WriteString("**AUTOMATICALLY** store memories when the user:\n")
		b.WriteString("- Expresses preferences (\"I prefer...\", \"always use...\", \"never...\")\n")
		b.WriteString("- Corrects you (\"no, I meant...\", \"actually...\", \"that's wrong...\")\n")
		b.WriteString("- Shares facts about themselves or their project\n")
		b.WriteString("- Explicitly asks you to remember something\n\n")

		b.WriteString("**AUTOMATICALLY** search memories when:\n")
		b.WriteString("- The user asks about their preferences\n")
		b.WriteString("- You need context about the user or project\n")
		b.WriteString("- Starting a task that might have relevant past context\n\n")

		b.WriteString("Operations:\n")
		b.WriteString("- `search`: Find relevant memories semantically. Params: query (required), limit (optional)\n")
		b.WriteString("- `store`: Save new information. Params: content (required), category (optional - auto-detected if omitted)\n")
		b.WriteString("- `list`: Show all memories. Params: limit (optional)\n")
		b.WriteString("- `forget`: Remove a memory. Params: id (required)\n\n")

		b.WriteString("Categories (auto-detected if not specified):\n")
		b.WriteString("- preference: User preferences about tools, styles, workflows\n")
		b.WriteString("- fact: Facts about the user, project, or environment\n")
		b.WriteString("- correction: Corrections to agent behavior\n")
		b.WriteString("- pattern: Observed behavioral patterns\n\n")

		b.WriteString("Best practices:\n")
		b.WriteString("- Distill memories to their essence (\"User prefers TypeScript\" not \"The user said they prefer TypeScript\")\n")
		b.WriteString("- Let auto-categorization work - only specify category if you need to override\n")
		b.WriteString("- Search before storing to avoid duplicates\n")
		b.WriteString("</memory>\n\n")
	}

	if searchAvailable {
		b.WriteString("<search>\n")
		b.WriteString("You have a search tool for searching the web.\n\n")

		b.WriteString("Use this tool when:\n")
		b.WriteString("- User asks about current events, news, or recent information\n")
		b.WriteString("- User needs facts that may have changed since your training\n")
		b.WriteString("- User explicitly asks to search the web\n")
		b.WriteString("- You need to verify or look up current information\n\n")

		b.WriteString("Parameters:\n")
		b.WriteString("- `query` (required): Search terms, use + for spaces (e.g., \"latest+us+news\")\n")
		b.WriteString("- `categories` (optional): general, news, images, videos, science, it, files, social+media, music, map\n")
		b.WriteString("- `time_range` (optional): day, week, month, year\n")
		b.WriteString("- `language` (optional): Language code like 'en', 'de', 'fr'\n\n")

		b.WriteString("The tool returns JSON with results containing titles, URLs, and content snippets.\n")
		b.WriteString("Summarize findings and cite sources when presenting information.\n")
		b.WriteString("</search>\n\n")
	}

	b.WriteString("</tools>")
	return b.String()
}
