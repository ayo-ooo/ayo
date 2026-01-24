package agent

import (
	"strings"

	"github.com/alexcabrera/ayo/internal/builtin"
)

// BuildToolsPrompt returns system prompt instructions for tool usage
func BuildToolsPrompt(allowedTools []string) string {
	// Default to bash if no tools specified
	if len(allowedTools) == 0 {
		allowedTools = []string{"bash", "agent_call"}
	}

	hasBash := false
	hasAgentCall := false
	for _, t := range allowedTools {
		switch t {
		case "bash":
			hasBash = true
		case "agent_call":
			hasAgentCall = true
		}
	}

	if !hasBash && !hasAgentCall {
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

		b.WriteString("**CRITICAL - Check <delegate_context> first:**\n")
		b.WriteString("If a coding delegate is configured, use agent_call instead of bash for:\n")
		b.WriteString("- Creating projects or applications (even with scaffolding tools like create-react-app, vite, etc.)\n")
		b.WriteString("- Writing any source code files\n")
		b.WriteString("- Implementing features or making code changes\n")
		b.WriteString("The coding delegate handles the ENTIRE task including any scaffolding.\n\n")

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
		if len(agentInfos) == 0 {
			// No builtin agents available, skip agent_call section
		} else {
			b.WriteString("<agent_call>\n")
			b.WriteString("You have access to specialized builtin agents via the agent_call tool.\n\n")

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

			b.WriteString("The called agent runs as a subprocess and returns its complete response.\n")
			b.WriteString("You can then incorporate or process that response as needed.\n\n")

			b.WriteString("Required parameters:\n")
			b.WriteString("- `agent`: The agent handle (e.g., '@ayo')\n")
			b.WriteString("- `prompt`: The prompt/question to send to the agent\n\n")

			b.WriteString("Optional: `timeout_seconds` (default 120, max 300)\n")
			b.WriteString("</agent_call>\n\n")
		}
	}

	b.WriteString("</tools>")
	return b.String()
}
