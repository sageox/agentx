package agents

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sageox/agentx"
)

func TestCodexDetect(t *testing.T) {
	ctx := context.Background()
	agent := NewCodexAgent()

	tests := []struct {
		name     string
		env      *agentx.MockEnvironment
		expected bool
	}{
		{
			name:     "AGENT_ENV=codex",
			env:      agentx.NewMockEnvironment(map[string]string{"AGENT_ENV": "codex"}),
			expected: true,
		},
		{
			name:     "CODEX_CI set",
			env:      agentx.NewMockEnvironment(map[string]string{"CODEX_CI": "1"}),
			expected: true,
		},
		{
			name:     "CODEX_SANDBOX set",
			env:      agentx.NewMockEnvironment(map[string]string{"CODEX_SANDBOX": "workspace-write"}),
			expected: true,
		},
		{
			name:     "CODEX_THREAD_ID set",
			env:      agentx.NewMockEnvironment(map[string]string{"CODEX_THREAD_ID": "thread_123"}),
			expected: true,
		},
		{
			name: "project .codex directory",
			env: &agentx.MockEnvironment{
				ExistingDirs: map[string]bool{".codex": true},
			},
			expected: true,
		},
		{
			name: "PWD .codex directory",
			env: &agentx.MockEnvironment{
				EnvVars:      map[string]string{"PWD": "/repo"},
				ExistingDirs: map[string]bool{"/repo/.codex": true},
			},
			expected: true,
		},
		{
			// CodeRabbit review on #9: a non-"codex" AGENT_ENV must not
			// suppress genuine Codex runtime signals. Previously Detect
			// returned false here, hiding the CODEX_CI / CODEX_THREAD_ID
			// evidence from DetectByType callers that bypass the registry's
			// two-phase flow. Post-fix, Codex honors its runtime evidence
			// regardless of a foreign AGENT_ENV; cross-agent priority is
			// resolved at the Detector layer.
			name: "AGENT_ENV non-codex does NOT hide Codex runtime signals",
			env: agentx.NewMockEnvironment(map[string]string{
				"AGENT_ENV":       "claude-code",
				"CODEX_CI":        "1",
				"CODEX_THREAD_ID": "thread_123",
			}),
			expected: true,
		},
		{
			name:     "no signals",
			env:      agentx.NewMockEnvironment(map[string]string{}),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detected, err := agent.Detect(ctx, tt.env)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, detected)
		})
	}
}

func TestCodexMetadata(t *testing.T) {
	agent := NewCodexAgent()

	assert.Equal(t, agentx.AgentTypeCodex, agent.Type())
	assert.Equal(t, "Codex", agent.Name())
	assert.Equal(t, "https://github.com/openai/codex", agent.URL())
	assert.Equal(t, agentx.RoleAgent, agent.Role())
	assert.False(t, agent.SupportsXDGConfig())
	assert.Contains(t, agent.ContextFiles(), "AGENTS.md")
}

func TestCodexCapabilities(t *testing.T) {
	agent := NewCodexAgent()
	caps := agent.Capabilities()

	assert.False(t, caps.Hooks)
	assert.True(t, caps.ProjectContext)
	assert.True(t, caps.SystemPrompt)
}

func TestCodexUserConfigPath(t *testing.T) {
	agent := NewCodexAgent()
	env := &agentx.MockEnvironment{Home: "/home/test"}

	path, err := agent.UserConfigPath(env)
	require.NoError(t, err)
	assert.Equal(t, "/home/test/.codex", path)
}

func TestCodexIsInstalled(t *testing.T) {
	ctx := context.Background()
	agent := NewCodexAgent()

	t.Run("found in PATH", func(t *testing.T) {
		env := &agentx.MockEnvironment{
			PathBinaries: map[string]string{"codex": "/usr/local/bin/codex"},
		}
		installed, err := agent.IsInstalled(ctx, env)
		require.NoError(t, err)
		assert.True(t, installed)
	})

	t.Run("config dir exists", func(t *testing.T) {
		env := &agentx.MockEnvironment{
			Home:         "/home/test",
			ExistingDirs: map[string]bool{"/home/test/.codex": true},
		}
		installed, err := agent.IsInstalled(ctx, env)
		require.NoError(t, err)
		assert.True(t, installed)
	})

	t.Run("not installed", func(t *testing.T) {
		env := &agentx.MockEnvironment{
			Home: "/home/test",
		}
		installed, err := agent.IsInstalled(ctx, env)
		require.NoError(t, err)
		assert.False(t, installed)
	})
}

func TestCodexSessionID(t *testing.T) {
	agent := NewCodexAgent()
	assert.True(t, agent.SupportsSession())

	t.Run("returns thread ID from env var", func(t *testing.T) {
		env := agentx.NewMockEnvironment(map[string]string{
			"CODEX_THREAD_ID": "thread_xyz789",
		})
		assert.Equal(t, "thread_xyz789", agent.SessionID(env))
	})

	t.Run("returns empty when env var not set", func(t *testing.T) {
		env := agentx.NewMockEnvironment(nil)
		assert.Equal(t, "", agent.SessionID(env))
	})
}
