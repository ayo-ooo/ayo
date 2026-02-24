#!/bin/bash
# Clean Slate Preparation Script
# 
# Removes all ayo state, config, and services to start fresh.
# Run this before beginning development on a clean environment.
#
# Usage: ./scripts/clean-slate.sh
#
# WARNING: This will destroy ALL ayo data, sandboxes, and config!

set -e

echo "=== Ayo Clean Slate Preparation ==="
echo ""

# Stop daemon if running
echo "Stopping any running daemons..."
if command -v ayo &> /dev/null; then
    ayo daemon stop 2>/dev/null || true
fi

# Kill any running sandboxes
echo "Destroying any running sandboxes..."
if command -v ayo &> /dev/null; then
    for sb in $(ayo sandbox list --json 2>/dev/null | jq -r '.[].id' 2>/dev/null); do
        echo "  Destroying sandbox: $sb"
        ayo sandbox destroy "$sb" 2>/dev/null || true
    done
fi

# Remove local state
echo "Removing local state (~/.local/share/ayo)..."
rm -rf ~/.local/share/ayo

# Remove config
echo "Removing config (~/.config/ayo)..."
rm -rf ~/.config/ayo

# macOS: Remove launchd service
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "Removing launchd service (macOS)..."
    launchctl unload ~/Library/LaunchAgents/land.charm.ayod.plist 2>/dev/null || true
    rm -f ~/Library/LaunchAgents/land.charm.ayod.plist
fi

# Linux: Remove systemd service
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo "Removing systemd service (Linux)..."
    systemctl --user stop ayo-daemon 2>/dev/null || true
    systemctl --user disable ayo-daemon 2>/dev/null || true
    rm -f ~/.config/systemd/user/ayo-daemon.service
    systemctl --user daemon-reload 2>/dev/null || true
fi

# Remove stale sockets
echo "Removing any stale sockets..."
rm -f /tmp/ayo*.sock 2>/dev/null || true
rm -f ~/.local/share/ayo/daemon.sock 2>/dev/null || true

echo ""
echo "=== Clean slate complete ==="
echo ""
echo "Verification (all should fail or show empty):"
echo "  ls ~/.local/share/ayo"
echo "  ls ~/.config/ayo"
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "  launchctl list | grep ayo"
fi
echo "  ayo daemon status"
echo ""
echo "Run 'ayo doctor' to verify clean state."
