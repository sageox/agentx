package agentx

import (
	"context"
	"path/filepath"

)

// WindsurfAgent implements Agent for Windsurf (Codeium).
type WindsurfAgent struct{}

// NewWindsurfAgent creates a new Windsurf agent.
func NewWindsurfAgent() *WindsurfAgent {
	return &WindsurfAgent{}
}

func (a *WindsurfAgent) Type() AgentType {
	return AgentTypeWindsurf
}

func (a *WindsurfAgent) Name() string {
	return "Windsurf"
}

func (a *WindsurfAgent) URL() string {
	return "https://github.com/codeium/windsurf"
}

// Detect checks if Windsurf is the active agent.
//
// Detection methods:
//   - WINDSURF_AGENT=1 (future standard)
//   - CODEIUM_AGENT=1 (alternative)
//   - AGENT_ENV=windsurf or codeium
func (a *WindsurfAgent) Detect(ctx context.Context, env Environment) (bool, error) {
	// Check explicit WINDSURF_AGENT env var
	if env.GetEnv("WINDSURF_AGENT") == "1" {
		return true, nil
	}

	// Check CODEIUM_AGENT (Windsurf was formerly Codeium)
	if env.GetEnv("CODEIUM_AGENT") == "1" {
		return true, nil
	}

	// Check AGENT_ENV
	agentEnv := env.GetEnv("AGENT_ENV")
	switch agentEnv {
	case "windsurf", "codeium":
		return true, nil
	}

	return false, nil
}

// UserConfigPath returns the Windsurf user configuration directory (~/.codeium).
func (a *WindsurfAgent) UserConfigPath(env Environment) (string, error) {
	home, err := env.HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".codeium"), nil
}

// ProjectConfigPath returns the Windsurf project configuration directory.
func (a *WindsurfAgent) ProjectConfigPath() string {
	return ".windsurf"
}

// ContextFiles returns the context/instruction files Windsurf supports.
func (a *WindsurfAgent) ContextFiles() []string {
	return []string{".windsurfrules"}
}

// IsInstalled checks if Windsurf is installed on the system.
func (a *WindsurfAgent) IsInstalled(ctx context.Context, env Environment) (bool, error) {
	// Check if windsurf CLI is in PATH
	if _, err := env.LookPath("windsurf"); err == nil {
		return true, nil
	}

	// Check for macOS application bundle
	if env.GOOS() == "darwin" {
		if env.IsDir("/Applications/Windsurf.app") {
			return true, nil
		}
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

var _ Agent = (*WindsurfAgent)(nil)
