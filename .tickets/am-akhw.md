---
id: am-akhw
status: open
deps: []
links: []
created: 2026-02-02T02:56:30Z
type: task
priority: 2
assignee: Alex Cabrera
---
# Refactor run.Runner to reduce coupling

run.Runner has 13 fields and depends on 7+ packages (config, session, memory, smallmodel, ui, agent, plugins). Consider extracting:
- MemoryManager for memory operations
- SessionManager for session operations
- ToolManager for tool loading

Also runChatWithHistory() is 166 lines - extract buildFantasyAgent() and handleStreamCallbacks().

