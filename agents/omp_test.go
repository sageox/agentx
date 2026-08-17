package agents

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sageox/agentx"
)

func TestOMPDetect(t *testing.T) {
	ctx := context.Background()
	agent := NewOMPAgent()

	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{"AGENT_ENV=omp", map[string]string{"AGENT_ENV": "omp"}, true},
		{"AGENT_ENV=oh-my-pi", map[string]string{"AGENT_ENV": "oh-my-pi"}, true},
		{"exec path is omp binary", map[string]string{"_": "/usr/local/bin/omp"}, true},
		{"exec path contains oh-my-pi", map[string]string{"_": "/home/user/.npm/bin/oh-my-pi"}, true},
		// OMP inherits Pi's runtime markers from shared cli.ts ancestry; those
		// identify Pi, so OMP must NOT claim them (mirrors ox-adapter-omp).
		{"PI_CODING_AGENT=true is Pi, not OMP", map[string]string{"PI_CODING_AGENT": "true"}, false},
		{"PI_CODING_AGENT_DIR is Pi, not OMP", map[string]string{"PI_CODING_AGENT_DIR": "/custom/pi"}, false},
		{"AGENT_ENV=pi is Pi, not OMP", map[string]string{"AGENT_ENV": "pi"}, false},
		{"bare pi binary should not match", map[string]string{"_": "/usr/local/bin/pi"}, false},
		{"substring compose should not match", map[string]string{"_": "/usr/local/bin/compose"}, false},
		{"no env vars", map[string]string{}, false},
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

// TestOMPPiMutualExclusion locks in the core disambiguation contract: OMP and
// Pi runtime detectors must never both fire for the same environment, because
// the registry is a map with no ordering guarantee to break a tie. OMP inherits
// Pi's PI_CODING_AGENT marker, so the split turns on the launching binary.
func TestOMPPiMutualExclusion(t *testing.T) {
	ctx := context.Background()
	omp := NewOMPAgent()
	pi := NewPiAgent()

	cases := []struct {
		name            string
		env             map[string]string
		wantOMP, wantPi bool
	}{
		{
			name:    "genuine Pi: pi-coding-agent launcher + PI_CODING_AGENT",
			env:     map[string]string{"_": "/usr/local/bin/pi-coding-agent", "PI_CODING_AGENT": "true"},
			wantOMP: false, wantPi: true,
		},
		{
			// The pathological overlap: omp launcher AND inherited PI_CODING_AGENT.
			// Pi must yield so exactly one detector fires regardless of map order.
			name:    "OMP fork: omp launcher despite inherited PI_CODING_AGENT",
			env:     map[string]string{"_": "/usr/local/bin/omp", "PI_CODING_AGENT": "true"},
			wantOMP: true, wantPi: false,
		},
		{
			name:    "OMP fork: oh-my-pi launcher",
			env:     map[string]string{"_": "/home/user/.npm/bin/oh-my-pi", "PI_CODING_AGENT": "1"},
			wantOMP: true, wantPi: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := agentx.NewMockEnvironment(tc.env)

			gotOMP, err := omp.DetectRuntime(ctx, env)
			require.NoError(t, err)
			gotPi, err := pi.DetectRuntime(ctx, env)
			require.NoError(t, err)

			assert.Equal(t, tc.wantOMP, gotOMP, "OMP.DetectRuntime")
			assert.Equal(t, tc.wantPi, gotPi, "Pi.DetectRuntime")
			assert.False(t, gotOMP && gotPi, "OMP and Pi must be mutually exclusive")
		})
	}
}

func TestOMPMetadata(t *testing.T) {
	agent := NewOMPAgent()

	assert.Equal(t, agentx.AgentTypeOMP, agent.Type())
	assert.Equal(t, "OMP", agent.Name())
	assert.Equal(t, "https://github.com/can1357/oh-my-pi", agent.URL())
	assert.Equal(t, agentx.RoleAgent, agent.Role())
	assert.False(t, agent.SupportsXDGConfig())
	assert.Equal(t, ".omp", agent.ProjectConfigPath())
	assert.Contains(t, agent.ContextFiles(), "AGENTS.md")
	assert.Contains(t, agent.ContextFiles(), "CLAUDE.md")
	assert.Contains(t, agent.ContextFiles(), "SYSTEM.md")
	assert.Equal(t, []string{"omp", "oh-my-pi"}, agent.AgentENVAliases())
}

// TestOMPIsNotLifecycleAgent locks in that OMP is tail-mode: the SageOx
// integration drives it via .omp/AGENTS.md, not an extension-bridge hook.
func TestOMPIsNotLifecycleAgent(t *testing.T) {
	agent := NewOMPAgent()
	_, isLifecycle := any(agent).(agentx.LifecycleEventMapper)
	assert.False(t, isLifecycle, "OMP has no lifecycle hooks")
	assert.False(t, agent.Capabilities().Hooks)
}

func TestOMPCapabilities(t *testing.T) {
	agent := NewOMPAgent()
	caps := agent.Capabilities()

	assert.False(t, caps.Hooks)
	assert.False(t, caps.CustomCommands)
	assert.True(t, caps.SystemPrompt)
	assert.True(t, caps.ProjectContext)
}

func TestOMPUserConfigPath(t *testing.T) {
	agent := NewOMPAgent()

	env := &agentx.MockEnvironment{
		Home: "/home/test",
	}
	path, err := agent.UserConfigPath(env)
	require.NoError(t, err)
	assert.Equal(t, "/home/test/.omp", path)
}

func TestOMPDetectVersion(t *testing.T) {
	ctx := context.Background()
	agent := NewOMPAgent()

	t.Run("detects version from cli", func(t *testing.T) {
		env := &agentx.MockEnvironment{
			ExecOutputs: map[string][]byte{
				"omp": []byte("omp 17.3.5\n"),
			},
		}
		assert.Equal(t, "17.3.5", agent.DetectVersion(ctx, env))
	})

	t.Run("binary not found", func(t *testing.T) {
		env := &agentx.MockEnvironment{}
		assert.Equal(t, "", agent.DetectVersion(ctx, env))
	})
}

func TestOMPIsInstalled(t *testing.T) {
	ctx := context.Background()
	agent := NewOMPAgent()

	t.Run("found in PATH", func(t *testing.T) {
		env := &agentx.MockEnvironment{
			PathBinaries: map[string]string{"omp": "/usr/local/bin/omp"},
		}
		installed, err := agent.IsInstalled(ctx, env)
		require.NoError(t, err)
		assert.True(t, installed)
	})

	t.Run("config dir exists", func(t *testing.T) {
		env := &agentx.MockEnvironment{
			Home:         "/home/test",
			ExistingDirs: map[string]bool{"/home/test/.omp": true},
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

func TestOMPSession(t *testing.T) {
	agent := NewOMPAgent()
	assert.True(t, agent.SupportsSession())

	emptyEnv := agentx.NewMockEnvironment(nil)
	assert.Equal(t, "", agent.SessionID(emptyEnv))

	ompEnv := agentx.NewMockEnvironment(map[string]string{"OMP_SESSION_ID": "sess_omp123"})
	assert.Equal(t, "sess_omp123", agent.SessionID(ompEnv))

	// Falls back to Pi's variable when OMP's own is unset (shared ancestry).
	piEnv := agentx.NewMockEnvironment(map[string]string{"PI_SESSION_ID": "sess_pi456"})
	assert.Equal(t, "sess_pi456", agent.SessionID(piEnv))
}

// TestOMPRegisteredInDefaultRegistry confirms OMP is discoverable by type via
// the default registry (order is irrelevant — see TestOMPPiMutualExclusion).
func TestOMPRegisteredInDefaultRegistry(t *testing.T) {
	agent, ok := agentx.DefaultRegistry.Get(agentx.AgentTypeOMP)
	require.True(t, ok, "OMP must be registered in the default registry")
	assert.Equal(t, "OMP", agent.Name())
}
