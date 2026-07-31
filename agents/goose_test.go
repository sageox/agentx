package agents

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sageox/agentx"
)

// TestGooseContextFiles pins the load ORDER, not just membership. Goose reads
// AGENTS.md first and .goosehints second, and a local file wins over the global
// ~/.config/goose/.goosehints. Callers that inject a prime marker need the first
// entry to be the file they should write to.
func TestGooseContextFiles(t *testing.T) {
	agent := NewGooseAgent()
	files := agent.ContextFiles()

	require.Len(t, files, 2)
	assert.Equal(t, "AGENTS.md", files[0], "AGENTS.md must come first — Goose loads it before .goosehints")
	assert.Equal(t, ".goosehints", files[1])
}

func TestGooseCapabilities(t *testing.T) {
	agent := NewGooseAgent()
	caps := agent.Capabilities()

	assert.True(t, caps.Hooks, "Goose ships lifecycle hooks via the plugin system")
	assert.True(t, caps.MCPServers)
	assert.True(t, caps.SystemPrompt)
	assert.True(t, caps.ProjectContext)
	assert.False(t, caps.CustomCommands, "recipes are not slash commands")
	assert.Equal(t, "1.38.0", caps.MinVersion, "hooks landed in goose 1.38")
}

func TestGooseEventPhases(t *testing.T) {
	agent := NewGooseAgent()

	mapper, ok := agentx.Agent(agent).(agentx.LifecycleEventMapper)
	require.True(t, ok, "Goose must implement LifecycleEventMapper")

	phases := mapper.EventPhases()

	expected := agentx.EventPhaseMap{
		agentx.HookEventSessionStart:       agentx.PhaseStart,
		agentx.HookEventSessionEnd:         agentx.PhaseEnd,
		agentx.HookEventPreToolUse:         agentx.PhaseBeforeTool,
		agentx.HookEventPostToolUse:        agentx.PhaseAfterTool,
		agentx.HookEventPostToolUseFailure: agentx.PhaseAfterTool,
		agentx.HookEventUserPromptSubmit:   agentx.PhasePrompt,
		agentx.HookEventStop:               agentx.PhaseStop,
	}
	assert.Equal(t, expected, phases)

	// Goose has no compaction event. Anything relying on PhaseCompact to
	// re-inject context after a context reset will never fire under Goose —
	// that limitation is upstream and cannot be worked around by a caller.
	for _, phase := range phases {
		assert.NotEqual(t, agentx.PhaseCompact, phase,
			"Goose fires no compaction event; mapping one would silently promise re-priming that never happens")
	}

	assert.Equal(t, []string{"goose"}, mapper.AgentENVAliases())
}

// TestGooseSession documents why SupportsSession is true while SessionID is
// always empty: Goose puts session_id on the hook payload (stdin), never in the
// environment. Callers must read it from the payload.
func TestGooseSession(t *testing.T) {
	agent := NewGooseAgent()

	assert.True(t, agent.SupportsSession())
	assert.Equal(t, "", agent.SessionID(agentx.NewMockEnvironment(nil)))
	assert.Equal(t, "", agent.SessionID(agentx.NewMockEnvironment(map[string]string{
		"GOOSE_SESSION_ID": "should-be-ignored",
	})), "Goose exposes no session-ID env var")
}

// TestGooseRegisteredInEventPhaseMap guards the reason this mapping exists:
// downstream tools resolve a hook event to a phase by AGENT_ENV alias. Before
// Goose implemented LifecycleEventMapper, "goose" was absent from that map and
// resolution fell through to a scan of every other agent.
func TestGooseRegisteredInEventPhaseMap(t *testing.T) {
	byAlias := agentx.BuildEventPhaseMap()

	phases, ok := byAlias["goose"]
	require.True(t, ok, "AGENT_ENV=goose must resolve to a phase map")
	assert.Equal(t, agentx.PhaseStart, phases[agentx.HookEventSessionStart])
	assert.Equal(t, agentx.PhaseEnd, phases[agentx.HookEventSessionEnd])
	assert.Equal(t, agentx.PhasePrompt, phases[agentx.HookEventUserPromptSubmit])
}
