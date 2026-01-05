package agentx

import (
	
	"context"
	"path/filepath"

)

// CodyAgent implements Agent for Sourcegraph Cody.
type CodyAgent struct{}

// NewCodyAgent creates a new Cody agent.
func NewCodyAgent() *CodyAgent {
	return &CodyAgent{}
}

func (a *CodyAgent) Type() AgentType {
	return AgentTypeCody
}

func (a *CodyAgent) Name() string {
	return "Cody"
}

func (a *CodyAgent) URL() string {
	return "https://github.com/sourcegraph/cody"
}

// Detect checks if Cody is the active agent.
//
// Detection methods:
//   - CODY_AGENT=1
//   - AGENT_ENV=cody
func (a *CodyAgent) Detect(ctx context.Context, env Environment) (bool, error) {
	// Check CODY env var
	if env.GetEnv("CODY_AGENT") == "1" {
		return true, nil
	}

	// Check AGENT_ENV
	if env.GetEnv("AGENT_ENV") == "cody" {
		return true, nil
	}

	return false, nil
}

// UserConfigPath returns the Cody user configuration directory.
func (a *CodyAgent) UserConfigPath(env Environment) (string, error) {
	configDir, err := env.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "cody"), nil
}

// ProjectConfigPath returns the Cody project configuration directory.
func (a *CodyAgent) ProjectConfigPath() string {
	return ".cody"
}

// ContextFiles returns the context/instruction files Cody supports.
func (a *CodyAgent) ContextFiles() []string {
	return []string{".cody/cody.json"}
}

// IsInstalled checks if Cody is installed on the system.
func (a *CodyAgent) IsInstalled(ctx context.Context, env Environment) (bool, error) {
	// Check if cody is in PATH
	if _, err := env.LookPath("cody"); err == nil {
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

var _ Agent = (*CodyAgent)(nil)
