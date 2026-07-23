package orchestrators

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sageox/agentx"
)

func TestBuzzAgent_Detect(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		ancestry    []string
		ancestryErr bool
		want        bool
		why         string
	}{
		{
			name: "ORCHESTRATOR_ENV=buzz",
			env:  map[string]string{"ORCHESTRATOR_ENV": "buzz"},
			want: true,
			why:  "generic self-ID hatch must detect",
		},
		{
			name: "BUZZ_ACP_AGENT_COMMAND set",
			env:  map[string]string{"BUZZ_ACP_AGENT_COMMAND": "claude-agent-acp"},
			want: true,
			why:  "buzz-acp harness config var (env-form) is a valid secondary signal",
		},
		{
			name:     "buzz-acp direct parent",
			ancestry: []string{"claude-agent-acp", "buzz-acp"},
			want:     true,
			why:      "the reliable spawn-scoped signal must fire with no env vars set",
		},
		{
			name:     "buzz-acp deeper in ancestry chain",
			ancestry: []string{"bash", "claude", "buzz-acp", "login"},
			want:     true,
			why:      "detection must walk past intermediate shells/agents",
		},
		{
			name:     "buzz-acp-worker prefix",
			ancestry: []string{"buzz-acp-worker"},
			want:     true,
			why:      "HasPrefix must tolerate suffixed/truncated harness names",
		},
		{
			name: "BUZZ_RELAY_URL alone does NOT detect",
			env:  map[string]string{"BUZZ_RELAY_URL": "ws://localhost:3000"},
			want: false,
			why:  "workstation-scoped var must never false-tag a plain agent session (guards against re-adding the rejected heuristic)",
		},
		{
			name: "BUZZ_PRIVATE_KEY alone does NOT detect",
			env:  map[string]string{"BUZZ_PRIVATE_KEY": "nsec1abc"},
			want: false,
			why:  "workstation-scoped var must never false-tag a plain agent session",
		},
		{
			name:     "bare buzz in ancestry does NOT detect",
			ancestry: []string{"buzz", "bash"},
			want:     false,
			why:      "the Buzz chat CLI is not the buzz-acp spawner; only buzz-acp counts",
		},
		{
			name: "wrong ORCHESTRATOR_ENV",
			env:  map[string]string{"ORCHESTRATOR_ENV": "conductor"},
			want: false,
			why:  "another orchestrator's declaration must not be claimed as Buzz",
		},
		{
			name:     "unrelated ancestry",
			ancestry: []string{"claude", "bash", "zsh"},
			want:     false,
			why:      "a normal Claude Code terminal session must not detect as Buzz",
		},
		{
			name: "no signals",
			want: false,
			why:  "clean environment is the negative baseline",
		},
		{
			name: "ORCHESTRATOR_ENV case-sensitive",
			env:  map[string]string{"ORCHESTRATOR_ENV": "Buzz"},
			want: false,
			why:  "match is exact (no case-fold), mirroring the other orchestrators",
		},
		{
			name:        "ancestry probe error falls through cleanly",
			ancestryErr: true,
			want:        false,
			why:         "a ProcessAncestry error must not panic or false-positive; detection just yields false",
		},
		{
			name:        "env signal wins even when ancestry probe errors",
			env:         map[string]string{"ORCHESTRATOR_ENV": "buzz"},
			ancestryErr: true,
			want:        true,
			why:         "cheap env checks run before ancestry, so a probe failure never masks an explicit signal",
		},
		{
			name:     "buzz-acp ancestry detects despite a conflicting ORCHESTRATOR_ENV",
			env:      map[string]string{"ORCHESTRATOR_ENV": "conductor"},
			ancestry: []string{"claude-agent-acp", "buzz-acp"},
			want:     true,
			why:      "ancestry is ground truth: buzz-acp literally launched us, so a stale/wrong env label must not hide it",
		},
	}

	agent := NewBuzzAgent()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := agentx.NewMockEnvironment(tt.env)
			env.Ancestry = tt.ancestry
			if tt.ancestryErr {
				env.AncestryError = errors.New("ps snapshot failed")
			}
			got, err := agent.Detect(ctx, env)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got, tt.why)
		})
	}
}

func TestBuzzAgent_Identity(t *testing.T) {
	agent := NewBuzzAgent()
	assert.Equal(t, agentx.AgentTypeBuzz, agent.Type())
	assert.Equal(t, "Buzz", agent.Name())
	assert.Equal(t, agentx.RoleOrchestrator, agent.Role())
	assert.Equal(t, "https://github.com/block/buzz", agent.URL())
}

func TestBuzzAgent_PureOrchestrator(t *testing.T) {
	agent := NewBuzzAgent()
	// Buzz manages no ox-native hooks/commands/rules and exposes no session id.
	assert.Nil(t, agent.HookManager())
	assert.Nil(t, agent.CommandManager())
	assert.Nil(t, agent.RulesManager())
	assert.False(t, agent.SupportsSession())
	assert.Equal(t, "", agent.ProjectConfigPath())
}

func TestBuzzAgent_Config(t *testing.T) {
	agent := NewBuzzAgent()
	ctx := context.Background()

	// Capabilities: MCP is the load-bearing one (persona-pack MCP forwarding).
	caps := agent.Capabilities()
	assert.True(t, caps.MCPServers, "MCPServers must be true — Buzz's durable integration surface")
	assert.True(t, caps.SystemPrompt)
	assert.True(t, caps.ProjectContext)
	assert.False(t, caps.Hooks, "ox installs no Buzz-native hooks")

	assert.Equal(t, []string{"AGENTS.md"}, agent.ContextFiles())
	assert.False(t, agent.SupportsXDGConfig())

	// UserConfigPath is a best-effort ~/.buzz.
	env := agentx.NewMockEnvironment(nil) // Home defaults to /home/test
	p, err := agent.UserConfigPath(env)
	assert.NoError(t, err)
	assert.Equal(t, "/home/test/.buzz", p)

	// IsInstalled: either `buzz` or `buzz-acp` on PATH counts.
	assert.NoError(t, err)
	notInstalled, err := agent.IsInstalled(ctx, agentx.NewMockEnvironment(nil))
	assert.NoError(t, err)
	assert.False(t, notInstalled, "no binary on PATH -> not installed")

	withHarness := agentx.NewMockEnvironment(nil)
	withHarness.PathBinaries = map[string]string{"buzz-acp": "/usr/local/bin/buzz-acp"}
	installed, err := agent.IsInstalled(ctx, withHarness)
	assert.NoError(t, err)
	assert.True(t, installed, "buzz-acp on PATH -> installed")

	// DetectVersion parses `buzz --version` output; absent binary -> "".
	verEnv := agentx.NewMockEnvironment(nil)
	verEnv.ExecOutputs = map[string][]byte{"buzz": []byte("buzz version 1.2.3\n")}
	assert.Equal(t, "1.2.3", agent.DetectVersion(ctx, verEnv))
	assert.Equal(t, "", agent.DetectVersion(ctx, agentx.NewMockEnvironment(nil)))
}
