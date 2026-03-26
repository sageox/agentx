package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sageox/agentx"
)

// ClaudeCodeCommandManager implements CommandManager for Claude Code.
// Commands are installed to .claude/commands/ in the project root.
type ClaudeCodeCommandManager struct {
	StampPrefix string
	env         agentx.Environment
}

// NewClaudeCodeCommandManager creates a new Claude Code command manager.
func NewClaudeCodeCommandManager() *ClaudeCodeCommandManager {
	return &ClaudeCodeCommandManager{
		StampPrefix: agentx.DefaultStampPrefix,
		env:         agentx.NewSystemEnvironment(),
	}
}

// NewClaudeCodeCommandManagerWithEnv creates a command manager with a custom environment.
func NewClaudeCodeCommandManagerWithEnv(env agentx.Environment) *ClaudeCodeCommandManager {
	return &ClaudeCodeCommandManager{
		StampPrefix: agentx.DefaultStampPrefix,
		env:         env,
	}
}

// CommandDir returns the path to the Claude Code command directory.
func (m *ClaudeCodeCommandManager) CommandDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".claude", "commands")
}

// Install writes command files to .claude/commands/.
// When overwrite is false, existing files are skipped entirely (safe for first-time init).
// When overwrite is true, existing files are replaced only if content differs AND the
// installed version is not newer than ours (prevents older binaries from downgrading).
// Each file is stamped with a content hash and version on the first line.
func (m *ClaudeCodeCommandManager) Install(_ context.Context, projectRoot string, commands []agentx.CommandFile, overwrite bool) ([]string, error) {
	cmdDir := m.CommandDir(projectRoot)
	if err := m.env.MkdirAll(cmdDir, 0o755); err != nil {
		return nil, fmt.Errorf("create command directory: %w", err)
	}

	var written []string
	for _, cmd := range commands {
		dstPath := filepath.Join(cmdDir, cmd.Name)

		var existing []byte
		if data, err := m.env.ReadFile(dstPath); err == nil {
			existing = data
		}

		if !agentx.ShouldWriteCommand(existing, cmd, overwrite, m.StampPrefix) {
			continue
		}

		stamped := agentx.StampedContent(cmd.Content, cmd.Version, m.StampPrefix)
		if err := m.env.WriteFile(dstPath, stamped, 0o644); err != nil {
			return written, fmt.Errorf("write command file %s: %w", cmd.Name, err)
		}
		written = append(written, cmd.Name)
	}

	return written, nil
}

// Uninstall removes command files matching the prefix from .claude/commands/.
func (m *ClaudeCodeCommandManager) Uninstall(_ context.Context, projectRoot string, prefix string) ([]string, error) {
	cmdDir := m.CommandDir(projectRoot)

	entries, err := m.env.ReadDir(cmdDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read command directory: %w", err)
	}

	var removed []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		if err := m.env.Remove(filepath.Join(cmdDir, name)); err != nil {
			return removed, fmt.Errorf("remove command file %s: %w", name, err)
		}
		removed = append(removed, name)
	}

	// remove empty command directory (best-effort)
	remaining, _ := m.env.ReadDir(cmdDir)
	if len(remaining) == 0 {
		m.env.Remove(cmdDir)
	}

	return removed, nil
}

// Validate checks which expected command files are missing or stale.
// A file is stale if its content hash differs from expected AND the installed
// version is not newer than ours (same downgrade guard as Install).
// Files without a hash stamp are not considered stale (user-managed).
func (m *ClaudeCodeCommandManager) Validate(_ context.Context, projectRoot string, commands []agentx.CommandFile) (missing []string, stale []string, err error) {
	cmdDir := m.CommandDir(projectRoot)

	for _, cmd := range commands {
		dstPath := filepath.Join(cmdDir, cmd.Name)
		existing, err := m.env.ReadFile(dstPath)
		if err != nil {
			if os.IsNotExist(err) {
				missing = append(missing, cmd.Name)
				continue
			}
			return missing, stale, fmt.Errorf("read command file %s: %w", cmd.Name, err)
		}

		if agentx.IsCommandStale(existing, cmd, m.StampPrefix) {
			stale = append(stale, cmd.Name)
		}
	}

	return missing, stale, nil
}

// Ensure ClaudeCodeCommandManager implements CommandManager.
var _ agentx.CommandManager = (*ClaudeCodeCommandManager)(nil)
