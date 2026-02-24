---
id: ayo-clns
status: open
deps: []
links: []
created: 2026-02-24T01:30:00Z
type: task
priority: 0
assignee: Alex Cabrera
parent: ayo-6h19
tags: [cleanup, setup]
---
# Clean slate preparation

Before beginning implementation, clean up all existing state to start fresh.

## Why

- No production agents or configs to preserve
- Clean state avoids interference from old code
- Simplifies debugging during development
- Removes any conflicting daemons or services

## Cleanup Script

```bash
#!/bin/bash
set -e

echo "Stopping any running daemons..."
ayo daemon stop 2>/dev/null || true

echo "Killing any running sandboxes..."
for sb in $(ayo sandbox list --json 2>/dev/null | jq -r '.[].id' 2>/dev/null); do
    echo "  Destroying $sb"
    ayo sandbox destroy "$sb" 2>/dev/null || true
done

echo "Removing local state..."
rm -rf ~/.local/share/ayo

echo "Removing config..."
rm -rf ~/.config/ayo

echo "Removing launchd service (macOS)..."
launchctl unload ~/Library/LaunchAgents/land.charm.ayod.plist 2>/dev/null || true
rm -f ~/Library/LaunchAgents/land.charm.ayod.plist

echo "Removing systemd service (Linux)..."
systemctl --user stop ayo-daemon 2>/dev/null || true
systemctl --user disable ayo-daemon 2>/dev/null || true
rm -f ~/.config/systemd/user/ayo-daemon.service

echo "Removing any stale sockets..."
rm -f /tmp/ayo*.sock 2>/dev/null || true

echo "Done. Run 'ayo doctor' to verify clean state."
```

## Save Script Location

Save to: `scripts/clean-slate.sh`

## Verification

After running:
- `ls ~/.local/share/ayo` should fail (not found)
- `ls ~/.config/ayo` should fail (not found)
- `launchctl list | grep ayo` should show nothing
- `ayo daemon status` should report "not running"

## Acceptance Criteria

- [ ] Script created at `scripts/clean-slate.sh`
- [ ] Script tested on macOS
- [ ] Script tested on Linux (if available)
- [ ] Clean state verified
