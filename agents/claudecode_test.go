package agents

import (
	"context"
	"testing"

	"github.com/sageox/agentx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaudeCodeEventPhases(t *testing.T) {
	agent := NewClaudeCodeAgent()
	phases := agent.EventPhases()

	// all 7 Claude Code lifecycle events should be mapped
	assert.Len(t, phases, 7)

	assert.Equal(t, agentx.PhaseStart, phases[agentx.HookEventSessionStart])
	assert.Equal(t, agentx.PhaseEnd, phases[agentx.HookEventSessionEnd])
	assert.Equal(t, agentx.PhaseBeforeTool, phases[agentx.HookEventPreToolUse])
	assert.Equal(t, agentx.PhaseAfterTool, phases[agentx.HookEventPostToolUse])
	assert.Equal(t, agentx.PhasePrompt, phases[agentx.HookEventUserPromptSubmit])
	assert.Equal(t, agentx.PhaseStop, phases[agentx.HookEventStop])
	assert.Equal(t, agentx.PhaseCompact, phases[agentx.HookEventPreCompact])
}

func TestClaudeCodeAgentENVAliases(t *testing.T) {
	agent := NewClaudeCodeAgent()
	aliases := agent.AgentENVAliases()

	assert.Contains(t, aliases, "claude-code")
	assert.Contains(t, aliases, "claudecode")
	assert.Contains(t, aliases, "claude")
	assert.Equal(t, "claude-code", aliases[0], "first alias should be canonical")
}

func TestClaudeCodeImplementsLifecycleEventMapper(t *testing.T) {
	var _ agentx.LifecycleEventMapper = (*ClaudeCodeAgent)(nil)
}

func TestClaudeCodeSessionID(t *testing.T) {
	agent := NewClaudeCodeAgent()
	assert.True(t, agent.SupportsSession())

	t.Run("returns session ID from env var", func(t *testing.T) {
		env := agentx.NewMockEnvironment(map[string]string{
			"CLAUDE_CODE_SESSION_ID": "sess_abc123",
		})
		assert.Equal(t, "sess_abc123", agent.SessionID(env))
	})

	t.Run("returns empty when env var not set", func(t *testing.T) {
		env := agentx.NewMockEnvironment(nil)
		assert.Equal(t, "", agent.SessionID(env))
	})
}

// TestClaudeCodeImplementsRuntimeDetector confirms Claude Code exposes
// runtime-only detection so the two-phase Detector can prefer it over
// AGENT_ENV overrides (#527).
func TestClaudeCodeImplementsRuntimeDetector(t *testing.T) {
	var _ agentx.RuntimeDetector = (*ClaudeCodeAgent)(nil)
}

// TestClaudeCodeDetectRuntime_SkipsAgentEnv is the per-agent half of the
// #527 guard: DetectRuntime must NOT claim Claude Code merely because
// AGENT_ENV=claude-code is set — that's an override, not a runtime signal.
// Runtime detection requires a signal Claude Code itself sets.
//
// Failure prevented: DetectRuntime returning true on AGENT_ENV would
// collapse the two-phase priority back to the pre-fix shape, where
// registration order silently decided which agent claimed a session.
func TestClaudeCodeDetectRuntime_SkipsAgentEnv(t *testing.T) {
	agent := NewClaudeCodeAgent()

	// AGENT_ENV alone must not trigger DetectRuntime
	env := agentx.NewMockEnvironment(map[string]string{
		"AGENT_ENV": "claude-code",
	})
	detected, err := agent.DetectRuntime(context.Background(), env)
	require.NoError(t, err)
	assert.False(t, detected, "DetectRuntime must ignore AGENT_ENV; that's the override path")

	// But a real runtime signal (CLAUDECODE=1) must be honored
	envRuntime := agentx.NewMockEnvironment(map[string]string{
		"CLAUDECODE": "1",
	})
	detected, err = agent.DetectRuntime(context.Background(), envRuntime)
	require.NoError(t, err)
	assert.True(t, detected, "DetectRuntime must honor CLAUDECODE=1")
}
