package agents

import (
	"context"
	"path/filepath"

	"github.com/sageox/agentx"
)

// GeminiAgent implements Agent for Gemini CLI (https://github.com/google-gemini/gemini-cli).
type GeminiAgent struct{}

// NewGeminiAgent creates a new Gemini CLI agent.
func NewGeminiAgent() *GeminiAgent {
	return &GeminiAgent{}
}

func (a *GeminiAgent) Type() agentx.AgentType {
	return agentx.AgentTypeGemini
}

func (a *GeminiAgent) Name() string {
	return "Gemini CLI"
}

func (a *GeminiAgent) URL() string {
	return "https://github.com/google-gemini/gemini-cli"
}

// Detect checks if Gemini CLI is the active agent.
//
// Detection methods:
//   - GEMINI_AGENT=1
//   - AGENT_ENV=gemini
func (a *GeminiAgent) Detect(ctx context.Context, env agentx.Environment) (bool, error) {
	// Check GEMINI env var
	if env.GetEnv("GEMINI_AGENT") == "1" {
		return true, nil
	}

	// Check AGENT_ENV
	if env.GetEnv("AGENT_ENV") == "gemini" {
		return true, nil
	}

	return false, nil
}

// UserConfigPath returns the Gemini CLI user configuration directory (~/.gemini).
func (a *GeminiAgent) UserConfigPath(env agentx.Environment) (string, error) {
	home, err := env.HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gemini"), nil
}

// ProjectConfigPath returns the Gemini CLI project configuration directory.
func (a *GeminiAgent) ProjectConfigPath() string {
	return ".gemini"
}

// ContextFiles returns the context/instruction files Gemini CLI supports.
func (a *GeminiAgent) ContextFiles() []string {
	return []string{"GEMINI.md", "AGENTS.md"}
}

// IsInstalled checks if Gemini CLI is installed on the system.
func (a *GeminiAgent) IsInstalled(ctx context.Context, env agentx.Environment) (bool, error) {
	// Check if gemini is in PATH
	if _, err := env.LookPath("gemini"); err == nil {
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

var _ agentx.Agent = (*GeminiAgent)(nil)
