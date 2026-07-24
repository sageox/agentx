package agents

import (
	"context"
	"path/filepath"

	"github.com/sageox/agentx"
)

// CodexAgent implements Agent for Codex (https://github.com/openai/codex).
type CodexAgent struct{}

// NewCodexAgent creates a new Codex agent.
func NewCodexAgent() *CodexAgent {
	return &CodexAgent{}
}

func (a *CodexAgent) Type() agentx.AgentType {
	return agentx.AgentTypeCodex
}

func (a *CodexAgent) Name() string {
	return "Codex"
}

func (a *CodexAgent) URL() string {
	return "https://github.com/openai/codex"
}

// Detect checks if Codex is the active agent.
//
// Detection methods:
//   - CODEX_AGENT=1
//   - AGENT_ENV=codex
func (a *CodexAgent) Detect(ctx context.Context, env agentx.Environment) (bool, error) {
	// Check CODEX env var
	if env.GetEnv("CODEX_AGENT") == "1" {
		return true, nil
	}

	// Check AGENT_ENV
	if env.GetEnv("AGENT_ENV") == "codex" {
		return true, nil
	}

	return false, nil
}

// UserConfigPath returns the Codex user configuration directory (~/.codex).
func (a *CodexAgent) UserConfigPath(env agentx.Environment) (string, error) {
	home, err := env.HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".codex"), nil
}

// ProjectConfigPath returns empty as Codex does not use project-level configuration.
func (a *CodexAgent) ProjectConfigPath() string {
	return ""
}

// ContextFiles returns the context/instruction files Codex supports.
func (a *CodexAgent) ContextFiles() []string {
	return []string{"AGENTS.md"}
}

// IsInstalled checks if Codex is installed on the system.
func (a *CodexAgent) IsInstalled(ctx context.Context, env agentx.Environment) (bool, error) {
	// Check if codex is in PATH
	if _, err := env.LookPath("codex"); err == nil {
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

var _ agentx.Agent = (*CodexAgent)(nil)
