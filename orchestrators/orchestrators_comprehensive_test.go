package orchestrators

import (
	"context"
	"errors"
	"os/exec"
	"testing"

	"github.com/sageox/agentx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- ConductorAgent additional tests ---

func TestConductorAgent_UserConfigPath(t *testing.T) {
	agent := NewConductorAgent()
	env := agentx.NewMockEnvironment(nil)
	path, err := agent.UserConfigPath(env)
	require.NoError(t, err)
	assert.Equal(t, "/home/test/Library/Application Support/com.conductor.app", path)
}

func TestConductorAgent_UserConfigPath_HomeError(t *testing.T) {
	agent := NewConductorAgent()
	env := agentx.NewMockEnvironment(nil)
	env.HomeError = errors.New("no home")
	_, err := agent.UserConfigPath(env)
	assert.Error(t, err)
}

func TestConductorAgent_ProjectConfigPath(t *testing.T) {
	agent := NewConductorAgent()
	assert.Equal(t, ".context", agent.ProjectConfigPath())
}

func TestConductorAgent_ContextFiles(t *testing.T) {
	agent := NewConductorAgent()
	assert.Nil(t, agent.ContextFiles())
}

func TestConductorAgent_Capabilities(t *testing.T) {
	agent := NewConductorAgent()
	caps := agent.Capabilities()
	assert.False(t, caps.Hooks)
	assert.False(t, caps.MCPServers)
	assert.False(t, caps.SystemPrompt)
	assert.False(t, caps.ProjectContext)
	assert.False(t, caps.CustomCommands)
}

func TestConductorAgent_IsInstalled(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{"CONDUCTOR_BIN_DIR set", map[string]string{"CONDUCTOR_BIN_DIR": "/usr/local/bin"}, true},
		{"CFBundleIdentifier match", map[string]string{"__CFBundleIdentifier": "com.conductor.app"}, true},
		{"nothing set", map[string]string{}, false},
		{"wrong CFBundleIdentifier", map[string]string{"__CFBundleIdentifier": "com.other.app"}, false},
	}

	agent := NewConductorAgent()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := agentx.NewMockEnvironment(tt.env)
			got, err := agent.IsInstalled(ctx, env)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConductorAgent_SupportsSession(t *testing.T) {
	agent := NewConductorAgent()
	assert.False(t, agent.SupportsSession())
}

func TestConductorAgent_SessionID(t *testing.T) {
	agent := NewConductorAgent()
	env := agentx.NewMockEnvironment(nil)
	assert.Equal(t, "", agent.SessionID(env))
}

func TestConductorAgent_HookManager(t *testing.T) {
	agent := NewConductorAgent()
	assert.Nil(t, agent.HookManager())
}

func TestConductorAgent_CommandManager(t *testing.T) {
	agent := NewConductorAgent()
	assert.Nil(t, agent.CommandManager())
}

func TestConductorAgent_DetectVersion(t *testing.T) {
	agent := NewConductorAgent()
	ctx := context.Background()
	env := agentx.NewMockEnvironment(nil)
	assert.Equal(t, "", agent.DetectVersion(ctx, env))
}

func TestConductorAgent_URL(t *testing.T) {
	agent := NewConductorAgent()
	assert.Equal(t, "https://conductor.build", agent.URL())
}

func TestConductorAgent_SupportsXDGConfig(t *testing.T) {
	agent := NewConductorAgent()
	assert.False(t, agent.SupportsXDGConfig())
}

// --- OpenClawAgent comprehensive tests ---

func TestOpenClawAgent_Detect_Comprehensive(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{"ORCHESTRATOR_ENV=openclaw", map[string]string{"ORCHESTRATOR_ENV": "openclaw"}, true},
		{"OPENCLAW_STATE_DIR set", map[string]string{"OPENCLAW_STATE_DIR": "/tmp/state"}, true},
		{"OPENCLAW_HOME set", map[string]string{"OPENCLAW_HOME": "/opt/openclaw"}, true},
		{"OPENCLAW_GATEWAY_TOKEN set", map[string]string{"OPENCLAW_GATEWAY_TOKEN": "tok-123"}, true},
		{"no env vars", map[string]string{}, false},
		{"wrong ORCHESTRATOR_ENV", map[string]string{"ORCHESTRATOR_ENV": "conductor"}, false},
		{"unrelated env vars", map[string]string{"HOME": "/home/user", "PATH": "/usr/bin"}, false},
	}

	agent := NewOpenClawAgent()
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

func TestOpenClawAgent_UserConfigPath_WithStateDir(t *testing.T) {
	agent := NewOpenClawAgent()
	env := agentx.NewMockEnvironment(map[string]string{
		"OPENCLAW_STATE_DIR": "/custom/state",
	})
	path, err := agent.UserConfigPath(env)
	require.NoError(t, err)
	assert.Equal(t, "/custom/state", path)
}

func TestOpenClawAgent_UserConfigPath_Default(t *testing.T) {
	agent := NewOpenClawAgent()
	env := agentx.NewMockEnvironment(nil)
	path, err := agent.UserConfigPath(env)
	require.NoError(t, err)
	assert.Equal(t, "/home/test/.openclaw", path)
}

func TestOpenClawAgent_UserConfigPath_HomeError(t *testing.T) {
	agent := NewOpenClawAgent()
	env := agentx.NewMockEnvironment(nil)
	env.HomeError = errors.New("no home")
	_, err := agent.UserConfigPath(env)
	assert.Error(t, err)
}

func TestOpenClawAgent_ProjectConfigPath(t *testing.T) {
	agent := NewOpenClawAgent()
	assert.Equal(t, "", agent.ProjectConfigPath())
}

func TestOpenClawAgent_ContextFiles(t *testing.T) {
	agent := NewOpenClawAgent()
	files := agent.ContextFiles()
	assert.Equal(t, []string{"AGENTS.md"}, files)
}

func TestOpenClawAgent_Capabilities(t *testing.T) {
	agent := NewOpenClawAgent()
	caps := agent.Capabilities()
	assert.True(t, caps.MCPServers)
	assert.True(t, caps.SystemPrompt)
	assert.True(t, caps.ProjectContext)
	assert.False(t, caps.Hooks)
	assert.False(t, caps.CustomCommands)
}

func TestOpenClawAgent_IsInstalled_InPath(t *testing.T) {
	agent := NewOpenClawAgent()
	env := agentx.NewMockEnvironment(nil)
	env.PathBinaries = map[string]string{"openclaw": "/usr/local/bin/openclaw"}

	installed, err := agent.IsInstalled(context.Background(), env)
	assert.NoError(t, err)
	assert.True(t, installed)
}

func TestOpenClawAgent_IsInstalled_DirExists(t *testing.T) {
	agent := NewOpenClawAgent()
	env := agentx.NewMockEnvironment(nil)
	env.ExistingDirs = map[string]bool{"/home/test/.openclaw": true}

	installed, err := agent.IsInstalled(context.Background(), env)
	assert.NoError(t, err)
	assert.True(t, installed)
}

func TestOpenClawAgent_IsInstalled_Nothing(t *testing.T) {
	agent := NewOpenClawAgent()
	env := agentx.NewMockEnvironment(nil)

	installed, err := agent.IsInstalled(context.Background(), env)
	assert.NoError(t, err)
	assert.False(t, installed)
}

func TestOpenClawAgent_IsInstalled_HomeError(t *testing.T) {
	agent := NewOpenClawAgent()
	env := agentx.NewMockEnvironment(nil)
	env.HomeError = errors.New("no home")
	// no binary in path either

	installed, err := agent.IsInstalled(context.Background(), env)
	assert.NoError(t, err)
	assert.False(t, installed)
}

func TestOpenClawAgent_DetectVersion_Success(t *testing.T) {
	agent := NewOpenClawAgent()
	env := agentx.NewMockEnvironment(nil)
	env.ExecOutputs = map[string][]byte{
		"openclaw": []byte("openclaw version 1.2.3\n"),
	}

	version := agent.DetectVersion(context.Background(), env)
	assert.Equal(t, "1.2.3", version)
}

func TestOpenClawAgent_DetectVersion_CommandNotFound(t *testing.T) {
	agent := NewOpenClawAgent()
	env := agentx.NewMockEnvironment(nil)
	env.ExecErrors = map[string]error{
		"openclaw": exec.ErrNotFound,
	}

	version := agent.DetectVersion(context.Background(), env)
	assert.Equal(t, "", version)
}

func TestOpenClawAgent_DetectVersion_NoSemver(t *testing.T) {
	agent := NewOpenClawAgent()
	env := agentx.NewMockEnvironment(nil)
	env.ExecOutputs = map[string][]byte{
		"openclaw": []byte("openclaw development build\n"),
	}

	version := agent.DetectVersion(context.Background(), env)
	assert.Equal(t, "", version)
}

func TestOpenClawAgent_HookManager(t *testing.T) {
	agent := NewOpenClawAgent()
	assert.Nil(t, agent.HookManager())
}

func TestOpenClawAgent_CommandManager(t *testing.T) {
	agent := NewOpenClawAgent()
	assert.Nil(t, agent.CommandManager())
}

func TestOpenClawAgent_SupportsSession(t *testing.T) {
	agent := NewOpenClawAgent()
	assert.False(t, agent.SupportsSession())
}

func TestOpenClawAgent_SessionID(t *testing.T) {
	agent := NewOpenClawAgent()
	env := agentx.NewMockEnvironment(nil)
	assert.Equal(t, "", agent.SessionID(env))
}

func TestOpenClawAgent_URL(t *testing.T) {
	agent := NewOpenClawAgent()
	assert.Equal(t, "https://github.com/openclaw/openclaw", agent.URL())
}

func TestOpenClawAgent_SupportsXDGConfig(t *testing.T) {
	agent := NewOpenClawAgent()
	assert.False(t, agent.SupportsXDGConfig())
}

// --- versionFromCommand tests ---

func TestVersionFromCommand_Success(t *testing.T) {
	env := agentx.NewMockEnvironment(nil)
	env.ExecOutputs = map[string][]byte{
		"mytool": []byte("mytool v2.5.10\n"),
	}

	result := versionFromCommand(env, "mytool", "--version")
	assert.Equal(t, "2.5.10", result)
}

func TestVersionFromCommand_CommandNotFound(t *testing.T) {
	env := agentx.NewMockEnvironment(nil)
	// no exec outputs configured, so Exec returns ErrNotFound

	result := versionFromCommand(env, "nonexistent", "--version")
	assert.Equal(t, "", result)
}

func TestVersionFromCommand_NoSemver(t *testing.T) {
	env := agentx.NewMockEnvironment(nil)
	env.ExecOutputs = map[string][]byte{
		"mytool": []byte("mytool development build (no version)\n"),
	}

	result := versionFromCommand(env, "mytool", "--version")
	assert.Equal(t, "", result)
}

func TestVersionFromCommand_MultipleLinesFirstMatch(t *testing.T) {
	env := agentx.NewMockEnvironment(nil)
	env.ExecOutputs = map[string][]byte{
		"mytool": []byte("mytool 3.1.4\nBuilt with Go 1.21.0\n"),
	}

	result := versionFromCommand(env, "mytool", "--version")
	// semverRe.FindString returns the first match across the entire text
	assert.Equal(t, "3.1.4", result)
}

func TestVersionFromCommand_ExecError(t *testing.T) {
	env := agentx.NewMockEnvironment(nil)
	env.ExecErrors = map[string]error{
		"mytool": errors.New("permission denied"),
	}

	result := versionFromCommand(env, "mytool", "--version")
	assert.Equal(t, "", result)
}
