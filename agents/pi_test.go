package agents

import (
	"context"
	"testing"

	"github.com/sageox/agentx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPiDetect(t *testing.T) {
	ctx := context.Background()
	agent := NewPiAgent()

	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{"PI_CODING_AGENT_DIR set", map[string]string{"PI_CODING_AGENT_DIR": "/custom/pi"}, true},
		{"AGENT_ENV=pi", map[string]string{"AGENT_ENV": "pi"}, true},
		{"exec path heuristic", map[string]string{"_": "/usr/local/bin/pi-coding-agent"}, true},
		{"exec path heuristic npm global", map[string]string{"_": "/home/user/.npm/bin/pi-coding-agent"}, true},
		{"bare pi binary should not match", map[string]string{"_": "/usr/local/bin/pi"}, false},
		{"no env vars", map[string]string{}, false},
		{"unrelated env", map[string]string{"CURSOR_AGENT": "1"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := agentx.NewMockEnvironment(tt.envVars)
			detected, err := agent.Detect(ctx, env)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, detected)
		})
	}
}

func TestPiMetadata(t *testing.T) {
	agent := NewPiAgent()

	assert.Equal(t, agentx.AgentTypePi, agent.Type())
	assert.Equal(t, "Pi", agent.Name())
	assert.Equal(t, "https://shittycodingagent.ai/", agent.URL())
	assert.Equal(t, agentx.RoleAgent, agent.Role())
	assert.False(t, agent.SupportsXDGConfig())
	assert.Equal(t, ".pi", agent.ProjectConfigPath())
	assert.Contains(t, agent.ContextFiles(), "AGENTS.md")
	assert.Contains(t, agent.ContextFiles(), "CLAUDE.md")
	assert.Contains(t, agent.ContextFiles(), "SYSTEM.md")
}

func TestPiCapabilities(t *testing.T) {
	agent := NewPiAgent()
	caps := agent.Capabilities()

	assert.False(t, caps.Hooks)
	assert.True(t, caps.MCPServers)
	assert.True(t, caps.SystemPrompt)
	assert.True(t, caps.ProjectContext)
	assert.False(t, caps.CustomCommands)
}

func TestPiUserConfigPath(t *testing.T) {
	agent := NewPiAgent()

	env := &agentx.MockEnvironment{
		Home: "/home/test",
	}
	path, err := agent.UserConfigPath(env)
	require.NoError(t, err)
	assert.Equal(t, "/home/test/.pi", path)
}

func TestPiDetectVersion(t *testing.T) {
	ctx := context.Background()
	agent := NewPiAgent()

	t.Run("detects version from cli", func(t *testing.T) {
		env := &agentx.MockEnvironment{
			ExecOutputs: map[string][]byte{
				"pi": []byte("pi-coding-agent 1.2.3\n"),
			},
		}
		assert.Equal(t, "1.2.3", agent.DetectVersion(ctx, env))
	})

	t.Run("binary not found", func(t *testing.T) {
		env := &agentx.MockEnvironment{}
		assert.Equal(t, "", agent.DetectVersion(ctx, env))
	})
}

func TestPiIsInstalled(t *testing.T) {
	ctx := context.Background()
	agent := NewPiAgent()

	t.Run("found in PATH", func(t *testing.T) {
		env := &agentx.MockEnvironment{
			PathBinaries: map[string]string{"pi": "/usr/local/bin/pi"},
		}
		installed, err := agent.IsInstalled(ctx, env)
		require.NoError(t, err)
		assert.True(t, installed)
	})

	t.Run("config dir exists", func(t *testing.T) {
		env := &agentx.MockEnvironment{
			Home:         "/home/test",
			ExistingDirs: map[string]bool{"/home/test/.pi": true},
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

func TestPiSession(t *testing.T) {
	agent := NewPiAgent()
	assert.False(t, agent.SupportsSession())

	env := agentx.NewMockEnvironment(nil)
	assert.Equal(t, "", agent.SessionID(env))
}
