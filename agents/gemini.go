package agents

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/sageox/agentx"
)

// GeminiAgent implements Agent for Gemini CLI (https://github.com/google-gemini/gemini-cli).
// Gemini CLI is Google's terminal-based coding agent powered by Gemini models.
// The ox adapter sets AGENT_ENV=gemini; native detection also checks GEMINI=1
// and GEMINI_AGENT=1.
type GeminiAgent struct {
	hookManager    agentx.HookManager
	commandManager agentx.CommandManager
}

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

func (a *GeminiAgent) Role() agentx.AgentRole { return agentx.RoleAgent }

// Detect checks if Gemini CLI is the active agent.
//
// Detection methods:
//   - GEMINI=1 or GEMINI_AGENT=1
//   - AGENT_ENV=gemini
//   - Running from gemini binary (heuristic on $_)
func (a *GeminiAgent) Detect(ctx context.Context, env agentx.Environment) (bool, error) {
	// Check explicit Gemini env vars
	if env.GetEnv("GEMINI") == "1" || env.GetEnv("GEMINI_AGENT") == "1" {
		return true, nil
	}

	// Check AGENT_ENV
	if strings.ToLower(env.GetEnv("AGENT_ENV")) == "gemini" {
		return true, nil
	}

	// Heuristic: check if running from gemini CLI
	if execPath := env.GetEnv("_"); strings.Contains(strings.ToLower(execPath), "gemini") {
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

// SupportsXDGConfig returns false; Gemini CLI uses ~/.gemini.
func (a *GeminiAgent) SupportsXDGConfig() bool {
	return false
}

// Capabilities returns Gemini CLI's supported features.
func (a *GeminiAgent) Capabilities() agentx.Capabilities {
	return agentx.Capabilities{
		Hooks:          false,
		MCPServers:     true,
		SystemPrompt:   true,
		ProjectContext: true,
		CustomCommands: false,
		MinVersion:     "",
	}
}

// HookManager returns the hook manager for Gemini CLI.
func (a *GeminiAgent) HookManager() agentx.HookManager {
	return a.hookManager
}

// SetHookManager sets the hook manager.
func (a *GeminiAgent) SetHookManager(hm agentx.HookManager) {
	a.hookManager = hm
}

func (a *GeminiAgent) CommandManager() agentx.CommandManager {
	return a.commandManager
}

// RulesManager returns the rules manager (nil if not supported).
func (a *GeminiAgent) RulesManager() agentx.RulesManager {
	return nil
}

func (a *GeminiAgent) SetCommandManager(cm agentx.CommandManager) {
	a.commandManager = cm
}

// DetectVersion attempts to determine the installed Gemini CLI version.
// Runs: gemini --version
func (a *GeminiAgent) DetectVersion(ctx context.Context, env agentx.Environment) string {
	return versionFromCommand(ctx, env, "gemini", "--version")
}

// IsInstalled checks if Gemini CLI is installed on the system.
func (a *GeminiAgent) IsInstalled(ctx context.Context, env agentx.Environment) (bool, error) {
	// Check if gemini CLI is in PATH
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

// SupportsSession returns false; Gemini CLI does not expose session IDs.
func (a *GeminiAgent) SupportsSession() bool                 { return false }
func (a *GeminiAgent) SessionID(_ agentx.Environment) string { return "" }

// Ensure GeminiAgent implements Agent.
var _ agentx.Agent = (*GeminiAgent)(nil)
