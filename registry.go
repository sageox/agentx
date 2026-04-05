package agentx

import (
	"context"
	"fmt"
	"sync"
)

// DefaultRegistry is the global registry with all supported agents.
var DefaultRegistry = NewRegistry()

// Registry manages available agents and provides detection.
type Registry interface {
	// Register adds an agent to the registry.
	Register(agent Agent) error

	// Get retrieves an agent by type.
	Get(agentType AgentType) (Agent, bool)

	// List returns all registered agents.
	List() []Agent
}

// registry implements Registry with thread-safe agent management.
type registry struct {
	mu     sync.RWMutex
	agents map[AgentType]Agent
}

// NewRegistry creates a new empty agent registry.
func NewRegistry() Registry {
	return &registry{
		agents: make(map[AgentType]Agent),
	}
}

func (r *registry) Register(agent Agent) error {
	if agent == nil {
		return fmt.Errorf("agent cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.agents[agent.Type()] = agent
	return nil
}

func (r *registry) Get(agentType AgentType) (Agent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, ok := r.agents[agentType]
	return agent, ok
}

func (r *registry) List() []Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]Agent, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}
	return agents
}

// detector implements detection using a registry.
type detector struct {
	registry *registry
	env      Environment
}

func (d *detector) getEnv() Environment {
	if d.env != nil {
		return d.env
	}
	return NewSystemEnvironment()
}

// agentEnvAliases maps AGENT_ENV values to canonical agent types.
// We use the canonical agent type slug (e.g., "claude") as the standard.
var agentEnvAliases = map[string]AgentType{
	"claude": AgentTypeClaudeCode,
	"cursor":      AgentTypeCursor,
	"windsurf":    AgentTypeWindsurf,
	"copilot":     AgentTypeCopilot,
	"aider":       AgentTypeAider,
	"cody":        AgentTypeCody,
	"continue":    AgentTypeContinue,
	"code-puppy":  AgentTypeCodePuppy,
	"kiro":        AgentTypeKiro,
	"opencode":    AgentTypeOpenCode,
	"goose":       AgentTypeGoose,
	"amp":         AgentTypeAmp,
	"gemini":      AgentTypeGemini,
}

func (d *detector) detect(ctx context.Context) (Agent, error) {
	env := d.getEnv()

	// First check explicit AGENT_ENV - this is the definitive answer if set
	if agentEnv := env.GetEnv("AGENT_ENV"); agentEnv != "" {
		if agentType, ok := agentEnvAliases[agentEnv]; ok {
			if agent, ok := d.registry.Get(agentType); ok {
				return agent, nil
			}
		}
		// AGENT_ENV set but not recognized - don't fall through to native detection
		return nil, nil
	}

	// Then check each agent's native detection (env vars, heuristics)
	for _, agent := range d.registry.List() {
		detected, err := agent.Detect(ctx, env)
		if err != nil {
			continue
		}
		if detected {
			return agent, nil
		}
	}

	return nil, nil
}

func (d *detector) detectAll(ctx context.Context) ([]Agent, error) {
	env := d.getEnv()
	var detected []Agent

	for _, agent := range d.registry.List() {
		ok, err := agent.Detect(ctx, env)
		if err != nil {
			continue
		}
		if ok {
			detected = append(detected, agent)
		}
	}

	return detected, nil
}

func (d *detector) detectByType(ctx context.Context, agentType AgentType) (bool, error) {
	agent, ok := d.registry.Get(agentType)
	if !ok {
		return false, fmt.Errorf("agent type %s not registered", agentType)
	}

	return agent.Detect(ctx, d.getEnv())
}
