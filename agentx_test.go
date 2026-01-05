package agentx_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/sageox/agentx"
	
)

func TestDetect_AGENT_ENV(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		wantType agentx.AgentType
	}{
		{
			name:     "claude via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "claude"},
			wantType: agentx.AgentTypeClaudeCode,
		},
		{
			name:     "cursor via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "cursor"},
			wantType: agentx.AgentTypeCursor,
		},
		{
			name:     "aider via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "aider"},
			wantType: agentx.AgentTypeAider,
		},
		{
			name:     "windsurf via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "windsurf"},
			wantType: agentx.AgentTypeWindsurf,
		},
		{
			name:     "copilot via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "copilot"},
			wantType: agentx.AgentTypeCopilot,
		},
		{
			name:     "cody via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "cody"},
			wantType: agentx.AgentTypeCody,
		},
		{
			name:     "continue via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "continue"},
			wantType: agentx.AgentTypeContinue,
		},
		{
			name:     "code-puppy via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "code-puppy"},
			wantType: agentx.AgentTypeCodePuppy,
		},
		{
			name:     "kiro via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "kiro"},
			wantType: agentx.AgentTypeKiro,
		},
		{
			name:     "opencode via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "opencode"},
			wantType: agentx.AgentTypeOpenCode,
		},
		{
			name:     "goose via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "goose"},
			wantType: agentx.AgentTypeGoose,
		},
		{
			name:     "amp via AGENT_ENV",
			envVars:  map[string]string{"AGENT_ENV": "amp"},
			wantType: agentx.AgentTypeAmp,
		},
		{
			name:     "no agent detected",
			envVars:  map[string]string{},
			wantType: agentx.AgentTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := agentx.NewMockEnvironment(tt.envVars)
			agent := agentx.DetectWithEnv(env)

			if tt.wantType == agentx.AgentTypeUnknown {
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
		wantType agentx.AgentType
	}{
		{
			name:     "CLAUDECODE=1",
			envVars:  map[string]string{"CLAUDECODE": "1"},
			wantType: agentx.AgentTypeClaudeCode,
		},
		{
			name:     "CURSOR_AGENT=1",
			envVars:  map[string]string{"CURSOR_AGENT": "1"},
			wantType: agentx.AgentTypeCursor,
		},
		{
			name:     "AIDER=1",
			envVars:  map[string]string{"AIDER": "1"},
			wantType: agentx.AgentTypeAider,
		},
		{
			name:     "WINDSURF_AGENT=1",
			envVars:  map[string]string{"WINDSURF_AGENT": "1"},
			wantType: agentx.AgentTypeWindsurf,
		},
		{
			name:     "CODEIUM_AGENT=1",
			envVars:  map[string]string{"CODEIUM_AGENT": "1"},
			wantType: agentx.AgentTypeWindsurf,
		},
		{
			name:     "COPILOT_AGENT=1",
			envVars:  map[string]string{"COPILOT_AGENT": "1"},
			wantType: agentx.AgentTypeCopilot,
		},
		{
			name:     "CODY_AGENT=1",
			envVars:  map[string]string{"CODY_AGENT": "1"},
			wantType: agentx.AgentTypeCody,
		},
		{
			name:     "CONTINUE_AGENT=1",
			envVars:  map[string]string{"CONTINUE_AGENT": "1"},
			wantType: agentx.AgentTypeContinue,
		},
		{
			name:     "CODE_PUPPY=1",
			envVars:  map[string]string{"CODE_PUPPY": "1"},
			wantType: agentx.AgentTypeCodePuppy,
		},
		{
			name:     "KIRO=1",
			envVars:  map[string]string{"KIRO": "1"},
			wantType: agentx.AgentTypeKiro,
		},
		{
			name:     "OPENCODE=1",
			envVars:  map[string]string{"OPENCODE": "1"},
			wantType: agentx.AgentTypeOpenCode,
		},
		{
			name:     "GOOSE=1",
			envVars:  map[string]string{"GOOSE": "1"},
			wantType: agentx.AgentTypeGoose,
		},
		{
			name:     "AMP=1",
			envVars:  map[string]string{"AMP": "1"},
			wantType: agentx.AgentTypeAmp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := agentx.NewMockEnvironment(tt.envVars)
			agent := agentx.DetectWithEnv(env)

			if agent == nil {
				t.Fatalf("expected agent %s, got nil", tt.wantType)
			}

			if agent.Type() != tt.wantType {
				t.Errorf("Type() = %s, want %s", agent.Type(), tt.wantType)
			}
		})
	}
}

func TestDetect_HeuristicDetection(t *testing.T) {
	t.Run("cursor detected via $_ exec path", func(t *testing.T) {
		env := agentx.NewMockEnvironment(map[string]string{
			"_": "/Applications/Cursor.app/Contents/MacOS/cursor",
		})
		agent := agentx.DetectWithEnv(env)

		if agent == nil {
			t.Fatal("expected cursor agent to be detected via $_ heuristic")
		}
		if agent.Type() != agentx.AgentTypeCursor {
			t.Errorf("Type() = %s, want %s", agent.Type(), agentx.AgentTypeCursor)
		}
	})

	t.Run("cursor detected via lowercase $_ path", func(t *testing.T) {
		env := agentx.NewMockEnvironment(map[string]string{
			"_": "/usr/local/bin/cursor",
		})
		agent := agentx.DetectWithEnv(env)

		if agent == nil {
			t.Fatal("expected cursor agent to be detected via $_ heuristic")
		}
		if agent.Type() != agentx.AgentTypeCursor {
			t.Errorf("Type() = %s, want %s", agent.Type(), agentx.AgentTypeCursor)
		}
	})

	t.Run("no detection from unrelated $_ path", func(t *testing.T) {
		env := agentx.NewMockEnvironment(map[string]string{
			"_": "/usr/bin/bash",
		})
		agent := agentx.DetectWithEnv(env)

		if agent != nil {
			t.Errorf("expected nil agent, got %s", agent.Type())
		}
	})
}

func TestIsAgentContext(t *testing.T) {
	t.Run("returns true when agent detected", func(t *testing.T) {
		env := agentx.NewMockEnvironment(map[string]string{"AGENT_ENV": "claude"})
		result := agentx.IsAgentContextWithEnv(env)
		if !result {
			t.Error("IsAgentContextWithEnv() = false, want true")
		}
	})

	t.Run("returns false when no agent", func(t *testing.T) {
		env := agentx.NewMockEnvironment(map[string]string{})
		result := agentx.IsAgentContextWithEnv(env)
		if result {
			t.Error("IsAgentContextWithEnv() = true, want false")
		}
	})

	t.Run("returns true for native env var", func(t *testing.T) {
		env := agentx.NewMockEnvironment(map[string]string{"CLAUDECODE": "1"})
		result := agentx.IsAgentContextWithEnv(env)
		if !result {
			t.Error("IsAgentContextWithEnv() = false, want true for CLAUDECODE=1")
		}
	})
}

func TestRequireAgent(t *testing.T) {
	t.Run("returns empty when agent present", func(t *testing.T) {
		env := agentx.NewMockEnvironment(map[string]string{"AGENT_ENV": "cursor"})
		msg := agentx.RequireAgentWithEnv("my-command", env)
		if msg != "" {
			t.Errorf("RequireAgentWithEnv() = %q, want empty string", msg)
		}
	})

	t.Run("returns error message when no agent", func(t *testing.T) {
		env := agentx.NewMockEnvironment(map[string]string{})
		msg := agentx.RequireAgentWithEnv("my-command", env)
		if msg == "" {
			t.Error("RequireAgentWithEnv() = empty, want error message")
		}
		if !strings.Contains(msg, "my-command") {
			t.Errorf("error message should contain command name, got %q", msg)
		}
		if !strings.Contains(msg, "AGENT_ENV") {
			t.Errorf("error message should mention AGENT_ENV, got %q", msg)
		}
	})
}

func TestRegistry(t *testing.T) {
	t.Run("List returns all supported agents", func(t *testing.T) {
		agents := agentx.DefaultRegistry.List()
		if len(agents) != len(agentx.SupportedAgents) {
			t.Errorf("List() returned %d agents, want %d (len(SupportedAgents))", len(agents), len(agentx.SupportedAgents))
		}
	})

	t.Run("all SupportedAgents are registered", func(t *testing.T) {
		for _, agentType := range agentx.SupportedAgents {
			_, ok := agentx.DefaultRegistry.Get(agentType)
			if !ok {
				t.Errorf("SupportedAgent %s is not registered in DefaultRegistry", agentType)
			}
		}
	})

	t.Run("Get returns agent by type", func(t *testing.T) {
		agent, ok := agentx.DefaultRegistry.Get(agentx.AgentTypeClaudeCode)
		if !ok {
			t.Fatal("Get(agentx.AgentTypeClaudeCode) returned false")
		}
		if agent.Type() != agentx.AgentTypeClaudeCode {
			t.Errorf("Type() = %s, want %s", agent.Type(), agentx.AgentTypeClaudeCode)
		}
		if agent.Name() != "Claude Code" {
			t.Errorf("Name() = %s, want 'Claude Code'", agent.Name())
		}
	})

	t.Run("Get returns false for unknown type", func(t *testing.T) {
		_, ok := agentx.DefaultRegistry.Get("nonexistent")
		if ok {
			t.Error("Get('nonexistent') returned true, want false")
		}
	})
}

func TestRegistry_Concurrent(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	iterations := 100

	// concurrent reads should not race
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = agentx.DefaultRegistry.List()
			_, _ = agentx.DefaultRegistry.Get(agentx.AgentTypeClaudeCode)
			_, _ = agentx.DefaultRegistry.Get(agentx.AgentTypeCursor)
		}()
	}
	wg.Wait()
}

func TestAgentContextFiles(t *testing.T) {
	t.Run("all agents have context files", func(t *testing.T) {
		for _, agentType := range agentx.SupportedAgents {
			agent, ok := agentx.DefaultRegistry.Get(agentType)
			if !ok {
				t.Fatalf("agent %s not registered", agentType)
			}

			files := agent.ContextFiles()
			// context files should be non-empty for most agents
			// some agents may legitimately have no context files
			for _, f := range files {
				if f == "" {
					t.Errorf("agent %s has empty context file entry", agentType)
				}
				// verify context files have reasonable names (contain dot for extension or known patterns)
				if !strings.Contains(f, ".") && !strings.HasSuffix(f, "rules") {
					t.Errorf("agent %s context file %q appears to be missing extension", agentType, f)
				}
			}
		}
	})

	// spot check specific agents that have well-known context files
	t.Run("claude has CLAUDE.md", func(t *testing.T) {
		agent, _ := agentx.DefaultRegistry.Get(agentx.AgentTypeClaudeCode)
		files := agent.ContextFiles()
		hasClaudeMD := false
		for _, f := range files {
			if f == "CLAUDE.md" {
				hasClaudeMD = true
				break
			}
		}
		if !hasClaudeMD {
			t.Errorf("Claude Code should have CLAUDE.md in context files, got %v", files)
		}
	})

	t.Run("cursor has .cursorrules", func(t *testing.T) {
		agent, _ := agentx.DefaultRegistry.Get(agentx.AgentTypeCursor)
		files := agent.ContextFiles()
		hasCursorRules := false
		for _, f := range files {
			if f == ".cursorrules" {
				hasCursorRules = true
				break
			}
		}
		if !hasCursorRules {
			t.Errorf("Cursor should have .cursorrules in context files, got %v", files)
		}
	})
}

func TestAgentConfigPaths(t *testing.T) {
	env := agentx.NewMockEnvironment(nil)
	env.Home = "/home/user"
	env.Config = "/home/user/.config"

	tests := []struct {
		agentType      agentx.AgentType
		wantUserConfig string
		wantProjConfig string
	}{
		{agentx.AgentTypeClaudeCode, "/home/user/.claude", ".claude"},
		{agentx.AgentTypeCursor, "/home/user/.cursor", ".cursor"},
		{agentx.AgentTypeAider, "/home/user/.aider", ".aider"},
		{agentx.AgentTypeCody, "/home/user/.config/cody", ".cody"},
	}

	for _, tt := range tests {
		t.Run(string(tt.agentType), func(t *testing.T) {
			agent, ok := agentx.DefaultRegistry.Get(tt.agentType)
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

func TestAgentConfigPaths_ErrorHandling(t *testing.T) {
	t.Run("UserConfigPath returns error when HomeDir fails", func(t *testing.T) {
		env := agentx.NewMockEnvironment(nil)
		env.HomeError = errors.New("home directory not found")

		agent := agentx.NewClaudeCodeAgent()
		_, err := agent.UserConfigPath(env)
		if err == nil {
			t.Error("UserConfigPath() should return error when HomeDir fails")
		}
	})
}

func TestAgentIsInstalled(t *testing.T) {
	ctx := context.Background()

	t.Run("detects binary in PATH", func(t *testing.T) {
		env := agentx.NewMockEnvironment(nil)
		env.PathBinaries = map[string]string{"claude": "/usr/local/bin/claude"}

		agent := agentx.NewClaudeCodeAgent()
		installed, err := agent.IsInstalled(ctx, env)
		if err != nil {
			t.Fatalf("IsInstalled() error: %v", err)
		}
		if !installed {
			t.Error("IsInstalled() = false, want true (binary in PATH)")
		}
	})

	t.Run("detects config directory", func(t *testing.T) {
		env := agentx.NewMockEnvironment(nil)
		env.Home = "/home/user"
		env.ExistingDirs = map[string]bool{"/home/user/.claude": true}

		agent := agentx.NewClaudeCodeAgent()
		installed, err := agent.IsInstalled(ctx, env)
		if err != nil {
			t.Fatalf("IsInstalled() error: %v", err)
		}
		if !installed {
			t.Error("IsInstalled() = false, want true (config dir exists)")
		}
	})

	t.Run("not installed", func(t *testing.T) {
		env := agentx.NewMockEnvironment(nil)
		env.Home = "/home/user"

		agent := agentx.NewClaudeCodeAgent()
		installed, err := agent.IsInstalled(ctx, env)
		if err != nil {
			t.Fatalf("IsInstalled() error: %v", err)
		}
		if installed {
			t.Error("IsInstalled() = true, want false")
		}
	})

	t.Run("returns false gracefully when HomeDir errors", func(t *testing.T) {
		env := agentx.NewMockEnvironment(nil)
		env.HomeError = errors.New("home directory not found")

		agent := agentx.NewClaudeCodeAgent()
		installed, err := agent.IsInstalled(ctx, env)
		if err != nil {
			t.Fatalf("IsInstalled() should not return error, got: %v", err)
		}
		if installed {
			t.Error("IsInstalled() = true, want false when HomeDir errors")
		}
	})
}

func TestCursorIsInstalled_MacOS(t *testing.T) {
	ctx := context.Background()

	t.Run("detects macOS application bundle", func(t *testing.T) {
		env := agentx.NewMockEnvironment(nil)
		env.OS = "darwin"
		env.ExistingDirs = map[string]bool{"/Applications/Cursor.app": true}

		agent := agentx.NewCursorAgent()
		installed, err := agent.IsInstalled(ctx, env)
		if err != nil {
			t.Fatalf("IsInstalled() error: %v", err)
		}
		if !installed {
			t.Error("IsInstalled() = false, want true (macOS app bundle exists)")
		}
	})
}

func TestInit(t *testing.T) {
	// save and restore AGENT_ENV
	originalAgentEnv := os.Getenv("AGENT_ENV")
	defer os.Setenv("AGENT_ENV", originalAgentEnv)

	t.Run("sets AGENT_ENV when agent detected", func(t *testing.T) {
		os.Unsetenv("AGENT_ENV")

		env := agentx.NewMockEnvironment(map[string]string{
			"CURSOR_AGENT": "1",
		})
		agent := agentx.InitWithEnv(env)

		if agent == nil {
			t.Fatal("expected agent to be detected")
		}
		if agent.Type() != agentx.AgentTypeCursor {
			t.Errorf("Type() = %s, want cursor", agent.Type())
		}

		// verify AGENT_ENV was set
		if got := os.Getenv("AGENT_ENV"); got != "cursor" {
			t.Errorf("AGENT_ENV = %q, want %q", got, "cursor")
		}
	})

	t.Run("respects existing AGENT_ENV in environment", func(t *testing.T) {
		os.Unsetenv("AGENT_ENV")

		env := agentx.NewMockEnvironment(map[string]string{
			"AGENT_ENV":    "claude",
			"CURSOR_AGENT": "1", // would detect cursor without AGENT_ENV
		})
		agent := agentx.InitWithEnv(env)

		if agent == nil {
			t.Fatal("expected agent to be detected")
		}
		// should respect existing AGENT_ENV=claude, not override with cursor
		if agent.Type() != agentx.AgentTypeClaudeCode {
			t.Errorf("Type() = %s, want claude (should respect existing AGENT_ENV)", agent.Type())
		}
	})

	t.Run("does not set AGENT_ENV when no agent detected", func(t *testing.T) {
		os.Unsetenv("AGENT_ENV")

		env := agentx.NewMockEnvironment(map[string]string{})
		agent := agentx.InitWithEnv(env)

		if agent != nil {
			t.Errorf("expected nil agent, got %s", agent.Type())
		}

		// AGENT_ENV should remain unset
		if got := os.Getenv("AGENT_ENV"); got != "" {
			t.Errorf("AGENT_ENV = %q, want empty", got)
		}
	})
}
