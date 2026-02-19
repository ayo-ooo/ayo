#!/bin/bash
# =============================================================================
# collect-all.sh - Collect all diagnostic information into a single report
# =============================================================================
#
# DESCRIPTION:
#   Runs all diagnostic scripts and combines their output into a comprehensive
#   report. Use this when reporting bugs or asking for help - it captures
#   everything needed to understand the system state.
#
# USAGE:
#   ./debug/collect-all.sh [--json] [--output <file>]
#
# OPTIONS:
#   --json           Output in JSON format
#   --output <file>  Write to file instead of stdout
#
# OUTPUT:
#   Combined output from:
#   - system-info.sh
#   - sandbox-status.sh (with --verbose)
#   - daemon-status.sh (with --logs)
#   - ayo configuration summary
#   - Recent session information
#
# EXAMPLES:
#   ./debug/collect-all.sh                              # Print to terminal
#   ./debug/collect-all.sh --output debug-report.txt    # Save to file
#   ./debug/collect-all.sh | pbcopy                     # Copy to clipboard
#
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
JSON_OUTPUT=false
OUTPUT_FILE=""

for arg in "$@"; do
    case $arg in
        --json)
            JSON_OUTPUT=true
            ;;
        --output)
            shift
            OUTPUT_FILE="$1"
            ;;
    esac
done

# Redirect output if file specified
if [[ -n "$OUTPUT_FILE" ]]; then
    exec > "$OUTPUT_FILE"
fi

if $JSON_OUTPUT; then
    echo "{"
    echo '  "report_timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",'
    echo '  "report_version": "1.0",'
    
    echo '  "system": '
    "$SCRIPT_DIR/system-info.sh" --json
    echo ','
    
    echo '  "sandbox": '
    "$SCRIPT_DIR/sandbox-status.sh" --json
    echo ','
    
    echo '  "daemon": '
    "$SCRIPT_DIR/daemon-status.sh" --json
    
    echo "}"
else
    cat <<'HEADER'
╔═══════════════════════════════════════════════════════════════════════════════╗
║                                                                               ║
║                         AYO DIAGNOSTIC REPORT                                 ║
║                                                                               ║
╚═══════════════════════════════════════════════════════════════════════════════╝

  This report contains comprehensive diagnostic information for debugging
  ayo issues. Include the entire report when asking for help.

HEADER
    echo "  Generated: $(date)"
    echo ""

    # System Info
    "$SCRIPT_DIR/system-info.sh"

    # Sandbox Status
    "$SCRIPT_DIR/sandbox-status.sh" --verbose

    # Daemon Status
    "$SCRIPT_DIR/daemon-status.sh" --logs

    # Ayo Configuration
    echo ""
    echo "═══════════════════════════════════════════════════════════════════════════════"
    echo "  AYO CONFIGURATION"
    echo "═══════════════════════════════════════════════════════════════════════════════"
    
    CONFIG_FILE="${XDG_CONFIG_HOME:-$HOME/.config}/ayo/ayo.json"
    echo ""
    echo "───────────────────────────────────────────────────────────────────────────────"
    echo "  Config File"
    echo "───────────────────────────────────────────────────────────────────────────────"
    if [[ -f "$CONFIG_FILE" ]]; then
        echo "  Path: $CONFIG_FILE"
        echo "  Contents (sensitive values redacted):"
        # Redact API keys and secrets
        cat "$CONFIG_FILE" 2>/dev/null | sed -E 's/("api_key"|"secret"|"token"|"password"):\s*"[^"]+"/\1: "[REDACTED]"/g' | sed 's/^/  /' || echo "  Cannot read config file"
    else
        echo "  Config file not found at: $CONFIG_FILE"
    fi

    # Agents
    echo ""
    echo "───────────────────────────────────────────────────────────────────────────────"
    echo "  Installed Agents"
    echo "───────────────────────────────────────────────────────────────────────────────"
    if command -v ayo &>/dev/null; then
        ayo agents list --quiet 2>&1 | sed 's/^/  /' || echo "  Error listing agents"
    else
        echo "  ayo command not found"
    fi

    # Recent Sessions
    echo ""
    echo "───────────────────────────────────────────────────────────────────────────────"
    echo "  Recent Sessions"
    echo "───────────────────────────────────────────────────────────────────────────────"
    if command -v ayo &>/dev/null; then
        ayo sessions list -n 5 2>&1 | sed 's/^/  /' || echo "  Error listing sessions"
    else
        echo "  ayo command not found"
    fi

    # Docker Images
    echo ""
    echo "───────────────────────────────────────────────────────────────────────────────"
    echo "  Ayo Docker Images"
    echo "───────────────────────────────────────────────────────────────────────────────"
    docker images --filter "reference=*ayo*" --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}" 2>/dev/null | sed 's/^/  /' || echo "  Cannot list Docker images"

    # Mounts configuration
    echo ""
    echo "───────────────────────────────────────────────────────────────────────────────"
    echo "  Mount Permissions"
    echo "───────────────────────────────────────────────────────────────────────────────"
    MOUNTS_FILE="${XDG_DATA_HOME:-$HOME/.local/share}/ayo/mounts.json"
    if [[ -f "$MOUNTS_FILE" ]]; then
        cat "$MOUNTS_FILE" 2>/dev/null | sed 's/^/  /' || echo "  Cannot read mounts file"
    else
        echo "  No mounts.json found (no persistent mounts configured)"
    fi

    echo ""
    echo "═══════════════════════════════════════════════════════════════════════════════"
    echo "  END OF DIAGNOSTIC REPORT"
    echo "═══════════════════════════════════════════════════════════════════════════════"
    echo ""
fi
