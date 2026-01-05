package agentx

import (
	"context"
	"path/filepath"
	"strings"

)

// GooseAgent implements Agent for Goose by Block (https://block.github.io/goose/).
type GooseAgent struct{}

// NewGooseAgent creates a new Goose agent.
func NewGooseAgent() *GooseAgent {
	return &GooseAgent{}
}

func (a *GooseAgent) Type() AgentType {
	return AgentTypeGoose
}

func (a *GooseAgent) Name() string {
	return "Goose"
}

func (a *GooseAgent) URL() string {
	return "https://github.com/block/goose"
}

// Detect checks if Goose is the active agent.
//
// Detection methods:
//   - GOOSE_AGENT=1 or GOOSE=1
//   - AGENT_ENV=goose
//   - Running from goose command (heuristic)
func (a *GooseAgent) Detect(ctx context.Context, env Environment) (bool, error) {
	// Check GOOSE env vars
	if env.GetEnv("GOOSE") == "1" || env.GetEnv("GOOSE_AGENT") == "1" {
		return true, nil
	}

	// Check AGENT_ENV
	if env.GetEnv("AGENT_ENV") == "goose" {
		return true, nil
	}

	// Heuristic: check if running from goose CLI
	if execPath := env.GetEnv("_"); strings.Contains(strings.ToLower(execPath), "goose") {
		return true, nil
	}

	return false, nil
}

// UserConfigPath returns the Goose user configuration directory.
// Goose uses XDG-compliant paths (~/.config/goose).
func (a *GooseAgent) UserConfigPath(env Environment) (string, error) {
	configDir, err := env.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "goose"), nil
}

// ProjectConfigPath returns empty as Goose is primarily user-level configuration.
func (a *GooseAgent) ProjectConfigPath() string {
	return ""
}

// ContextFiles returns the context/instruction files Goose supports.
func (a *GooseAgent) ContextFiles() []string {
	return []string{".goose/config.yaml", ".goosehints"}
}

// IsInstalled checks if Goose is installed on the system.
func (a *GooseAgent) IsInstalled(ctx context.Context, env Environment) (bool, error) {
	// Check if goose is in PATH
	if _, err := env.LookPath("goose"); err == nil {
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

var _ Agent = (*GooseAgent)(nil)
