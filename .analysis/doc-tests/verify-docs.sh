#!/bin/bash
# Documentation Verification Script
# Validates internal links and structure of ayo documentation

set -e

DOCS_DIR="docs"
ERRORS=0

echo "=== Ayo Documentation Verification ==="
echo ""

# Check all expected files exist
echo "Checking file structure..."
EXPECTED_FILES=(
    "getting-started.md"
    "concepts.md"
    "tutorials/first-agent.md"
    "tutorials/squads.md"
    "tutorials/triggers.md"
    "tutorials/memory.md"
    "tutorials/plugins.md"
    "guides/agents.md"
    "guides/squads.md"
    "guides/triggers.md"
    "guides/tools.md"
    "guides/sandbox.md"
    "guides/security.md"
    "reference/cli.md"
    "reference/ayo-json.md"
    "reference/prompts.md"
    "reference/rpc.md"
    "reference/plugins.md"
    "advanced/architecture.md"
    "advanced/extending.md"
    "advanced/troubleshooting.md"
)

for file in "${EXPECTED_FILES[@]}"; do
    if [ -f "$DOCS_DIR/$file" ]; then
        echo "  ✓ $file"
    else
        echo "  ✗ $file MISSING"
        ((ERRORS++))
    fi
done
echo ""

# Validate internal links
echo "Checking internal links..."
LINK_ERRORS=0

# Find all relative links in markdown files
while IFS= read -r -d '' file; do
    dir=$(dirname "$file")
    
    # Extract relative links (../path/file.md or ./path/file.md)
    links=$(grep -oE '\]\([^)]+\.md\)' "$file" 2>/dev/null | sed 's/\](//;s/)$//' | grep -E '^\.\.' || true)
    
    for link in $links; do
        # Resolve the link relative to the file's directory
        target="$dir/$link"
        if [ ! -f "$target" ]; then
            echo "  ✗ Broken link in $file: $link"
            ((LINK_ERRORS++))
        fi
    done
done < <(find "$DOCS_DIR" -name "*.md" -print0)

if [ $LINK_ERRORS -eq 0 ]; then
    echo "  ✓ All internal links valid"
else
    ((ERRORS+=LINK_ERRORS))
fi
echo ""

# Check for common documentation issues
echo "Checking documentation quality..."

# Check for TODO markers
TODOS=$(grep -r "TODO" "$DOCS_DIR" --include="*.md" 2>/dev/null | wc -l | tr -d ' ')
if [ "$TODOS" -gt 0 ]; then
    echo "  ! Found $TODOS TODO markers"
fi

# Check for placeholder text
PLACEHOLDERS=$(grep -rE '\[placeholder\]|\[coming soon\]|\[TBD\]' "$DOCS_DIR" --include="*.md" 2>/dev/null | wc -l | tr -d ' ')
if [ "$PLACEHOLDERS" -gt 0 ]; then
    echo "  ! Found $PLACEHOLDERS placeholder markers"
fi

echo ""

# Summary
echo "=== Summary ==="
TOTAL_FILES=$(find "$DOCS_DIR" -name "*.md" | wc -l | tr -d ' ')
echo "Total documentation files: $TOTAL_FILES"
echo "Expected files present: ${#EXPECTED_FILES[@]}/${#EXPECTED_FILES[@]}"

if [ $ERRORS -eq 0 ]; then
    echo ""
    echo "✓ Documentation verification passed!"
    exit 0
else
    echo ""
    echo "✗ Documentation verification failed with $ERRORS error(s)"
    exit 1
fi
