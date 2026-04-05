package agents

import (
	"context"
	"path/filepath"

	"github.com/sageox/agentx"
)

// PiAgent implements Agent for Pi (https://github.com/anthropics/pi).
type PiAgent struct{}

// NewPiAgent creates a new Pi agent.
func NewPiAgent() *PiAgent {
	return &PiAgent{}
}

func (a *PiAgent) Type() agentx.AgentType {
	return agentx.AgentTypePi
}

func (a *PiAgent) Name() string {
	return "Pi"
}

func (a *PiAgent) URL() string {
	return "https://github.com/anthropics/pi"
}

// Detect checks if Pi is the active agent.
//
// Detection methods:
//   - PI_AGENT=1
//   - AGENT_ENV=pi
func (a *PiAgent) Detect(ctx context.Context, env agentx.Environment) (bool, error) {
	// Check PI env var
	if env.GetEnv("PI_AGENT") == "1" {
		return true, nil
	}

	// Check AGENT_ENV
	if env.GetEnv("AGENT_ENV") == "pi" {
		return true, nil
	}

	return false, nil
}

// UserConfigPath returns the Pi user configuration directory (~/.pi).
func (a *PiAgent) UserConfigPath(env agentx.Environment) (string, error) {
	home, err := env.HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".pi"), nil
}

// ProjectConfigPath returns empty as Pi does not use project-level configuration.
func (a *PiAgent) ProjectConfigPath() string {
	return ""
}

// ContextFiles returns the context/instruction files Pi supports.
func (a *PiAgent) ContextFiles() []string {
	return []string{"AGENTS.md"}
}

// IsInstalled checks if Pi is installed on the system.
func (a *PiAgent) IsInstalled(ctx context.Context, env agentx.Environment) (bool, error) {
	// Check if pi is in PATH
	if _, err := env.LookPath("pi"); err == nil {
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

var _ agentx.Agent = (*PiAgent)(nil)
