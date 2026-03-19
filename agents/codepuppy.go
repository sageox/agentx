package agents

import (
	"github.com/sageox/agentx"
	"context"
	"path/filepath"

)

// CodePuppyAgent implements Agent for Code Puppy.
type CodePuppyAgent struct{}

// NewCodePuppyAgent creates a new Code Puppy agent.
func NewCodePuppyAgent() *CodePuppyAgent {
	return &CodePuppyAgent{}
}

func (a *CodePuppyAgent) Type() agentx.AgentType {
	return agentx.AgentTypeCodePuppy
}

func (a *CodePuppyAgent) Name() string {
	return "Code Puppy"
}

func (a *CodePuppyAgent) URL() string {
	return "https://github.com/codepuppy-ai/codepuppy"
}

// Detect checks if Code Puppy is the active agent.
//
// Detection methods:
//   - CODE_PUPPY=1 or CODE_PUPPY_AGENT=1
//   - AGENT_ENV=code-puppy or codepuppy
func (a *CodePuppyAgent) Detect(ctx context.Context, env agentx.Environment) (bool, error) {
	// Check CODE_PUPPY env vars
	if env.GetEnv("CODE_PUPPY") == "1" || env.GetEnv("CODE_PUPPY_AGENT") == "1" {
		return true, nil
	}

	// Check AGENT_ENV
	agentEnv := env.GetEnv("AGENT_ENV")
	switch agentEnv {
	case "code-puppy", "codepuppy":
		return true, nil
	}

	return false, nil
}

// UserConfigPath returns the Code Puppy user configuration directory.
func (a *CodePuppyAgent) UserConfigPath(env agentx.Environment) (string, error) {
	configDir, err := env.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "code-puppy"), nil
}

// ProjectConfigPath returns the Code Puppy project configuration directory.
func (a *CodePuppyAgent) ProjectConfigPath() string {
	return ".codepuppy"
}

// ContextFiles returns the context/instruction files Code Puppy supports.
func (a *CodePuppyAgent) ContextFiles() []string {
	return []string{".codepuppy/config.json"}
}

// IsInstalled checks if Code Puppy is installed on the system.
func (a *CodePuppyAgent) IsInstalled(ctx context.Context, env agentx.Environment) (bool, error) {
	// Check if codepuppy is in PATH
	if _, err := env.LookPath("codepuppy"); err == nil {
		return true, nil
	}

	// Also check code-puppy
	if _, err := env.LookPath("code-puppy"); err == nil {
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

var _ agentx.Agent = (*CodePuppyAgent)(nil)
