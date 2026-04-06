package rules

import "github.com/sageox/agentx"

// CursorConfig is the rules configuration for Cursor.
var CursorConfig = Config{
	Dir:              ".cursor/rules",
	Extension:        ".mdc",
	GlobField:        "globs",
	AlwaysApplyField: "alwaysApply",
}

// NewCursorRulesManager creates a rules manager for Cursor (.cursor/rules/*.mdc).
func NewCursorRulesManager() *BaseRulesManager {
	return NewBaseRulesManager(CursorConfig)
}

// NewCursorRulesManagerWithEnv creates a Cursor rules manager with a custom environment.
func NewCursorRulesManagerWithEnv(env agentx.Environment) *BaseRulesManager {
	return NewBaseRulesManagerWithEnv(CursorConfig, env)
}
