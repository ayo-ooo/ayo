#!/bin/bash
# =============================================================================
# sandbox-status.sh - Check sandbox container status and health
# =============================================================================
#
# DESCRIPTION:
#   Reports on the ayo sandbox container status, including running state,
#   resource usage, network configuration, and installed services (IRC, etc).
#   Use this to verify the sandbox is properly initialized and healthy.
#
# USAGE:
#   ./debug/sandbox-status.sh [--json] [--verbose]
#
# OPTIONS:
#   --json      Output in JSON format
#   --verbose   Include detailed container inspection
#
# OUTPUT:
#   - Container existence and running state
#   - Container resource usage (CPU, memory)
#   - Network configuration
#   - Mounted volumes
#   - Running processes inside container
#   - IRC server status
#
# EXAMPLES:
#   ./debug/sandbox-status.sh                 # Quick status check
#   ./debug/sandbox-status.sh --verbose       # Full details
#   ./debug/sandbox-status.sh | pbcopy        # Copy to clipboard
#
# =============================================================================

set -euo pipefail

JSON_OUTPUT=false
VERBOSE=false

for arg in "$@"; do
    case $arg in
        --json) JSON_OUTPUT=true ;;
        --verbose) VERBOSE=true ;;
    esac
done

# Find ayo sandbox containers
SANDBOX_CONTAINERS=$(docker ps -a --filter "name=ayo-sandbox" --format "{{.ID}}" 2>/dev/null || echo "")
RUNNING_CONTAINERS=$(docker ps --filter "name=ayo-sandbox" --format "{{.ID}}" 2>/dev/null || echo "")

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
    # Build JSON output
    containers_json="[]"
    if [[ -n "$SANDBOX_CONTAINERS" ]]; then
        containers_json=$(docker ps -a --filter "name=ayo-sandbox" --format '{"id":"{{.ID}}","name":"{{.Names}}","status":"{{.Status}}","image":"{{.Image}}","created":"{{.CreatedAt}}"}' | jq -s '.')
    fi
    
    cat <<EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "source": "host",
  "script": "sandbox-status.sh",
  "sandbox": {
    "containers_exist": $([ -n "$SANDBOX_CONTAINERS" ] && echo true || echo false),
    "containers_running": $([ -n "$RUNNING_CONTAINERS" ] && echo true || echo false),
    "container_count": $(echo "$SANDBOX_CONTAINERS" | grep -c . || echo 0),
    "running_count": $(echo "$RUNNING_CONTAINERS" | grep -c . || echo 0)
  },
  "containers": $containers_json
}
EOF
else
    divider "SANDBOX STATUS"
    echo "  Generated: $(date)"
    echo "  Source: HOST → CONTAINER"

    section "Container Overview"
    if [[ -z "$SANDBOX_CONTAINERS" ]]; then
        echo "  Status:       No ayo sandbox containers found"
        echo "  Action:       Run 'ayo sandbox list' or start an agent to create one"
    else
        echo "  Total:        $(echo "$SANDBOX_CONTAINERS" | wc -l | tr -d ' ') container(s)"
        echo "  Running:      $(echo "$RUNNING_CONTAINERS" | grep -c . || echo 0) container(s)"
        echo ""
        docker ps -a --filter "name=ayo-sandbox" --format "table {{.ID}}\t{{.Names}}\t{{.Status}}\t{{.Image}}" 2>/dev/null || echo "  Error listing containers"
    fi

    if [[ -n "$RUNNING_CONTAINERS" ]]; then
        for container_id in $RUNNING_CONTAINERS; do
            section "Container: $container_id"
            
            # Basic info
            echo "  Name:         $(docker inspect --format '{{.Name}}' "$container_id" | sed 's/^\///')"
            echo "  Image:        $(docker inspect --format '{{.Config.Image}}' "$container_id")"
            echo "  Created:      $(docker inspect --format '{{.Created}}' "$container_id" | cut -d'T' -f1)"
            echo "  Uptime:       $(docker inspect --format '{{.State.StartedAt}}' "$container_id" | cut -d'T' -f1)"
            
            # Resource usage
            echo ""
            echo "  Resource Usage:"
            docker stats --no-stream --format "    CPU: {{.CPUPerc}}  Memory: {{.MemUsage}}" "$container_id" 2>/dev/null || echo "    Stats unavailable"
            
            # Mounts
            echo ""
            echo "  Volumes/Mounts:"
            docker inspect --format '{{range .Mounts}}    {{.Type}}: {{.Source}} → {{.Destination}}{{"\n"}}{{end}}' "$container_id" 2>/dev/null || echo "    None"
            
            # Network
            echo ""
            echo "  Network:"
            docker inspect --format '{{range $k, $v := .NetworkSettings.Networks}}    {{$k}}: {{$v.IPAddress}}{{"\n"}}{{end}}' "$container_id" 2>/dev/null || echo "    None"
            
            if $VERBOSE; then
                # Processes inside container
                echo ""
                echo "  Processes (inside container):"
                docker exec "$container_id" ps aux 2>/dev/null | head -15 || echo "    Cannot execute ps in container"
                
                # IRC server check
                echo ""
                echo "  IRC Server (ngircd):"
                if docker exec "$container_id" pgrep ngircd &>/dev/null; then
                    echo "    Status: Running"
                    docker exec "$container_id" cat /etc/ngircd/ngircd.conf 2>/dev/null | grep -E "^(Name|Listen|Ports)" | sed 's/^/    /' || true
                else
                    echo "    Status: Not running"
                fi
                
                # Agent users
                echo ""
                echo "  Agent Users:"
                docker exec "$container_id" cat /etc/passwd 2>/dev/null | grep -E "^agent-" | cut -d: -f1,6 | sed 's/^/    /' || echo "    None found"
            fi
        done
    fi
    
    echo ""
fi
