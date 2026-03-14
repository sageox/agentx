package sessions

import (
	"errors"
	"sync"
)

var (
	// ErrProviderNotFound is returned when a provider for an agent type is not found.
	ErrProviderNotFound = errors.New("session provider not found for agent type")

	// ErrSessionNotFound is returned when a session cannot be found.
	ErrSessionNotFound = errors.New("session not found")

	// ErrProjectNotFound is returned when a project has no sessions.
	ErrProjectNotFound = errors.New("project not found")
)

// registry holds the session providers.
var (
	registry   = make(map[string]SessionProvider)
	registryMu sync.RWMutex
)

// Register registers a session provider for the given agent type.
func Register(agentType string, provider SessionProvider) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[agentType] = provider
}

// ProviderFor returns the session provider for the given agent type.
func ProviderFor(agentType string) (SessionProvider, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	provider, ok := registry[agentType]
	if !ok {
		return nil, ErrProviderNotFound
	}
	return provider, nil
}

// DefaultProviders returns all registered providers.
func DefaultProviders() map[string]SessionProvider {
	registryMu.RLock()
	defer registryMu.RUnlock()

	result := make(map[string]SessionProvider)
	for k, v := range registry {
		result[k] = v
	}
	return result
}

// ListAllSessions returns all sessions across all providers and projects.
func ListAllSessions() (map[string][]SessionSummary, error) {
	registryMu.RLock()
	providers := DefaultProviders()
	registryMu.RUnlock()

	result := make(map[string][]SessionSummary)

	for agentType, provider := range providers {
		projects, err := provider.ListProjects()
		if err != nil {
			continue
		}

		for _, project := range projects {
			sessions, err := provider.ListSessions(project)
			if err != nil {
				continue
			}
			result[agentType] = append(result[agentType], sessions...)
		}
	}

	return result, nil
}

func init() {
	// Register Claude Code provider
	Register("claude-code", &claudeCodeProvider{})

	// Register Gemini provider
	Register("gemini", &geminiProvider{})
}
