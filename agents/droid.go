package agents

import (
	"context"
	"path/filepath"

	"github.com/sageox/agentx"
)

// DroidAgent implements Agent for Droid (https://github.com/Factory-AI/factory).
type DroidAgent struct{}

// NewDroidAgent creates a new Droid agent.
func NewDroidAgent() *DroidAgent {
	return &DroidAgent{}
}

func (a *DroidAgent) Type() agentx.AgentType {
	return agentx.AgentTypeDroid
}

func (a *DroidAgent) Name() string {
	return "Droid"
}

func (a *DroidAgent) URL() string {
	return "https://github.com/Factory-AI/factory"
}

// Detect checks if Droid is the active agent.
//
// Detection methods:
//   - AGENT_ENV=droid
func (a *DroidAgent) Detect(ctx context.Context, env agentx.Environment) (bool, error) {
	// Check AGENT_ENV
	if env.GetEnv("AGENT_ENV") == "droid" {
		return true, nil
	}

	return false, nil
}

// UserConfigPath returns the Droid user configuration directory (~/.factory).
func (a *DroidAgent) UserConfigPath(env agentx.Environment) (string, error) {
	home, err := env.HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".factory"), nil
}

// ProjectConfigPath returns empty as Droid does not use project-level configuration.
func (a *DroidAgent) ProjectConfigPath() string {
	return ""
}

// ContextFiles returns the context/instruction files Droid supports.
func (a *DroidAgent) ContextFiles() []string {
	return []string{"AGENTS.md"}
}

// IsInstalled checks if Droid is installed on the system.
func (a *DroidAgent) IsInstalled(ctx context.Context, env agentx.Environment) (bool, error) {
	// Check if droid is in PATH
	if _, err := env.LookPath("droid"); err == nil {
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

var _ agentx.Agent = (*DroidAgent)(nil)
