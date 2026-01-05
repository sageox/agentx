package agentx

import (
	"context"
	"os"
)

// Init detects the current coding agent and sets AGENT_ENV in the process
// environment. This enables downstream code and child processes to simply
// check AGENT_ENV instead of implementing their own detection logic.
//
// Call Init() early in your program (e.g., in main or init) to propagate
// agent context throughout your application.
//
// Returns the detected agent, or nil if no agent was detected.
// If AGENT_ENV is already set, Init() respects the existing value.
//
// Example:
//
//	func main() {
//	    agent := agentx.Init()
//	    if agent != nil {
//	        log.Printf("Running in %s", agent.Name())
//	    }
//	    // Child processes will inherit AGENT_ENV
//	}
func Init() Agent {
	// If AGENT_ENV is already set, respect it
	if os.Getenv("AGENT_ENV") != "" {
		return Detect()
	}

	agent := Detect()
	if agent != nil {
		os.Setenv("AGENT_ENV", string(agent.Type()))
	}
	return agent
}

// InitWithEnv is like Init but uses a custom environment for detection.
// Checks the provided environment for existing AGENT_ENV before detecting.
// Note: This still sets the real AGENT_ENV in os environment if detected.
func InitWithEnv(env Environment) Agent {
	// If AGENT_ENV is already set in the provided env, respect it
	if env.GetEnv("AGENT_ENV") != "" {
		return DetectWithEnv(env)
	}

	agent := DetectWithEnv(env)
	if agent != nil {
		os.Setenv("AGENT_ENV", string(agent.Type()))
	}
	return agent
}

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

// IsAgentContextWithEnv returns true if running inside any coding agent.
// Uses the provided environment for detection (useful for testing).
func IsAgentContextWithEnv(env Environment) bool {
	return DetectWithEnv(env) != nil
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

// RequireAgentWithEnv returns an error message if not running in an agent context.
// Uses the provided environment for detection (useful for testing).
func RequireAgentWithEnv(commandName string, env Environment) string {
	if IsAgentContextWithEnv(env) {
		return ""
	}
	return "'" + commandName + "' must be run from within a coding agent (Claude Code, Cursor, etc.).\n" +
		"If your agent doesn't set standard env vars, set AGENT_ENV=<agent-name> before running."
}
