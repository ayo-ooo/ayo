#!/bin/bash
# Build verification tests for all examples
# Exits 0 on success, 1 on any failure

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EXAMPLES_DIR="$(dirname "$SCRIPT_DIR")/examples"
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

errors=0
total=0

echo "Testing all examples build successfully..."
echo ""

for example_dir in "$EXAMPLES_DIR"/*/; do
    example_name=$(basename "$example_dir")
    total=$((total + 1))
    
    printf "Building %-20s " "$example_name..."
    
    if go run ./cmd/ayo runthat "$example_dir" -o "$TMPDIR/$example_name" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗${NC}"
        echo "  Error building $example_name"
        go run ./cmd/ayo runthat "$example_dir" -o "$TMPDIR/$example_name" 2>&1 | sed 's/^/  /'
        errors=$((errors + 1))
    fi
done

echo ""
echo "Results: $((total - errors))/$total passed"

if [ $errors -gt 0 ]; then
    exit 1
fi

exit 0
