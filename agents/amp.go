package agents

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/sageox/agentx"
)

// AmpAgent implements Agent for Amp by Sourcegraph (https://ampcode.com).
type AmpAgent struct {
	hookManager    agentx.HookManager
	commandManager agentx.CommandManager
}

// NewAmpAgent creates a new Amp agent.
func NewAmpAgent() *AmpAgent {
	return &AmpAgent{}
}

func (a *AmpAgent) Type() agentx.AgentType {
	return agentx.AgentTypeAmp
}

func (a *AmpAgent) Name() string {
	return "Amp"
}

func (a *AmpAgent) URL() string {
	return "https://ampcode.com"
}

func (a *AmpAgent) Role() agentx.AgentRole { return agentx.RoleAgent }

// Detect checks if Amp is the active agent.
//
// Detection methods:
//   - AMP_AGENT=1 or AMP=1
//   - AMP_THREAD_URL is set
//   - AGENT_ENV=amp
//   - AMP_ORB is set (inside an Amp orb)
func (a *AmpAgent) Detect(ctx context.Context, env agentx.Environment) (bool, error) {
	if ok, _ := a.DetectRuntime(ctx, env); ok {
		return true, nil
	}

	// Check AGENT_ENV (caller-provided override)
	if env.GetEnv("AGENT_ENV") == "amp" {
		return true, nil
	}

	if ok, _ := a.DetectVenue(ctx, env); ok {
		return true, nil
	}

	return false, nil
}

// DetectVenue reports an Amp orb. Amp sets AMP_ORB for every process in
// an orb. An orb's agent shell was seen with AMP_ORB=1 and none of the
// runtime variables DetectRuntime checks, so without this an orb session
// was never recognised as Amp's (sageox/agentx#17).
func (a *AmpAgent) DetectVenue(_ context.Context, env agentx.Environment) (bool, error) {
	return env.GetEnv("AMP_ORB") != "", nil
}

// DetectRuntime reports Amp presence from runtime signals only —
// no AGENT_ENV consultation. See RuntimeDetector in agentx for the
// two-phase priority this enables (#527).
func (a *AmpAgent) DetectRuntime(_ context.Context, env agentx.Environment) (bool, error) {
	// Check AMP env vars
	if env.GetEnv("AMP") == "1" || env.GetEnv("AMP_AGENT") == "1" {
		return true, nil
	}

	// The thread's own variables are per-agent evidence. Amp exports
	// AMP_THREAD_URL on a laptop; an orb's agent shell has AMP_THREAD_ID
	// instead (sageox/agentx#17).
	if env.GetEnv("AMP_THREAD_URL") != "" || env.GetEnv("AMP_THREAD_ID") != "" {
		return true, nil
	}

	return false, nil
}

// UserConfigPath returns the Amp user configuration directory (~/.config/amp).
func (a *AmpAgent) UserConfigPath(env agentx.Environment) (string, error) {
	configDir, err := env.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "amp"), nil
}

// ProjectConfigPath returns empty as Amp is primarily user-level configuration.
func (a *AmpAgent) ProjectConfigPath() string {
	return ""
}

// ContextFiles returns the context/instruction files Amp supports.
func (a *AmpAgent) ContextFiles() []string {
	return []string{"AGENTS.md"}
}

// SupportsXDGConfig returns true as Amp uses ~/.config/amp (XDG-compliant).
func (a *AmpAgent) SupportsXDGConfig() bool {
	return true
}

// Capabilities returns Amp's supported features.
func (a *AmpAgent) Capabilities() agentx.Capabilities {
	return agentx.Capabilities{
		Hooks:          true, // amp.hooks in .vscode/settings.json
		MCPServers:     true, // supports MCP
		SystemPrompt:   true, // custom instructions
		ProjectContext: true, // AGENTS.md
		CustomCommands: true, // toolboxes and skills
		MinVersion:     "",
	}
}

func (a *AmpAgent) HookManager() agentx.HookManager {
	return a.hookManager
}

func (a *AmpAgent) SetHookManager(hm agentx.HookManager) {
	a.hookManager = hm
}

func (a *AmpAgent) CommandManager() agentx.CommandManager {
	return a.commandManager
}

// RulesManager returns the rules manager (nil if not supported).
func (a *AmpAgent) RulesManager() agentx.RulesManager {
	return nil
}

func (a *AmpAgent) SetCommandManager(cm agentx.CommandManager) {
	a.commandManager = cm
}

// DetectVersion attempts to determine the installed Amp version.
// Runs: amp --version
func (a *AmpAgent) DetectVersion(ctx context.Context, env agentx.Environment) string {
	return versionFromCommand(ctx, env, "amp", "--version")
}

// IsInstalled checks if Amp is installed on the system.
// Checks: amp binary in PATH or config directory exists.
func (a *AmpAgent) IsInstalled(ctx context.Context, env agentx.Environment) (bool, error) {
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

// EventPhases returns Amp's native event-to-phase mapping.
// Reference: https://ampcode.com/news/hooks
func (a *AmpAgent) EventPhases() agentx.EventPhaseMap {
	return agentx.EventPhaseMap{
		agentx.AmpEventToolPreExecute:  agentx.PhaseBeforeTool,
		agentx.AmpEventToolPostExecute: agentx.PhaseAfterTool,
	}
}

// AgentENVAliases returns the AGENT_ENV values that identify Amp.
func (a *AmpAgent) AgentENVAliases() []string {
	return []string{"amp"}
}

func (a *AmpAgent) SupportsSession() bool { return true }

// SessionID returns the thread URL that recordings of this thread are keyed
// by. An orb's agent shell has no AMP_THREAD_URL, only AMP_THREAD_ID and
// AMP_URL (sageox/agentx#17), so the URL is rebuilt from those in the shape
// the ox-bridge plugin sends at session end, <AMP_URL>/threads/<id>.
// Without it an orb recording has no thread to match and never uploads.
func (a *AmpAgent) SessionID(env agentx.Environment) string {
	if threadURL := env.GetEnv("AMP_THREAD_URL"); threadURL != "" {
		return threadURL
	}
	threadID := strings.TrimSpace(env.GetEnv("AMP_THREAD_ID"))
	// a thread id is one URL path segment; anything else isn't one we can use
	if threadID == "" || strings.ContainsAny(threadID, "/?# ") {
		return ""
	}
	base := strings.TrimRight(strings.TrimSpace(env.GetEnv("AMP_URL")), "/")
	if base == "" {
		base = "https://ampcode.com"
	}
	return base + "/threads/" + threadID
}

var _ agentx.Agent = (*AmpAgent)(nil)
var _ agentx.LifecycleEventMapper = (*AmpAgent)(nil)
