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

// TestBaseRulesManager_CursorConfig verifies the base manager works with Cursor's .mdc format.
func TestBaseRulesManager_CursorConfig(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	m := NewCursorRulesManager()

	assert.Equal(t, filepath.Join(dir, ".cursor", "rules"), m.RulesDir(dir))

	rules := []agentx.RuleFile{
		{
			Name:        "ox.mdc",
			Content:     []byte("Use ox commands for team context."),
			Version:     "0.7.0",
			Globs:       "**/*.ts,**/*.tsx",
			Description: "SageOx guidance for TypeScript",
		},
	}

	written, err := m.Install(ctx, dir, rules, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"ox.mdc"}, written)

	data, err := os.ReadFile(filepath.Join(dir, ".cursor", "rules", "ox.mdc"))
	require.NoError(t, err)
	content := string(data)
	assert.True(t, strings.HasPrefix(content, "---\n"))
	assert.Contains(t, content, `globs: "**/*.ts,**/*.tsx"`)
	assert.Contains(t, content, "description: SageOx guidance for TypeScript")
	assert.Contains(t, content, "agentx-hash:")
}

// TestBaseRulesManager_KiroConfig verifies Kiro's steering directory and field names.
func TestBaseRulesManager_KiroConfig(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	m := NewKiroRulesManager()

	assert.Equal(t, filepath.Join(dir, ".kiro", "steering"), m.RulesDir(dir))

	rules := []agentx.RuleFile{
		{
			Name:    "ox.md",
			Content: []byte("Use ox for team context."),
			Version: "0.7.0",
			Globs:   "**/*.go",
		},
	}

	written, err := m.Install(ctx, dir, rules, false)
	require.NoError(t, err)
	assert.Len(t, written, 1)

	data, err := os.ReadFile(filepath.Join(dir, ".kiro", "steering", "ox.md"))
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, `fileMatchPattern: "**/*.go"`)
}

// TestBaseRulesManager_CopilotConfig verifies Copilot's .github/instructions/ directory.
func TestBaseRulesManager_CopilotConfig(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	m := NewCopilotRulesManager()

	assert.Equal(t, filepath.Join(dir, ".github", "instructions"), m.RulesDir(dir))

	rules := []agentx.RuleFile{
		{
			Name:    "ox.md",
			Content: []byte("Use ox for team context."),
			Version: "0.7.0",
			Globs:   "**/*.py",
		},
	}

	written, err := m.Install(ctx, dir, rules, false)
	require.NoError(t, err)
	assert.Len(t, written, 1)

	data, err := os.ReadFile(filepath.Join(dir, ".github", "instructions", "ox.md"))
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, `applyTo: "**/*.py"`)
}

// TestBaseRulesManager_NoGlobField verifies agents without glob support skip the field.
func TestBaseRulesManager_NoGlobField(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	m := NewDroidRulesManager() // Droid has no glob support

	rules := []agentx.RuleFile{
		{
			Name:    "ox.md",
			Content: []byte("Droid rules."),
			Version: "0.7.0",
			Globs:   "**/*.go", // should be ignored since GlobField is empty
		},
	}

	written, err := m.Install(ctx, dir, rules, false)
	require.NoError(t, err)
	assert.Len(t, written, 1)

	data, err := os.ReadFile(filepath.Join(dir, ".factory", "rules", "ox.md"))
	require.NoError(t, err)
	content := string(data)
	// should NOT have frontmatter since glob field is empty and no description
	assert.True(t, strings.HasPrefix(content, "<!-- agentx-hash:"), "should start with stamp, not frontmatter")
}

// TestBaseRulesManager_UninstallRespectsExtension verifies uninstall only targets the right extension.
func TestBaseRulesManager_UninstallRespectsExtension(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	m := NewCursorRulesManager()

	// install .mdc file
	rules := []agentx.RuleFile{
		{Name: "ox.mdc", Content: []byte("content"), Version: "0.7.0"},
	}
	_, err := m.Install(ctx, dir, rules, false)
	require.NoError(t, err)

	// create a .md file with matching prefix (should not be removed by Cursor manager)
	rulesDir := m.RulesDir(dir)
	require.NoError(t, os.WriteFile(filepath.Join(rulesDir, "ox-notes.md"), []byte("notes"), 0o644))

	removed, err := m.Uninstall(ctx, dir, "ox")
	require.NoError(t, err)
	assert.Equal(t, []string{"ox.mdc"}, removed)

	// .md file should still exist
	_, err = os.Stat(filepath.Join(rulesDir, "ox-notes.md"))
	assert.NoError(t, err)
}
