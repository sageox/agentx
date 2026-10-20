package agents

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sageox/agentx"
)

func TestAmpDetect(t *testing.T) {
	ctx := context.Background()
	agent := NewAmpAgent()

	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{"AMP=1", map[string]string{"AMP": "1"}, true},
		{"AMP_AGENT=1", map[string]string{"AMP_AGENT": "1"}, true},
		{"AMP_THREAD_URL set", map[string]string{"AMP_THREAD_URL": "https://ampcode.com/threads/abc123"}, true},
		{"AGENT_ENV=amp", map[string]string{"AGENT_ENV": "amp"}, true},
		{"AMP_ORB=1 (inside an orb)", map[string]string{"AMP_ORB": "1"}, true},
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

// AMP_ORB and AMP_THREAD_ID are inherited by everything under an orb's
// thread, so they must stay out of DetectRuntime: a runtime match outranks
// every other agent's, and Claude Code run inside an orb would otherwise be
// reported as Amp (sageox/agentx#17).
func TestAmpOrbIsAVenueSignalNotARuntimeSignal(t *testing.T) {
	ctx := context.Background()
	agent := NewAmpAgent()

	tests := []struct {
		name        string
		envVars     map[string]string
		wantRuntime bool
		wantVenue   bool
	}{
		{"orb only", map[string]string{"AMP_ORB": "1"}, false, true},
		{"orb with thread URL", map[string]string{"AMP_ORB": "1", "AMP_THREAD_URL": "https://ampcode.com/threads/T-1"}, true, true},
		{"real orb agent shell (thread id, no URL)", map[string]string{"AMP_ORB": "1", "AMP_THREAD_ID": "T-1", "AMP_URL": "https://ampcode.com"}, false, true},
		{"thread id alone", map[string]string{"AMP_THREAD_ID": "T-1"}, false, true},
		{"thread URL only (laptop)", map[string]string{"AMP_THREAD_URL": "https://ampcode.com/threads/T-1"}, true, false},
		{"empty AMP_ORB", map[string]string{"AMP_ORB": ""}, false, false},
		{"AGENT_ENV=amp is neither", map[string]string{"AGENT_ENV": "amp"}, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := agentx.NewMockEnvironment(tt.envVars)
			runtime, err := agent.DetectRuntime(ctx, env)
			require.NoError(t, err)
			assert.Equal(t, tt.wantRuntime, runtime, "DetectRuntime")
			venue, err := agent.DetectVenue(ctx, env)
			require.NoError(t, err)
			assert.Equal(t, tt.wantVenue, venue, "DetectVenue")
		})
	}
}

func TestAmpMetadata(t *testing.T) {
	agent := NewAmpAgent()

	assert.Equal(t, agentx.AgentTypeAmp, agent.Type())
	assert.Equal(t, "Amp", agent.Name())
	assert.Equal(t, "https://ampcode.com", agent.URL())
	assert.Equal(t, agentx.RoleAgent, agent.Role())
	assert.True(t, agent.SupportsXDGConfig())
	assert.Contains(t, agent.ContextFiles(), "AGENTS.md")
}

func TestAmpCapabilities(t *testing.T) {
	agent := NewAmpAgent()
	caps := agent.Capabilities()

	assert.True(t, caps.MCPServers)
	assert.True(t, caps.SystemPrompt)
	assert.True(t, caps.ProjectContext)
	assert.True(t, caps.CustomCommands)
}

func TestAmpUserConfigPath(t *testing.T) {
	agent := NewAmpAgent()

	env := &agentx.MockEnvironment{
		Config: "/home/test/.config",
	}
	path, err := agent.UserConfigPath(env)
	require.NoError(t, err)
	assert.Equal(t, "/home/test/.config/amp", path)
}

func TestAmpDetectVersion(t *testing.T) {
	ctx := context.Background()
	agent := NewAmpAgent()

	t.Run("detects version from cli", func(t *testing.T) {
		env := &agentx.MockEnvironment{
			ExecOutputs: map[string][]byte{
				"amp": []byte("0.3.1\n"),
			},
		}
		assert.Equal(t, "0.3.1", agent.DetectVersion(ctx, env))
	})

	t.Run("binary not found", func(t *testing.T) {
		env := &agentx.MockEnvironment{}
		assert.Equal(t, "", agent.DetectVersion(ctx, env))
	})
}

func TestAmpIsInstalled(t *testing.T) {
	ctx := context.Background()
	agent := NewAmpAgent()

	t.Run("found in PATH", func(t *testing.T) {
		env := &agentx.MockEnvironment{
			PathBinaries: map[string]string{"amp": "/usr/local/bin/amp"},
		}
		installed, err := agent.IsInstalled(ctx, env)
		require.NoError(t, err)
		assert.True(t, installed)
	})

	t.Run("config dir exists", func(t *testing.T) {
		env := &agentx.MockEnvironment{
			Config:       "/home/test/.config",
			ExistingDirs: map[string]bool{"/home/test/.config/amp": true},
		}
		installed, err := agent.IsInstalled(ctx, env)
		require.NoError(t, err)
		assert.True(t, installed)
	})

	t.Run("not installed", func(t *testing.T) {
		env := &agentx.MockEnvironment{
			Config: "/home/test/.config",
		}
		installed, err := agent.IsInstalled(ctx, env)
		require.NoError(t, err)
		assert.False(t, installed)
	})
}

func TestAmpSessionID(t *testing.T) {
	agent := NewAmpAgent()
	assert.True(t, agent.SupportsSession())

	t.Run("returns thread URL from env var", func(t *testing.T) {
		env := agentx.NewMockEnvironment(map[string]string{
			"AMP_THREAD_URL": "https://ampcode.com/threads/abc123",
		})
		assert.Equal(t, "https://ampcode.com/threads/abc123", agent.SessionID(env))
	})

	t.Run("returns empty when env var not set", func(t *testing.T) {
		env := agentx.NewMockEnvironment(nil)
		assert.Equal(t, "", agent.SessionID(env))
	})

	// An orb's agent shell has AMP_THREAD_ID and AMP_URL but no
	// AMP_THREAD_URL (sageox/agentx#17). The rebuilt URL must match what
	// the ox-bridge plugin sends at session end, or the recording is never
	// found and never uploads.
	orb := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"orb: id and AMP_URL", map[string]string{"AMP_THREAD_ID": "T-01a1", "AMP_URL": "https://ampcode.com"}, "https://ampcode.com/threads/T-01a1"},
		{"orb: AMP_URL with trailing slash", map[string]string{"AMP_THREAD_ID": "T-01a1", "AMP_URL": "https://ampcode.com/"}, "https://ampcode.com/threads/T-01a1"},
		{"orb: no AMP_URL falls back to ampcode.com", map[string]string{"AMP_THREAD_ID": "T-01a1"}, "https://ampcode.com/threads/T-01a1"},
		{"self-hosted Amp server", map[string]string{"AMP_THREAD_ID": "T-9", "AMP_URL": "https://amp.example.test"}, "https://amp.example.test/threads/T-9"},
		{"thread URL wins over id", map[string]string{"AMP_THREAD_URL": "https://ampcode.com/threads/T-url", "AMP_THREAD_ID": "T-id"}, "https://ampcode.com/threads/T-url"},
		{"id with a slash is refused", map[string]string{"AMP_THREAD_ID": "T-1/../x"}, ""},
		{"blank id", map[string]string{"AMP_THREAD_ID": "  "}, ""},
	}
	for _, tt := range orb {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, agent.SessionID(agentx.NewMockEnvironment(tt.env)))
		})
	}
}
