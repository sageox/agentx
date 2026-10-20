package agentx

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// DefaultRegistry is the global registry with all supported agents.
var DefaultRegistry = NewRegistry()

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

// List returns the registered agents sorted by type. Detection takes the
// first agent that matches, so walking the map directly made the winner
// random whenever two agents matched (sageox/agentx#17).
func (r *registry) List() []Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]Agent, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}
	sort.Slice(agents, func(i, j int) bool { return agents[i].Type() < agents[j].Type() })
	return agents
}

func (r *registry) Detector() Detector {
	return &detector{registry: r}
}

// detector implements Detector using a registry.
type detector struct {
	registry *registry
	env      Environment
}

// NewDetector creates a detector with the default registry.
func NewDetector() Detector {
	return DefaultRegistry.Detector()
}

// NewDetectorWithEnv creates a detector with a custom environment.
func NewDetectorWithEnv(env Environment) Detector {
	reg, ok := DefaultRegistry.(*registry)
	if !ok {
		// Fallback to empty registry if type assertion fails
		reg = &registry{
			agents: make(map[AgentType]Agent),
		}
	}
	return &detector{
		registry: reg,
		env:      env,
	}
}

func (d *detector) getEnv() Environment {
	if d.env != nil {
		return d.env
	}
	return NewSystemEnvironment()
}

func (d *detector) Detect(ctx context.Context) (Agent, error) {
	return d.detectByRole(ctx, RoleAgent)
}

func (d *detector) DetectOrchestrator(ctx context.Context) (Agent, error) {
	return d.detectByRole(ctx, RoleOrchestrator)
}

// detectByRole finds the first detected agent matching the given role.
//
// Five phases, strongest evidence first; within a phase agents are tried in
// List() order, so the result never depends on map iteration:
//  1. Runtime signals (authoritative). Agents that implement RuntimeDetector
//     are asked whether any of their non-AGENT_ENV evidence is present
//     (CLAUDECODE=1, CLAUDE_CODE_ENTRYPOINT, PI_CODING_AGENT_DIR, etc.).
//     The first match wins regardless of what AGENT_ENV says.
//  2. An explicit AGENT_ENV naming a registered agent, by type or alias.
//  3. Detect() of agents that don't implement RuntimeDetector. Detect() is
//     their only evidence, and it is their own marker (GOOSE=1,
//     CURSOR_AGENT=1), not a filesystem hint.
//  4. Venue signals (VenueDetector): the host belongs to an agent, such as
//     an Amp orb.
//  5. Everything else Detect() reports, including filesystem hints like a
//     checked-in .codex/ directory.
//
// Phase 1 before 2 resolves the #527 class of bug: a Claude Code session
// where a stray CLAUDE.md instruction sets AGENT_ENV=pi no longer silently
// mis-routes through the Pi adapter. Phases 2 to 4 before 5 resolve
// sageox/agentx#17: inside an Amp orb, a repository's .codex/ directory
// made Codex win, so the orb's sessions were never recorded as Amp's.
// Phase 3 before 4 keeps an agent run inside an orb, such as Goose, from
// being reported as the orb's Amp.
func (d *detector) detectByRole(ctx context.Context, role AgentRole) (Agent, error) {
	env := d.getEnv()
	agents := d.registry.List()
	agentEnv := strings.ToLower(strings.TrimSpace(env.GetEnv("AGENT_ENV")))

	// Phase 1: runtime signals only
	for _, agent := range agents {
		if agent.Role() != role {
			continue
		}
		rd, ok := agent.(RuntimeDetector)
		if !ok {
			continue
		}
		if detected, err := rd.DetectRuntime(ctx, env); err == nil && detected {
			return agent, nil
		}
	}

	// Phase 2: an explicit AGENT_ENV
	if agentEnv != "" {
		for _, agent := range agents {
			if agent.Role() == role && namedByAgentENV(agent, agentEnv) {
				return agent, nil
			}
		}
	}

	// Phase 3: agents whose Detect() is all the evidence they have
	for _, agent := range agents {
		if agent.Role() != role {
			continue
		}
		if _, ok := agent.(RuntimeDetector); ok {
			continue
		}
		if detected, err := agent.Detect(ctx, env); err == nil && detected {
			return agent, nil
		}
	}

	// Phase 4: venue signals
	for _, agent := range agents {
		if agent.Role() != role {
			continue
		}
		vd, ok := agent.(VenueDetector)
		if !ok {
			continue
		}
		if detected, err := vd.DetectVenue(ctx, env); err == nil && detected {
			return agent, nil
		}
	}

	// Phase 5: fallback to AGENT_ENV-inclusive Detect, filesystem hints included
	for _, agent := range agents {
		if agent.Role() != role {
			continue
		}
		if detected, err := agent.Detect(ctx, env); err == nil && detected {
			return agent, nil
		}
	}

	return nil, nil
}

// agentENVAliaser is the part of LifecycleEventMapper that names an agent's
// AGENT_ENV values. Some agents declare aliases without mapping lifecycle
// events (OMP accepts "oh-my-pi" but has no EventPhases), so phase 2 checks
// this alone.
type agentENVAliaser interface {
	AgentENVAliases() []string
}

// namedByAgentENV reports whether a lowercased, trimmed AGENT_ENV names
// agent: its type, or one of its aliases. The type counts too because some
// agents (Codex) declare no aliases and match AGENT_ENV against their type
// inside Detect.
func namedByAgentENV(agent Agent, agentEnv string) bool {
	if strings.ToLower(string(agent.Type())) == agentEnv {
		return true
	}
	aliaser, ok := agent.(agentENVAliaser)
	if !ok {
		return false
	}
	for _, alias := range aliaser.AgentENVAliases() {
		if strings.ToLower(alias) == agentEnv {
			return true
		}
	}
	return false
}

// DetectAll returns all detected agents and orchestrators regardless of role.
// Unlike Detect (agents only) and DetectOrchestrator (orchestrators only),
// this does not filter by role.
func (d *detector) DetectAll(ctx context.Context) ([]Agent, error) {
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

func (d *detector) DetectByType(ctx context.Context, agentType AgentType) (bool, error) {
	agent, ok := d.registry.Get(agentType)
	if !ok {
		return false, fmt.Errorf("agent type %s not registered", agentType)
	}

	return agent.Detect(ctx, d.getEnv())
}

// BuildEventPhaseMap builds a map from AGENT_ENV alias -> (native event -> phase)
// by iterating all registered agents that implement LifecycleEventMapper.
// The returned map uses AGENT_ENV aliases as keys (e.g., "claude-code"),
// not AgentType slugs (e.g., "claude").
func BuildEventPhaseMap() map[string]EventPhaseMap {
	return DefaultRegistry.BuildEventPhaseMap()
}

func (r *registry) BuildEventPhaseMap() map[string]EventPhaseMap {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]EventPhaseMap)
	for _, agent := range r.agents {
		mapper, ok := agent.(LifecycleEventMapper)
		if !ok {
			continue
		}
		phases := mapper.EventPhases()
		for _, alias := range mapper.AgentENVAliases() {
			result[alias] = phases
		}
	}
	return result
}

// ResolveAgentENV resolves an AGENT_ENV string to an AgentType.
// Returns AgentTypeUnknown if not found.
func ResolveAgentENV(agentEnv string) AgentType {
	return DefaultRegistry.ResolveAgentENV(agentEnv)
}

func (r *registry) ResolveAgentENV(agentEnv string) AgentType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, agent := range r.agents {
		mapper, ok := agent.(LifecycleEventMapper)
		if !ok {
			continue
		}
		for _, alias := range mapper.AgentENVAliases() {
			if alias == agentEnv {
				return agent.Type()
			}
		}
	}
	return AgentTypeUnknown
}

// HookSupportEntry represents one agent's lifecycle hook support.
// Phases maps each canonical phase to the agent's native event name.
type HookSupportEntry struct {
	AgentType AgentType
	AgentName string
	Phases    map[Phase][]HookEvent // phase -> native event names (may have multiple)
}

// HookSupportMatrix returns the hook support matrix for all registered agents
// that implement LifecycleEventMapper. Each entry shows which phases the agent
// supports and the native event name for each phase.
func HookSupportMatrix() []HookSupportEntry {
	return DefaultRegistry.HookSupportMatrix()
}

func (r *registry) HookSupportMatrix() []HookSupportEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var entries []HookSupportEntry
	for _, agent := range r.agents {
		mapper, ok := agent.(LifecycleEventMapper)
		if !ok {
			continue
		}
		// Invert the event->phase map to phase->events
		phaseEvents := make(map[Phase][]HookEvent)
		for event, phase := range mapper.EventPhases() {
			phaseEvents[phase] = append(phaseEvents[phase], event)
		}
		entries = append(entries, HookSupportEntry{
			AgentType: agent.Type(),
			AgentName: agent.Name(),
			Phases:    phaseEvents,
		})
	}
	return entries
}

// IsAgentContext returns true if running inside any coding agent.
// This is a convenience function using the default registry.
func IsAgentContext() bool {
	ctx := context.Background()
	agent, _ := NewDetector().Detect(ctx)
	return agent != nil
}

// RequireAgent returns an error message if not running in an agent context.
// Returns empty string if in agent context.
func RequireAgent(commandName string) string {
	if IsAgentContext() {
		return ""
	}
	return fmt.Sprintf("'%s' must be run from within a coding agent (Claude Code, Cursor, etc.).\n"+
		"If your agent doesn't set standard env vars, set AGENT_ENV=<agent-name> before running.", commandName)
}

// CurrentAgent returns the currently detected coding agent, or nil if none.
func CurrentAgent() Agent {
	ctx := context.Background()
	agent, _ := NewDetector().Detect(ctx)
	return agent
}

// CurrentOrchestrator returns the currently detected orchestrator, or nil if none.
func CurrentOrchestrator() Agent {
	ctx := context.Background()
	orch, _ := NewDetector().DetectOrchestrator(ctx)
	return orch
}

// OrchestratorType returns the type string of the detected orchestrator,
// or empty string if none detected.
func OrchestratorType() string {
	orch := CurrentOrchestrator()
	if orch == nil {
		return ""
	}
	return string(orch.Type())
}
