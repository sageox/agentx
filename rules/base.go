package rules

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sageox/agentx"
)

// Config holds per-agent rules directory configuration.
type Config struct {
	// Dir is the rules directory relative to project root (e.g., ".claude/rules").
	Dir string

	// Extension is the file extension including dot (e.g., ".md", ".mdc").
	Extension string

	// GlobField is the YAML frontmatter field name for glob patterns.
	// Examples: "globs" (Claude, Cursor), "paths" (Cline), "applyTo" (Copilot),
	// "fileMatchPattern" (Kiro). Empty means glob scoping is not supported.
	GlobField string

	// AlwaysApplyField is an optional YAML field for "always load this rule"
	// behavior. Example: "alwaysApply" (Cursor). Empty if not supported.
	AlwaysApplyField string
}

// BaseRulesManager implements RulesManager with configurable directory paths,
// file extensions, and frontmatter field names. It handles stamped content,
// version guards, and safe uninstall for any agent that supports modular rules.
type BaseRulesManager struct {
	config      Config
	stampPrefix string
	env         agentx.Environment
}

// NewBaseRulesManager creates a rules manager with the given agent-specific config.
func NewBaseRulesManager(cfg Config) *BaseRulesManager {
	return &BaseRulesManager{
		config:      cfg,
		stampPrefix: agentx.DefaultStampPrefix,
		env:         agentx.NewSystemEnvironment(),
	}
}

// NewBaseRulesManagerWithEnv creates a rules manager with a custom environment.
func NewBaseRulesManagerWithEnv(cfg Config, env agentx.Environment) *BaseRulesManager {
	return &BaseRulesManager{
		config:      cfg,
		stampPrefix: agentx.DefaultStampPrefix,
		env:         env,
	}
}

// RulesDir returns the absolute path to the rules directory for a project.
func (m *BaseRulesManager) RulesDir(projectRoot string) string {
	return filepath.Join(projectRoot, m.config.Dir)
}

// Install writes rule files to the agent's rules directory.
// Each file is stamped with a content hash and version on the first line.
// If the rule has Globs set and the agent supports glob scoping,
// YAML frontmatter is prepended with the agent-specific field name.
func (m *BaseRulesManager) Install(_ context.Context, projectRoot string, rules []agentx.RuleFile, overwrite bool) ([]string, error) {
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

		if !agentx.ShouldWriteRule(existing, rule, overwrite, m.stampPrefix) {
			continue
		}

		content := m.buildContent(rule)
		if err := m.env.WriteFile(dstPath, content, 0o644); err != nil {
			return written, fmt.Errorf("write rule file %s: %w", rule.Name, err)
		}
		written = append(written, rule.Name)
	}

	return written, nil
}

// Uninstall removes rule files matching the prefix from the rules directory.
// Only removes files with a recognized stamp (preserves user-created files).
func (m *BaseRulesManager) Uninstall(_ context.Context, projectRoot string, prefix string) ([]string, error) {
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
		if !strings.HasSuffix(name, m.config.Extension) {
			continue
		}
		// only remove stamped files (ours)
		data, err := m.env.ReadFile(filepath.Join(rulesDir, name))
		if err != nil {
			continue
		}
		if agentx.ExtractCommandHash(data, m.stampPrefix) == "" {
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
func (m *BaseRulesManager) Validate(_ context.Context, projectRoot string, rules []agentx.RuleFile) (missing []string, stale []string, err error) {
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

		if agentx.IsRuleStale(existing, rule, m.stampPrefix) {
			stale = append(stale, rule.Name)
		}
	}

	return missing, stale, nil
}

// buildContent constructs the full file content with optional YAML
// frontmatter (for glob-scoped rules) and a stamp header.
func (m *BaseRulesManager) buildContent(rule agentx.RuleFile) []byte {
	var parts []string

	// add YAML frontmatter if description or globs are set
	needsFrontmatter := rule.Description != "" || (rule.Globs != "" && m.config.GlobField != "")
	if needsFrontmatter {
		parts = append(parts, "---")
		if rule.Description != "" {
			parts = append(parts, fmt.Sprintf("description: %s", rule.Description))
		}
		if rule.Globs != "" && m.config.GlobField != "" {
			parts = append(parts, fmt.Sprintf("%s: %q", m.config.GlobField, rule.Globs))
		}
		parts = append(parts, "---")
		parts = append(parts, "")
	}

	frontmatter := ""
	if len(parts) > 0 {
		frontmatter = strings.Join(parts, "\n")
	}

	// stamp the rule content (without frontmatter) for hash calculation
	stamped := agentx.StampedContent(rule.Content, rule.Version, m.stampPrefix)

	if frontmatter != "" {
		return append([]byte(frontmatter), stamped...)
	}
	return stamped
}

// Ensure BaseRulesManager implements RulesManager.
var _ agentx.RulesManager = (*BaseRulesManager)(nil)
