#!/bin/bash
# =============================================================================
# system-info.sh - Collect comprehensive system information for debugging
# =============================================================================
#
# DESCRIPTION:
#   Gathers system-level diagnostic information about the host environment
#   and ayo configuration. This is the first script to run when diagnosing
#   any ayo issue.
#
# USAGE:
#   ./debug/system-info.sh [--json]
#
# OPTIONS:
#   --json    Output in JSON format (for machine parsing)
#
# OUTPUT:
#   - OS and architecture
#   - Sandbox provider status
#   - ayo version and config paths
#   - Resource usage (memory, disk)
#   - Environment variables relevant to ayo
#
# EXAMPLES:
#   ./debug/system-info.sh                    # Human-readable output
#   ./debug/system-info.sh | pbcopy           # Copy to clipboard
#   ./debug/system-info.sh --json | jq .      # Parse as JSON
#
# =============================================================================

set -euo pipefail

# Detect dev mode: if running from project directory with .local/bin/ayo
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DEV_MODE=false
AYO_BINARY=""
CONFIG_DIR=""
DATA_DIR=""

if [[ -x "$PROJECT_DIR/.local/bin/ayo" ]]; then
    DEV_MODE=true
    AYO_BINARY="$PROJECT_DIR/.local/bin/ayo"
    CONFIG_DIR="$PROJECT_DIR/.config/ayo"
    DATA_DIR="$PROJECT_DIR/.local/share/ayo"
elif command -v ayo &>/dev/null; then
    AYO_BINARY="$(which ayo)"
    CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/ayo"
    DATA_DIR="${XDG_DATA_HOME:-$HOME/.local/share}/ayo"
fi

# Detect sandbox provider
SANDBOX_PROVIDER="none"
if [[ "$(uname -s)" == "Darwin" ]]; then
    if command -v container &>/dev/null; then
        SANDBOX_PROVIDER="apple-container"
    fi
elif [[ "$(uname -s)" == "Linux" ]]; then
    if command -v systemd-nspawn &>/dev/null; then
        SANDBOX_PROVIDER="systemd-nspawn"
    fi
fi

JSON_OUTPUT=false
if [[ "${1:-}" == "--json" ]]; then
    JSON_OUTPUT=true
fi

divider() {
    if ! $JSON_OUTPUT; then
        echo ""
        echo "═══════════════════════════════════════════════════════════════════════════════"
        echo "  $1"
        echo "═══════════════════════════════════════════════════════════════════════════════"
    fi
}

section() {
    if ! $JSON_OUTPUT; then
        echo ""
        echo "───────────────────────────────────────────────────────────────────────────────"
        echo "  $1"
        echo "───────────────────────────────────────────────────────────────────────────────"
    fi
}

if $JSON_OUTPUT; then
    # JSON output mode
    cat <<EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "source": "host",
  "script": "system-info.sh",
  "dev_mode": $DEV_MODE,
  "system": {
    "os": "$(uname -s)",
    "os_version": "$(uname -r)",
    "arch": "$(uname -m)",
    "hostname": "$(hostname)",
    "user": "$(whoami)"
  },
  "sandbox": {
    "provider": "$SANDBOX_PROVIDER",
    "available": $([ "$SANDBOX_PROVIDER" != "none" ] && echo true || echo false)
  },
  "ayo": {
    "version": "$($AYO_BINARY --version 2>/dev/null | head -1 || echo 'N/A')",
    "config_dir": "$CONFIG_DIR",
    "data_dir": "$DATA_DIR",
    "binary_path": "$AYO_BINARY"
  },
  "resources": {
    "memory_total_mb": $(sysctl -n hw.memsize 2>/dev/null | awk '{print int($1/1024/1024)}' || free -m 2>/dev/null | awk '/Mem:/{print $2}' || echo 0),
    "disk_available_gb": $(df -g "$HOME" 2>/dev/null | awk 'NR==2{print $4}' || df -BG "$HOME" 2>/dev/null | awk 'NR==2{print $4}' | tr -d 'G' || echo 0)
  }
}
EOF
else
    # Human-readable output
    divider "SYSTEM DIAGNOSTIC INFO"
    echo "  Generated: $(date)"
    echo "  Source: HOST"
    if $DEV_MODE; then
        echo "  Mode: DEV (using project-local paths)"
    fi

    section "Operating System"
    echo "  OS:           $(uname -s)"
    echo "  Version:      $(uname -r)"
    echo "  Architecture: $(uname -m)"
    echo "  Hostname:     $(hostname)"
    echo "  User:         $(whoami)"

    section "Sandbox Provider"
    echo "  Provider:     $SANDBOX_PROVIDER"
    if [[ "$SANDBOX_PROVIDER" == "apple-container" ]]; then
        echo "  Binary:       $(which container 2>/dev/null || echo 'not found')"
    elif [[ "$SANDBOX_PROVIDER" == "systemd-nspawn" ]]; then
        echo "  Binary:       $(which systemd-nspawn 2>/dev/null || echo 'not found')"
    else
        echo "  Status:       No sandbox provider available"
    fi

    section "Ayo Installation"
    if [[ -n "$AYO_BINARY" ]]; then
        echo "  Version:      $($AYO_BINARY --version 2>/dev/null | head -1 || echo 'unknown')"
        echo "  Binary:       $AYO_BINARY"
        echo "  Config Dir:   $CONFIG_DIR"
        echo "  Data Dir:     $DATA_DIR"
    else
        echo "  Installed:    No (ayo not found)"
    fi

    section "Resource Usage"
    if [[ "$(uname -s)" == "Darwin" ]]; then
        mem_total=$(sysctl -n hw.memsize | awk '{print int($1/1024/1024)}')
        echo "  Memory:       ${mem_total} MB total"
    else
        free -h 2>/dev/null | head -2 || echo "  Memory info unavailable"
    fi
    echo "  Disk (HOME):  $(df -h "$HOME" 2>/dev/null | awk 'NR==2{print $4 " available"}' || echo 'unavailable')"

    section "Environment Variables"
    echo "  AYO_CONFIG:   ${AYO_CONFIG:-<not set>}"
    echo "  AYO_DEBUG:    ${AYO_DEBUG:-<not set>}"
    echo "  XDG_CONFIG:   ${XDG_CONFIG_HOME:-<not set>}"
    echo "  XDG_DATA:     ${XDG_DATA_HOME:-<not set>}"

    echo ""
fi
