#!/bin/bash
# =============================================================================
# sandbox-exec.sh - Execute commands inside the sandbox and capture output
# =============================================================================
#
# DESCRIPTION:
#   Executes a command inside the ayo sandbox container and captures output.
#   Useful for debugging what the agent would see/experience inside the sandbox.
#   Can run as a specific agent user or as root.
#
# USAGE:
#   ./debug/sandbox-exec.sh [--as <agent>] [--json] <command>
#
# OPTIONS:
#   --as <agent>   Run as specified agent user (e.g., --as ayo)
#   --json         Wrap output in JSON with metadata
#
# ARGUMENTS:
#   <command>      Command to execute inside sandbox (default: sh)
#
# OUTPUT:
#   - Command output from inside sandbox
#   - Exit code
#   - Execution metadata (with --json)
#
# EXAMPLES:
#   ./debug/sandbox-exec.sh pwd                      # Run pwd as root
#   ./debug/sandbox-exec.sh --as ayo ls -la          # Run as agent-ayo
#   ./debug/sandbox-exec.sh --as ayo whoami          # Check user context
#   ./debug/sandbox-exec.sh cat /etc/os-release      # Check OS info
#
# =============================================================================

set -euo pipefail

JSON_OUTPUT=false
AGENT=""
COMMAND=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --json)
            JSON_OUTPUT=true
            shift
            ;;
        --as)
            AGENT="$2"
            shift 2
            ;;
        *)
            COMMAND="$*"
            break
            ;;
    esac
done

# Default command
if [[ -z "$COMMAND" ]]; then
    COMMAND="sh"
fi

# Find running sandbox container
CONTAINER_ID=$(docker ps --filter "name=ayo-sandbox" --format "{{.ID}}" | head -1)

if [[ -z "$CONTAINER_ID" ]]; then
    if $JSON_OUTPUT; then
        cat <<EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "source": "host",
  "script": "sandbox-exec.sh",
  "success": false,
  "error": "No running ayo sandbox container found"
}
EOF
    else
        echo "ERROR: No running ayo sandbox container found" >&2
        echo "Start one with: ayo @agent or ayo sandbox service start" >&2
    fi
    exit 1
fi

# Build docker exec command
EXEC_OPTS=()
if [[ -n "$AGENT" ]]; then
    # Run as agent user (agent-<name>)
    EXEC_OPTS+=("-u" "agent-$AGENT")
    EXEC_OPTS+=("-w" "/home/agent-$AGENT")
fi

if $JSON_OUTPUT; then
    # Capture output and wrap in JSON
    START_TIME=$(date +%s.%N)
    OUTPUT=$(docker exec "${EXEC_OPTS[@]}" "$CONTAINER_ID" sh -c "$COMMAND" 2>&1) || EXIT_CODE=$?
    EXIT_CODE=${EXIT_CODE:-0}
    END_TIME=$(date +%s.%N)
    DURATION=$(echo "$END_TIME - $START_TIME" | bc)
    
    # Escape output for JSON
    ESCAPED_OUTPUT=$(echo "$OUTPUT" | jq -Rs .)
    
    cat <<EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "source": "container",
  "script": "sandbox-exec.sh",
  "container_id": "$CONTAINER_ID",
  "command": $(echo "$COMMAND" | jq -Rs .),
  "user": "${AGENT:-root}",
  "exit_code": $EXIT_CODE,
  "duration_seconds": $DURATION,
  "output": $ESCAPED_OUTPUT
}
EOF
else
    echo "═══════════════════════════════════════════════════════════════════════════════"
    echo "  SANDBOX EXEC"
    echo "═══════════════════════════════════════════════════════════════════════════════"
    echo "  Container: $CONTAINER_ID"
    echo "  User:      ${AGENT:-root}"
    echo "  Command:   $COMMAND"
    echo "───────────────────────────────────────────────────────────────────────────────"
    echo ""
    
    docker exec "${EXEC_OPTS[@]}" "$CONTAINER_ID" sh -c "$COMMAND"
    EXIT_CODE=$?
    
    echo ""
    echo "───────────────────────────────────────────────────────────────────────────────"
    echo "  Exit Code: $EXIT_CODE"
    echo "═══════════════════════════════════════════════════════════════════════════════"
fi
