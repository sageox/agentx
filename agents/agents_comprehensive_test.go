package agents

import (
	"context"
	"testing"

	"github.com/sageox/agentx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// agentSpec captures the expected values for a single agent implementation,
// enabling table-driven tests across all agents with a single test function.
type agentSpec struct {
	name             string
	constructor      func() agentx.Agent
	agentType        agentx.AgentType
	displayName      string
	url              string
	supportsXDG      bool
	projectConfig    string
	contextFiles     []string
	capabilities     agentx.Capabilities
	binaryName       string   // primary binary for LookPath; empty if none
	extraBinaries    []string // additional binaries to check (e.g., code-puppy)
	macOSApp         string   // e.g., "/Applications/Cursor.app"; empty if none
	supportsSession  bool
	sessionEnvVar    string // env var for SessionID; empty if session comes from hooks
	usesHomeConfig   bool   // true = UserConfigPath uses HomeDir, false = uses ConfigDir
	configSubdir     string // subdir under home or config (e.g., ".claude", "amp")
	versionCmd       string // binary name for versionFromCommand; empty if not supported
	versionPkgJSON   string // relative path under configPath for package.json version detection
	detectEnvVars    []detectCase
	isLifecycleAgent bool
	envAliases       []string
	eventPhaseCount  int
}

type detectCase struct {
	name     string
	envVars  map[string]string
	expected bool
}

// allAgentSpecs returns the specification for every agent in the package.
// Each spec is derived from reading the source code, not guessing.
func allAgentSpecs() []agentSpec {
	return []agentSpec{
		{
			name:            "ClaudeCode",
			constructor:     func() agentx.Agent { return NewClaudeCodeAgent() },
			agentType:       agentx.AgentTypeClaudeCode,
			displayName:     "Claude Code",
			url:             "https://github.com/anthropics/claude-code",
			supportsXDG:     false,
			projectConfig:   ".claude",
			contextFiles:    []string{"CLAUDE.md", "AGENTS.md"},
			capabilities:    agentx.Capabilities{Hooks: true, MCPServers: true, SystemPrompt: true, ProjectContext: true, CustomCommands: true, MinVersion: "1.0"},
			binaryName:      "claude",
			supportsSession: true,
			sessionEnvVar:   "CLAUDE_CODE_SESSION_ID",
			usesHomeConfig:  true,
			configSubdir:    ".claude",
			versionCmd:      "claude",
			detectEnvVars: []detectCase{
				{"CLAUDECODE=1", map[string]string{"CLAUDECODE": "1"}, true},
				{"CLAUDE_CODE_ENTRYPOINT set", map[string]string{"CLAUDE_CODE_ENTRYPOINT": "cli"}, true},
				{"CLAUDE_CODE_SESSION_ID set", map[string]string{"CLAUDE_CODE_SESSION_ID": "sess_123"}, true},
				{"AGENT_ENV=claude-code", map[string]string{"AGENT_ENV": "claude-code"}, true},
				{"AGENT_ENV=claudecode", map[string]string{"AGENT_ENV": "claudecode"}, true},
				{"AGENT_ENV=claude", map[string]string{"AGENT_ENV": "claude"}, true},
				{"no env vars", map[string]string{}, false},
				{"unrelated env", map[string]string{"CURSOR_AGENT": "1"}, false},
			},
			isLifecycleAgent: true,
			envAliases:       []string{"claude-code", "claudecode", "claude"},
			eventPhaseCount:  7,
		},
		{
			name:            "Cursor",
			constructor:     func() agentx.Agent { return NewCursorAgent() },
			agentType:       agentx.AgentTypeCursor,
			displayName:     "Cursor",
			url:             "https://github.com/getcursor/cursor",
			supportsXDG:     false,
			projectConfig:   ".cursor",
			contextFiles:    []string{".cursorrules"},
			capabilities:    agentx.Capabilities{Hooks: true, MCPServers: true, SystemPrompt: true, ProjectContext: true, CustomCommands: false},
			binaryName:      "cursor",
			macOSApp:        "/Applications/Cursor.app",
			supportsSession: true,
			sessionEnvVar:   "",
			usesHomeConfig:  true,
			configSubdir:    ".cursor",
			versionPkgJSON:  "package.json",
			detectEnvVars: []detectCase{
				{"CURSOR_AGENT=1", map[string]string{"CURSOR_AGENT": "1"}, true},
				{"AGENT_ENV=cursor", map[string]string{"AGENT_ENV": "cursor"}, true},
				{"exec path heuristic", map[string]string{"_": "/usr/bin/cursor"}, true},
				{"no env vars", map[string]string{}, false},
			},
			isLifecycleAgent: true,
			envAliases:       []string{"cursor"},
			eventPhaseCount:  7,
		},
		{
			name:            "Windsurf",
			constructor:     func() agentx.Agent { return NewWindsurfAgent() },
			agentType:       agentx.AgentTypeWindsurf,
			displayName:     "Windsurf",
			url:             "https://github.com/codeium/windsurf",
			supportsXDG:     false,
			projectConfig:   ".windsurf",
			contextFiles:    []string{".windsurfrules"},
			capabilities:    agentx.Capabilities{Hooks: true, MCPServers: false, SystemPrompt: true, ProjectContext: true, CustomCommands: false},
			binaryName:      "windsurf",
			macOSApp:        "/Applications/Windsurf.app",
			supportsSession: true,
			sessionEnvVar:   "",
			usesHomeConfig:  true,
			configSubdir:    ".codeium",
			versionPkgJSON:  "windsurf/package.json",
			detectEnvVars: []detectCase{
				{"WINDSURF_AGENT=1", map[string]string{"WINDSURF_AGENT": "1"}, true},
				{"CODEIUM_AGENT=1", map[string]string{"CODEIUM_AGENT": "1"}, true},
				{"AGENT_ENV=windsurf", map[string]string{"AGENT_ENV": "windsurf"}, true},
				{"AGENT_ENV=codeium", map[string]string{"AGENT_ENV": "codeium"}, true},
				{"no env vars", map[string]string{}, false},
			},
			isLifecycleAgent: true,
			envAliases:       []string{"windsurf", "codeium"},
			eventPhaseCount:  6,
		},
		{
			name:            "Copilot",
			constructor:     func() agentx.Agent { return NewCopilotAgent() },
			agentType:       agentx.AgentTypeCopilot,
			displayName:     "GitHub Copilot",
			url:             "https://github.com/features/copilot",
			supportsXDG:     true,
			projectConfig:   ".github",
			contextFiles:    []string{".github/copilot-instructions.md"},
			capabilities:    agentx.Capabilities{Hooks: true, MCPServers: false, SystemPrompt: true, ProjectContext: true, CustomCommands: false},
			binaryName:      "gh",
			supportsSession: true,
			sessionEnvVar:   "",
			usesHomeConfig:  true,
			configSubdir:    ".config/github-copilot",
			detectEnvVars: []detectCase{
				{"COPILOT_AGENT=1", map[string]string{"COPILOT_AGENT": "1"}, true},
				{"AGENT_ENV=copilot", map[string]string{"AGENT_ENV": "copilot"}, true},
				{"AGENT_ENV=github-copilot", map[string]string{"AGENT_ENV": "github-copilot"}, true},
				{"no env vars", map[string]string{}, false},
			},
			isLifecycleAgent: true,
			envAliases:       []string{"copilot", "github-copilot"},
			eventPhaseCount:  5,
		},
		{
			name:            "Aider",
			constructor:     func() agentx.Agent { return NewAiderAgent() },
			agentType:       agentx.AgentTypeAider,
			displayName:     "Aider",
			url:             "https://github.com/Aider-AI/aider",
			supportsXDG:     false,
			projectConfig:   ".aider",
			contextFiles:    []string{".aider.conf.yml", "CONVENTIONS.md"},
			capabilities:    agentx.Capabilities{Hooks: false, MCPServers: false, SystemPrompt: true, ProjectContext: true, CustomCommands: true},
			binaryName:      "aider",
			supportsSession: false,
			usesHomeConfig:  true,
			configSubdir:    ".aider",
			versionCmd:      "aider",
			detectEnvVars: []detectCase{
				{"AIDER=1", map[string]string{"AIDER": "1"}, true},
				{"AIDER_AGENT=1", map[string]string{"AIDER_AGENT": "1"}, true},
				{"AGENT_ENV=aider", map[string]string{"AGENT_ENV": "aider"}, true},
				{"exec path heuristic", map[string]string{"_": "/usr/local/bin/aider"}, true},
				{"no env vars", map[string]string{}, false},
			},
		},
		{
			name:            "Cody",
			constructor:     func() agentx.Agent { return NewCodyAgent() },
			agentType:       agentx.AgentTypeCody,
			displayName:     "Cody",
			url:             "https://github.com/sourcegraph/cody",
			supportsXDG:     true,
			projectConfig:   ".cody",
			contextFiles:    []string{".cody/cody.json"},
			capabilities:    agentx.Capabilities{Hooks: false, MCPServers: false, SystemPrompt: true, ProjectContext: true, CustomCommands: true},
			binaryName:      "cody",
			supportsSession: false,
			usesHomeConfig:  false,
			configSubdir:    "cody",
			detectEnvVars: []detectCase{
				{"CODY_AGENT=1", map[string]string{"CODY_AGENT": "1"}, true},
				{"AGENT_ENV=cody", map[string]string{"AGENT_ENV": "cody"}, true},
				{"no env vars", map[string]string{}, false},
			},
		},
		{
			name:            "Continue",
			constructor:     func() agentx.Agent { return NewContinueAgent() },
			agentType:       agentx.AgentTypeContinue,
			displayName:     "Continue",
			url:             "https://github.com/continuedev/continue",
			supportsXDG:     false,
			projectConfig:   ".continue",
			contextFiles:    []string{".continuerc.json"},
			capabilities:    agentx.Capabilities{Hooks: false, MCPServers: false, SystemPrompt: true, ProjectContext: true, CustomCommands: true},
			binaryName:      "continue",
			supportsSession: false,
			usesHomeConfig:  true,
			configSubdir:    ".continue",
			detectEnvVars: []detectCase{
				{"CONTINUE_AGENT=1", map[string]string{"CONTINUE_AGENT": "1"}, true},
				{"AGENT_ENV=continue", map[string]string{"AGENT_ENV": "continue"}, true},
				{"no env vars", map[string]string{}, false},
			},
		},
		{
			name:            "CodePuppy",
			constructor:     func() agentx.Agent { return NewCodePuppyAgent() },
			agentType:       agentx.AgentTypeCodePuppy,
			displayName:     "Code Puppy",
			url:             "https://github.com/codepuppy-ai/codepuppy",
			supportsXDG:     true,
			projectConfig:   ".codepuppy",
			contextFiles:    []string{".codepuppy/config.json"},
			capabilities:    agentx.Capabilities{Hooks: false, MCPServers: false, SystemPrompt: true, ProjectContext: true, CustomCommands: false},
			binaryName:      "codepuppy",
			extraBinaries:   []string{"code-puppy"},
			supportsSession: false,
			usesHomeConfig:  false,
			configSubdir:    "code-puppy",
			detectEnvVars: []detectCase{
				{"CODE_PUPPY=1", map[string]string{"CODE_PUPPY": "1"}, true},
				{"CODE_PUPPY_AGENT=1", map[string]string{"CODE_PUPPY_AGENT": "1"}, true},
				{"AGENT_ENV=code-puppy", map[string]string{"AGENT_ENV": "code-puppy"}, true},
				{"AGENT_ENV=codepuppy", map[string]string{"AGENT_ENV": "codepuppy"}, true},
				{"no env vars", map[string]string{}, false},
			},
		},
		{
			name:            "Kiro",
			constructor:     func() agentx.Agent { return NewKiroAgent() },
			agentType:       agentx.AgentTypeKiro,
			displayName:     "Kiro",
			url:             "https://kiro.dev",
			supportsXDG:     false,
			projectConfig:   ".kiro",
			contextFiles:    []string{".kiro/rules", ".amazonq/rules"},
			capabilities:    agentx.Capabilities{Hooks: true, MCPServers: true, SystemPrompt: true, ProjectContext: true, CustomCommands: false},
			binaryName:      "kiro",
			macOSApp:        "/Applications/Kiro.app",
			supportsSession: true,
			sessionEnvVar:   "",
			usesHomeConfig:  true,
			configSubdir:    ".kiro",
			detectEnvVars: []detectCase{
				{"KIRO_AGENT=1", map[string]string{"KIRO_AGENT": "1"}, true},
				{"KIRO=1", map[string]string{"KIRO": "1"}, true},
				{"AGENT_ENV=kiro", map[string]string{"AGENT_ENV": "kiro"}, true},
				{"no env vars", map[string]string{}, false},
			},
			isLifecycleAgent: true,
			envAliases:       []string{"kiro"},
			eventPhaseCount:  4,
		},
		{
			name:            "OpenCode",
			constructor:     func() agentx.Agent { return NewOpenCodeAgent() },
			agentType:       agentx.AgentTypeOpenCode,
			displayName:     "OpenCode",
			url:             "https://github.com/opencode-ai/opencode",
			supportsXDG:     false,
			projectConfig:   ".opencode",
			contextFiles:    []string{"AGENTS.md"},
			capabilities:    agentx.Capabilities{Hooks: true, MCPServers: true, SystemPrompt: true, ProjectContext: true, CustomCommands: false},
			binaryName:      "opencode",
			supportsSession: true,
			sessionEnvVar:   "",
			usesHomeConfig:  true,
			configSubdir:    ".opencode",
			versionCmd:      "opencode",
			detectEnvVars: []detectCase{
				{"OPENCODE=1", map[string]string{"OPENCODE": "1"}, true},
				{"OPENCODE_AGENT=1", map[string]string{"OPENCODE_AGENT": "1"}, true},
				{"AGENT_ENV=opencode", map[string]string{"AGENT_ENV": "opencode"}, true},
				{"no env vars", map[string]string{}, false},
			},
			isLifecycleAgent: true,
			envAliases:       []string{"opencode"},
			eventPhaseCount:  5,
		},
		{
			name:            "Goose",
			constructor:     func() agentx.Agent { return NewGooseAgent() },
			agentType:       agentx.AgentTypeGoose,
			displayName:     "Goose",
			url:             "https://github.com/block/goose",
			supportsXDG:     true,
			projectConfig:   "",
			contextFiles:    []string{".goose/config.yaml", ".goosehints"},
			capabilities:    agentx.Capabilities{Hooks: false, MCPServers: true, SystemPrompt: true, ProjectContext: true, CustomCommands: false},
			binaryName:      "goose",
			supportsSession: false,
			usesHomeConfig:  false,
			configSubdir:    "goose",
			versionCmd:      "goose",
			detectEnvVars: []detectCase{
				{"GOOSE=1", map[string]string{"GOOSE": "1"}, true},
				{"GOOSE_AGENT=1", map[string]string{"GOOSE_AGENT": "1"}, true},
				{"AGENT_ENV=goose", map[string]string{"AGENT_ENV": "goose"}, true},
				{"exec path heuristic", map[string]string{"_": "/usr/local/bin/goose"}, true},
				{"no env vars", map[string]string{}, false},
			},
		},
		{
			name:            "Amp",
			constructor:     func() agentx.Agent { return NewAmpAgent() },
			agentType:       agentx.AgentTypeAmp,
			displayName:     "Amp",
			url:             "https://ampcode.com",
			supportsXDG:     true,
			projectConfig:   "",
			contextFiles:    []string{"AGENTS.md"},
			capabilities:    agentx.Capabilities{Hooks: true, MCPServers: true, SystemPrompt: true, ProjectContext: true, CustomCommands: true},
			binaryName:      "amp",
			supportsSession: true,
			sessionEnvVar:   "AMP_THREAD_URL",
			usesHomeConfig:  false,
			configSubdir:    "amp",
			versionCmd:      "amp",
			detectEnvVars: []detectCase{
				{"AMP=1", map[string]string{"AMP": "1"}, true},
				{"AMP_AGENT=1", map[string]string{"AMP_AGENT": "1"}, true},
				{"AMP_THREAD_URL set", map[string]string{"AMP_THREAD_URL": "https://ampcode.com/threads/abc"}, true},
				{"AGENT_ENV=amp", map[string]string{"AGENT_ENV": "amp"}, true},
				{"no env vars", map[string]string{}, false},
			},
			isLifecycleAgent: true,
			envAliases:       []string{"amp"},
			eventPhaseCount:  2,
		},
		{
			name:            "Pi",
			constructor:     func() agentx.Agent { return NewPiAgent() },
			agentType:       agentx.AgentTypePi,
			displayName:     "Pi",
			url:             "https://shittycodingagent.ai/",
			supportsXDG:     false,
			projectConfig:   ".pi",
			contextFiles:    []string{"AGENTS.md", "CLAUDE.md", "SYSTEM.md"},
			capabilities:    agentx.Capabilities{Hooks: false, MCPServers: true, SystemPrompt: true, ProjectContext: true, CustomCommands: false},
			binaryName:      "pi",
			supportsSession: false,
			usesHomeConfig:  true,
			configSubdir:    ".pi",
			versionCmd:      "pi",
			detectEnvVars: []detectCase{
				{"PI_CODING_AGENT_DIR set", map[string]string{"PI_CODING_AGENT_DIR": "/custom/pi"}, true},
				{"AGENT_ENV=pi", map[string]string{"AGENT_ENV": "pi"}, true},
				{"exec path heuristic", map[string]string{"_": "/usr/local/bin/pi-coding-agent"}, true},
				{"bare pi should not match", map[string]string{"_": "/usr/local/bin/pi"}, false},
				{"no env vars", map[string]string{}, false},
			},
		},
		{
			name:            "Cline",
			constructor:     func() agentx.Agent { return NewClineAgent() },
			agentType:       agentx.AgentTypeCline,
			displayName:     "Cline",
			url:             "https://github.com/cline/cline",
			supportsXDG:     true,
			projectConfig:   ".cline",
			contextFiles:    []string{".clinerules", ".cline/instructions.md"},
			capabilities:    agentx.Capabilities{Hooks: true, MCPServers: true, SystemPrompt: true, ProjectContext: true, CustomCommands: false},
			supportsSession: true,
			sessionEnvVar:   "",
			detectEnvVars: []detectCase{
				{"CLINE=1", map[string]string{"CLINE": "1"}, true},
				{"CLINE_AGENT=1", map[string]string{"CLINE_AGENT": "1"}, true},
				{"AGENT_ENV=cline", map[string]string{"AGENT_ENV": "cline"}, true},
				{"AGENT_ENV=claude-dev", map[string]string{"AGENT_ENV": "claude-dev"}, true},
				{"no env vars", map[string]string{}, false},
			},
			isLifecycleAgent: true,
			envAliases:       []string{"cline", "claude-dev"},
			eventPhaseCount:  8,
		},
		{
			name:            "Droid",
			constructor:     func() agentx.Agent { return NewDroidAgent() },
			agentType:       agentx.AgentTypeDroid,
			displayName:     "Droid",
			url:             "https://factory.ai/",
			supportsXDG:     true,
			projectConfig:   ".factory",
			contextFiles:    []string{"DROID.md", ".factory/instructions.md", "AGENTS.md"},
			capabilities:    agentx.Capabilities{Hooks: true, MCPServers: true, SystemPrompt: true, ProjectContext: true, CustomCommands: true},
			binaryName:      "droid",
			supportsSession: true,
			sessionEnvVar:   "",
			versionCmd:      "droid",
			detectEnvVars: []detectCase{
				{"DROID=1", map[string]string{"DROID": "1"}, true},
				{"DROID_AGENT=1", map[string]string{"DROID_AGENT": "1"}, true},
				{"FACTORY_DROID=1", map[string]string{"FACTORY_DROID": "1"}, true},
				{"AGENT_ENV=droid", map[string]string{"AGENT_ENV": "droid"}, true},
				{"AGENT_ENV=factory-droid", map[string]string{"AGENT_ENV": "factory-droid"}, true},
				{"AGENT_ENV=factory", map[string]string{"AGENT_ENV": "factory"}, true},
				{"exec path heuristic", map[string]string{"_": "/usr/local/bin/droid"}, true},
				{"no env vars", map[string]string{}, false},
			},
			isLifecycleAgent: true,
			envAliases:       []string{"droid", "factory-droid", "factory"},
			eventPhaseCount:  7,
		},
		{
			name:            "Gemini CLI",
			constructor:     func() agentx.Agent { return NewGeminiAgent() },
			agentType:       agentx.AgentTypeGemini,
			displayName:     "Gemini CLI",
			url:             "https://github.com/google-gemini/gemini-cli",
			supportsXDG:     false,
			projectConfig:   ".gemini",
			contextFiles:    []string{"GEMINI.md", "AGENTS.md"},
			capabilities:    agentx.Capabilities{Hooks: false, MCPServers: true, SystemPrompt: true, ProjectContext: true, CustomCommands: false},
			binaryName:      "gemini",
			supportsSession: false,
			usesHomeConfig:  true,
			configSubdir:    ".gemini",
			versionCmd:      "gemini",
			detectEnvVars: []detectCase{
				{"GEMINI=1", map[string]string{"GEMINI": "1"}, true},
				{"GEMINI_AGENT=1", map[string]string{"GEMINI_AGENT": "1"}, true},
				{"AGENT_ENV=gemini", map[string]string{"AGENT_ENV": "gemini"}, true},
				{"exec path heuristic", map[string]string{"_": "/usr/local/bin/gemini"}, true},
				{"no env vars", map[string]string{}, false},
			},
		},
		{
			name:            "Codex",
			constructor:     func() agentx.Agent { return NewCodexAgent() },
			agentType:       agentx.AgentTypeCodex,
			displayName:     "Codex",
			url:             "https://github.com/openai/codex",
			supportsXDG:     false,
			projectConfig:   ".codex",
			contextFiles:    []string{"AGENTS.md"},
			capabilities:    agentx.Capabilities{Hooks: false, MCPServers: false, SystemPrompt: true, ProjectContext: true, CustomCommands: false},
			binaryName:      "codex",
			supportsSession: true,
			sessionEnvVar:   "CODEX_THREAD_ID",
			usesHomeConfig:  true,
			configSubdir:    ".codex",
			versionCmd:      "codex",
			detectEnvVars: []detectCase{
				{"AGENT_ENV=codex", map[string]string{"AGENT_ENV": "codex"}, true},
				{"CODEX_CI set", map[string]string{"CODEX_CI": "1"}, true},
				{"CODEX_SANDBOX set", map[string]string{"CODEX_SANDBOX": "workspace-write"}, true},
				{"CODEX_THREAD_ID set", map[string]string{"CODEX_THREAD_ID": "thread_123"}, true},
				{"AGENT_ENV non-codex overrides native vars", map[string]string{"AGENT_ENV": "claude-code", "CODEX_CI": "1"}, false},
				{"no env vars", map[string]string{}, false},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// 1. TestAllAgents_Identity
// ---------------------------------------------------------------------------

func TestAllAgents_Identity(t *testing.T) {
	for _, spec := range allAgentSpecs() {
		t.Run(spec.name, func(t *testing.T) {
			agent := spec.constructor()

			assert.Equal(t, spec.agentType, agent.Type(), "Type()")
			assert.Equal(t, spec.displayName, agent.Name(), "Name()")
			assert.NotEmpty(t, agent.Name(), "Name() must be non-empty")
			assert.Equal(t, spec.url, agent.URL(), "URL()")
			assert.NotEmpty(t, agent.URL(), "URL() must be non-empty")
			assert.Equal(t, agentx.RoleAgent, agent.Role(), "Role()")
		})
	}
}

// ---------------------------------------------------------------------------
// 2. TestAllAgents_Detect
// ---------------------------------------------------------------------------

func TestAllAgents_Detect(t *testing.T) {
	ctx := context.Background()

	for _, spec := range allAgentSpecs() {
		t.Run(spec.name, func(t *testing.T) {
			agent := spec.constructor()

			for _, dc := range spec.detectEnvVars {
				t.Run(dc.name, func(t *testing.T) {
					env := agentx.NewMockEnvironment(dc.envVars)
					detected, err := agent.Detect(ctx, env)
					require.NoError(t, err)
					assert.Equal(t, dc.expected, detected)
				})
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 3. TestAllAgents_UserConfigPath
// ---------------------------------------------------------------------------

func TestAllAgents_UserConfigPath(t *testing.T) {
	for _, spec := range allAgentSpecs() {
		// skip agents with special config path logic (Cline, Droid)
		if spec.configSubdir == "" && spec.name != "Goose" && spec.name != "Amp" {
			continue
		}

		t.Run(spec.name, func(t *testing.T) {
			agent := spec.constructor()

			var env *agentx.MockEnvironment
			if spec.usesHomeConfig {
				env = &agentx.MockEnvironment{Home: "/home/test"}
			} else {
				env = &agentx.MockEnvironment{
					Home:   "/home/test",
					Config: "/home/test/.config",
				}
			}

			path, err := agent.UserConfigPath(env)
			require.NoError(t, err)
			assert.NotEmpty(t, path, "UserConfigPath should return non-empty path")

			if spec.usesHomeConfig {
				assert.Equal(t, "/home/test/"+spec.configSubdir, path)
			} else {
				assert.Equal(t, "/home/test/.config/"+spec.configSubdir, path)
			}
		})
	}
}

func TestClineUserConfigPath_PlatformSpecific(t *testing.T) {
	agent := NewClineAgent()

	tests := []struct {
		name     string
		os       string
		expected string
	}{
		{"darwin", "darwin", "/home/test/Library/Application Support/Code/User/globalStorage/saoudrizwan.claude-dev"},
		{"linux", "linux", "/home/test/.config/Code/User/globalStorage/saoudrizwan.claude-dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := &agentx.MockEnvironment{Home: "/home/test", OS: tt.os}
			path, err := agent.UserConfigPath(env)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, path)
		})
	}
}

func TestDroidUserConfigPath_PlatformSpecific(t *testing.T) {
	agent := NewDroidAgent()

	tests := []struct {
		name     string
		os       string
		envVars  map[string]string
		expected string
	}{
		{"darwin default", "darwin", nil, "/home/test/.config/factory"},
		{"linux default", "linux", nil, "/home/test/.config/factory"},
		{"XDG override", "linux", map[string]string{"XDG_CONFIG_HOME": "/custom/config"}, "/custom/config/factory"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := &agentx.MockEnvironment{
				Home:    "/home/test",
				OS:      tt.os,
				EnvVars: tt.envVars,
			}
			path, err := agent.UserConfigPath(env)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, path)
		})
	}
}

// ---------------------------------------------------------------------------
// 4. TestAllAgents_ProjectConfigPath
// ---------------------------------------------------------------------------

func TestAllAgents_ProjectConfigPath(t *testing.T) {
	for _, spec := range allAgentSpecs() {
		t.Run(spec.name, func(t *testing.T) {
			agent := spec.constructor()
			assert.Equal(t, spec.projectConfig, agent.ProjectConfigPath())
		})
	}
}

// ---------------------------------------------------------------------------
// 5. TestAllAgents_ContextFiles
// ---------------------------------------------------------------------------

func TestAllAgents_ContextFiles(t *testing.T) {
	for _, spec := range allAgentSpecs() {
		t.Run(spec.name, func(t *testing.T) {
			agent := spec.constructor()
			files := agent.ContextFiles()

			assert.Equal(t, spec.contextFiles, files, "ContextFiles()")
			assert.NotEmpty(t, files, "ContextFiles() should return at least one file")
		})
	}
}

// ---------------------------------------------------------------------------
// 6. TestAllAgents_Capabilities
// ---------------------------------------------------------------------------

func TestAllAgents_Capabilities(t *testing.T) {
	for _, spec := range allAgentSpecs() {
		t.Run(spec.name, func(t *testing.T) {
			agent := spec.constructor()
			caps := agent.Capabilities()

			assert.Equal(t, spec.capabilities.Hooks, caps.Hooks, "Hooks")
			assert.Equal(t, spec.capabilities.MCPServers, caps.MCPServers, "MCPServers")
			assert.Equal(t, spec.capabilities.SystemPrompt, caps.SystemPrompt, "SystemPrompt")
			assert.Equal(t, spec.capabilities.ProjectContext, caps.ProjectContext, "ProjectContext")
			assert.Equal(t, spec.capabilities.CustomCommands, caps.CustomCommands, "CustomCommands")
			assert.Equal(t, spec.capabilities.MinVersion, caps.MinVersion, "MinVersion")
		})
	}
}

// ---------------------------------------------------------------------------
// 7. TestAllAgents_IsInstalled
// ---------------------------------------------------------------------------

func TestAllAgents_IsInstalled(t *testing.T) {
	ctx := context.Background()

	for _, spec := range allAgentSpecs() {
		t.Run(spec.name, func(t *testing.T) {
			agent := spec.constructor()

			// binary in PATH
			if spec.binaryName != "" {
				t.Run("binary in PATH", func(t *testing.T) {
					env := &agentx.MockEnvironment{
						Home:         "/home/test",
						Config:       "/home/test/.config",
						PathBinaries: map[string]string{spec.binaryName: "/usr/local/bin/" + spec.binaryName},
					}
					installed, err := agent.IsInstalled(ctx, env)
					require.NoError(t, err)
					assert.True(t, installed)
				})
			}

			// extra binary names (e.g., code-puppy)
			for _, bin := range spec.extraBinaries {
				t.Run("extra binary "+bin, func(t *testing.T) {
					env := &agentx.MockEnvironment{
						Home:         "/home/test",
						Config:       "/home/test/.config",
						PathBinaries: map[string]string{bin: "/usr/local/bin/" + bin},
					}
					installed, err := agent.IsInstalled(ctx, env)
					require.NoError(t, err)
					assert.True(t, installed)
				})
			}

			// macOS app bundle
			if spec.macOSApp != "" {
				t.Run("macOS app bundle", func(t *testing.T) {
					env := &agentx.MockEnvironment{
						Home:         "/home/test",
						Config:       "/home/test/.config",
						OS:           "darwin",
						ExistingDirs: map[string]bool{spec.macOSApp: true},
					}
					installed, err := agent.IsInstalled(ctx, env)
					require.NoError(t, err)
					assert.True(t, installed)
				})
			}

			// config dir exists
			t.Run("config dir exists", func(t *testing.T) {
				configPath := buildExpectedConfigPath(spec)
				if configPath == "" {
					t.Skip("agent has custom config path logic")
				}
				env := &agentx.MockEnvironment{
					Home:         "/home/test",
					Config:       "/home/test/.config",
					ExistingDirs: map[string]bool{configPath: true},
				}
				installed, err := agent.IsInstalled(ctx, env)
				require.NoError(t, err)
				assert.True(t, installed)
			})

			// nothing found
			t.Run("not installed", func(t *testing.T) {
				env := &agentx.MockEnvironment{
					Home:   "/home/test",
					Config: "/home/test/.config",
				}
				installed, err := agent.IsInstalled(ctx, env)
				require.NoError(t, err)
				assert.False(t, installed)
			})
		})
	}
}

// buildExpectedConfigPath computes the expected config path for an agent.
func buildExpectedConfigPath(spec agentSpec) string {
	if spec.configSubdir == "" {
		return ""
	}
	if spec.usesHomeConfig {
		return "/home/test/" + spec.configSubdir
	}
	return "/home/test/.config/" + spec.configSubdir
}

// ---------------------------------------------------------------------------
// 8. TestAllAgents_DetectVersion
// ---------------------------------------------------------------------------

func TestAllAgents_DetectVersion(t *testing.T) {
	ctx := context.Background()

	for _, spec := range allAgentSpecs() {
		t.Run(spec.name, func(t *testing.T) {
			agent := spec.constructor()

			if spec.versionCmd != "" {
				t.Run("from command", func(t *testing.T) {
					env := &agentx.MockEnvironment{
						Home: "/home/test",
						ExecOutputs: map[string][]byte{
							spec.versionCmd: []byte("2.5.1\n"),
						},
					}
					assert.Equal(t, "2.5.1", agent.DetectVersion(ctx, env))
				})
			}

			if spec.versionPkgJSON != "" {
				t.Run("from package.json", func(t *testing.T) {
					configBase := "/home/test/" + spec.configSubdir
					pkgPath := configBase + "/" + spec.versionPkgJSON
					env := &agentx.MockEnvironment{
						Home: "/home/test",
						Files: map[string][]byte{
							pkgPath: []byte(`{"version": "3.14.0"}`),
						},
					}
					assert.Equal(t, "3.14.0", agent.DetectVersion(ctx, env))
				})
			}

			// no version available
			t.Run("nothing found", func(t *testing.T) {
				env := &agentx.MockEnvironment{
					Home:   "/home/test",
					Config: "/home/test/.config",
				}
				assert.Equal(t, "", agent.DetectVersion(ctx, env))
			})
		})
	}
}

// ---------------------------------------------------------------------------
// 9. TestAllAgents_HookAndCommandManager
// ---------------------------------------------------------------------------

func TestAllAgents_HookAndCommandManager(t *testing.T) {
	for _, spec := range allAgentSpecs() {
		t.Run(spec.name, func(t *testing.T) {
			agent := spec.constructor()

			// default nil via the Agent interface
			assert.Nil(t, agent.HookManager(), "HookManager() should be nil by default")
			assert.Nil(t, agent.CommandManager(), "CommandManager() should be nil by default")
		})
	}
}

// TestAllAgents_SetHookAndCommandManager tests the SetHookManager/SetCommandManager
// methods on concrete agent types (these are not part of the Agent interface).
func TestAllAgents_SetHookAndCommandManager(t *testing.T) {
	// each concrete agent has SetHookManager/SetCommandManager; test them individually
	type hookSetter interface {
		SetHookManager(agentx.HookManager)
		HookManager() agentx.HookManager
	}
	type cmdSetter interface {
		SetCommandManager(agentx.CommandManager)
		CommandManager() agentx.CommandManager
	}

	agents := []struct {
		name  string
		agent agentx.Agent
	}{
		{"ClaudeCode", NewClaudeCodeAgent()},
		{"Cursor", NewCursorAgent()},
		{"Windsurf", NewWindsurfAgent()},
		{"Copilot", NewCopilotAgent()},
		{"Aider", NewAiderAgent()},
		{"Cody", NewCodyAgent()},
		{"Continue", NewContinueAgent()},
		{"CodePuppy", NewCodePuppyAgent()},
		{"Kiro", NewKiroAgent()},
		{"OpenCode", NewOpenCodeAgent()},
		{"Goose", NewGooseAgent()},
		{"Amp", NewAmpAgent()},
		{"Pi", NewPiAgent()},
		{"Cline", NewClineAgent()},
		{"Droid", NewDroidAgent()},
		{"Codex", NewCodexAgent()},
		{"Gemini CLI", NewGeminiAgent()},
	}

	for _, tt := range agents {
		t.Run(tt.name, func(t *testing.T) {
			if hs, ok := tt.agent.(hookSetter); ok {
				hs.SetHookManager(nil)
				assert.Nil(t, hs.HookManager())
			}
			if cs, ok := tt.agent.(cmdSetter); ok {
				cs.SetCommandManager(nil)
				assert.Nil(t, cs.CommandManager())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 10. TestAllAgents_SessionSupport
// ---------------------------------------------------------------------------

func TestAllAgents_SessionSupport(t *testing.T) {
	for _, spec := range allAgentSpecs() {
		t.Run(spec.name, func(t *testing.T) {
			agent := spec.constructor()

			assert.Equal(t, spec.supportsSession, agent.SupportsSession(), "SupportsSession()")

			emptyEnv := agentx.NewMockEnvironment(nil)

			if spec.sessionEnvVar != "" {
				// agent reads session ID from a specific env var
				envWithSession := agentx.NewMockEnvironment(map[string]string{
					spec.sessionEnvVar: "test_session_value",
				})
				assert.Equal(t, "test_session_value", agent.SessionID(envWithSession))
				assert.Equal(t, "", agent.SessionID(emptyEnv))
			} else {
				// agent returns empty (session comes from hook stdin or not supported)
				assert.Equal(t, "", agent.SessionID(emptyEnv))
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 11. TestLifecycleEventMapper
// ---------------------------------------------------------------------------

func TestLifecycleEventMapper(t *testing.T) {
	for _, spec := range allAgentSpecs() {
		if !spec.isLifecycleAgent {
			continue
		}

		t.Run(spec.name, func(t *testing.T) {
			agent := spec.constructor()

			// verify the agent implements LifecycleEventMapper
			mapper, ok := agent.(agentx.LifecycleEventMapper)
			require.True(t, ok, "%s should implement LifecycleEventMapper", spec.name)

			// verify EventPhases returns expected count
			phases := mapper.EventPhases()
			assert.Len(t, phases, spec.eventPhaseCount,
				"EventPhases() should have %d entries", spec.eventPhaseCount)

			// verify all phase values are valid (non-empty)
			for event, phase := range phases {
				assert.NotEmpty(t, event, "event key should be non-empty")
				assert.NotEmpty(t, phase, "phase value should be non-empty")
			}

			// verify AgentENVAliases
			aliases := mapper.AgentENVAliases()
			assert.Equal(t, spec.envAliases, aliases, "AgentENVAliases()")
			assert.NotEmpty(t, aliases, "AgentENVAliases() should return at least one alias")
		})
	}
}

func TestNonLifecycleAgents_DoNotImplementMapper(t *testing.T) {
	nonLifecycleAgents := []struct {
		name  string
		agent agentx.Agent
	}{
		{"Aider", NewAiderAgent()},
		{"Cody", NewCodyAgent()},
		{"Continue", NewContinueAgent()},
		{"CodePuppy", NewCodePuppyAgent()},
		{"Goose", NewGooseAgent()},
		{"Pi", NewPiAgent()},
		{"Codex", NewCodexAgent()},
		{"Gemini CLI", NewGeminiAgent()},
	}

	for _, tt := range nonLifecycleAgents {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := tt.agent.(agentx.LifecycleEventMapper)
			assert.False(t, ok, "%s should NOT implement LifecycleEventMapper", tt.name)
		})
	}
}

// ---------------------------------------------------------------------------
// 12. TestAllAgents_SupportsXDGConfig
// ---------------------------------------------------------------------------

func TestAllAgents_SupportsXDGConfig(t *testing.T) {
	for _, spec := range allAgentSpecs() {
		t.Run(spec.name, func(t *testing.T) {
			agent := spec.constructor()
			assert.Equal(t, spec.supportsXDG, agent.SupportsXDGConfig(), "SupportsXDGConfig()")
		})
	}
}

// ---------------------------------------------------------------------------
// 13. TestAllAgents_ImplementInterface (compile-time + runtime)
// ---------------------------------------------------------------------------

func TestAllAgents_ImplementInterface(t *testing.T) {
	for _, spec := range allAgentSpecs() {
		t.Run(spec.name, func(t *testing.T) {
			agent := spec.constructor()
			// runtime check that every constructor returns a valid Agent
			var _ agentx.Agent = agent
			assert.Implements(t, (*agentx.Agent)(nil), agent)
		})
	}
}

// ---------------------------------------------------------------------------
// 14. TestCodexDetect_DirectoryFallback (Codex-specific detection via .codex dir)
// ---------------------------------------------------------------------------

func TestCodexDetect_DirectoryFallback(t *testing.T) {
	ctx := context.Background()
	agent := NewCodexAgent()

	t.Run("cwd .codex directory", func(t *testing.T) {
		env := &agentx.MockEnvironment{
			ExistingDirs: map[string]bool{".codex": true},
		}
		detected, err := agent.Detect(ctx, env)
		require.NoError(t, err)
		assert.True(t, detected)
	})

	t.Run("PWD .codex directory", func(t *testing.T) {
		env := &agentx.MockEnvironment{
			EnvVars:      map[string]string{"PWD": "/repo"},
			ExistingDirs: map[string]bool{"/repo/.codex": true},
		}
		detected, err := agent.Detect(ctx, env)
		require.NoError(t, err)
		assert.True(t, detected)
	})
}

// ---------------------------------------------------------------------------
// 15. TestClineIsInstalled_VSCodeCheck
// ---------------------------------------------------------------------------

func TestClineIsInstalled_VSCodeCheck(t *testing.T) {
	ctx := context.Background()
	agent := NewClineAgent()

	t.Run("VS Code in PATH but no extension dir returns false", func(t *testing.T) {
		env := &agentx.MockEnvironment{
			Home:         "/home/test",
			OS:           "linux",
			PathBinaries: map[string]string{"code": "/usr/bin/code"},
		}
		installed, err := agent.IsInstalled(ctx, env)
		require.NoError(t, err)
		assert.False(t, installed, "VS Code in PATH alone is insufficient for Cline")
	})

	t.Run("extension storage dir exists", func(t *testing.T) {
		env := &agentx.MockEnvironment{
			Home: "/home/test",
			OS:   "linux",
			ExistingDirs: map[string]bool{
				"/home/test/.config/Code/User/globalStorage/saoudrizwan.claude-dev": true,
			},
		}
		installed, err := agent.IsInstalled(ctx, env)
		require.NoError(t, err)
		assert.True(t, installed)
	})
}
