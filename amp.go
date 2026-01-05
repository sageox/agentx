package agentx

import (
	"context"
	"path/filepath"

)

// AmpAgent implements Agent for Amp by Sourcegraph (https://ampcode.io).
type AmpAgent struct{}

// NewAmpAgent creates a new Amp agent.
func NewAmpAgent() *AmpAgent {
	return &AmpAgent{}
}

func (a *AmpAgent) Type() AgentType {
	return AgentTypeAmp
}

func (a *AmpAgent) Name() string {
	return "Amp"
}

func (a *AmpAgent) URL() string {
	return "https://github.com/sourcegraph/amp"
}

// Detect checks if Amp is the active agent.
//
// Detection methods:
//   - AMP_AGENT=1 or AMP=1
//   - AGENT_ENV=amp
func (a *AmpAgent) Detect(ctx context.Context, env Environment) (bool, error) {
	// Check AMP env vars
	if env.GetEnv("AMP") == "1" || env.GetEnv("AMP_AGENT") == "1" {
		return true, nil
	}

	// Check AGENT_ENV
	if env.GetEnv("AGENT_ENV") == "amp" {
		return true, nil
	}

	return false, nil
}

// UserConfigPath returns the Amp user configuration directory (~/.amp).
func (a *AmpAgent) UserConfigPath(env Environment) (string, error) {
	home, err := env.HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".amp"), nil
}

// ProjectConfigPath returns empty as Amp is primarily user-level configuration.
func (a *AmpAgent) ProjectConfigPath() string {
	return ""
}

// ContextFiles returns the context/instruction files Amp supports.
func (a *AmpAgent) ContextFiles() []string {
	return []string{"AGENTS.md"}
}

// IsInstalled checks if Amp is installed on the system.
func (a *AmpAgent) IsInstalled(ctx context.Context, env Environment) (bool, error) {
	// Check if amp is in PATH
	if _, err := env.LookPath("amp"); err == nil {
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

var _ Agent = (*AmpAgent)(nil)
