package rules

import "github.com/sageox/agentx"

// KiroConfig is the rules configuration for Kiro.
// Kiro calls rules "steering" and uses .kiro/steering/ directory.
var KiroConfig = Config{
	Dir:       ".kiro/steering",
	Extension: ".md",
	GlobField: "fileMatchPattern",
}

// NewKiroRulesManager creates a rules manager for Kiro (.kiro/steering/*.md).
func NewKiroRulesManager() *BaseRulesManager {
	return NewBaseRulesManager(KiroConfig)
}

// NewKiroRulesManagerWithEnv creates a Kiro rules manager with a custom environment.
func NewKiroRulesManagerWithEnv(env agentx.Environment) *BaseRulesManager {
	return NewBaseRulesManagerWithEnv(KiroConfig, env)
}
