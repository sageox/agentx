package rules

import "github.com/sageox/agentx"

// DroidConfig is the rules configuration for Factory Droid.
var DroidConfig = Config{
	Dir:       ".factory/rules",
	Extension: ".md",
	GlobField: "", // glob scoping not yet supported by Droid
}

// NewDroidRulesManager creates a rules manager for Droid (.factory/rules/*.md).
func NewDroidRulesManager() *BaseRulesManager {
	return NewBaseRulesManager(DroidConfig)
}

// NewDroidRulesManagerWithEnv creates a Droid rules manager with a custom environment.
func NewDroidRulesManagerWithEnv(env agentx.Environment) *BaseRulesManager {
	return NewBaseRulesManagerWithEnv(DroidConfig, env)
}
