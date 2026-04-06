package rules

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sageox/agentx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaudeCodeRulesManager_RulesDir(t *testing.T) {
	m := NewClaudeCodeRulesManager()
	assert.Equal(t, "/project/.claude/rules", m.RulesDir("/project"))
}

func TestClaudeCodeRulesManager_Install(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	m := NewClaudeCodeRulesManager()

	rules := []agentx.RuleFile{
		{Name: "ox.md", Content: []byte("# ox rules\nUse ox commands."), Version: "0.7.0"},
	}

	written, err := m.Install(ctx, dir, rules, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"ox.md"}, written)

	// verify file exists with stamp
	data, err := os.ReadFile(filepath.Join(dir, ".claude", "rules", "ox.md"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "agentx-hash:")
	assert.Contains(t, string(data), "ver: 0.7.0")
	assert.Contains(t, string(data), "# ox rules")
}

func TestClaudeCodeRulesManager_InstallIdempotent(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	m := NewClaudeCodeRulesManager()

	rules := []agentx.RuleFile{
		{Name: "ox.md", Content: []byte("content"), Version: "0.7.0"},
	}

	// first install
	written, err := m.Install(ctx, dir, rules, false)
	require.NoError(t, err)
	assert.Len(t, written, 1)

	// second install (overwrite=false) — skipped
	written, err = m.Install(ctx, dir, rules, false)
	require.NoError(t, err)
	assert.Len(t, written, 0)

	// third install (overwrite=true, same content) — skipped
	written, err = m.Install(ctx, dir, rules, true)
	require.NoError(t, err)
	assert.Len(t, written, 0)
}

func TestClaudeCodeRulesManager_InstallWithGlobs(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	m := NewClaudeCodeRulesManager()

	rules := []agentx.RuleFile{
		{
			Name:        "ox-team-go.md",
			Content:     []byte("Use slog for logging."),
			Version:     "0.7.0",
			Globs:       "**/*.go",
			Description: "Team Go conventions",
		},
	}

	written, err := m.Install(ctx, dir, rules, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"ox-team-go.md"}, written)

	data, err := os.ReadFile(filepath.Join(dir, ".claude", "rules", "ox-team-go.md"))
	require.NoError(t, err)
	content := string(data)
	assert.True(t, strings.HasPrefix(content, "---\n"), "should start with YAML frontmatter")
	assert.Contains(t, content, `globs: "**/*.go"`)
	assert.Contains(t, content, "description: Team Go conventions")
	assert.Contains(t, content, "Use slog for logging.")
}

func TestClaudeCodeRulesManager_Uninstall(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	m := NewClaudeCodeRulesManager()

	// install first
	rules := []agentx.RuleFile{
		{Name: "ox.md", Content: []byte("content"), Version: "0.7.0"},
		{Name: "ox-team.md", Content: []byte("team"), Version: "0.7.0"},
	}
	_, err := m.Install(ctx, dir, rules, false)
	require.NoError(t, err)

	// also create a user-managed file (no stamp)
	userFile := filepath.Join(m.RulesDir(dir), "my-rules.md")
	require.NoError(t, os.WriteFile(userFile, []byte("user rules"), 0o644))

	// uninstall ox-prefixed files
	removed, err := m.Uninstall(ctx, dir, "ox")
	require.NoError(t, err)
	assert.Len(t, removed, 2)
	assert.Contains(t, removed, "ox.md")
	assert.Contains(t, removed, "ox-team.md")

	// user file should still exist
	_, err = os.Stat(userFile)
	assert.NoError(t, err, "user-managed file should not be removed")
}

func TestClaudeCodeRulesManager_UninstallSkipsUnstamped(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	m := NewClaudeCodeRulesManager()

	// create an unstamped file with matching prefix
	rulesDir := m.RulesDir(dir)
	require.NoError(t, os.MkdirAll(rulesDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(rulesDir, "ox-custom.md"),
		[]byte("user-created ox rules without stamp"),
		0o644,
	))

	removed, err := m.Uninstall(ctx, dir, "ox")
	require.NoError(t, err)
	assert.Len(t, removed, 0, "should not remove unstamped files")
}

func TestClaudeCodeRulesManager_Validate(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	m := NewClaudeCodeRulesManager()

	rules := []agentx.RuleFile{
		{Name: "ox.md", Content: []byte("content v1"), Version: "0.7.0"},
		{Name: "ox-team.md", Content: []byte("team content"), Version: "0.7.0"},
	}

	// nothing installed — both missing
	missing, stale, err := m.Validate(ctx, dir, rules)
	require.NoError(t, err)
	assert.Equal(t, []string{"ox.md", "ox-team.md"}, missing)
	assert.Nil(t, stale)

	// install v1
	_, err = m.Install(ctx, dir, rules, false)
	require.NoError(t, err)

	// validate — nothing missing or stale
	missing, stale, err = m.Validate(ctx, dir, rules)
	require.NoError(t, err)
	assert.Nil(t, missing)
	assert.Nil(t, stale)

	// update expected content — should be stale
	rules[0].Content = []byte("content v2")
	missing, stale, err = m.Validate(ctx, dir, rules)
	require.NoError(t, err)
	assert.Nil(t, missing)
	assert.Equal(t, []string{"ox.md"}, stale)
}

func TestClaudeCodeRulesManager_DowngradeGuard(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	m := NewClaudeCodeRulesManager()

	// install with v0.8.0
	rules := []agentx.RuleFile{
		{Name: "ox.md", Content: []byte("new content"), Version: "0.8.0"},
	}
	_, err := m.Install(ctx, dir, rules, false)
	require.NoError(t, err)

	// try to overwrite with v0.7.0 — should be skipped
	oldRules := []agentx.RuleFile{
		{Name: "ox.md", Content: []byte("old content"), Version: "0.7.0"},
	}
	written, err := m.Install(ctx, dir, oldRules, true)
	require.NoError(t, err)
	assert.Len(t, written, 0, "downgrade should be prevented")

	// verify content is still v0.8.0
	data, err := os.ReadFile(filepath.Join(m.RulesDir(dir), "ox.md"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "new content")
}

// Ensure BaseRulesManager implements RulesManager at compile time.
var _ agentx.RulesManager = (*BaseRulesManager)(nil)
