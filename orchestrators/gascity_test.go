package orchestrators

import (
	"context"
	"testing"

	"github.com/sageox/agentx"
	"github.com/stretchr/testify/assert"
)

func TestGasCityAgent_Detect(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{"ORCHESTRATOR_ENV=gascity", map[string]string{"ORCHESTRATOR_ENV": "gascity"}, true},
		{"GASCITY=1", map[string]string{"GASCITY": "1"}, true},
		{"GC_VERSION set", map[string]string{"GC_VERSION": "1.0.0"}, true},
		{"GC_RIG set", map[string]string{"GC_RIG": "/path/to/rig"}, true},
		{"GC_PACK set", map[string]string{"GC_PACK": "sageox"}, true},
		{"GC_RUN_ID set", map[string]string{"GC_RUN_ID": "abc123"}, true},
		{"no env vars", map[string]string{}, false},
		{"wrong ORCHESTRATOR_ENV", map[string]string{"ORCHESTRATOR_ENV": "other"}, false},
		{"GASCITY=0", map[string]string{"GASCITY": "0"}, false},
	}

	agent := NewGasCityAgent()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := agentx.NewMockEnvironment(tt.env)
			got, err := agent.Detect(ctx, env)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGasCityAgent_Identity(t *testing.T) {
	agent := NewGasCityAgent()
	assert.Equal(t, agentx.AgentTypeGasCity, agent.Type())
	assert.Equal(t, "Gas City", agent.Name())
	assert.Equal(t, agentx.RoleOrchestrator, agent.Role())
	assert.Equal(t, "https://github.com/gastownhall/gascity", agent.URL())
}

func TestGasCityAgent_SessionID(t *testing.T) {
	agent := NewGasCityAgent()
	assert.True(t, agent.SupportsSession())

	env := agentx.NewMockEnvironment(map[string]string{"GC_RUN_ID": "run-abc"})
	assert.Equal(t, "run-abc", agent.SessionID(env))

	empty := agentx.NewMockEnvironment(map[string]string{})
	assert.Equal(t, "", agent.SessionID(empty))
}

func TestGasCityAgent_DetectVersion(t *testing.T) {
	agent := NewGasCityAgent()
	env := agentx.NewMockEnvironment(map[string]string{"GC_VERSION": "1.2.3"})
	assert.Equal(t, "1.2.3", agent.DetectVersion(context.Background(), env))
}

func TestGasCityAgent_UserConfigPath(t *testing.T) {
	agent := NewGasCityAgent()

	// GC_STATE_DIR override
	env := agentx.NewMockEnvironment(map[string]string{"GC_STATE_DIR": "/custom/gc/state"})
	p, err := agent.UserConfigPath(env)
	assert.NoError(t, err)
	assert.Equal(t, "/custom/gc/state", p)
}
