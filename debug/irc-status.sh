#!/bin/bash
# =============================================================================
# irc-status.sh - Check IRC server status and agent connections
# =============================================================================
#
# DESCRIPTION:
#   Checks the ngircd IRC server running inside the sandbox container.
#   Shows connected clients, channels, and message activity. Use this to
#   debug inter-agent communication issues.
#
# USAGE:
#   ./debug/irc-status.sh [--json] [--messages <n>]
#
# OPTIONS:
#   --json          Output in JSON format
#   --messages <n>  Show last n IRC log messages (default: 20)
#
# OUTPUT:
#   - IRC server running status
#   - Connected clients (agents)
#   - Active channels
#   - Recent message activity
#
# EXAMPLES:
#   ./debug/irc-status.sh                     # Quick status
#   ./debug/irc-status.sh --messages 50       # More message history
#   ./debug/irc-status.sh | pbcopy            # Copy to clipboard
#
# =============================================================================

set -euo pipefail

JSON_OUTPUT=false
MESSAGE_COUNT=20

while [[ $# -gt 0 ]]; do
    case $1 in
        --json)
            JSON_OUTPUT=true
            shift
            ;;
        --messages)
            MESSAGE_COUNT="$2"
            shift 2
            ;;
        *)
            shift
            ;;
    esac
done

# Find running sandbox container
CONTAINER_ID=$(docker ps --filter "name=ayo-sandbox" --format "{{.ID}}" | head -1)

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

if [[ -z "$CONTAINER_ID" ]]; then
    if $JSON_OUTPUT; then
        cat <<EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "source": "host",
  "script": "irc-status.sh",
  "success": false,
  "error": "No running ayo sandbox container found"
}
EOF
    else
        echo "ERROR: No running ayo sandbox container found" >&2
    fi
    exit 1
fi

# Check if ngircd is running
IRC_RUNNING=false
if docker exec "$CONTAINER_ID" pgrep ngircd &>/dev/null; then
    IRC_RUNNING=true
fi

if $JSON_OUTPUT; then
    # Get IRC info
    clients="[]"
    if $IRC_RUNNING; then
        # Try to get connected clients via ngircd stats if available
        clients=$(docker exec "$CONTAINER_ID" cat /var/log/ngircd.log 2>/dev/null | grep -E "(Client|connection)" | tail -10 | jq -Rs 'split("\n") | map(select(length > 0))') || clients="[]"
    fi
    
    cat <<EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "source": "container",
  "script": "irc-status.sh",
  "container_id": "$CONTAINER_ID",
  "irc": {
    "running": $IRC_RUNNING,
    "server": "ngircd"
  }
}
EOF
else
    divider "IRC SERVER STATUS"
    echo "  Generated: $(date)"
    echo "  Source: CONTAINER ($CONTAINER_ID)"

    section "Server Status"
    if $IRC_RUNNING; then
        echo "  ngircd:     Running"
        docker exec "$CONTAINER_ID" ps aux 2>/dev/null | grep ngircd | grep -v grep | sed 's/^/  /' || true
    else
        echo "  ngircd:     Not running"
        echo ""
        echo "  Start IRC server with: docker exec $CONTAINER_ID ngircd"
    fi

    section "Configuration"
    if docker exec "$CONTAINER_ID" test -f /etc/ngircd/ngircd.conf 2>/dev/null; then
        echo "  Config file exists: /etc/ngircd/ngircd.conf"
        echo ""
        docker exec "$CONTAINER_ID" grep -E "^(Name|Listen|Ports|MaxConnections)" /etc/ngircd/ngircd.conf 2>/dev/null | sed 's/^/  /' || true
    else
        echo "  Config file not found"
    fi

    section "IRC Log (last $MESSAGE_COUNT lines)"
    if docker exec "$CONTAINER_ID" test -f /var/log/ngircd.log 2>/dev/null; then
        docker exec "$CONTAINER_ID" tail -"$MESSAGE_COUNT" /var/log/ngircd.log 2>/dev/null | sed 's/^/  /' || echo "  Cannot read log"
    else
        echo "  Log file not found at /var/log/ngircd.log"
        echo "  Checking alternative locations..."
        docker exec "$CONTAINER_ID" find /var/log -name "*ngircd*" -o -name "*irc*" 2>/dev/null | sed 's/^/  /' || echo "  No IRC logs found"
    fi

    section "Network Ports"
    echo "  Listening ports inside container:"
    docker exec "$CONTAINER_ID" netstat -tlnp 2>/dev/null | grep -E "(6667|6697)" | sed 's/^/  /' || docker exec "$CONTAINER_ID" ss -tlnp 2>/dev/null | grep -E "(6667|6697)" | sed 's/^/  /' || echo "  Cannot check ports"

    echo ""
fi
