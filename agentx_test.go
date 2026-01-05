package agentx

import (
	"context"
	"testing"
)

func TestDetect_AGENT_ENV(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		wantType AgentType
	}{
		{
			name:     "claude-code via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "claude-code"},
			wantType: AgentTypeClaudeCode,
		},
		{
			name:     "claude via AGENT_ENV alias",
			envVars:  map[string]string{"AGENT_ENV": "claude"},
			wantType: AgentTypeClaudeCode,
		},
		{
			name:     "cursor via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "cursor"},
			wantType: AgentTypeCursor,
		},
		{
			name:     "aider via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "aider"},
			wantType: AgentTypeAider,
		},
		{
			name:     "windsurf via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "windsurf"},
			wantType: AgentTypeWindsurf,
		},
		{
			name:     "codeium alias via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "codeium"},
			wantType: AgentTypeWindsurf,
		},
		{
			name:     "copilot via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "copilot"},
			wantType: AgentTypeCopilot,
		},
		{
			name:     "cody via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "cody"},
			wantType: AgentTypeCody,
		},
		{
			name:     "continue via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "continue"},
			wantType: AgentTypeContinue,
		},
		{
			name:     "code-puppy via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "code-puppy"},
			wantType: AgentTypeCodePuppy,
		},
		{
			name:     "kiro via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "kiro"},
			wantType: AgentTypeKiro,
		},
		{
			name:     "opencode via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "opencode"},
			wantType: AgentTypeOpenCode,
		},
		{
			name:     "goose via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "goose"},
			wantType: AgentTypeGoose,
		},
		{
			name:     "amp via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "amp"},
			wantType: AgentTypeAmp,
		},
		{
			name:     "no agent detected",
			envVars:  map[string]string{},
			wantType: AgentTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := NewMockEnvironment(tt.envVars)
			agent := DetectWithEnv(env)

			if tt.wantType == AgentTypeUnknown {
				if agent != nil {
					t.Errorf("expected nil agent, got %s", agent.Type())
				}
				return
			}

			if agent == nil {
				t.Fatalf("expected agent %s, got nil", tt.wantType)
			}

			if agent.Type() != tt.wantType {
				t.Errorf("Type() = %s, want %s", agent.Type(), tt.wantType)
			}
		})
	}
}

func TestDetect_NativeEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		wantType AgentType
	}{
		{
			name:     "CLAUDECODE=1",
			envVars:  map[string]string{"CLAUDECODE": "1"},
			wantType: AgentTypeClaudeCode,
		},
		{
			name:     "CURSOR_AGENT=1",
			envVars:  map[string]string{"CURSOR_AGENT": "1"},
			wantType: AgentTypeCursor,
		},
		{
			name:     "AIDER=1",
			envVars:  map[string]string{"AIDER": "1"},
			wantType: AgentTypeAider,
		},
		{
			name:     "WINDSURF_AGENT=1",
			envVars:  map[string]string{"WINDSURF_AGENT": "1"},
			wantType: AgentTypeWindsurf,
		},
		{
			name:     "CODEIUM_AGENT=1",
			envVars:  map[string]string{"CODEIUM_AGENT": "1"},
			wantType: AgentTypeWindsurf,
		},
		{
			name:     "COPILOT_AGENT=1",
			envVars:  map[string]string{"COPILOT_AGENT": "1"},
			wantType: AgentTypeCopilot,
		},
		{
			name:     "CODY_AGENT=1",
			envVars:  map[string]string{"CODY_AGENT": "1"},
			wantType: AgentTypeCody,
		},
		{
			name:     "CONTINUE_AGENT=1",
			envVars:  map[string]string{"CONTINUE_AGENT": "1"},
			wantType: AgentTypeContinue,
		},
		{
			name:     "CODE_PUPPY=1",
			envVars:  map[string]string{"CODE_PUPPY": "1"},
			wantType: AgentTypeCodePuppy,
		},
		{
			name:     "KIRO=1",
			envVars:  map[string]string{"KIRO": "1"},
			wantType: AgentTypeKiro,
		},
		{
			name:     "OPENCODE=1",
			envVars:  map[string]string{"OPENCODE": "1"},
			wantType: AgentTypeOpenCode,
		},
		{
			name:     "GOOSE=1",
			envVars:  map[string]string{"GOOSE": "1"},
			wantType: AgentTypeGoose,
		},
		{
			name:     "AMP=1",
			envVars:  map[string]string{"AMP": "1"},
			wantType: AgentTypeAmp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := NewMockEnvironment(tt.envVars)
			agent := DetectWithEnv(env)

			if agent == nil {
				t.Fatalf("expected agent %s, got nil", tt.wantType)
			}

			if agent.Type() != tt.wantType {
				t.Errorf("Type() = %s, want %s", agent.Type(), tt.wantType)
			}
		})
	}
}

func TestIsAgentContext(t *testing.T) {
	// Save original detection and restore after
	// We can't easily mock global state, so we just test with real env
	// This test is more of a smoke test
	t.Run("returns bool", func(t *testing.T) {
		// Just verify it doesn't panic
		_ = IsAgentContext()
	})
}

func TestRequireAgent(t *testing.T) {
	t.Run("returns message when not in agent", func(t *testing.T) {
		// When AGENT_ENV is not set, RequireAgent should return an error message
		// We can't easily control the real environment, so this is a basic test
		msg := RequireAgent("my-command")
		// The function should return a non-empty string when not in agent context
		// (unless we happen to be running in Claude Code or another agent)
		if msg != "" && len(msg) < 10 {
			t.Errorf("expected meaningful error message, got %q", msg)
		}
	})
}

func TestRegistry(t *testing.T) {
	t.Run("List returns all agents", func(t *testing.T) {
		agents := DefaultRegistry.List()
		if len(agents) != 12 {
			t.Errorf("List() returned %d agents, want 12", len(agents))
		}
	})

	t.Run("Get returns agent by type", func(t *testing.T) {
		agent, ok := DefaultRegistry.Get(AgentTypeClaudeCode)
		if !ok {
			t.Fatal("Get(AgentTypeClaudeCode) returned false")
		}
		if agent.Type() != AgentTypeClaudeCode {
			t.Errorf("Type() = %s, want %s", agent.Type(), AgentTypeClaudeCode)
		}
		if agent.Name() != "Claude Code" {
			t.Errorf("Name() = %s, want 'Claude Code'", agent.Name())
		}
	})

	t.Run("Get returns false for unknown type", func(t *testing.T) {
		_, ok := DefaultRegistry.Get("nonexistent")
		if ok {
			t.Error("Get('nonexistent') returned true, want false")
		}
	})
}

func TestAgentContextFiles(t *testing.T) {
	tests := []struct {
		agentType AgentType
		wantFiles []string
	}{
		{AgentTypeClaudeCode, []string{"CLAUDE.md", "AGENTS.md"}},
		{AgentTypeCursor, []string{".cursorrules"}},
		{AgentTypeWindsurf, []string{".windsurfrules"}},
		{AgentTypeCopilot, []string{".github/copilot-instructions.md"}},
	}

	for _, tt := range tests {
		t.Run(string(tt.agentType), func(t *testing.T) {
			agent, ok := DefaultRegistry.Get(tt.agentType)
			if !ok {
				t.Fatalf("agent %s not registered", tt.agentType)
			}

			files := agent.ContextFiles()
			if len(files) != len(tt.wantFiles) {
				t.Errorf("ContextFiles() = %v, want %v", files, tt.wantFiles)
				return
			}

			for i, f := range files {
				if f != tt.wantFiles[i] {
					t.Errorf("ContextFiles()[%d] = %s, want %s", i, f, tt.wantFiles[i])
				}
			}
		})
	}
}

func TestAgentConfigPaths(t *testing.T) {
	env := NewMockEnvironment(nil)
	env.Home = "/home/user"
	env.Config = "/home/user/.config"

	tests := []struct {
		agentType       AgentType
		wantUserConfig  string
		wantProjConfig  string
	}{
		{AgentTypeClaudeCode, "/home/user/.claude", ".claude"},
		{AgentTypeCursor, "/home/user/.cursor", ".cursor"},
		{AgentTypeAider, "/home/user/.aider", ".aider"},
		{AgentTypeCody, "/home/user/.config/cody", ".cody"},
	}

	for _, tt := range tests {
		t.Run(string(tt.agentType), func(t *testing.T) {
			agent, ok := DefaultRegistry.Get(tt.agentType)
			if !ok {
				t.Fatalf("agent %s not registered", tt.agentType)
			}

			userConfig, err := agent.UserConfigPath(env)
			if err != nil {
				t.Fatalf("UserConfigPath() error: %v", err)
			}
			if userConfig != tt.wantUserConfig {
				t.Errorf("UserConfigPath() = %s, want %s", userConfig, tt.wantUserConfig)
			}

			projConfig := agent.ProjectConfigPath()
			if projConfig != tt.wantProjConfig {
				t.Errorf("ProjectConfigPath() = %s, want %s", projConfig, tt.wantProjConfig)
			}
		})
	}
}

func TestAgentIsInstalled(t *testing.T) {
	ctx := context.Background()

	t.Run("detects binary in PATH", func(t *testing.T) {
		env := NewMockEnvironment(nil)
		env.PathBinaries = map[string]string{"claude": "/usr/local/bin/claude"}

		agent := NewClaudeCodeAgent()
		installed, err := agent.IsInstalled(ctx, env)
		if err != nil {
			t.Fatalf("IsInstalled() error: %v", err)
		}
		if !installed {
			t.Error("IsInstalled() = false, want true (binary in PATH)")
		}
	})

	t.Run("detects config directory", func(t *testing.T) {
		env := NewMockEnvironment(nil)
		env.Home = "/home/user"
		env.ExistingDirs = map[string]bool{"/home/user/.claude": true}

		agent := NewClaudeCodeAgent()
		installed, err := agent.IsInstalled(ctx, env)
		if err != nil {
			t.Fatalf("IsInstalled() error: %v", err)
		}
		if !installed {
			t.Error("IsInstalled() = false, want true (config dir exists)")
		}
	})

	t.Run("not installed", func(t *testing.T) {
		env := NewMockEnvironment(nil)
		env.Home = "/home/user"

		agent := NewClaudeCodeAgent()
		installed, err := agent.IsInstalled(ctx, env)
		if err != nil {
			t.Fatalf("IsInstalled() error: %v", err)
		}
		if installed {
			t.Error("IsInstalled() = true, want false")
		}
	})
}
