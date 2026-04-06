package rules

import "github.com/sageox/agentx"

// ClineConfig is the rules configuration for Cline.
var ClineConfig = Config{
	Dir:       ".clinerules",
	Extension: ".md",
	GlobField: "paths",
}

// NewClineRulesManager creates a rules manager for Cline (.clinerules/*.md).
func NewClineRulesManager() *BaseRulesManager {
	return NewBaseRulesManager(ClineConfig)
}

// NewClineRulesManagerWithEnv creates a Cline rules manager with a custom environment.
func NewClineRulesManagerWithEnv(env agentx.Environment) *BaseRulesManager {
	return NewBaseRulesManagerWithEnv(ClineConfig, env)
}
