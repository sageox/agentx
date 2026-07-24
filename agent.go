// Package agentx provides coding agent detection for CLI tools.
//
// agentx helps tools understand which coding agent is invoking them,
// enabling agent-aware behavior and better integration with the AI
// coding ecosystem.
//
// Supported agents: Claude Code, Cursor, Windsurf, GitHub Copilot, Aider,
// Cody, Continue, Code Puppy, Kiro, OpenCode, Goose, Amp, Codex, Gemini CLI,
// Pi, and Droid.
//
// Detection uses the AGENT_ENV environment variable as the standard
// mechanism for explicit agent identification. Native environment
// variables and binary path heuristics are used as fallbacks.
//
// Usage:
//
//	import "github.com/sageox/agentx"
//
//	func main() {
//	    agent := agentx.Detect()
//	    if agent != nil {
//	        fmt.Printf("Running in %s\n", agent.Name())
//	    }
//	}
package agentx

import "context"

// AgentType identifies a coding agent.
type AgentType string

const (
	AgentTypeUnknown    AgentType = ""
	AgentTypeClaudeCode AgentType = "claude"
	AgentTypeCursor     AgentType = "cursor"
	AgentTypeWindsurf  AgentType = "windsurf"
	AgentTypeCopilot   AgentType = "copilot"
	AgentTypeAider     AgentType = "aider"
	AgentTypeCody      AgentType = "cody"
	AgentTypeContinue  AgentType = "continue"
	AgentTypeCodePuppy AgentType = "code-puppy"
	AgentTypeKiro      AgentType = "kiro"
	AgentTypeOpenCode  AgentType = "opencode"
	AgentTypeGoose     AgentType = "goose"
	AgentTypeAmp       AgentType = "amp"
	AgentTypeCodex     AgentType = "codex"
	AgentTypeGemini    AgentType = "gemini"
	AgentTypePi        AgentType = "pi"
	AgentTypeDroid     AgentType = "droid"
)

// SupportedAgents is the canonical list of coding agents that agentx supports.
var SupportedAgents = []AgentType{
	AgentTypeClaudeCode,
	AgentTypeCursor,
	AgentTypeWindsurf,
	AgentTypeCopilot,
	AgentTypeAider,
	AgentTypeCody,
	AgentTypeContinue,
	AgentTypeCodePuppy,
	AgentTypeKiro,
	AgentTypeOpenCode,
	AgentTypeGoose,
	AgentTypeAmp,
	AgentTypeCodex,
	AgentTypeGemini,
	AgentTypePi,
	AgentTypeDroid,
}

// Agent represents a coding agent with detection and configuration capabilities.
type Agent interface {
	// Type returns the agent type slug (e.g., "claude-code", "cursor").
	Type() AgentType

	// Name returns the human-readable agent name (e.g., "Claude Code", "Cursor").
	Name() string

	// URL returns the official project URL.
	URL() string

	// Detect checks if this agent is currently running.
	// Returns true if the agent's environment markers are present.
	Detect(ctx context.Context, env Environment) (bool, error)

	// IsInstalled checks if this agent is installed on the system.
	IsInstalled(ctx context.Context, env Environment) (bool, error)

	// UserConfigPath returns the user-level configuration directory.
	// Examples: ~/.claude, ~/.cursor, ~/.aider
	UserConfigPath(env Environment) (string, error)

	// ProjectConfigPath returns the project-level configuration directory.
	// Examples: .claude, .cursor, .aider
	// Returns empty string if the agent doesn't support project-level config.
	ProjectConfigPath() string

	// ContextFiles returns the context/instruction files this agent supports.
	// Examples: CLAUDE.md, .cursorrules, .windsurfrules
	ContextFiles() []string
}

// DetectionSource indicates how an agent was detected.
type DetectionSource string

const (
	// SourceAgentEnv means detection via AGENT_ENV environment variable.
	SourceAgentEnv DetectionSource = "AGENT_ENV"

	// SourceNative means detection via agent's native environment variable.
	SourceNative DetectionSource = "native"

	// SourceHeuristic means detection via binary path or other heuristics.
	SourceHeuristic DetectionSource = "heuristic"
)

// DetectionResult contains information about a detected agent.
type DetectionResult struct {
	// Agent is the detected agent.
	Agent Agent

	// Source indicates how the agent was detected.
	Source DetectionSource
}
