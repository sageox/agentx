package agents

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/sageox/agentx"
)

// OMPAgent implements Agent for Oh My Pi (https://github.com/can1357/oh-my-pi),
// a fork of Mario Zechner's Pi coding agent (see PiAgent). OMP keeps Pi's
// transcript lineage and PI_* environment overrides but installs under ~/.omp
// and ships its own `omp` binary.
//
// Because OMP inherits Pi's runtime markers (PI_CODING_AGENT, PI_CODING_AGENT_DIR)
// from the shared cli.ts ancestry, those signals identify *Pi*, not OMP — so
// detection deliberately does NOT consult them. OMP is recognized only by its
// own authoritative signals: its `omp` binary in the exec/ancestor chain, and
// the AGENT_ENV=omp override that the ox integration always sets. This mirrors
// the ox-adapter-omp detector and keeps OMP from stealing Pi sessions.
type OMPAgent struct {
	hookManager    agentx.HookManager
	commandManager agentx.CommandManager
}

// NewOMPAgent creates a new Oh My Pi agent.
func NewOMPAgent() *OMPAgent {
	return &OMPAgent{}
}

func (a *OMPAgent) Type() agentx.AgentType {
	return agentx.AgentTypeOMP
}

func (a *OMPAgent) Name() string {
	return "OMP"
}

func (a *OMPAgent) URL() string {
	return "https://github.com/can1357/oh-my-pi"
}

func (a *OMPAgent) Role() agentx.AgentRole { return agentx.RoleAgent }

// Detect checks if OMP is the active agent.
//
// Detection methods:
//   - OMP's `omp` binary in the exec path (runtime signal, see DetectRuntime)
//   - AGENT_ENV=omp or AGENT_ENV=oh-my-pi (standard agentx override)
func (a *OMPAgent) Detect(ctx context.Context, env agentx.Environment) (bool, error) {
	if ok, _ := a.DetectRuntime(ctx, env); ok {
		return true, nil
	}

	switch env.GetEnv("AGENT_ENV") {
	case "omp", "oh-my-pi":
		return true, nil
	}

	return false, nil
}

// DetectRuntime reports OMP presence from runtime signals only — no AGENT_ENV
// consultation. See RuntimeDetector in agentx for the two-phase priority this
// enables.
//
// Deliberately does NOT check PI_CODING_AGENT / PI_CODING_AGENT_DIR: those are
// Pi's markers, and OMP inherits them from the shared cli.ts ancestry. Claiming
// them here would let OMP hijack real Pi sessions. The reciprocal guard lives
// in PiAgent.DetectRuntime, which yields when the launcher is `omp` — so the
// two forks' runtime detectors are mutually exclusive and the map-ordered
// registry never has to break a tie between them.
//
// LIMITATION: when a live OMP process is launched indirectly (e.g. via node,
// so `_` is not `omp`) AND exports the inherited PI_CODING_AGENT, phase 1 can
// still resolve to Pi. AGENT_ENV=omp (which the ox integration always sets) or
// Detector.DetectByType is the authoritative path in that case.
func (a *OMPAgent) DetectRuntime(_ context.Context, env agentx.Environment) (bool, error) {
	// Heuristic: the launching binary is OMP's `omp` command (matched on the
	// exact basename, not a substring, to avoid false positives like "compose").
	execPath := env.GetEnv("_")
	if strings.Contains(strings.ToLower(execPath), "oh-my-pi") {
		return true, nil
	}
	if strings.EqualFold(filepath.Base(execPath), "omp") {
		return true, nil
	}

	return false, nil
}

// UserConfigPath returns the OMP user configuration directory (~/.omp).
func (a *OMPAgent) UserConfigPath(env agentx.Environment) (string, error) {
	home, err := env.HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".omp"), nil
}

// ProjectConfigPath returns the OMP project configuration directory.
func (a *OMPAgent) ProjectConfigPath() string {
	return ".omp"
}

// ContextFiles returns the context/instruction files OMP supports.
// Like Pi, OMP loads AGENTS.md, CLAUDE.md, and SYSTEM.md from CWD and parents.
func (a *OMPAgent) ContextFiles() []string {
	return []string{"AGENTS.md", "CLAUDE.md", "SYSTEM.md"}
}

// SupportsXDGConfig returns false as OMP uses ~/.omp (home-relative dotdir).
func (a *OMPAgent) SupportsXDGConfig() bool {
	return false
}

// Capabilities returns OMP's supported features.
//
// Hooks/CustomCommands are false: unlike Pi, the SageOx integration drives OMP
// through native project guidance (.omp/AGENTS.md) and tails its JSONL
// transcript rather than installing an extension-bridge hook. It is not an
// agentx LifecycleEventMapper for the same reason.
func (a *OMPAgent) Capabilities() agentx.Capabilities {
	return agentx.Capabilities{
		Hooks:          false,
		MCPServers:     false,
		SystemPrompt:   true, // SYSTEM.md, CLAUDE.md
		ProjectContext: true, // AGENTS.md, .omp/
		CustomCommands: false,
		MinVersion:     "",
	}
}

func (a *OMPAgent) HookManager() agentx.HookManager {
	return a.hookManager
}

func (a *OMPAgent) SetHookManager(hm agentx.HookManager) {
	a.hookManager = hm
}

func (a *OMPAgent) CommandManager() agentx.CommandManager {
	return a.commandManager
}

func (a *OMPAgent) SetCommandManager(cm agentx.CommandManager) {
	a.commandManager = cm
}

// RulesManager returns the rules manager (nil if not supported).
func (a *OMPAgent) RulesManager() agentx.RulesManager {
	return nil
}

// DetectVersion attempts to determine the installed OMP version.
// Runs: omp --version
func (a *OMPAgent) DetectVersion(ctx context.Context, env agentx.Environment) string {
	return versionFromCommand(ctx, env, "omp", "--version")
}

// IsInstalled checks if OMP is installed on the system.
// Checks: omp binary in PATH or ~/.omp config directory exists.
func (a *OMPAgent) IsInstalled(ctx context.Context, env agentx.Environment) (bool, error) {
	if _, err := env.LookPath("omp"); err == nil {
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

func (a *OMPAgent) SupportsSession() bool { return true }

// SessionID returns OMP's native session identifier from the environment.
// OMP does not export a dedicated variable in every release, so this prefers
// OMP_SESSION_ID and falls back to Pi's PI_SESSION_ID (shared ancestry).
// Returns "" when neither is set; sessions are otherwise discovered by file.
func (a *OMPAgent) SessionID(env agentx.Environment) string {
	if id := env.GetEnv("OMP_SESSION_ID"); id != "" {
		return id
	}
	return env.GetEnv("PI_SESSION_ID")
}

// AgentENVAliases returns the AGENT_ENV values that identify OMP.
func (a *OMPAgent) AgentENVAliases() []string {
	return []string{"omp", "oh-my-pi"}
}

var _ agentx.Agent = (*OMPAgent)(nil)
var _ agentx.RuntimeDetector = (*OMPAgent)(nil)
