package orchestrators

import (
	"context"
	"path/filepath"

	"github.com/sageox/agentx"
)

type GasCityAgent struct{}

func NewGasCityAgent() *GasCityAgent { return &GasCityAgent{} }

func (a *GasCityAgent) Type() agentx.AgentType { return agentx.AgentTypeGasCity }
func (a *GasCityAgent) Name() string           { return "Gas City" }
func (a *GasCityAgent) URL() string            { return "https://github.com/gastownhall/gascity" }
func (a *GasCityAgent) Role() agentx.AgentRole { return agentx.RoleOrchestrator }

func (a *GasCityAgent) Detect(_ context.Context, env agentx.Environment) (bool, error) {
	if env.GetEnv("ORCHESTRATOR_ENV") == "gascity" {
		return true, nil
	}
	if env.GetEnv("GASCITY") == "1" {
		return true, nil
	}
	if env.GetEnv("GC_VERSION") != "" {
		return true, nil
	}
	if env.GetEnv("GC_RIG") != "" {
		return true, nil
	}
	if env.GetEnv("GC_PACK") != "" {
		return true, nil
	}
	if env.GetEnv("GC_RUN_ID") != "" {
		return true, nil
	}
	return false, nil
}

func (a *GasCityAgent) UserConfigPath(env agentx.Environment) (string, error) {
	if stateDir := env.GetEnv("GC_STATE_DIR"); stateDir != "" {
		return stateDir, nil
	}
	home, err := env.HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gascity"), nil
}

func (a *GasCityAgent) ProjectConfigPath() string { return "" }
func (a *GasCityAgent) ContextFiles() []string    { return []string{"AGENTS.md"} }
func (a *GasCityAgent) SupportsXDGConfig() bool   { return false }

func (a *GasCityAgent) Capabilities() agentx.Capabilities {
	return agentx.Capabilities{MCPServers: true, SystemPrompt: true, ProjectContext: true}
}

func (a *GasCityAgent) HookManager() agentx.HookManager       { return nil }
func (a *GasCityAgent) CommandManager() agentx.CommandManager { return nil }
func (a *GasCityAgent) RulesManager() agentx.RulesManager     { return nil }

func (a *GasCityAgent) DetectVersion(_ context.Context, env agentx.Environment) string {
	if v := env.GetEnv("GC_VERSION"); v != "" {
		return v
	}
	return versionFromCommand(env, "gc", "--version")
}

func (a *GasCityAgent) IsInstalled(_ context.Context, env agentx.Environment) (bool, error) {
	if _, err := env.LookPath("gc"); err == nil {
		return true, nil
	}
	home, err := env.HomeDir()
	if err != nil {
		return false, nil
	}
	if env.IsDir(filepath.Join(home, ".gascity")) {
		return true, nil
	}
	return false, nil
}

func (a *GasCityAgent) SupportsSession() bool { return true }

func (a *GasCityAgent) SessionID(env agentx.Environment) string {
	return env.GetEnv("GC_RUN_ID")
}

var _ agentx.Agent = (*GasCityAgent)(nil)
