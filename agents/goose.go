package agents

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/sageox/agentx"
)

// GooseAgent implements Agent for Goose by Block (https://block.github.io/goose/).
type GooseAgent struct {
	hookManager    agentx.HookManager
	commandManager agentx.CommandManager
}

// NewGooseAgent creates a new Goose agent.
func NewGooseAgent() *GooseAgent {
	return &GooseAgent{}
}

func (a *GooseAgent) Type() agentx.AgentType {
	return agentx.AgentTypeGoose
}

func (a *GooseAgent) Name() string {
	return "Goose"
}

func (a *GooseAgent) URL() string {
	return "https://github.com/block/goose"

}

func (a *GooseAgent) Role() agentx.AgentRole { return agentx.RoleAgent }

// Detect checks if Goose is the active agent.
//
// Detection methods:
//   - GOOSE_AGENT=1 or GOOSE=1
//   - AGENT_ENV=goose
//   - Running from goose command (heuristic)
func (a *GooseAgent) Detect(ctx context.Context, env agentx.Environment) (bool, error) {
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
func (a *GooseAgent) UserConfigPath(env agentx.Environment) (string, error) {
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

// ContextFiles returns the context/instruction files Goose loads into the
// system prompt, in Goose's own load order: AGENTS.md first, then .goosehints.
// Goose resolves both hierarchically from the working directory up to the
// repository root, and also reads the global ~/.config/goose/.goosehints.
// The set is overridable via the CONTEXT_FILE_NAMES environment variable.
//
// Reference: https://block.github.io/goose/docs/guides/context-engineering/using-goosehints
func (a *GooseAgent) ContextFiles() []string {
	return []string{"AGENTS.md", ".goosehints"}
}

// SupportsXDGConfig returns true as Goose uses ~/.config/goose.
func (a *GooseAgent) SupportsXDGConfig() bool {
	return true
}

// Capabilities returns Goose's supported features.
func (a *GooseAgent) Capabilities() agentx.Capabilities {
	return agentx.Capabilities{
		Hooks:          true,  // lifecycle hooks via plugins (goose >= 1.38)
		MCPServers:     true,  // supports MCP extensions
		SystemPrompt:   true,  // config.yaml
		ProjectContext: true,  // AGENTS.md, .goosehints
		CustomCommands: false, // recipes are not slash commands
		MinVersion:     "1.38.0",
	}
}

func (a *GooseAgent) HookManager() agentx.HookManager {
	return a.hookManager
}

func (a *GooseAgent) SetHookManager(hm agentx.HookManager) {
	a.hookManager = hm
}

func (a *GooseAgent) CommandManager() agentx.CommandManager {
	return a.commandManager
}

// RulesManager returns the rules manager (nil if not supported).
func (a *GooseAgent) RulesManager() agentx.RulesManager {
	return nil
}

func (a *GooseAgent) SetCommandManager(cm agentx.CommandManager) {
	a.commandManager = cm
}

// DetectVersion attempts to determine the installed Goose version.
// Runs: goose --version
func (a *GooseAgent) DetectVersion(ctx context.Context, env agentx.Environment) string {
	return versionFromCommand(ctx, env, "goose", "--version")
}

// IsInstalled checks if Goose is installed on the system.
// Checks: goose binary in PATH or config directory exists.
func (a *GooseAgent) IsInstalled(ctx context.Context, env agentx.Environment) (bool, error) {
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

// EventPhases returns Goose's native event-to-phase mapping.
//
// Goose follows the Open Plugins hooks specification. It fires several more
// events than are mapped here (PostToolUseFailure, BeforeReadFile,
// AfterFileEdit, BeforeShellExecution, AfterShellExecution); those have no
// canonical phase equivalent and are deliberately omitted.
//
// Goose has no compaction event, so PhaseCompact is unreachable — context
// injected at session start does not survive a Goose compaction.
//
// Reference: https://block.github.io/goose/docs/guides/context-engineering/hooks
func (a *GooseAgent) EventPhases() agentx.EventPhaseMap {
	return agentx.EventPhaseMap{
		agentx.HookEventSessionStart:     agentx.PhaseStart,
		agentx.HookEventSessionEnd:       agentx.PhaseEnd,
		agentx.HookEventPreToolUse:       agentx.PhaseBeforeTool,
		agentx.HookEventPostToolUse:      agentx.PhaseAfterTool,
		agentx.HookEventUserPromptSubmit: agentx.PhasePrompt,
		agentx.HookEventStop:             agentx.PhaseStop,
	}
}

// AgentENVAliases returns the AGENT_ENV values that identify Goose.
func (a *GooseAgent) AgentENVAliases() []string {
	return []string{"goose"}
}

// SupportsSession returns true; Goose supplies a session_id on every hook
// payload. There is no session-ID environment variable, so SessionID always
// returns empty and callers must read it from the hook payload instead.
func (a *GooseAgent) SupportsSession() bool                 { return true }
func (a *GooseAgent) SessionID(_ agentx.Environment) string { return "" }

var _ agentx.Agent = (*GooseAgent)(nil)
var _ agentx.LifecycleEventMapper = (*GooseAgent)(nil)
