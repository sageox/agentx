package agentx

import (
	"context"
	"path/filepath"
	"strings"

)

// AiderAgent implements Agent for Aider (https://aider.chat).
type AiderAgent struct{}

// NewAiderAgent creates a new Aider agent.
func NewAiderAgent() *AiderAgent {
	return &AiderAgent{}
}

func (a *AiderAgent) Type() AgentType {
	return AgentTypeAider
}

func (a *AiderAgent) Name() string {
	return "Aider"
}

func (a *AiderAgent) URL() string {
	return "https://github.com/Aider-AI/aider"
}

// Detect checks if Aider is the active agent.
//
// Detection methods:
//   - AIDER=1 or AIDER_AGENT=1
//   - AGENT_ENV=aider
//   - Running from aider command
func (a *AiderAgent) Detect(ctx context.Context, env Environment) (bool, error) {
	// Check AIDER env vars
	if env.GetEnv("AIDER") == "1" || env.GetEnv("AIDER_AGENT") == "1" {
		return true, nil
	}

	// Check AGENT_ENV
	if env.GetEnv("AGENT_ENV") == "aider" {
		return true, nil
	}

	// Heuristic: check if running from aider
	if execPath := env.GetEnv("_"); strings.Contains(strings.ToLower(execPath), "aider") {
		return true, nil
	}

	return false, nil
}

// UserConfigPath returns the Aider user configuration directory.
func (a *AiderAgent) UserConfigPath(env Environment) (string, error) {
	home, err := env.HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".aider"), nil
}

// ProjectConfigPath returns the Aider project configuration directory.
func (a *AiderAgent) ProjectConfigPath() string {
	return ".aider"
}

// ContextFiles returns the context/instruction files Aider supports.
func (a *AiderAgent) ContextFiles() []string {
	return []string{".aider.conf.yml", "CONVENTIONS.md"}
}

// IsInstalled checks if Aider is installed on the system.
func (a *AiderAgent) IsInstalled(ctx context.Context, env Environment) (bool, error) {
	// Check if aider is in PATH
	if _, err := env.LookPath("aider"); err == nil {
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

var _ Agent = (*AiderAgent)(nil)
