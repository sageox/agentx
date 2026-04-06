package rules

import "github.com/sageox/agentx"

// ClaudeCodeConfig is the rules configuration for Claude Code.
var ClaudeCodeConfig = Config{
	Dir:       ".claude/rules",
	Extension: ".md",
	GlobField: "globs",
}

// NewClaudeCodeRulesManager creates a rules manager for Claude Code (.claude/rules/*.md).
func NewClaudeCodeRulesManager() *BaseRulesManager {
	return NewBaseRulesManager(ClaudeCodeConfig)
}

// NewClaudeCodeRulesManagerWithEnv creates a Claude Code rules manager with a custom environment.
func NewClaudeCodeRulesManagerWithEnv(env agentx.Environment) *BaseRulesManager {
	return NewBaseRulesManagerWithEnv(ClaudeCodeConfig, env)
}
