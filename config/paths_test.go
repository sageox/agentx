package config

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/sageox/agentx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPaths(t *testing.T) {
	t.Run("nil env defaults to system", func(t *testing.T) {
		p := NewPaths(nil)
		assert.NotNil(t, p)
		assert.NotNil(t, p.env)
	})

	t.Run("custom env works", func(t *testing.T) {
		env := agentx.NewMockEnvironment(map[string]string{"TEST": "val"})
		p := NewPaths(env)
		assert.NotNil(t, p)
		assert.Equal(t, "val", p.env.GetEnv("TEST"))
	})
}

func TestPaths_ConfigHome(t *testing.T) {
	env := &agentx.MockEnvironment{Config: "/custom/config"}
	p := NewPaths(env)
	result, err := p.ConfigHome()
	require.NoError(t, err)
	assert.Equal(t, "/custom/config", result)
}

func TestPaths_DataHome(t *testing.T) {
	env := &agentx.MockEnvironment{Data: "/custom/data"}
	p := NewPaths(env)
	result, err := p.DataHome()
	require.NoError(t, err)
	assert.Equal(t, "/custom/data", result)
}

func TestPaths_CacheHome(t *testing.T) {
	env := &agentx.MockEnvironment{Cache: "/custom/cache"}
	p := NewPaths(env)
	result, err := p.CacheHome()
	require.NoError(t, err)
	assert.Equal(t, "/custom/cache", result)
}

func TestPaths_AppConfig(t *testing.T) {
	t.Run("returns filepath.Join of configDir and appName", func(t *testing.T) {
		env := &agentx.MockEnvironment{Config: "/home/user/.config"}
		p := NewPaths(env)
		result, err := p.AppConfig("myapp")
		require.NoError(t, err)
		assert.Equal(t, filepath.Join("/home/user/.config", "myapp"), result)
	})
}

func TestPaths_AppData(t *testing.T) {
	t.Run("returns filepath.Join of dataDir and appName", func(t *testing.T) {
		env := &agentx.MockEnvironment{Data: "/home/user/.local/share"}
		p := NewPaths(env)
		result, err := p.AppData("myapp")
		require.NoError(t, err)
		assert.Equal(t, filepath.Join("/home/user/.local/share", "myapp"), result)
	})
}

func TestPaths_AppCache(t *testing.T) {
	t.Run("returns filepath.Join of cacheDir and appName", func(t *testing.T) {
		env := &agentx.MockEnvironment{Cache: "/home/user/.cache"}
		p := NewPaths(env)
		result, err := p.AppCache("myapp")
		require.NoError(t, err)
		assert.Equal(t, filepath.Join("/home/user/.cache", "myapp"), result)
	})
}

func TestPaths_AgentConfigPath(t *testing.T) {
	t.Run("claude returns .claude", func(t *testing.T) {
		env := &agentx.MockEnvironment{Home: "/home/user"}
		p := NewPaths(env)
		result, err := p.AgentConfigPath(agentx.AgentTypeClaudeCode)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join("/home/user", ".claude"), result)
	})

	t.Run("cursor returns .cursor", func(t *testing.T) {
		env := &agentx.MockEnvironment{Home: "/home/user"}
		p := NewPaths(env)
		result, err := p.AgentConfigPath(agentx.AgentTypeCursor)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join("/home/user", ".cursor"), result)
	})

	t.Run("windsurf returns .codeium", func(t *testing.T) {
		env := &agentx.MockEnvironment{Home: "/home/user"}
		p := NewPaths(env)
		result, err := p.AgentConfigPath(agentx.AgentTypeWindsurf)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join("/home/user", ".codeium"), result)
	})

	t.Run("copilot returns .config/github-copilot", func(t *testing.T) {
		env := &agentx.MockEnvironment{Home: "/home/user"}
		p := NewPaths(env)
		result, err := p.AgentConfigPath(agentx.AgentTypeCopilot)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join("/home/user", ".config/github-copilot"), result)
	})

	t.Run("kiro returns .kiro", func(t *testing.T) {
		env := &agentx.MockEnvironment{Home: "/home/user"}
		p := NewPaths(env)
		result, err := p.AgentConfigPath(agentx.AgentTypeKiro)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join("/home/user", ".kiro"), result)
	})

	t.Run("unknown agent falls back to configDir/agentType", func(t *testing.T) {
		env := &agentx.MockEnvironment{
			Home:   "/home/user",
			Config: "/home/user/.config",
		}
		p := NewPaths(env)
		result, err := p.AgentConfigPath(agentx.AgentType("unknown-agent"))
		require.NoError(t, err)
		assert.Equal(t, filepath.Join("/home/user/.config", "unknown-agent"), result)
	})

	t.Run("HomeDir error propagates", func(t *testing.T) {
		env := &agentx.MockEnvironment{HomeError: fmt.Errorf("no home")}
		p := NewPaths(env)
		_, err := p.AgentConfigPath(agentx.AgentTypeClaudeCode)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no home")
	})
}

func TestAgentConfigPaths(t *testing.T) {
	expected := map[agentx.AgentType]string{
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

	for agentType, expectedPath := range expected {
		t.Run(string(agentType), func(t *testing.T) {
			actual, ok := AgentConfigPaths[agentType]
			assert.True(t, ok, "expected %s to be in AgentConfigPaths", agentType)
			assert.Equal(t, expectedPath, actual)
		})
	}

	assert.Len(t, AgentConfigPaths, len(expected), "AgentConfigPaths should have exactly the expected entries")
}
