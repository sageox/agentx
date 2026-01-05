package agentx

import "context"

// Detect returns the currently active coding agent, or nil if none detected.
//
// Detection priority:
//  1. AGENT_ENV environment variable (explicit override)
//  2. Native agent-specific environment variables (CLAUDECODE, CURSOR_AGENT, etc.)
//  3. Binary path heuristics (checking $_ for agent name)
//
// This is a convenience function using the default registry.
func Detect() Agent {
	ctx := context.Background()
	reg, ok := DefaultRegistry.(*registry)
	if !ok {
		return nil
	}
	d := &detector{registry: reg}
	agent, _ := d.detect(ctx)
	return agent
}

// DetectWithEnv returns the currently active coding agent using a custom environment.
// This is useful for testing or when environment abstraction is needed.
func DetectWithEnv(env Environment) Agent {
	ctx := context.Background()
	reg, ok := DefaultRegistry.(*registry)
	if !ok {
		return nil
	}
	d := &detector{registry: reg, env: env}
	agent, _ := d.detect(ctx)
	return agent
}

// DetectAll returns all detected agents (some may run simultaneously).
func DetectAll() []Agent {
	ctx := context.Background()
	reg, ok := DefaultRegistry.(*registry)
	if !ok {
		return nil
	}
	d := &detector{registry: reg}
	agents, _ := d.detectAll(ctx)
	return agents
}

// DetectByType checks if a specific agent type is active.
func DetectByType(agentType AgentType) bool {
	ctx := context.Background()
	reg, ok := DefaultRegistry.(*registry)
	if !ok {
		return false
	}
	d := &detector{registry: reg}
	detected, _ := d.detectByType(ctx, agentType)
	return detected
}

// IsAgentContext returns true if running inside any coding agent.
func IsAgentContext() bool {
	return Detect() != nil
}

// CurrentAgent is an alias for Detect() for semantic clarity.
func CurrentAgent() Agent {
	return Detect()
}

// RequireAgent returns an error message if not running in an agent context.
// Returns empty string if in agent context.
// The commandName parameter is used in the error message to help users.
func RequireAgent(commandName string) string {
	if IsAgentContext() {
		return ""
	}
	return "'" + commandName + "' must be run from within a coding agent (Claude Code, Cursor, etc.).\n" +
		"If your agent doesn't set standard env vars, set AGENT_ENV=<agent-name> before running."
}
