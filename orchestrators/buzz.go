package orchestrators

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/sageox/agentx"
)

// BuzzAgent detects Block's Buzz (https://github.com/block/buzz), a chat/agent
// platform whose buzz-acp harness spawns Claude Code / Codex / Goose as
// sub-agents over ACP. Buzz sets no fixed "spawned by Buzz" marker on the child,
// so detection leans on process ancestry (buzz-acp is always an ancestor) with
// env-var signals as cheaper first checks.
type BuzzAgent struct{}

func NewBuzzAgent() *BuzzAgent { return &BuzzAgent{} }

func (a *BuzzAgent) Type() agentx.AgentType { return agentx.AgentTypeBuzz }
func (a *BuzzAgent) Name() string           { return "Buzz" }
func (a *BuzzAgent) URL() string            { return "https://github.com/block/buzz" }
func (a *BuzzAgent) Role() agentx.AgentRole { return agentx.RoleOrchestrator }

func (a *BuzzAgent) Detect(_ context.Context, env agentx.Environment) (bool, error) {
	// Generic self-identification hatch. Not emitted by Buzz itself today, but an
	// operator or wrapper can set it, and it stays consistent with the other
	// orchestrators (Conductor/Gas City/OpenClaw).
	if env.GetEnv("ORCHESTRATOR_ENV") == "buzz" {
		return true, nil
	}

	// Best-effort env secondary: buzz-acp's own harness config var. Because
	// buzz-acp never clears the environment, a child inherits it when the
	// operator configured buzz-acp via the env var (rather than the
	// --agent-command flag).
	if env.GetEnv("BUZZ_ACP_AGENT_COMMAND") != "" {
		return true, nil
	}

	// Reliable, config-independent signal: buzz-acp is always an ancestor of the
	// agent process it launches, so it appears in the ancestry chain regardless
	// of how the operator configured Buzz. This is the reason ProcessAncestry
	// was added to the Environment interface.
	if ancestry, err := env.ProcessAncestry(); err == nil {
		for _, name := range ancestry {
			if strings.HasPrefix(name, "buzz-acp") {
				return true, nil
			}
		}
	}

	// Deliberately NOT sniffing BUZZ_RELAY_URL / BUZZ_PRIVATE_KEY: those are
	// workstation-scoped Quick-Start exports users keep in their shell profile,
	// so they would false-positive on any plain agent session in a Buzz user's
	// shell — mis-answering the discriminating question "did Buzz launch THIS
	// process?".
	return false, nil
}

func (a *BuzzAgent) UserConfigPath(env agentx.Environment) (string, error) {
	// Best-effort; nothing in agentx writes here (managers are nil). Buzz is
	// configured via CLI flags / env / relay and installs persona packs out of
	// band, so there is no canonical per-user config dir to point at.
	home, err := env.HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".buzz"), nil
}

// Buzz's "project" is a relay-hosted chat channel/workspace, not a git-adjacent
// directory, so there is no project-level config path (matches Gas City/OpenClaw).
func (a *BuzzAgent) ProjectConfigPath() string { return "" }

// Project-root context file the underlying coding agent may read; also ox's
// universal <!-- ox:prime --> injection target.
func (a *BuzzAgent) ContextFiles() []string  { return []string{"AGENTS.md"} }
func (a *BuzzAgent) SupportsXDGConfig() bool { return false }

func (a *BuzzAgent) Capabilities() agentx.Capabilities {
	// MCPServers is load-bearing: buzz-acp forwards persona-pack MCP servers to
	// the child agent over the ACP session/new request (protocol-level, identical
	// across Claude Code / Codex / Goose). That is Buzz's durable integration
	// surface.
	return agentx.Capabilities{MCPServers: true, SystemPrompt: true, ProjectContext: true}
}

// ox does not manage Buzz-native hooks/commands/rules — pure orchestrator.
func (a *BuzzAgent) HookManager() agentx.HookManager       { return nil }
func (a *BuzzAgent) CommandManager() agentx.CommandManager { return nil }
func (a *BuzzAgent) RulesManager() agentx.RulesManager     { return nil }

func (a *BuzzAgent) DetectVersion(_ context.Context, env agentx.Environment) string {
	// Best-effort; `buzz` is frequently absent from a spawned agent's PATH, so
	// this usually returns "" in the child, which is harmless.
	return versionFromCommand(env, "buzz", "--version")
}

func (a *BuzzAgent) IsInstalled(_ context.Context, env agentx.Environment) (bool, error) {
	if _, err := env.LookPath("buzz"); err == nil {
		return true, nil
	}
	if _, err := env.LookPath("buzz-acp"); err == nil {
		return true, nil
	}
	return false, nil
}

// buzz-acp assigns an ACP session id internally but exposes no per-session id
// env var to the child today. Flip to true and read that var if/when Buzz
// exports one (cf. Gas City's GC_RUN_ID).
func (a *BuzzAgent) SupportsSession() bool                 { return false }
func (a *BuzzAgent) SessionID(_ agentx.Environment) string { return "" }

var _ agentx.Agent = (*BuzzAgent)(nil)
