#!/usr/bin/env bash
#
# version-bump.sh - Update CHANGELOG.md for a new release
#
# For Go modules, versioning is done via git tags. This script:
# 1. Updates CHANGELOG.md with the new version and date
# 2. Moves [Unreleased] changes to the new version section
#
# Usage: ./scripts/version-bump.sh <version>
# Example: ./scripts/version-bump.sh 0.3.0

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CHANGELOG="$PROJECT_ROOT/CHANGELOG.md"

usage() {
    echo "Usage: $0 <version>"
    echo "Example: $0 0.3.0"
    echo ""
    echo "Updates CHANGELOG.md with the new version."
    echo "Version should be in semver format (e.g., 0.3.0, 1.0.0-rc1)"
    exit 1
}

if [[ $# -ne 1 ]]; then
    usage
fi

VERSION="$1"

# Validate version format (basic semver check)
if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$ ]]; then
    echo "ERROR: Invalid version format: $VERSION"
    echo "Expected semver format: X.Y.Z or X.Y.Z-prerelease"
    exit 1
fi

# Check if CHANGELOG.md exists
if [[ ! -f "$CHANGELOG" ]]; then
    echo "ERROR: CHANGELOG.md not found at $CHANGELOG"
    exit 1
fi

# Check if version already exists in CHANGELOG
if grep -q "## \[$VERSION\]" "$CHANGELOG"; then
    echo "ERROR: Version $VERSION already exists in CHANGELOG.md"
    exit 1
fi

# Get today's date in YYYY-MM-DD format
TODAY=$(date +%Y-%m-%d)

echo "Updating CHANGELOG.md for version $VERSION ($TODAY)..."

# Check if there's an [Unreleased] section with content
if grep -q "## \[Unreleased\]" "$CHANGELOG"; then
    # Replace [Unreleased] with the new version
    sed -i.bak "s/## \[Unreleased\]/## [Unreleased]\n\n## [$VERSION] - $TODAY/" "$CHANGELOG"
    rm -f "$CHANGELOG.bak"
    echo "Converted [Unreleased] section to [$VERSION]"
else
    # No [Unreleased] section, check if we need to add version header
    # Find the first ## [ line and insert before it
    if grep -q "^## \[" "$CHANGELOG"; then
        # Get line number of first version entry
        LINE_NUM=$(grep -n "^## \[" "$CHANGELOG" | head -1 | cut -d: -f1)

        # Insert new version section before the first version
        sed -i.bak "${LINE_NUM}i\\
## [$VERSION] - $TODAY\\
\\
### Added\\
\\
### Changed\\
\\
### Fixed\\
\\
" "$CHANGELOG"
        rm -f "$CHANGELOG.bak"
        echo "Added new version section [$VERSION]"
        echo "NOTE: Please fill in the changes for this version"
    else
        echo "ERROR: Could not find version sections in CHANGELOG.md"
        exit 1
    fi
fi

echo ""
echo "CHANGELOG.md updated for version $VERSION"
echo ""
echo "Next steps:"
echo "  1. Review and edit CHANGELOG.md if needed"
echo "  2. Commit the changes: git add CHANGELOG.md && git commit -m 'chore: prepare release $VERSION'"
echo "  3. Create and push the tag: git tag v$VERSION && git push origin v$VERSION"
