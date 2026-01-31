# Work In Progress: Agent Call Tool Rendering

## Summary

Implementing proper rendering of `agent_call` tool output in the TUI. The goal is to show the sub-agent conversation (prompt + response) threaded under the tool call indicator, similar to how bash shows command output.

## Changes Made

### 1. Agent Call Restrictions (`internal/agent/agent.go`)
- Added `AllowedAgents []string` field to `Config` struct
- Added `CanCallAgent(targetHandle)` method for checking permissions
- Supports exact handles (`@crush`) and namespace patterns (`@ayo.*`)
- Empty `allowed_agents` = all agents allowed (default for `@ayo`)

### 2. Agent Call Executor (`internal/run/run.go`)
- Updated `agentCallExecutor` to use `CanCallAgent` for restrictions
- Removed hardcoded builtin/plugin-only restriction
- Sub-runner now uses `NullWriter` to silence streaming output
- Removed direct `ui.PrintSubAgentStart/End` calls that were printing to terminal in TUI mode
- Returns structured JSON response: `{"agent":"@handle", "prompt":"...", "response":"...", "duration":"..."}`

### 3. Shared Renderer (`internal/ui/shared/toolrender_agentcall.go`)
- Created `AgentCallRenderer` following the bash pattern
- Parses `RawOutput` as JSON (not metadata)
- Sets label to agent handle, header shows truncated prompt
- Body shows response as markdown section
- Added debug logging to `/tmp/ayo_debug.log`

### 4. TUI Renderer (`internal/ui/chat/messages/renderers.go`)
- Updated `agentCallRenderer` to use shared renderer
- Registered as `agent_call` (was incorrectly `agent`)
- Added debug logging to trace rendering

### 5. Tool Call Ordering (`internal/ui/chat/chat.go`)
- Tool calls now render inline with their associated messages
- Tool calls appear BEFORE assistant response (correct execution order)
- Added `ToolCallIDs` to message struct for association
- Added `pendingToolCallIDs` to track in-progress tool calls

## Issues Being Debugged

From the screenshot, several issues were observed:
1. Multiple `@woofers` tool calls showing with just prompt, no response
2. Stray `?` and `completed (1s)` text floating in UI
3. `sub-agent` label appearing incorrectly

### Root Causes Identified
- `ui.PrintSubAgentStart/End` were printing directly to terminal even in TUI mode - **FIXED**
- Renderer was registered as `agent` instead of `agent_call` - **FIXED**

### Still Investigating
- Whether `RawOutput` is populated correctly with JSON
- Whether JSON parsing succeeds

## Debug Logging

Debug logging has been added to trace the rendering flow. After running ayo with an agent_call:

```bash
cat /tmp/ayo_debug.log
```

The log will show:
- Tool call ID, name, input
- Result content and metadata
- Parsed render output (label, header params, sections)
- Whether JSON parsing succeeded

## Testing Steps

1. Run `ayo` in interactive mode
2. Ask it to call another agent (e.g., "create an agent called @test and send it a message")
3. Observe the tool call rendering
4. Check `/tmp/ayo_debug.log` for detailed trace

## Expected Behavior

```
● @test Hey test, how are you?

  [Sub-agent response rendered as markdown here]

✓ completed (1.2s)
```

## Files Modified

- `internal/agent/agent.go` - AllowedAgents config
- `internal/agent/agent_test.go` - CanCallAgent tests
- `internal/run/run.go` - agentCallExecutor changes
- `internal/run/fantasy_tools.go` - AgentCallParams description
- `internal/ui/shared/toolrender_agentcall.go` - New renderer
- `internal/ui/shared/toolrender.go` - SectionSubAgent type
- `internal/ui/chat/messages/renderers.go` - TUI renderer
- `internal/ui/chat/chat.go` - Tool call ordering
- `internal/ui/chat/messages/tree.go` - RenderByIDs method
