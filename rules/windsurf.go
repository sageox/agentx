package rules

import "github.com/sageox/agentx"

// WindsurfConfig is the rules configuration for Windsurf.
var WindsurfConfig = Config{
	Dir:       ".windsurf/rules",
	Extension: ".md",
	GlobField: "", // glob scoping not documented for Windsurf
}

// NewWindsurfRulesManager creates a rules manager for Windsurf (.windsurf/rules/*.md).
func NewWindsurfRulesManager() *BaseRulesManager {
	return NewBaseRulesManager(WindsurfConfig)
}

// NewWindsurfRulesManagerWithEnv creates a Windsurf rules manager with a custom environment.
func NewWindsurfRulesManagerWithEnv(env agentx.Environment) *BaseRulesManager {
	return NewBaseRulesManagerWithEnv(WindsurfConfig, env)
}
