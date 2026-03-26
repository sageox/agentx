package commands

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/sageox/agentx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaudeCodeCommandManager_Install(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewClaudeCodeCommandManager()
	ctx := context.Background()

	commands := []agentx.CommandFile{
		{Name: "ox-test.md", Content: []byte("# Test\n"), Version: "0.12.0"},
		{Name: "ox-other.md", Content: []byte("# Other\n"), Version: "0.12.0"},
	}

	// first install: both files should be written
	written, err := mgr.Install(ctx, tmpDir, commands, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"ox-test.md", "ox-other.md"}, written)

	// verify content has hash+version stamp
	content, err := os.ReadFile(filepath.Join(tmpDir, ".claude", "commands", "ox-test.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "<!-- agentx-hash: ")
	assert.Contains(t, string(content), " ver: 0.12.0 -->")
	assert.Contains(t, string(content), "# Test")

	// second install (overwrite=false): nothing should be written (files exist)
	written, err = mgr.Install(ctx, tmpDir, commands, false)
	require.NoError(t, err)
	assert.Empty(t, written)

	// third install (overwrite=true, same content+version): nothing written (hash match)
	written, err = mgr.Install(ctx, tmpDir, commands, true)
	require.NoError(t, err)
	assert.Empty(t, written)

	// fourth install (overwrite=true, different content, newer version): should overwrite
	updatedCommands := []agentx.CommandFile{
		{Name: "ox-test.md", Content: []byte("# Test v2\n"), Version: "0.13.0"},
		{Name: "ox-other.md", Content: []byte("# Other v2\n"), Version: "0.13.0"},
	}
	written, err = mgr.Install(ctx, tmpDir, updatedCommands, true)
	require.NoError(t, err)
	assert.Equal(t, []string{"ox-test.md", "ox-other.md"}, written)

	// verify new content with new hash
	content, err = os.ReadFile(filepath.Join(tmpDir, ".claude", "commands", "ox-test.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), " ver: 0.13.0 -->")
	assert.Contains(t, string(content), "# Test v2")
}

func TestClaudeCodeCommandManager_Install_PreservesUserManaged(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewClaudeCodeCommandManager()
	ctx := context.Background()

	// create a user-managed file (no hash stamp)
	cmdDir := mgr.CommandDir(tmpDir)
	require.NoError(t, os.MkdirAll(cmdDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(cmdDir, "ox-test.md"),
		[]byte("# My custom content\n"),
		0o644,
	))

	commands := []agentx.CommandFile{
		{Name: "ox-test.md", Content: []byte("# Embedded\n"), Version: "0.12.0"},
	}

	// overwrite=true but file has no hash stamp: should NOT overwrite
	written, err := mgr.Install(ctx, tmpDir, commands, true)
	require.NoError(t, err)
	assert.Empty(t, written)

	// verify user content preserved
	content, err := os.ReadFile(filepath.Join(cmdDir, "ox-test.md"))
	require.NoError(t, err)
	assert.Equal(t, "# My custom content\n", string(content))
}

func TestClaudeCodeCommandManager_Install_SameContentNoRewrite(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewClaudeCodeCommandManager()
	ctx := context.Background()

	commands := []agentx.CommandFile{
		{Name: "ox-test.md", Content: []byte("# Test\n"), Version: "0.12.0"},
	}

	// initial install
	written, err := mgr.Install(ctx, tmpDir, commands, false)
	require.NoError(t, err)
	assert.Len(t, written, 1)

	// overwrite with identical content: should skip (hash match)
	written, err = mgr.Install(ctx, tmpDir, commands, true)
	require.NoError(t, err)
	assert.Empty(t, written)
}

func TestClaudeCodeCommandManager_Install_DowngradeGuard(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewClaudeCodeCommandManager()
	ctx := context.Background()

	// install with v0.13.0
	newerCommands := []agentx.CommandFile{
		{Name: "ox-test.md", Content: []byte("# Newer content\n"), Version: "0.13.0"},
	}
	written, err := mgr.Install(ctx, tmpDir, newerCommands, false)
	require.NoError(t, err)
	assert.Len(t, written, 1)

	// attempt overwrite with v0.12.0 (older binary, different content)
	olderCommands := []agentx.CommandFile{
		{Name: "ox-test.md", Content: []byte("# Older content\n"), Version: "0.12.0"},
	}
	written, err = mgr.Install(ctx, tmpDir, olderCommands, true)
	require.NoError(t, err)
	assert.Empty(t, written, "older version should not overwrite newer")

	// verify newer content preserved
	content, err := os.ReadFile(filepath.Join(mgr.CommandDir(tmpDir), "ox-test.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Newer content")
	assert.Contains(t, string(content), "ver: 0.13.0")
}

func TestClaudeCodeCommandManager_Validate(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewClaudeCodeCommandManager()
	ctx := context.Background()

	commands := []agentx.CommandFile{
		{Name: "ox-test.md", Content: []byte("# Test\n"), Version: "0.12.0"},
		{Name: "ox-other.md", Content: []byte("# Other\n"), Version: "0.12.0"},
	}

	// nothing installed: both missing
	missing, stale, err := mgr.Validate(ctx, tmpDir, commands)
	require.NoError(t, err)
	assert.Equal(t, []string{"ox-test.md", "ox-other.md"}, missing)
	assert.Empty(t, stale)

	// install
	_, err = mgr.Install(ctx, tmpDir, commands, false)
	require.NoError(t, err)

	// validate with same content: nothing missing/stale
	missing, stale, err = mgr.Validate(ctx, tmpDir, commands)
	require.NoError(t, err)
	assert.Empty(t, missing)
	assert.Empty(t, stale)

	// validate with different content from newer version: should detect stale
	updatedCommands := []agentx.CommandFile{
		{Name: "ox-test.md", Content: []byte("# Test v2\n"), Version: "0.13.0"},
		{Name: "ox-other.md", Content: []byte("# Other v2\n"), Version: "0.13.0"},
	}
	missing, stale, err = mgr.Validate(ctx, tmpDir, updatedCommands)
	require.NoError(t, err)
	assert.Empty(t, missing)
	assert.Equal(t, []string{"ox-test.md", "ox-other.md"}, stale)
}

func TestClaudeCodeCommandManager_Validate_DowngradeNotStale(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewClaudeCodeCommandManager()
	ctx := context.Background()

	// install with v0.13.0
	newerCommands := []agentx.CommandFile{
		{Name: "ox-test.md", Content: []byte("# Newer\n"), Version: "0.13.0"},
	}
	_, err := mgr.Install(ctx, tmpDir, newerCommands, false)
	require.NoError(t, err)

	// validate from perspective of v0.12.0 (older binary)
	olderCommands := []agentx.CommandFile{
		{Name: "ox-test.md", Content: []byte("# Older\n"), Version: "0.12.0"},
	}
	missing, stale, err := mgr.Validate(ctx, tmpDir, olderCommands)
	require.NoError(t, err)
	assert.Empty(t, missing)
	assert.Empty(t, stale, "files installed by newer version should not appear stale")
}

func TestClaudeCodeCommandManager_Validate_UserManagedNotStale(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewClaudeCodeCommandManager()
	ctx := context.Background()

	// create a user-managed file (no hash stamp)
	cmdDir := mgr.CommandDir(tmpDir)
	require.NoError(t, os.MkdirAll(cmdDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(cmdDir, "ox-test.md"),
		[]byte("# My custom content\n"),
		0o644,
	))

	commands := []agentx.CommandFile{
		{Name: "ox-test.md", Content: []byte("# Different embedded\n"), Version: "0.12.0"},
	}

	// user-managed file should not be reported as stale
	missing, stale, err := mgr.Validate(ctx, tmpDir, commands)
	require.NoError(t, err)
	assert.Empty(t, missing)
	assert.Empty(t, stale)
}

func TestClaudeCodeCommandManager_Uninstall(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewClaudeCodeCommandManager()
	ctx := context.Background()

	commands := []agentx.CommandFile{
		{Name: "ox-test.md", Content: []byte("# Test\n"), Version: "0.12.0"},
		{Name: "ox-other.md", Content: []byte("# Other\n"), Version: "0.12.0"},
		{Name: "custom.md", Content: []byte("# Custom\n"), Version: "0.12.0"},
	}

	// install all
	_, err := mgr.Install(ctx, tmpDir, commands, false)
	require.NoError(t, err)

	// uninstall with "ox" prefix: should only remove ox-* files
	removed, err := mgr.Uninstall(ctx, tmpDir, "ox")
	require.NoError(t, err)
	assert.Equal(t, []string{"ox-other.md", "ox-test.md"}, removed)

	// custom.md should still exist
	_, err = os.Stat(filepath.Join(mgr.CommandDir(tmpDir), "custom.md"))
	assert.NoError(t, err)
}

func TestClaudeCodeCommandManager_CommandDir(t *testing.T) {
	mgr := NewClaudeCodeCommandManager()
	assert.Equal(t, "/project/.claude/commands", mgr.CommandDir("/project"))
}

func TestUninstall_SkipsNonMDFiles(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewClaudeCodeCommandManager()
	ctx := context.Background()

	cmdDir := mgr.CommandDir(tmpDir)
	require.NoError(t, os.MkdirAll(cmdDir, 0o755))

	// create a .json file with the matching prefix
	jsonPath := filepath.Join(cmdDir, "ox-data.json")
	require.NoError(t, os.WriteFile(jsonPath, []byte(`{"key":"val"}`), 0o644))

	removed, err := mgr.Uninstall(ctx, tmpDir, "ox")
	require.NoError(t, err)
	assert.Empty(t, removed, "non-.md file should not be removed")

	_, err = os.Stat(jsonPath)
	assert.NoError(t, err, ".json file should still exist")
}

func TestUninstall_SkipsMismatchedPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewClaudeCodeCommandManager()
	ctx := context.Background()

	cmdDir := mgr.CommandDir(tmpDir)
	require.NoError(t, os.MkdirAll(cmdDir, 0o755))

	// create .md files without the target prefix
	require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "other-cmd.md"), []byte("# Other"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "custom.md"), []byte("# Custom"), 0o644))

	removed, err := mgr.Uninstall(ctx, tmpDir, "ox")
	require.NoError(t, err)
	assert.Empty(t, removed, "files without matching prefix should not be removed")

	_, err = os.Stat(filepath.Join(cmdDir, "other-cmd.md"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(cmdDir, "custom.md"))
	assert.NoError(t, err)
}

func TestUninstall_SkipsDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewClaudeCodeCommandManager()
	ctx := context.Background()

	cmdDir := mgr.CommandDir(tmpDir)
	require.NoError(t, os.MkdirAll(cmdDir, 0o755))

	// create a subdirectory with a matching name
	subDir := filepath.Join(cmdDir, "ox-subdir.md")
	require.NoError(t, os.MkdirAll(subDir, 0o755))

	removed, err := mgr.Uninstall(ctx, tmpDir, "ox")
	require.NoError(t, err)
	assert.Empty(t, removed, "directories should be skipped")

	info, err := os.Stat(subDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir(), "directory should still exist")
}

func TestUninstall_CleansEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewClaudeCodeCommandManager()
	ctx := context.Background()

	// install a single file, then uninstall it
	commands := []agentx.CommandFile{
		{Name: "ox-only.md", Content: []byte("# Only"), Version: "0.12.0"},
	}
	_, err := mgr.Install(ctx, tmpDir, commands, false)
	require.NoError(t, err)

	removed, err := mgr.Uninstall(ctx, tmpDir, "ox")
	require.NoError(t, err)
	assert.Equal(t, []string{"ox-only.md"}, removed)

	// command directory should be removed since it's now empty
	_, err = os.Stat(mgr.CommandDir(tmpDir))
	assert.True(t, os.IsNotExist(err), "empty command directory should be removed")
}

func TestUninstall_NonExistentDir(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewClaudeCodeCommandManager()
	ctx := context.Background()

	// command dir doesn't exist at all
	removed, err := mgr.Uninstall(ctx, tmpDir, "ox")
	require.NoError(t, err)
	assert.Nil(t, removed)
}

func TestInstall_WriteError(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewClaudeCodeCommandManager()
	ctx := context.Background()

	cmdDir := mgr.CommandDir(tmpDir)
	require.NoError(t, os.MkdirAll(cmdDir, 0o755))

	// make the command directory read-only so WriteFile fails
	require.NoError(t, os.Chmod(cmdDir, 0o555))
	t.Cleanup(func() {
		os.Chmod(cmdDir, 0o755) // restore so cleanup can remove it
	})

	commands := []agentx.CommandFile{
		{Name: "ox-test.md", Content: []byte("# Test"), Version: "0.12.0"},
	}
	_, err := mgr.Install(ctx, tmpDir, commands, false)
	assert.Error(t, err, "should fail when directory is read-only")
	assert.Contains(t, err.Error(), "write command file")
}

func TestValidate_ReadError(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewClaudeCodeCommandManager()
	ctx := context.Background()

	cmdDir := mgr.CommandDir(tmpDir)
	require.NoError(t, os.MkdirAll(cmdDir, 0o755))

	// create a file but make it unreadable
	filePath := filepath.Join(cmdDir, "ox-test.md")
	require.NoError(t, os.WriteFile(filePath, []byte("content"), 0o644))
	require.NoError(t, os.Chmod(filePath, 0o000))
	t.Cleanup(func() {
		os.Chmod(filePath, 0o644)
	})

	commands := []agentx.CommandFile{
		{Name: "ox-test.md", Content: []byte("# Test"), Version: "0.12.0"},
	}
	_, _, err := mgr.Validate(ctx, tmpDir, commands)
	assert.Error(t, err, "should fail when file is unreadable")
	assert.Contains(t, err.Error(), "read command file")
}
