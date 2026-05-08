package setup

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sageox/agentx"
)

func TestRegisterDefaultAgents_RegistersAllSupportedAgents(t *testing.T) {
	orig := agentx.DefaultRegistry
	defer func() { agentx.DefaultRegistry = orig }()

	agentx.DefaultRegistry = agentx.NewRegistry()
	RegisterDefaultAgents()

	for _, agentType := range agentx.SupportedAgents {
		agent, ok := agentx.DefaultRegistry.Get(agentType)
		assert.True(t, ok, "expected %s to be registered", agentType)
		if ok {
			assert.Equal(t, agentType, agent.Type())
		}
	}
}
