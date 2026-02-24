---
id: ayo-ydub
status: closed
deps: []
links: []
created: 2026-02-23T22:15:02Z
type: task
priority: 1
assignee: Alex Cabrera
parent: ayo-6h19
tags: [removal, server]
---
# Remove internal/server/ package

Delete the REST API server package. This is the largest single removal (~2500 lines).

## Why Remove?

The REST API server was built for:
- Web interface (being removed)
- Mobile app connectivity (not a priority)
- External integrations (premature)

Ayo is a **CLI-first tool**. The daemon handles background tasks, but doesn't need HTTP endpoints.

## Files to Delete

```
internal/server/
├── server.go           # Main HTTP server
├── handlers.go         # Request handlers
├── chat.go             # Chat/streaming endpoints
├── sse.go              # Server-sent events
├── session_manager.go  # Session tracking
├── sandbox_manager.go  # Sandbox lifecycle via HTTP
├── web_client.go       # Static file serving
├── tunnel/             # Cloudflare tunnel integration
│   └── tunnel.go
├── qrcode.go           # QR code for mobile
└── *_test.go           # Tests
```

## Verification Steps

1. Delete `internal/server/` directory
2. Remove imports from `cmd/ayo/serve.go` (will be deleted separately)
3. Run `go build ./...` - should fail only on serve.go
4. Run `go test ./...` - should pass

## Acceptance Criteria

- [ ] `internal/server/` directory deleted
- [ ] No remaining imports of `internal/server` in codebase
- [ ] `go build ./cmd/ayo/...` passes (after serve.go removed)
- [ ] `go test ./...` passes
