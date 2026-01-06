# Release agentx Go Module

Release a new version of the agentx Go module library.

## Process

### 1. Pre-flight Checks

Verify the repository is ready for release.

```bash
# Check current branch is main
git branch --show-current

# Check for uncommitted changes
git status --porcelain

# Ensure we're up to date with remote
git fetch origin
git status -uno
```

**Requirements:**
- Must be on `main` branch
- Working directory must be clean (no uncommitted changes)
- Local branch must be up to date with remote

### 2. Run Tests and Linting

Ensure all quality checks pass.

```bash
# Run full check suite
make check
```

**All tests must pass before proceeding.**

### 3. Determine Version Number

Check the current version and recent changes.

```bash
# List existing tags
git tag -l 'v*' | sort -V | tail -5

# View recent commits since last tag
git log $(git describe --tags --abbrev=0 2>/dev/null || echo "HEAD~20")..HEAD --oneline
```

**Version Guidelines (Semantic Versioning):**
- **MAJOR** (1.0.0): Breaking API changes
- **MINOR** (0.X.0): New features, backward compatible
- **PATCH** (0.0.X): Bug fixes, backward compatible

### 4. Review and Update CHANGELOG.md

Analyze changes since the last release.

```bash
# View detailed changes
git log $(git describe --tags --abbrev=0 2>/dev/null || echo "HEAD~20")..HEAD --pretty=format:"%h %s"

# View files changed
git diff $(git describe --tags --abbrev=0 2>/dev/null || echo "HEAD~20")..HEAD --stat
```

**Update CHANGELOG.md with:**
- New version header with today's date
- Categorize changes under: Added, Changed, Deprecated, Removed, Fixed, Security
- Keep descriptions concise but informative

Example entry:
```markdown
## [0.3.0] - 2026-02-15

### Added
- New agent support for XYZ coding assistant

### Changed
- Improved detection accuracy for environment variables

### Fixed
- Race condition in hook manager initialization
```

### 5. Commit CHANGELOG Updates

```bash
git add CHANGELOG.md
git commit -m "chore: prepare release vX.Y.Z"
```

### 6. Create and Push Git Tag

For Go modules, the tag IS the version.

```bash
# Create annotated tag
git tag -a vX.Y.Z -m "Release vX.Y.Z"

# Push tag to trigger release workflow
git push origin vX.Y.Z
```

### 7. Verify Release

After pushing the tag:

1. **Check GitHub Actions**: The release workflow should run automatically
   - Validates tests pass
   - Runs linting
   - Creates GitHub Release with notes from CHANGELOG
   - Triggers pkg.go.dev indexing

2. **Verify GitHub Release**: https://github.com/sageox/agentx/releases

3. **Verify pkg.go.dev**: https://pkg.go.dev/github.com/sageox/agentx@vX.Y.Z
   (May take a few minutes to index)

### 8. Post-Release

Update CHANGELOG.md with [Unreleased] section for future changes:

```markdown
## [Unreleased]

### Added

### Changed

### Fixed
```

---

## Quick Reference

```bash
# Full release sequence (replace X.Y.Z with actual version)
make check
git log $(git describe --tags --abbrev=0)..HEAD --oneline
# Edit CHANGELOG.md
git add CHANGELOG.md
git commit -m "chore: prepare release vX.Y.Z"
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin main
git push origin vX.Y.Z
```

## Troubleshooting

### Release workflow failed
- Check GitHub Actions logs for specific error
- Most common: tests failing, lint errors, missing CHANGELOG entry

### pkg.go.dev not updating
- Manually trigger: `curl https://proxy.golang.org/github.com/sageox/agentx/@v/vX.Y.Z.info`
- Check https://sum.golang.org for checksum database status

### Need to fix a release
- For minor fixes: Create vX.Y.Z+1 patch release
- For critical issues: Delete the tag and release, fix, re-release
  ```bash
  git tag -d vX.Y.Z
  git push origin :refs/tags/vX.Y.Z
  # Delete GitHub release manually
  # Fix issues, then re-tag
  ```

## Pre-release Versions

For alpha/beta/rc releases:

```bash
git tag -a v1.0.0-alpha.1 -m "Release v1.0.0-alpha.1"
git tag -a v1.0.0-beta.1 -m "Release v1.0.0-beta.1"
git tag -a v1.0.0-rc.1 -m "Release v1.0.0-rc.1"
```

These are marked as pre-release in GitHub and won't be the "latest" on pkg.go.dev.
