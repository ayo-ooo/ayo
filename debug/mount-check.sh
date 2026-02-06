#!/bin/bash
# =============================================================================
# mount-check.sh - Verify mount permissions and file system access
# =============================================================================
#
# DESCRIPTION:
#   Checks mount configurations and verifies that the sandbox can access
#   host directories as expected. Use this when file access issues occur
#   or to verify mount setup.
#
# USAGE:
#   ./debug/mount-check.sh [--json] [--test-write]
#
# OPTIONS:
#   --json        Output in JSON format
#   --test-write  Attempt to write a test file (destructive test)
#
# OUTPUT:
#   - Configured mounts from mounts.json
#   - Active Docker mounts on container
#   - Read/write permission verification
#   - File count and accessibility
#
# EXAMPLES:
#   ./debug/mount-check.sh                    # Check mount status
#   ./debug/mount-check.sh --test-write       # Test write access
#   ./debug/mount-check.sh | pbcopy           # Copy to clipboard
#
# =============================================================================

set -euo pipefail

JSON_OUTPUT=false
TEST_WRITE=false

for arg in "$@"; do
    case $arg in
        --json) JSON_OUTPUT=true ;;
        --test-write) TEST_WRITE=true ;;
    esac
done

MOUNTS_FILE="${XDG_DATA_HOME:-$HOME/.local/share}/ayo/mounts.json"
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

if $JSON_OUTPUT; then
    mounts_config="null"
    if [[ -f "$MOUNTS_FILE" ]]; then
        mounts_config=$(cat "$MOUNTS_FILE")
    fi
    
    docker_mounts="[]"
    if [[ -n "$CONTAINER_ID" ]]; then
        docker_mounts=$(docker inspect --format '{{json .Mounts}}' "$CONTAINER_ID" 2>/dev/null || echo "[]")
    fi
    
    cat <<EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "source": "host",
  "script": "mount-check.sh",
  "mounts_file": "$MOUNTS_FILE",
  "mounts_file_exists": $([ -f "$MOUNTS_FILE" ] && echo true || echo false),
  "container_id": "${CONTAINER_ID:-null}",
  "configured_mounts": $mounts_config,
  "active_docker_mounts": $docker_mounts
}
EOF
else
    divider "MOUNT STATUS"
    echo "  Generated: $(date)"
    echo "  Source: HOST + CONTAINER"

    section "Configured Mounts (mounts.json)"
    if [[ -f "$MOUNTS_FILE" ]]; then
        echo "  File: $MOUNTS_FILE"
        echo ""
        cat "$MOUNTS_FILE" | jq -r 'to_entries[] | "  \(.key): \(.value.permissions // "rw") - \(.value.reason // "no reason")"' 2>/dev/null || cat "$MOUNTS_FILE" | sed 's/^/  /'
    else
        echo "  No mounts.json found at: $MOUNTS_FILE"
        echo "  Add mounts with: ayo mount add <path>"
    fi

    if [[ -z "$CONTAINER_ID" ]]; then
        section "Container Status"
        echo "  No running sandbox container found"
        echo "  Start one with: ayo @agent or ayo daemon start"
    else
        section "Active Docker Mounts"
        echo "  Container: $CONTAINER_ID"
        echo ""
        docker inspect --format '{{range .Mounts}}  {{.Type}}: {{.Source}} → {{.Destination}} ({{.Mode}}){{"\n"}}{{end}}' "$CONTAINER_ID" 2>/dev/null || echo "  Error inspecting mounts"

        section "Mount Accessibility Test"
        echo "  Testing read access from inside container..."
        echo ""
        
        # Get mount destinations
        MOUNT_DESTS=$(docker inspect --format '{{range .Mounts}}{{.Destination}}{{"\n"}}{{end}}' "$CONTAINER_ID" 2>/dev/null | grep -v '^$')
        
        for dest in $MOUNT_DESTS; do
            echo "  $dest:"
            if docker exec "$CONTAINER_ID" test -d "$dest" 2>/dev/null; then
                FILE_COUNT=$(docker exec "$CONTAINER_ID" find "$dest" -maxdepth 1 -type f 2>/dev/null | wc -l | tr -d ' ')
                DIR_COUNT=$(docker exec "$CONTAINER_ID" find "$dest" -maxdepth 1 -type d 2>/dev/null | wc -l | tr -d ' ')
                echo "    ✓ Accessible (${FILE_COUNT} files, ${DIR_COUNT} directories)"
                
                if $TEST_WRITE; then
                    TEST_FILE="$dest/.ayo-mount-test-$$"
                    if docker exec "$CONTAINER_ID" touch "$TEST_FILE" 2>/dev/null; then
                        echo "    ✓ Writable"
                        docker exec "$CONTAINER_ID" rm -f "$TEST_FILE" 2>/dev/null
                    else
                        echo "    ✗ Not writable (read-only mount)"
                    fi
                fi
            else
                echo "    ✗ Not accessible"
            fi
        done

        section "Working Copy Status"
        echo "  Checking session workspaces..."
        docker exec "$CONTAINER_ID" ls -la /workspace 2>/dev/null | head -10 | sed 's/^/  /' || echo "  No /workspace directory"
    fi

    echo ""
fi
