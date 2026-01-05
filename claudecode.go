package agentx

import (
	
	"context"
	"path/filepath"

)

// ClaudeCodeAgent implements Agent for Claude Code.
type ClaudeCodeAgent struct{}

// NewClaudeCodeAgent creates a new Claude Code agent.
func NewClaudeCodeAgent() *ClaudeCodeAgent {
	return &ClaudeCodeAgent{}
}

func (a *ClaudeCodeAgent) Type() AgentType {
	return AgentTypeClaudeCode
}

func (a *ClaudeCodeAgent) Name() string {
	return "Claude Code"
}

func (a *ClaudeCodeAgent) URL() string {
	return "https://github.com/anthropics/claude-code"
}

// Detect checks if Claude Code is the active agent.
//
// Detection methods:
//   - CLAUDECODE=1 (set by Claude Code)
//   - AGENT_ENV=claude
func (a *ClaudeCodeAgent) Detect(ctx context.Context, env Environment) (bool, error) {
	// Check CLAUDECODE env var (set by Claude Code itself)
	if env.GetEnv("CLAUDECODE") == "1" {
		return true, nil
	}

	// Check explicit AGENT_ENV
	if env.GetEnv("AGENT_ENV") == "claude" {
		return true, nil
	}

	return false, nil
}

// UserConfigPath returns the Claude Code user configuration directory (~/.claude).
func (a *ClaudeCodeAgent) UserConfigPath(env Environment) (string, error) {
	home, err := env.HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude"), nil
}

// ProjectConfigPath returns the Claude Code project configuration directory.
func (a *ClaudeCodeAgent) ProjectConfigPath() string {
	return ".claude"
}

// ContextFiles returns the context/instruction files Claude Code supports.
func (a *ClaudeCodeAgent) ContextFiles() []string {
	return []string{"CLAUDE.md", "AGENTS.md"}
}

// IsInstalled checks if Claude Code is installed on the system.
func (a *ClaudeCodeAgent) IsInstalled(ctx context.Context, env Environment) (bool, error) {
	// Check if claude CLI is in PATH
	if _, err := env.LookPath("claude"); err == nil {
		return true, nil
	}

	// Fallback: check if config directory exists
	configPath, err := a.UserConfigPath(env)
	if err != nil {
		return false, nil
	}
	if env.IsDir(configPath) {
		return true, nil
	}

	return false, nil
}

var _ Agent = (*ClaudeCodeAgent)(nil)
