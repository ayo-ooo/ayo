#!/bin/bash
# =============================================================================
# daemon-status.sh - Check ayo daemon status and connections
# =============================================================================
#
# DESCRIPTION:
#   Reports on the ayo daemon process, including running state, socket status,
#   managed sessions, triggers, and sandbox pool. Use this to diagnose issues
#   with background agent execution and triggers.
#
# USAGE:
#   ./debug/daemon-status.sh [--json] [--logs]
#
# OPTIONS:
#   --json    Output in JSON format
#   --logs    Include recent daemon logs
#
# OUTPUT:
#   - Daemon process status
#   - Unix socket health
#   - Active sessions
#   - Registered triggers
#   - Sandbox pool status
#   - Recent log entries (with --logs)
#
# EXAMPLES:
#   ./debug/daemon-status.sh                  # Quick status
#   ./debug/daemon-status.sh --logs           # Include log tail
#   ./debug/daemon-status.sh | pbcopy         # Copy to clipboard
#
# =============================================================================

set -euo pipefail

JSON_OUTPUT=false
SHOW_LOGS=false

for arg in "$@"; do
    case $arg in
        --json) JSON_OUTPUT=true ;;
        --logs) SHOW_LOGS=true ;;
    esac
done

SOCKET_PATH="${XDG_RUNTIME_DIR:-/tmp}/ayo/daemon.sock"
LOG_PATH="${XDG_STATE_HOME:-$HOME/.local/state}/ayo/daemon.log"
PID_PATH="${XDG_RUNTIME_DIR:-/tmp}/ayo/daemon.pid"

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

# Check if daemon is running
DAEMON_RUNNING=false
DAEMON_PID=""
if [[ -f "$PID_PATH" ]]; then
    DAEMON_PID=$(cat "$PID_PATH" 2>/dev/null || echo "")
    if [[ -n "$DAEMON_PID" ]] && kill -0 "$DAEMON_PID" 2>/dev/null; then
        DAEMON_RUNNING=true
    fi
fi

# Check socket
SOCKET_EXISTS=false
SOCKET_LISTENING=false
if [[ -S "$SOCKET_PATH" ]]; then
    SOCKET_EXISTS=true
    # Try to connect
    if command -v nc &>/dev/null; then
        if echo "" | nc -U "$SOCKET_PATH" &>/dev/null; then
            SOCKET_LISTENING=true
        fi
    fi
fi

if $JSON_OUTPUT; then
    # Get daemon status via ayo if available
    daemon_status="{}"
    if command -v ayo &>/dev/null && $DAEMON_RUNNING; then
        daemon_status=$(ayo status --json 2>/dev/null || echo '{}')
    fi
    
    cat <<EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "source": "host",
  "script": "daemon-status.sh",
  "daemon": {
    "running": $DAEMON_RUNNING,
    "pid": ${DAEMON_PID:-null},
    "socket_path": "$SOCKET_PATH",
    "socket_exists": $SOCKET_EXISTS,
    "log_path": "$LOG_PATH",
    "log_exists": $([ -f "$LOG_PATH" ] && echo true || echo false)
  },
  "ayo_status": $daemon_status
}
EOF
else
    divider "DAEMON STATUS"
    echo "  Generated: $(date)"
    echo "  Source: HOST"

    section "Daemon Process"
    echo "  Running:      $DAEMON_RUNNING"
    if [[ -n "$DAEMON_PID" ]]; then
        echo "  PID:          $DAEMON_PID"
        if $DAEMON_RUNNING; then
            echo "  Process:"
            ps -p "$DAEMON_PID" -o pid,ppid,%cpu,%mem,etime,command 2>/dev/null | tail -1 | sed 's/^/    /'
        fi
    else
        echo "  PID:          Not found (no pid file)"
    fi
    echo "  PID File:     $PID_PATH"

    section "Socket Status"
    echo "  Path:         $SOCKET_PATH"
    echo "  Exists:       $SOCKET_EXISTS"
    if $SOCKET_EXISTS; then
        ls -la "$SOCKET_PATH" 2>/dev/null | sed 's/^/  /'
    fi

    section "Ayo Status Command"
    if command -v ayo &>/dev/null; then
        if $DAEMON_RUNNING; then
            ayo status 2>&1 | sed 's/^/  /' || echo "  Error getting status"
        else
            echo "  Daemon not running - cannot get detailed status"
            echo "  Start with: ayo daemon start"
        fi
    else
        echo "  ayo command not found"
    fi

    if $DAEMON_RUNNING; then
        section "Active Sessions"
        ayo daemon sessions 2>&1 | head -20 | sed 's/^/  /' || echo "  Error listing sessions"

        section "Registered Triggers"
        ayo triggers list --json 2>/dev/null | jq -r '.[] | "  \(.id)  \(.type)  \(.agent)"' 2>/dev/null || echo "  Error listing triggers or none registered"
    fi

    if $SHOW_LOGS; then
        section "Recent Daemon Logs"
        echo "  Log file: $LOG_PATH"
        echo ""
        if [[ -f "$LOG_PATH" ]]; then
            tail -50 "$LOG_PATH" 2>/dev/null | sed 's/^/  /' || echo "  Cannot read log file"
        else
            echo "  Log file does not exist"
        fi
    fi

    section "Log File Info"
    if [[ -f "$LOG_PATH" ]]; then
        echo "  Path:     $LOG_PATH"
        echo "  Size:     $(du -h "$LOG_PATH" | cut -f1)"
        echo "  Lines:    $(wc -l < "$LOG_PATH")"
        echo "  Modified: $(stat -f %Sm "$LOG_PATH" 2>/dev/null || stat -c %y "$LOG_PATH" 2>/dev/null || echo 'unknown')"
    else
        echo "  Log file does not exist at: $LOG_PATH"
    fi

    echo ""
fi
