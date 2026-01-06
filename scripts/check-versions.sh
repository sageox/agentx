#!/usr/bin/env bash
#
# check-versions.sh - Verify CHANGELOG.md has an entry for the specified version
#
# For Go modules, this ensures the CHANGELOG is updated before creating a release tag.
#
# Usage: ./scripts/check-versions.sh <version>
# Example: ./scripts/check-versions.sh v0.3.0

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CHANGELOG="$PROJECT_ROOT/CHANGELOG.md"

usage() {
    echo "Usage: $0 <version>"
    echo "Example: $0 v0.3.0"
    echo ""
    echo "Verifies that CHANGELOG.md contains an entry for the specified version."
    exit 1
}

if [[ $# -ne 1 ]]; then
    usage
fi

VERSION="$1"
# Remove 'v' prefix if present for CHANGELOG check
VERSION_NUM="${VERSION#v}"

echo "Checking version $VERSION_NUM..."

# Check if CHANGELOG.md exists
if [[ ! -f "$CHANGELOG" ]]; then
    echo "ERROR: CHANGELOG.md not found at $CHANGELOG"
    exit 1
fi

# Check for version entry in CHANGELOG
if ! grep -q "## \[$VERSION_NUM\]" "$CHANGELOG"; then
    echo "ERROR: CHANGELOG.md does not contain an entry for version $VERSION_NUM"
    echo ""
    echo "Please update CHANGELOG.md before releasing."
    echo "Run: ./scripts/version-bump.sh $VERSION_NUM"
    exit 1
fi

# Extract and display the changelog entry
echo "Found CHANGELOG entry for $VERSION_NUM:"
echo "---"
awk -v ver="$VERSION_NUM" '
    /^## \[/ {
        if (found) exit
        if ($0 ~ "\\[" ver "\\]") {
            found=1
            print
            next
        }
    }
    found { print }
' "$CHANGELOG"
echo "---"

# Check if the version has a date
if grep -q "## \[$VERSION_NUM\] - [0-9]\{4\}-[0-9]\{2\}-[0-9]\{2\}" "$CHANGELOG"; then
    echo "Version $VERSION_NUM has a release date."
else
    echo "WARNING: Version $VERSION_NUM does not have a release date."
    echo "Expected format: ## [$VERSION_NUM] - YYYY-MM-DD"
fi

# Verify there's actual content in the version section
CONTENT=$(awk -v ver="$VERSION_NUM" '
    /^## \[/ {
        if (found) exit
        if ($0 ~ "\\[" ver "\\]") found=1
        next
    }
    found && /^###/ { has_sections=1 }
    found && !/^#/ && !/^[[:space:]]*$/ { has_content=1 }
    END { if (has_sections && has_content) print "OK" }
' "$CHANGELOG")

if [[ "$CONTENT" == "OK" ]]; then
    echo "Version $VERSION_NUM has content."
    echo ""
    echo "All checks passed for version $VERSION_NUM"
    exit 0
else
    echo "WARNING: Version $VERSION_NUM section appears to be empty or missing content."
    echo ""
    echo "Checks passed with warnings for version $VERSION_NUM"
    exit 0
fi
