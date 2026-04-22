package agentx

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDetectableAgent extends mockAgent with configurable detect behavior.
type mockDetectableAgent struct {
	agentType    AgentType
	name         string
	role         AgentRole
	detectFn     func(ctx context.Context, env Environment) (bool, error)
	capabilities Capabilities
}

func (a *mockDetectableAgent) Type() AgentType { return a.agentType }
func (a *mockDetectableAgent) Name() string    { return a.name }
func (a *mockDetectableAgent) URL() string     { return "" }
func (a *mockDetectableAgent) Role() AgentRole {
	if a.role != "" {
		return a.role
	}
	return RoleAgent
}
func (a *mockDetectableAgent) Detect(ctx context.Context, env Environment) (bool, error) {
	if a.detectFn != nil {
		return a.detectFn(ctx, env)
	}
	return false, nil
}
func (a *mockDetectableAgent) UserConfigPath(_ Environment) (string, error) { return "", nil }
func (a *mockDetectableAgent) ProjectConfigPath() string                    { return "" }
func (a *mockDetectableAgent) ContextFiles() []string                       { return nil }
func (a *mockDetectableAgent) SupportsXDGConfig() bool                      { return false }
func (a *mockDetectableAgent) Capabilities() Capabilities                   { return a.capabilities }
func (a *mockDetectableAgent) HookManager() HookManager                     { return nil }
func (a *mockDetectableAgent) CommandManager() CommandManager               { return nil }
func (a *mockDetectableAgent) RulesManager() RulesManager                   { return nil }
func (a *mockDetectableAgent) IsInstalled(_ context.Context, _ Environment) (bool, error) {
	return false, nil
}
func (a *mockDetectableAgent) DetectVersion(_ context.Context, _ Environment) string { return "" }
func (a *mockDetectableAgent) SupportsSession() bool                                 { return false }
func (a *mockDetectableAgent) SessionID(_ Environment) string                        { return "" }

// mockLifecycleDetectableAgent combines mockDetectableAgent with LifecycleEventMapper.
type mockLifecycleDetectableAgent struct {
	mockDetectableAgent
	phases  EventPhaseMap
	aliases []string
}

func (a *mockLifecycleDetectableAgent) EventPhases() EventPhaseMap { return a.phases }
func (a *mockLifecycleDetectableAgent) AgentENVAliases() []string  { return a.aliases }

// mockRuntimeDetectableAgent adds RuntimeDetector to mockDetectableAgent.
// Used to validate the two-phase detection priority enforced by the
// Detector — runtime signals outrank AGENT_ENV overrides (#527).
type mockRuntimeDetectableAgent struct {
	mockDetectableAgent
	detectRuntimeFn func(ctx context.Context, env Environment) (bool, error)
}

func (a *mockRuntimeDetectableAgent) DetectRuntime(ctx context.Context, env Environment) (bool, error) {
	if a.detectRuntimeFn != nil {
		return a.detectRuntimeFn(ctx, env)
	}
	return false, nil
}

// --- DetectAll tests ---

func TestDetectAll(t *testing.T) {
	reg := NewRegistry().(*registry)
	reg.Register(&mockDetectableAgent{
		agentType: AgentTypeClaudeCode,
		name:      "Agent A",
		detectFn:  func(_ context.Context, _ Environment) (bool, error) { return true, nil },
	})
	reg.Register(&mockDetectableAgent{
		agentType: AgentTypeCursor,
		name:      "Agent B",
		detectFn:  func(_ context.Context, _ Environment) (bool, error) { return true, nil },
	})
	reg.Register(&mockDetectableAgent{
		agentType: AgentTypeWindsurf,
		name:      "Agent C",
		detectFn:  func(_ context.Context, _ Environment) (bool, error) { return false, nil },
	})

	d := reg.Detector()
	ctx := context.Background()
	detected, err := d.DetectAll(ctx)
	require.NoError(t, err)
	assert.Len(t, detected, 2)

	types := make(map[AgentType]bool)
	for _, a := range detected {
		types[a.Type()] = true
	}
	assert.True(t, types[AgentTypeClaudeCode])
	assert.True(t, types[AgentTypeCursor])
	assert.False(t, types[AgentTypeWindsurf])
}

func TestDetectAll_Empty(t *testing.T) {
	reg := NewRegistry().(*registry)
	d := reg.Detector()

	detected, err := d.DetectAll(context.Background())
	require.NoError(t, err)
	assert.Empty(t, detected)
}

func TestDetectAll_WithErrors(t *testing.T) {
	reg := NewRegistry().(*registry)
	reg.Register(&mockDetectableAgent{
		agentType: AgentTypeClaudeCode,
		name:      "Good Agent",
		detectFn:  func(_ context.Context, _ Environment) (bool, error) { return true, nil },
	})
	reg.Register(&mockDetectableAgent{
		agentType: AgentTypeCursor,
		name:      "Error Agent",
		detectFn:  func(_ context.Context, _ Environment) (bool, error) { return false, errors.New("detect failed") },
	})

	d := reg.Detector()
	detected, err := d.DetectAll(context.Background())
	require.NoError(t, err)
	// error agents are skipped, only the good one remains
	assert.Len(t, detected, 1)
	assert.Equal(t, AgentTypeClaudeCode, detected[0].Type())
}

func TestDetectAll_IncludesBothRoles(t *testing.T) {
	reg := NewRegistry().(*registry)
	reg.Register(&mockDetectableAgent{
		agentType: AgentTypeClaudeCode,
		name:      "Agent",
		role:      RoleAgent,
		detectFn:  func(_ context.Context, _ Environment) (bool, error) { return true, nil },
	})
	reg.Register(&mockDetectableAgent{
		agentType: AgentTypeConductor,
		name:      "Orchestrator",
		role:      RoleOrchestrator,
		detectFn:  func(_ context.Context, _ Environment) (bool, error) { return true, nil },
	})

	d := reg.Detector()
	detected, err := d.DetectAll(context.Background())
	require.NoError(t, err)
	// DetectAll does not filter by role
	assert.Len(t, detected, 2)
}

// --- DetectByType tests ---

func TestDetectByType(t *testing.T) {
	reg := NewRegistry().(*registry)
	reg.Register(&mockDetectableAgent{
		agentType: AgentTypeClaudeCode,
		name:      "Agent",
		detectFn:  func(_ context.Context, _ Environment) (bool, error) { return true, nil },
	})

	d := reg.Detector()
	ok, err := d.DetectByType(context.Background(), AgentTypeClaudeCode)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestDetectByType_NotRegistered(t *testing.T) {
	reg := NewRegistry().(*registry)
	d := reg.Detector()

	ok, err := d.DetectByType(context.Background(), AgentTypeCursor)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not registered")
	assert.False(t, ok)
}

func TestDetectByType_NotDetected(t *testing.T) {
	reg := NewRegistry().(*registry)
	reg.Register(&mockDetectableAgent{
		agentType: AgentTypeClaudeCode,
		name:      "Agent",
		detectFn:  func(_ context.Context, _ Environment) (bool, error) { return false, nil },
	})

	d := reg.Detector()
	ok, err := d.DetectByType(context.Background(), AgentTypeClaudeCode)
	require.NoError(t, err)
	assert.False(t, ok)
}

// --- BuildEventPhaseMap tests (using private registry methods) ---

func TestBuildEventPhaseMap_AliasKeys(t *testing.T) {
	reg := NewRegistry().(*registry)
	reg.Register(&mockLifecycleDetectableAgent{
		mockDetectableAgent: mockDetectableAgent{
			agentType: AgentTypeClaudeCode,
			name:      "Agent",
		},
		phases: EventPhaseMap{
			HookEventSessionStart: PhaseStart,
			HookEventPostToolUse:  PhaseAfterTool,
		},
		aliases: []string{"claude-code", "claudecode"},
	})

	result := reg.BuildEventPhaseMap()
	// keys should be aliases, not agent type slug
	assert.Len(t, result, 2)
	assert.Contains(t, result, "claude-code")
	assert.Contains(t, result, "claudecode")
	assert.Equal(t, PhaseStart, result["claude-code"][HookEventSessionStart])
	assert.Equal(t, PhaseAfterTool, result["claudecode"][HookEventPostToolUse])
}

func TestBuildEventPhaseMap_NoMappers(t *testing.T) {
	reg := NewRegistry().(*registry)
	// register a plain agent without LifecycleEventMapper
	reg.Register(&mockDetectableAgent{
		agentType: AgentTypeCursor,
		name:      "Plain Agent",
	})

	result := reg.BuildEventPhaseMap()
	assert.Empty(t, result)
}

func TestBuildEventPhaseMap_MultipleAgents(t *testing.T) {
	reg := NewRegistry().(*registry)
	reg.Register(&mockLifecycleDetectableAgent{
		mockDetectableAgent: mockDetectableAgent{agentType: AgentTypeClaudeCode, name: "A"},
		phases:              EventPhaseMap{HookEventSessionStart: PhaseStart},
		aliases:             []string{"claude-code"},
	})
	reg.Register(&mockLifecycleDetectableAgent{
		mockDetectableAgent: mockDetectableAgent{agentType: AgentTypeWindsurf, name: "B"},
		phases:              EventPhaseMap{WindsurfEventPreReadCode: PhaseBeforeTool},
		aliases:             []string{"windsurf"},
	})

	result := reg.BuildEventPhaseMap()
	assert.Len(t, result, 2)
	assert.Equal(t, PhaseStart, result["claude-code"][HookEventSessionStart])
	assert.Equal(t, PhaseBeforeTool, result["windsurf"][WindsurfEventPreReadCode])
}

// --- ResolveAgentENV tests (using private registry methods) ---

func TestResolveAgentENV_Match(t *testing.T) {
	reg := NewRegistry().(*registry)
	reg.Register(&mockLifecycleDetectableAgent{
		mockDetectableAgent: mockDetectableAgent{agentType: AgentTypeClaudeCode, name: "Agent"},
		aliases:             []string{"claude-code", "claude"},
	})

	assert.Equal(t, AgentTypeClaudeCode, reg.ResolveAgentENV("claude-code"))
	assert.Equal(t, AgentTypeClaudeCode, reg.ResolveAgentENV("claude"))
}

func TestResolveAgentENV_Unknown(t *testing.T) {
	reg := NewRegistry().(*registry)
	reg.Register(&mockLifecycleDetectableAgent{
		mockDetectableAgent: mockDetectableAgent{agentType: AgentTypeClaudeCode, name: "Agent"},
		aliases:             []string{"claude-code"},
	})

	assert.Equal(t, AgentTypeUnknown, reg.ResolveAgentENV("unknown"))
	assert.Equal(t, AgentTypeUnknown, reg.ResolveAgentENV(""))
}

func TestResolveAgentENV_NoLifecycleAgents(t *testing.T) {
	reg := NewRegistry().(*registry)
	reg.Register(&mockDetectableAgent{
		agentType: AgentTypeCursor,
		name:      "Plain",
	})

	assert.Equal(t, AgentTypeUnknown, reg.ResolveAgentENV("cursor"))
}

// --- HookSupportMatrix tests (using private registry methods) ---

func TestHookSupportMatrix_SingleAgent(t *testing.T) {
	reg := NewRegistry().(*registry)
	reg.Register(&mockLifecycleDetectableAgent{
		mockDetectableAgent: mockDetectableAgent{agentType: AgentTypeClaudeCode, name: "Agent"},
		phases: EventPhaseMap{
			HookEventSessionStart: PhaseStart,
			HookEventPostToolUse:  PhaseAfterTool,
		},
		aliases: []string{"claude-code"},
	})

	matrix := reg.HookSupportMatrix()
	assert.Len(t, matrix, 1)
	assert.Equal(t, AgentTypeClaudeCode, matrix[0].AgentType)
	assert.Equal(t, "Agent", matrix[0].AgentName)
	assert.Contains(t, matrix[0].Phases[PhaseStart], HookEventSessionStart)
	assert.Contains(t, matrix[0].Phases[PhaseAfterTool], HookEventPostToolUse)
}

func TestHookSupportMatrix_NoLifecycleAgents(t *testing.T) {
	reg := NewRegistry().(*registry)
	reg.Register(&mockDetectableAgent{
		agentType: AgentTypeCursor,
		name:      "Plain",
	})

	matrix := reg.HookSupportMatrix()
	assert.Empty(t, matrix)
}

func TestHookSupportMatrix_MixedAgents(t *testing.T) {
	reg := NewRegistry().(*registry)
	// one with lifecycle, one without
	reg.Register(&mockLifecycleDetectableAgent{
		mockDetectableAgent: mockDetectableAgent{agentType: AgentTypeClaudeCode, name: "Lifecycle"},
		phases:              EventPhaseMap{HookEventSessionStart: PhaseStart},
		aliases:             []string{"claude-code"},
	})
	reg.Register(&mockDetectableAgent{
		agentType: AgentTypeCursor,
		name:      "Plain",
	})

	matrix := reg.HookSupportMatrix()
	assert.Len(t, matrix, 1)
	assert.Equal(t, AgentTypeClaudeCode, matrix[0].AgentType)
}

// --- NewDetectorWithEnv tests ---

func TestNewDetectorWithEnv(t *testing.T) {
	orig := DefaultRegistry
	defer func() { DefaultRegistry = orig }()

	DefaultRegistry = NewRegistry()

	envChecked := false
	DefaultRegistry.Register(&mockDetectableAgent{
		agentType: AgentTypeClaudeCode,
		name:      "Agent",
		detectFn: func(_ context.Context, env Environment) (bool, error) {
			envChecked = env.GetEnv("CUSTOM_VAR") == "yes"
			return envChecked, nil
		},
	})

	mockEnv := NewMockEnvironment(map[string]string{"CUSTOM_VAR": "yes"})
	d := NewDetectorWithEnv(mockEnv)

	agent, err := d.Detect(context.Background())
	require.NoError(t, err)
	require.NotNil(t, agent)
	assert.True(t, envChecked, "custom environment should have been used")
}

// --- detector.getEnv tests ---

func TestDetectorGetEnv_CustomEnv(t *testing.T) {
	mockEnv := NewMockEnvironment(map[string]string{"KEY": "val"})
	d := &detector{
		registry: NewRegistry().(*registry),
		env:      mockEnv,
	}

	env := d.getEnv()
	assert.Equal(t, "val", env.GetEnv("KEY"))
}

func TestDetectorGetEnv_FallbackToSystem(t *testing.T) {
	d := &detector{
		registry: NewRegistry().(*registry),
		env:      nil,
	}

	env := d.getEnv()
	// should not panic and should return a usable Environment
	assert.NotNil(t, env)
	// calling GOOS on a SystemEnvironment should return a non-empty value
	_, ok := env.(*SystemEnvironment)
	assert.True(t, ok, "should fall back to SystemEnvironment")
}

// --- DefaultRegistry-based function tests ---

func TestBuildEventPhaseMap_DefaultRegistry(t *testing.T) {
	orig := DefaultRegistry
	defer func() { DefaultRegistry = orig }()

	DefaultRegistry = NewRegistry()
	DefaultRegistry.Register(&mockLifecycleDetectableAgent{
		mockDetectableAgent: mockDetectableAgent{agentType: AgentTypeClaudeCode, name: "Agent"},
		phases:              EventPhaseMap{HookEventSessionStart: PhaseStart},
		aliases:             []string{"test-alias"},
	})

	result := BuildEventPhaseMap()
	assert.Contains(t, result, "test-alias")
	assert.Equal(t, PhaseStart, result["test-alias"][HookEventSessionStart])
}

func TestResolveAgentENV_DefaultRegistry(t *testing.T) {
	orig := DefaultRegistry
	defer func() { DefaultRegistry = orig }()

	DefaultRegistry = NewRegistry()
	DefaultRegistry.Register(&mockLifecycleDetectableAgent{
		mockDetectableAgent: mockDetectableAgent{agentType: AgentTypeClaudeCode, name: "Agent"},
		aliases:             []string{"test-alias"},
	})

	assert.Equal(t, AgentTypeClaudeCode, ResolveAgentENV("test-alias"))
	assert.Equal(t, AgentTypeUnknown, ResolveAgentENV("nonexistent"))
}

func TestHookSupportMatrix_DefaultRegistry(t *testing.T) {
	orig := DefaultRegistry
	defer func() { DefaultRegistry = orig }()

	DefaultRegistry = NewRegistry()
	DefaultRegistry.Register(&mockLifecycleDetectableAgent{
		mockDetectableAgent: mockDetectableAgent{agentType: AgentTypeClaudeCode, name: "Agent"},
		phases:              EventPhaseMap{HookEventSessionStart: PhaseStart},
		aliases:             []string{"test-alias"},
	})

	matrix := HookSupportMatrix()
	assert.Len(t, matrix, 1)
	assert.Equal(t, AgentTypeClaudeCode, matrix[0].AgentType)
}

// --- Convenience function tests ---

func TestCurrentAgent_None(t *testing.T) {
	orig := DefaultRegistry
	defer func() { DefaultRegistry = orig }()

	DefaultRegistry = NewRegistry()
	assert.Nil(t, CurrentAgent())
}

func TestCurrentAgent_Detected(t *testing.T) {
	orig := DefaultRegistry
	defer func() { DefaultRegistry = orig }()

	DefaultRegistry = NewRegistry()
	DefaultRegistry.Register(&mockDetectableAgent{
		agentType: AgentTypeClaudeCode,
		name:      "Agent",
		detectFn:  func(_ context.Context, _ Environment) (bool, error) { return true, nil },
	})

	agent := CurrentAgent()
	require.NotNil(t, agent)
	assert.Equal(t, AgentTypeClaudeCode, agent.Type())
}

// TestDetectByRole_RuntimeOutranksAgentEnv is the core #527 regression guard:
// when a runtime agent (e.g. Claude Code via CLAUDECODE=1) and an AGENT_ENV
// override (e.g. AGENT_ENV=pi from a stray CLAUDE.md instruction) disagree,
// the Detector MUST return the runtime-detected agent, not the one claimed
// by the override.
//
// Failure prevented: a Claude Code session silently registering as Pi in
// agent_instances.jsonl, routing through the wrong session adapter, and
// poisoning CLAUDE_ENV_FILE with AGENT_ENV=pi for every subsequent subprocess.
func TestDetectByRole_RuntimeOutranksAgentEnv(t *testing.T) {
	orig := DefaultRegistry
	defer func() { DefaultRegistry = orig }()

	DefaultRegistry = NewRegistry()

	// Pi-like agent: matches on AGENT_ENV=pi (override), no runtime signal.
	DefaultRegistry.Register(&mockRuntimeDetectableAgent{
		mockDetectableAgent: mockDetectableAgent{
			agentType: AgentTypePi,
			name:      "Pi",
			detectFn: func(_ context.Context, env Environment) (bool, error) {
				return env.GetEnv("AGENT_ENV") == "pi", nil
			},
		},
		// Pi has no runtime signal in this scenario.
		detectRuntimeFn: func(_ context.Context, _ Environment) (bool, error) { return false, nil },
	})

	// Claude-like agent: matches on CLAUDECODE=1 (runtime) or AGENT_ENV=claude-code.
	DefaultRegistry.Register(&mockRuntimeDetectableAgent{
		mockDetectableAgent: mockDetectableAgent{
			agentType: AgentTypeClaudeCode,
			name:      "Claude Code",
			detectFn: func(_ context.Context, env Environment) (bool, error) {
				if env.GetEnv("CLAUDECODE") == "1" {
					return true, nil
				}
				return env.GetEnv("AGENT_ENV") == "claude-code", nil
			},
		},
		detectRuntimeFn: func(_ context.Context, env Environment) (bool, error) {
			return env.GetEnv("CLAUDECODE") == "1", nil
		},
	})

	// Scenario: both signals are set and disagree.
	env := NewMockEnvironment(map[string]string{
		"CLAUDECODE": "1",
		"AGENT_ENV":  "pi",
	})

	detected, err := NewDetectorWithEnv(env).Detect(context.Background())
	require.NoError(t, err)
	require.NotNil(t, detected)
	assert.Equal(t, AgentTypeClaudeCode, detected.Type(),
		"runtime signal (CLAUDECODE=1) must outrank AGENT_ENV=pi")
}

// TestDetectByRole_AgentEnvFallback confirms phase 2: when no runtime
// signal matches, the AGENT_ENV override is honored as the fallback.
// This preserves the existing escape-hatch behavior for headless / CI
// contexts where only AGENT_ENV is available.
func TestDetectByRole_AgentEnvFallback(t *testing.T) {
	orig := DefaultRegistry
	defer func() { DefaultRegistry = orig }()

	DefaultRegistry = NewRegistry()

	DefaultRegistry.Register(&mockRuntimeDetectableAgent{
		mockDetectableAgent: mockDetectableAgent{
			agentType: AgentTypePi,
			name:      "Pi",
			detectFn: func(_ context.Context, env Environment) (bool, error) {
				return env.GetEnv("AGENT_ENV") == "pi", nil
			},
		},
		detectRuntimeFn: func(_ context.Context, _ Environment) (bool, error) { return false, nil },
	})

	DefaultRegistry.Register(&mockRuntimeDetectableAgent{
		mockDetectableAgent: mockDetectableAgent{
			agentType: AgentTypeClaudeCode,
			name:      "Claude Code",
			detectFn: func(_ context.Context, env Environment) (bool, error) {
				return env.GetEnv("CLAUDECODE") == "1", nil
			},
		},
		detectRuntimeFn: func(_ context.Context, env Environment) (bool, error) {
			return env.GetEnv("CLAUDECODE") == "1", nil
		},
	})

	// Only AGENT_ENV is set — no runtime signals from any agent.
	env := NewMockEnvironment(map[string]string{"AGENT_ENV": "pi"})

	detected, err := NewDetectorWithEnv(env).Detect(context.Background())
	require.NoError(t, err)
	require.NotNil(t, detected)
	assert.Equal(t, AgentTypePi, detected.Type(),
		"AGENT_ENV override must win when no runtime signal claims the process")
}

// TestDetectByRole_LegacyAgentHasNoRuntimeDetector confirms backward
// compatibility: agents that don't implement RuntimeDetector still work
// via the phase-2 fallback — they simply never match in phase 1.
func TestDetectByRole_LegacyAgentHasNoRuntimeDetector(t *testing.T) {
	orig := DefaultRegistry
	defer func() { DefaultRegistry = orig }()

	DefaultRegistry = NewRegistry()

	// Plain agent (no RuntimeDetector) that matches on AGENT_ENV only.
	DefaultRegistry.Register(&mockDetectableAgent{
		agentType: AgentTypeCursor,
		name:      "Cursor",
		detectFn: func(_ context.Context, env Environment) (bool, error) {
			return env.GetEnv("AGENT_ENV") == "cursor", nil
		},
	})

	env := NewMockEnvironment(map[string]string{"AGENT_ENV": "cursor"})
	detected, err := NewDetectorWithEnv(env).Detect(context.Background())
	require.NoError(t, err)
	require.NotNil(t, detected)
	assert.Equal(t, AgentTypeCursor, detected.Type(),
		"legacy agent without RuntimeDetector must still match via phase-2 fallback")
}

func TestCurrentOrchestrator_None(t *testing.T) {
	orig := DefaultRegistry
	defer func() { DefaultRegistry = orig }()

	DefaultRegistry = NewRegistry()
	assert.Nil(t, CurrentOrchestrator())
}

func TestCurrentOrchestrator_Detected(t *testing.T) {
	orig := DefaultRegistry
	defer func() { DefaultRegistry = orig }()

	DefaultRegistry = NewRegistry()
	DefaultRegistry.Register(&mockDetectableAgent{
		agentType: AgentTypeConductor,
		name:      "Conductor",
		role:      RoleOrchestrator,
		detectFn:  func(_ context.Context, _ Environment) (bool, error) { return true, nil },
	})

	orch := CurrentOrchestrator()
	require.NotNil(t, orch)
	assert.Equal(t, AgentTypeConductor, orch.Type())
}

func TestOrchestratorType_None(t *testing.T) {
	orig := DefaultRegistry
	defer func() { DefaultRegistry = orig }()

	DefaultRegistry = NewRegistry()
	assert.Equal(t, "", OrchestratorType())
}

func TestOrchestratorType_Detected(t *testing.T) {
	orig := DefaultRegistry
	defer func() { DefaultRegistry = orig }()

	DefaultRegistry = NewRegistry()
	DefaultRegistry.Register(&mockDetectableAgent{
		agentType: AgentTypeConductor,
		name:      "Conductor",
		role:      RoleOrchestrator,
		detectFn:  func(_ context.Context, _ Environment) (bool, error) { return true, nil },
	})

	assert.Equal(t, "conductor", OrchestratorType())
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	reg := NewRegistry()
	var wg sync.WaitGroup
	const n = 100

	for i := 0; i < n; i++ {
		wg.Add(3)
		go func(i int) {
			defer wg.Done()
			_ = reg.Register(&mockDetectableAgent{
				agentType: AgentType(fmt.Sprintf("stress-%d", i)),
				name:      fmt.Sprintf("stress-agent-%d", i),
				role:      RoleAgent,
			})
		}(i)
		go func() {
			defer wg.Done()
			_ = reg.List()
		}()
		go func(i int) {
			defer wg.Done()
			_, _ = reg.Get(AgentType(fmt.Sprintf("stress-%d", i)))
		}(i)
	}
	wg.Wait()

	agents := reg.List()
	assert.Equal(t, n, len(agents))
}

func TestRegistry_ConcurrentBuildEventPhaseMap(t *testing.T) {
	reg := NewRegistry()
	// register some lifecycle agents
	for i := 0; i < 10; i++ {
		_ = reg.Register(&mockLifecycleDetectableAgent{
			mockDetectableAgent: mockDetectableAgent{
				agentType: AgentType(fmt.Sprintf("lifecycle-%d", i)),
				name:      fmt.Sprintf("lifecycle-agent-%d", i),
				role:      RoleAgent,
			},
			phases:  EventPhaseMap{"event": PhaseStart},
			aliases: []string{fmt.Sprintf("alias-%d", i)},
		})
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			_ = reg.BuildEventPhaseMap()
		}()
		go func() {
			defer wg.Done()
			_ = reg.ResolveAgentENV("alias-5")
		}()
		go func() {
			defer wg.Done()
			_ = reg.HookSupportMatrix()
		}()
	}
	wg.Wait()
}
