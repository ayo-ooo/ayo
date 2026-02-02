---
id: am-8bfq
status: closed
deps: []
links: []
created: 2026-02-02T02:56:09Z
type: task
priority: 1
assignee: Alex Cabrera
---
# Remove dead code: unused exports and packages

Remove identified dead code:
- internal/ui/anim/ (entire package unused - no imports found)
- internal/ui/styles/ (entire package unused)
- internal/ui/selector/ (entire package unused)
- ui.SelectAgentResult type (never used)
- ui.ToolSpinner interface (orphaned)
- run.NewPrintStreamHandler (unused, only WithSpinner version used)
- server.GenerateQRCodeToStdout (unused)
- shared.debugLogShared (debug code writing to /tmp)
- config.GetProviderDefaultModel/GetProviderDefaultSmallModel (could be unexported)

