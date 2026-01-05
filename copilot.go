package agentx

import (
	
	"context"
	"path/filepath"

)

// CopilotAgent implements Agent for GitHub Copilot.
type CopilotAgent struct{}

// NewCopilotAgent creates a new GitHub Copilot agent.
func NewCopilotAgent() *CopilotAgent {
	return &CopilotAgent{}
}

func (a *CopilotAgent) Type() AgentType {
	return AgentTypeCopilot
}

func (a *CopilotAgent) Name() string {
	return "GitHub Copilot"
}

func (a *CopilotAgent) URL() string {
	return "https://github.com/features/copilot"
}

// Detect checks if GitHub Copilot is the active agent.
//
// Detection methods:
//   - COPILOT_AGENT=1 (future standard)
//   - AGENT_ENV=copilot or github-copilot
func (a *CopilotAgent) Detect(ctx context.Context, env Environment) (bool, error) {
	// Check explicit COPILOT_AGENT env var
	if env.GetEnv("COPILOT_AGENT") == "1" {
		return true, nil
	}

	// Check AGENT_ENV
	agentEnv := env.GetEnv("AGENT_ENV")
	switch agentEnv {
	case "copilot", "github-copilot":
		return true, nil
	}

	return false, nil
}

// UserConfigPath returns the GitHub Copilot user configuration directory.
func (a *CopilotAgent) UserConfigPath(env Environment) (string, error) {
	home, err := env.HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "github-copilot"), nil
}

// ProjectConfigPath returns the GitHub Copilot project configuration directory.
func (a *CopilotAgent) ProjectConfigPath() string {
	return ".github"
}

// ContextFiles returns the context/instruction files Copilot supports.
func (a *CopilotAgent) ContextFiles() []string {
	return []string{".github/copilot-instructions.md"}
}

// IsInstalled checks if GitHub Copilot is installed on the system.
func (a *CopilotAgent) IsInstalled(ctx context.Context, env Environment) (bool, error) {
	// Check if gh CLI is in PATH (Copilot is a gh extension)
	if _, err := env.LookPath("gh"); err == nil {
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

var _ Agent = (*CopilotAgent)(nil)
