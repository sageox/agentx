package rules

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sageox/agentx"
)

// ClaudeCodeRulesManager implements RulesManager for Claude Code.
// Rules are installed to .claude/rules/ in the project root as .md files.
// Files are stamped with a content hash and version for safe updates.
type ClaudeCodeRulesManager struct {
	StampPrefix string
	env         agentx.Environment
}

// NewClaudeCodeRulesManager creates a new Claude Code rules manager.
func NewClaudeCodeRulesManager() *ClaudeCodeRulesManager {
	return &ClaudeCodeRulesManager{
		StampPrefix: agentx.DefaultStampPrefix,
		env:         agentx.NewSystemEnvironment(),
	}
}

// NewClaudeCodeRulesManagerWithEnv creates a rules manager with a custom environment.
func NewClaudeCodeRulesManagerWithEnv(env agentx.Environment) *ClaudeCodeRulesManager {
	return &ClaudeCodeRulesManager{
		StampPrefix: agentx.DefaultStampPrefix,
		env:         env,
	}
}

// RulesDir returns the path to the Claude Code rules directory.
func (m *ClaudeCodeRulesManager) RulesDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".claude", "rules")
}

// Install writes rule files to .claude/rules/.
// Each file is stamped with a content hash and version on the first line.
// If the rule has Globs set, YAML frontmatter is prepended.
func (m *ClaudeCodeRulesManager) Install(_ context.Context, projectRoot string, rules []agentx.RuleFile, overwrite bool) ([]string, error) {
	rulesDir := m.RulesDir(projectRoot)
	if err := m.env.MkdirAll(rulesDir, 0o755); err != nil {
		return nil, fmt.Errorf("create rules directory: %w", err)
	}

	var written []string
	for _, rule := range rules {
		dstPath := filepath.Join(rulesDir, rule.Name)

		var existing []byte
		if data, err := m.env.ReadFile(dstPath); err == nil {
			existing = data
		}

		if !agentx.ShouldWriteRule(existing, rule, overwrite, m.StampPrefix) {
			continue
		}

		content := buildRuleContent(rule, m.StampPrefix)
		if err := m.env.WriteFile(dstPath, content, 0o644); err != nil {
			return written, fmt.Errorf("write rule file %s: %w", rule.Name, err)
		}
		written = append(written, rule.Name)
	}

	return written, nil
}

// Uninstall removes rule files matching the prefix from .claude/rules/.
func (m *ClaudeCodeRulesManager) Uninstall(_ context.Context, projectRoot string, prefix string) ([]string, error) {
	rulesDir := m.RulesDir(projectRoot)

	entries, err := m.env.ReadDir(rulesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read rules directory: %w", err)
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
		// only remove stamped files (ours)
		data, err := m.env.ReadFile(filepath.Join(rulesDir, name))
		if err != nil {
			continue
		}
		if agentx.ExtractCommandHash(data, m.StampPrefix) == "" {
			continue // not our file
		}
		if err := m.env.Remove(filepath.Join(rulesDir, name)); err != nil {
			return removed, fmt.Errorf("remove rule file %s: %w", name, err)
		}
		removed = append(removed, name)
	}

	// remove empty rules directory (best-effort)
	remaining, _ := m.env.ReadDir(rulesDir)
	if len(remaining) == 0 {
		m.env.Remove(rulesDir)
	}

	return removed, nil
}

// Validate checks which expected rule files are missing or stale.
func (m *ClaudeCodeRulesManager) Validate(_ context.Context, projectRoot string, rules []agentx.RuleFile) (missing []string, stale []string, err error) {
	rulesDir := m.RulesDir(projectRoot)

	for _, rule := range rules {
		dstPath := filepath.Join(rulesDir, rule.Name)
		existing, err := m.env.ReadFile(dstPath)
		if err != nil {
			if os.IsNotExist(err) {
				missing = append(missing, rule.Name)
				continue
			}
			return missing, stale, fmt.Errorf("read rule file %s: %w", rule.Name, err)
		}

		if agentx.IsRuleStale(existing, rule, m.StampPrefix) {
			stale = append(stale, rule.Name)
		}
	}

	return missing, stale, nil
}

// buildRuleContent constructs the full file content with optional YAML
// frontmatter (for glob-scoped rules) and a stamp header.
func buildRuleContent(rule agentx.RuleFile, stampPrefix string) []byte {
	var parts []string

	// add YAML frontmatter if globs or description are set
	if rule.Globs != "" || rule.Description != "" {
		parts = append(parts, "---")
		if rule.Description != "" {
			parts = append(parts, fmt.Sprintf("description: %s", rule.Description))
		}
		if rule.Globs != "" {
			parts = append(parts, fmt.Sprintf("globs: %q", rule.Globs))
		}
		parts = append(parts, "---")
		parts = append(parts, "")
	}

	frontmatter := ""
	if len(parts) > 0 {
		frontmatter = strings.Join(parts, "\n")
	}

	// stamp the rule content (without frontmatter) for hash calculation
	stamped := agentx.StampedContent(rule.Content, rule.Version, stampPrefix)

	if frontmatter != "" {
		return append([]byte(frontmatter), stamped...)
	}
	return stamped
}

// Ensure ClaudeCodeRulesManager implements RulesManager.
var _ agentx.RulesManager = (*ClaudeCodeRulesManager)(nil)
