package rules

import "github.com/sageox/agentx"

// CopilotConfig is the rules configuration for GitHub Copilot.
var CopilotConfig = Config{
	Dir:       ".github/instructions",
	Extension: ".md",
	GlobField: "applyTo",
}

// NewCopilotRulesManager creates a rules manager for Copilot (.github/instructions/*.md).
func NewCopilotRulesManager() *BaseRulesManager {
	return NewBaseRulesManager(CopilotConfig)
}

// NewCopilotRulesManagerWithEnv creates a Copilot rules manager with a custom environment.
func NewCopilotRulesManagerWithEnv(env agentx.Environment) *BaseRulesManager {
	return NewBaseRulesManagerWithEnv(CopilotConfig, env)
}
