---
id: ayo-1rw2
status: closed
deps: [ayo-15ou]
links: []
created: 2026-02-05T18:52:32Z
type: feature
priority: 2
assignee: Alex Cabrera
parent: ayo-3qkl
tags: [sandbox, files]
---
# File Transfer: ayo sandbox push/pull

Transfer files between host and sandbox. push copies host file into sandbox, pull copies sandbox file to host. Uses container cp or virtiofs mount.

## Acceptance Criteria

- ayo sandbox push <id> <local> <dest> works
- ayo sandbox pull <id> <src> <local> works
- Handles directories with -r flag
- Shows progress for large files
- Errors clearly on missing files

