#!/bin/bash
set -e

# Check for Crush and offer to install it
check_crush() {
    if command -v crush &> /dev/null; then
        echo "Crush detected: $(command -v crush)"
        return 0
    fi

    echo ""
    echo "Crush is not installed."
    echo ""
    echo "Crush is an AI-powered coding agent that ayo uses for all source code"
    echo "creation and modification tasks. Without it, coding features will be"
    echo "unavailable."
    echo ""
    read -p "Would you like to install Crush now? [Y/n] " -n 1 -r
    echo ""

    if [[ $REPLY =~ ^[Nn]$ ]]; then
        echo "Skipping Crush installation. You can install it later with:"
        echo "  go install github.com/charmbracelet/crush@latest"
        return 0
    fi

    echo "Installing Crush..."
    go install github.com/charmbracelet/crush@latest

    if command -v crush &> /dev/null; then
        echo "Crush installed successfully: $(command -v crush)"
    else
        echo ""
        echo "Warning: Crush was installed but is not in your PATH."
        echo "Make sure your Go bin directory is in your PATH:"
        echo "  export PATH=\"\$PATH:\$(go env GOPATH)/bin\""
    fi
}

# Check for Crush first
check_crush

echo ""

# Determine if we're on an unmodified main branch in sync with origin
branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
dirty=$(git status --porcelain 2>/dev/null)
behind_ahead=$(git rev-list --left-right --count origin/main...HEAD 2>/dev/null || echo "0 0")

if [[ "$branch" == "main" && -z "$dirty" && "$behind_ahead" == "0	0" ]]; then
    # Clean main branch in sync with origin - install to standard location
    echo "Installing to standard GOBIN location..."
    go install ./cmd/ayo
    echo ""
    ayo setup
else
    # Any other state - install to local .local/bin
    echo "Installing to .local/bin/ (branch: $branch, dirty: ${dirty:+yes}${dirty:-no})..."
    mkdir -p .local/bin
    GOBIN="$(pwd)/.local/bin" go install ./cmd/ayo
    echo ""
    .local/bin/ayo setup
fi
