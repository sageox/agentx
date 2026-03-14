package agents

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/sageox/agentx"
)

// PiAgent implements Agent for Pi Coding Agent (https://shittycodingagent.ai/).
// Pi is a minimal terminal coding harness by Mario Zechner, installed via
// npm (@mariozechner/pi-coding-agent). It sets zero self-identification env vars
// in child processes, so detection relies on PI_CODING_AGENT_DIR (config override)
// and the $_ heuristic matching "pi-coding-agent" (the npm package name).
type PiAgent struct {
	hookManager    agentx.HookManager
	commandManager agentx.CommandManager
}

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
	return "https://shittycodingagent.ai/"
}

func (a *PiAgent) Role() agentx.AgentRole { return agentx.RoleAgent }

// Detect checks if Pi is the active agent.
//
// Detection methods:
//   - PI_CODING_AGENT_DIR is set (real env var Pi checks for config override)
//   - AGENT_ENV=pi (standard agentx mechanism)
//   - Running from pi-coding-agent binary (heuristic on $_)
func (a *PiAgent) Detect(ctx context.Context, env agentx.Environment) (bool, error) {
	// PI_CODING_AGENT_DIR is a real env var Pi uses for config directory override;
	// its presence is a strong signal that Pi is the active agent
	if _, ok := env.LookupEnv("PI_CODING_AGENT_DIR"); ok {
		return true, nil
	}

	// Check AGENT_ENV
	if env.GetEnv("AGENT_ENV") == "pi" {
		return true, nil
	}

	// Heuristic: check for "pi-coding-agent" in exec path (the npm package name).
	// We match on "pi-coding-agent" rather than bare "pi" to avoid false positives.
	if execPath := env.GetEnv("_"); strings.Contains(strings.ToLower(execPath), "pi-coding-agent") {
		return true, nil
	}

	return false, nil
}

// UserConfigPath returns the Pi user configuration directory (~/.pi).
// Pi stores config under ~/.pi/agent/ but the root is ~/.pi.
func (a *PiAgent) UserConfigPath(env agentx.Environment) (string, error) {
	home, err := env.HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".pi"), nil
}

// ProjectConfigPath returns the Pi project configuration directory.
func (a *PiAgent) ProjectConfigPath() string {
	return ".pi"
}

// ContextFiles returns the context/instruction files Pi supports.
// Pi loads AGENTS.md, CLAUDE.md, and SYSTEM.md from CWD and parent directories.
func (a *PiAgent) ContextFiles() []string {
	return []string{"AGENTS.md", "CLAUDE.md", "SYSTEM.md"}
}

// SupportsXDGConfig returns false as Pi uses ~/.pi (home-relative dotdir).
func (a *PiAgent) SupportsXDGConfig() bool {
	return false
}

// Capabilities returns Pi's supported features.
func (a *PiAgent) Capabilities() agentx.Capabilities {
	return agentx.Capabilities{
		Hooks:          false, // extension-based, not shell hooks
		MCPServers:     true,  // Pi supports MCP via extensions
		SystemPrompt:   true,  // SYSTEM.md, CLAUDE.md
		ProjectContext: true,  // AGENTS.md, .pi/
		CustomCommands: false, // Pi uses skills/extensions, not slash commands
		MinVersion:     "",
	}
}

func (a *PiAgent) HookManager() agentx.HookManager {
	return a.hookManager
}

func (a *PiAgent) SetHookManager(hm agentx.HookManager) {
	a.hookManager = hm
}

func (a *PiAgent) CommandManager() agentx.CommandManager {
	return a.commandManager
}

func (a *PiAgent) SetCommandManager(cm agentx.CommandManager) {
	a.commandManager = cm
}

// DetectVersion attempts to determine the installed Pi version.
// Runs: pi --version
func (a *PiAgent) DetectVersion(ctx context.Context, env agentx.Environment) string {
	return versionFromCommand(ctx, env, "pi", "--version")
}

// IsInstalled checks if Pi is installed on the system.
// Checks: pi binary in PATH or ~/.pi config directory exists.
func (a *PiAgent) IsInstalled(ctx context.Context, env agentx.Environment) (bool, error) {
	if _, err := env.LookPath("pi"); err == nil {
		return true, nil
	}

	configPath, err := a.UserConfigPath(env)
	if err != nil {
		return false, nil
	}
	if env.IsDir(configPath) {
		return true, nil
	}

	return false, nil
}

func (a *PiAgent) SupportsSession() bool                 { return false }
func (a *PiAgent) SessionID(_ agentx.Environment) string { return "" }

var _ agentx.Agent = (*PiAgent)(nil)
