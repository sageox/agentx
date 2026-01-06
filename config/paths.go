// Package config provides configuration path resolution for agentx.
package config

import (
	"path/filepath"

	"github.com/sageox/agentx"
)

// Paths provides XDG-compliant path resolution for agent configuration.
type Paths struct {
	env agentx.Environment
}

// NewPaths creates a new path resolver.
func NewPaths(env agentx.Environment) *Paths {
	if env == nil {
		env = agentx.NewSystemEnvironment()
	}
	return &Paths{env: env}
}

// ConfigHome returns the XDG config home directory.
func (p *Paths) ConfigHome() (string, error) {
	return p.env.ConfigDir()
}

// DataHome returns the XDG data home directory.
func (p *Paths) DataHome() (string, error) {
	return p.env.DataDir()
}

// CacheHome returns the XDG cache home directory.
func (p *Paths) CacheHome() (string, error) {
	return p.env.CacheDir()
}

// AppConfig returns the config path for a specific application.
func (p *Paths) AppConfig(appName string) (string, error) {
	configDir, err := p.env.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, appName), nil
}

// AppData returns the data path for a specific application.
func (p *Paths) AppData(appName string) (string, error) {
	dataDir, err := p.env.DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, appName), nil
}

// AppCache returns the cache path for a specific application.
func (p *Paths) AppCache(appName string) (string, error) {
	cacheDir, err := p.env.CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, appName), nil
}

// AgentConfigPaths maps agent types to their known config directories.
// These are the standard locations where each agent stores its configuration.
var AgentConfigPaths = map[agentx.AgentType]string{
	agentx.AgentTypeClaudeCode: ".claude",
	agentx.AgentTypeCursor:     ".cursor",
	agentx.AgentTypeWindsurf:   ".codeium",
	agentx.AgentTypeCopilot:    ".config/github-copilot",
	agentx.AgentTypeAider:      ".aider",
	agentx.AgentTypeCody:       ".config/cody",
	agentx.AgentTypeContinue:   ".continue",
	agentx.AgentTypeCodePuppy:  ".config/code-puppy",
	agentx.AgentTypeKiro:       ".kiro",
}

// AgentConfigPath returns the config path for a specific agent type.
func (p *Paths) AgentConfigPath(agentType agentx.AgentType) (string, error) {
	home, err := p.env.HomeDir()
	if err != nil {
		return "", err
	}

	if relPath, ok := AgentConfigPaths[agentType]; ok {
		return filepath.Join(home, relPath), nil
	}

	// fallback to generic location
	configDir, err := p.env.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, string(agentType)), nil
}
