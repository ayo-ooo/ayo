#!/bin/bash
# =============================================================================
# system-info.sh - Collect comprehensive system information for debugging
# =============================================================================
#
# DESCRIPTION:
#   Gathers system-level diagnostic information about the host environment,
#   Docker installation, and ayo configuration. This is the first script to
#   run when diagnosing any ayo issue.
#
# USAGE:
#   ./debug/system-info.sh [--json]
#
# OPTIONS:
#   --json    Output in JSON format (for machine parsing)
#
# OUTPUT:
#   - OS and architecture
#   - Docker version and status
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
  "system": {
    "os": "$(uname -s)",
    "os_version": "$(uname -r)",
    "arch": "$(uname -m)",
    "hostname": "$(hostname)",
    "user": "$(whoami)"
  },
  "docker": {
    "installed": $(command -v docker &>/dev/null && echo true || echo false),
    "version": "$(docker --version 2>/dev/null | sed 's/Docker version //' | cut -d',' -f1 || echo 'N/A')",
    "running": $(docker info &>/dev/null && echo true || echo false),
    "containers_running": $(docker ps -q 2>/dev/null | wc -l | tr -d ' '),
    "containers_total": $(docker ps -aq 2>/dev/null | wc -l | tr -d ' ')
  },
  "ayo": {
    "version": "$(ayo --version 2>/dev/null | head -1 || echo 'N/A')",
    "config_dir": "${XDG_CONFIG_HOME:-$HOME/.config}/ayo",
    "data_dir": "${XDG_DATA_HOME:-$HOME/.local/share}/ayo",
    "binary_path": "$(which ayo 2>/dev/null || echo 'N/A')"
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

    section "Operating System"
    echo "  OS:           $(uname -s)"
    echo "  Version:      $(uname -r)"
    echo "  Architecture: $(uname -m)"
    echo "  Hostname:     $(hostname)"
    echo "  User:         $(whoami)"

    section "Docker Status"
    if command -v docker &>/dev/null; then
        echo "  Installed:    Yes"
        echo "  Version:      $(docker --version | sed 's/Docker version //' | cut -d',' -f1)"
        if docker info &>/dev/null; then
            echo "  Running:      Yes"
            echo "  Containers:   $(docker ps -q | wc -l | tr -d ' ') running / $(docker ps -aq | wc -l | tr -d ' ') total"
        else
            echo "  Running:      No (Docker daemon not responding)"
        fi
    else
        echo "  Installed:    No"
    fi

    section "Ayo Installation"
    if command -v ayo &>/dev/null; then
        echo "  Version:      $(ayo --version 2>/dev/null | head -1 || echo 'unknown')"
        echo "  Binary:       $(which ayo)"
        echo "  Config Dir:   ${XDG_CONFIG_HOME:-$HOME/.config}/ayo"
        echo "  Data Dir:     ${XDG_DATA_HOME:-$HOME/.local/share}/ayo"
    else
        echo "  Installed:    No (ayo not in PATH)"
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
    echo "  DOCKER_HOST:  ${DOCKER_HOST:-<not set>}"
    echo "  XDG_CONFIG:   ${XDG_CONFIG_HOME:-<not set>}"
    echo "  XDG_DATA:     ${XDG_DATA_HOME:-<not set>}"

    echo ""
fi
