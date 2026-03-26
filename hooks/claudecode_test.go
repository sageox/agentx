package hooks

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sageox/agentx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestManager creates a ClaudeCodeHookManager with configPath set to a temp directory.
func newTestManager(t *testing.T, env agentx.Environment) *ClaudeCodeHookManager {
	t.Helper()
	tmpDir := t.TempDir()
	if env == nil {
		env = agentx.NewSystemEnvironment()
	}
	m := NewClaudeCodeHookManager(env)
	m.configPath = tmpDir
	return m
}

// readJSON reads a JSON file and unmarshals it into a map.
func readJSON(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &result))
	return result
}

// writeJSON marshals a map to JSON and writes it to a file.
func writeJSON(t *testing.T, path string, data map[string]interface{}) {
	t.Helper()
	bytes, err := json.MarshalIndent(data, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, bytes, 0o644))
}

// --- TestNewClaudeCodeHookManager ---

func TestNewClaudeCodeHookManager(t *testing.T) {
	t.Run("nil env defaults to system environment", func(t *testing.T) {
		m := NewClaudeCodeHookManager(nil)
		assert.NotNil(t, m)
		assert.NotNil(t, m.env)
	})

	t.Run("custom env is used", func(t *testing.T) {
		env := agentx.NewMockEnvironment(map[string]string{"FOO": "bar"})
		m := NewClaudeCodeHookManager(env)
		assert.NotNil(t, m)
		assert.Equal(t, "bar", m.env.GetEnv("FOO"))
	})
}

// --- TestClaudeCodeHookManager_Install ---

func TestClaudeCodeHookManager_Install(t *testing.T) {
	ctx := context.Background()

	t.Run("install MCP servers creates mcp_config.json", func(t *testing.T) {
		m := newTestManager(t, nil)
		err := m.Install(ctx, agentx.HookConfig{
			MCPServers: map[string]agentx.MCPServerConfig{
				"sageox": {Command: "ox", Args: []string{"mcp"}},
			},
		})
		require.NoError(t, err)

		mcpPath := filepath.Join(m.configPath, "mcp_config.json")
		config := readJSON(t, mcpPath)
		servers := config["mcpServers"].(map[string]interface{})
		sageox := servers["sageox"].(map[string]interface{})
		assert.Equal(t, "ox", sageox["command"])
	})

	t.Run("install system instructions creates CLAUDE.md", func(t *testing.T) {
		m := newTestManager(t, nil)
		err := m.Install(ctx, agentx.HookConfig{
			SystemInstructions: "# Test Instructions",
		})
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(m.configPath, "CLAUDE.md"))
		require.NoError(t, err)
		assert.Equal(t, "# Test Instructions", string(content))
	})

	t.Run("install event hooks creates settings.json", func(t *testing.T) {
		m := newTestManager(t, nil)
		err := m.Install(ctx, agentx.HookConfig{
			EventHooks: agentx.EventHooks{
				agentx.HookEventPreToolUse: {
					{
						Matcher: "Bash",
						Hooks:   []agentx.HookAction{{Type: "command", Command: "echo hello"}},
					},
				},
			},
		})
		require.NoError(t, err)

		config := readJSON(t, filepath.Join(m.configPath, "settings.json"))
		hooks := config["hooks"].(map[string]interface{})
		assert.Contains(t, hooks, "PreToolUse")
	})

	t.Run("install all three together", func(t *testing.T) {
		m := newTestManager(t, nil)
		err := m.Install(ctx, agentx.HookConfig{
			MCPServers: map[string]agentx.MCPServerConfig{
				"sageox": {Command: "ox"},
			},
			SystemInstructions: "# Instructions",
			EventHooks: agentx.EventHooks{
				agentx.HookEventSessionStart: {
					{Hooks: []agentx.HookAction{{Type: "command", Command: "echo start"}}},
				},
			},
		})
		require.NoError(t, err)

		assert.FileExists(t, filepath.Join(m.configPath, "mcp_config.json"))
		assert.FileExists(t, filepath.Join(m.configPath, "CLAUDE.md"))
		assert.FileExists(t, filepath.Join(m.configPath, "settings.json"))
	})

	t.Run("merge mode preserves existing config", func(t *testing.T) {
		m := newTestManager(t, nil)

		// pre-populate mcp_config.json with an existing server
		mcpPath := filepath.Join(m.configPath, "mcp_config.json")
		writeJSON(t, mcpPath, map[string]interface{}{
			"mcpServers": map[string]interface{}{
				"other-tool": map[string]interface{}{
					"command": "other",
				},
			},
		})

		err := m.Install(ctx, agentx.HookConfig{
			MCPServers: map[string]agentx.MCPServerConfig{
				"sageox": {Command: "ox"},
			},
			Merge: true,
		})
		require.NoError(t, err)

		config := readJSON(t, mcpPath)
		servers := config["mcpServers"].(map[string]interface{})
		assert.Contains(t, servers, "other-tool", "existing server should be preserved")
		assert.Contains(t, servers, "sageox", "new server should be added")
	})

	t.Run("replace mode overwrites existing config", func(t *testing.T) {
		m := newTestManager(t, nil)

		mcpPath := filepath.Join(m.configPath, "mcp_config.json")
		writeJSON(t, mcpPath, map[string]interface{}{
			"mcpServers": map[string]interface{}{
				"other-tool": map[string]interface{}{
					"command": "other",
				},
			},
		})

		err := m.Install(ctx, agentx.HookConfig{
			MCPServers: map[string]agentx.MCPServerConfig{
				"sageox": {Command: "ox"},
			},
			Merge: false,
		})
		require.NoError(t, err)

		config := readJSON(t, mcpPath)
		servers := config["mcpServers"].(map[string]interface{})
		assert.NotContains(t, servers, "other-tool", "existing server should be overwritten")
		assert.Contains(t, servers, "sageox")
	})

	t.Run("error when config path is invalid", func(t *testing.T) {
		env := &agentx.MockEnvironment{HomeError: os.ErrPermission}
		m := NewClaudeCodeHookManager(env)
		// don't set configPath so getConfigPath falls through to HomeDir
		err := m.Install(ctx, agentx.HookConfig{
			MCPServers: map[string]agentx.MCPServerConfig{
				"sageox": {Command: "ox"},
			},
		})
		assert.Error(t, err)
	})
}

// --- TestClaudeCodeHookManager_InstallMCPServers ---

func TestClaudeCodeHookManager_InstallMCPServers(t *testing.T) {
	ctx := context.Background()

	t.Run("fresh install creates new mcp_config.json", func(t *testing.T) {
		m := newTestManager(t, nil)
		err := m.Install(ctx, agentx.HookConfig{
			MCPServers: map[string]agentx.MCPServerConfig{
				"sageox": {Command: "ox", Args: []string{"mcp", "serve"}},
			},
		})
		require.NoError(t, err)

		config := readJSON(t, filepath.Join(m.configPath, "mcp_config.json"))
		servers := config["mcpServers"].(map[string]interface{})
		sageox := servers["sageox"].(map[string]interface{})
		assert.Equal(t, "ox", sageox["command"])
		args := sageox["args"].([]interface{})
		assert.Equal(t, []interface{}{"mcp", "serve"}, args)
	})

	t.Run("merge preserves existing MCP servers and adds new", func(t *testing.T) {
		m := newTestManager(t, nil)
		mcpPath := filepath.Join(m.configPath, "mcp_config.json")

		writeJSON(t, mcpPath, map[string]interface{}{
			"mcpServers": map[string]interface{}{
				"existing": map[string]interface{}{"command": "existing-cmd"},
			},
		})

		err := m.Install(ctx, agentx.HookConfig{
			MCPServers: map[string]agentx.MCPServerConfig{
				"sageox": {Command: "ox"},
			},
			Merge: true,
		})
		require.NoError(t, err)

		config := readJSON(t, mcpPath)
		servers := config["mcpServers"].(map[string]interface{})
		assert.Contains(t, servers, "existing")
		assert.Contains(t, servers, "sageox")
	})

	t.Run("server with args and env", func(t *testing.T) {
		m := newTestManager(t, nil)
		err := m.Install(ctx, agentx.HookConfig{
			MCPServers: map[string]agentx.MCPServerConfig{
				"sageox": {
					Command: "ox",
					Args:    []string{"mcp", "--verbose"},
					Env:     map[string]string{"API_KEY": "test-key"},
				},
			},
		})
		require.NoError(t, err)

		config := readJSON(t, filepath.Join(m.configPath, "mcp_config.json"))
		servers := config["mcpServers"].(map[string]interface{})
		sageox := servers["sageox"].(map[string]interface{})
		assert.Equal(t, "ox", sageox["command"])
		assert.NotNil(t, sageox["args"])
		envMap := sageox["env"].(map[string]interface{})
		assert.Equal(t, "test-key", envMap["API_KEY"])
	})

	t.Run("server without optional fields", func(t *testing.T) {
		m := newTestManager(t, nil)
		err := m.Install(ctx, agentx.HookConfig{
			MCPServers: map[string]agentx.MCPServerConfig{
				"minimal": {Command: "minimal-cmd"},
			},
		})
		require.NoError(t, err)

		config := readJSON(t, filepath.Join(m.configPath, "mcp_config.json"))
		servers := config["mcpServers"].(map[string]interface{})
		server := servers["minimal"].(map[string]interface{})
		assert.Equal(t, "minimal-cmd", server["command"])
		assert.Nil(t, server["args"], "args should be absent when empty")
		assert.Nil(t, server["env"], "env should be absent when empty")
	})
}

// --- TestClaudeCodeHookManager_InstallSystemInstructions ---

func TestClaudeCodeHookManager_InstallSystemInstructions(t *testing.T) {
	ctx := context.Background()

	t.Run("fresh writes CLAUDE.md", func(t *testing.T) {
		m := newTestManager(t, nil)
		err := m.Install(ctx, agentx.HookConfig{
			SystemInstructions: "# Fresh instructions",
		})
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(m.configPath, "CLAUDE.md"))
		require.NoError(t, err)
		assert.Equal(t, "# Fresh instructions", string(content))
	})

	t.Run("merge appends to existing CLAUDE.md", func(t *testing.T) {
		m := newTestManager(t, nil)
		claudeMdPath := filepath.Join(m.configPath, "CLAUDE.md")
		require.NoError(t, os.WriteFile(claudeMdPath, []byte("# Existing"), 0o644))

		err := m.Install(ctx, agentx.HookConfig{
			SystemInstructions: "# New section",
			Merge:              true,
		})
		require.NoError(t, err)

		content, err := os.ReadFile(claudeMdPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "# Existing")
		assert.Contains(t, string(content), "# New section")
		assert.Contains(t, string(content), "\n\n", "should have separator between sections")
	})

	t.Run("replace overwrites existing CLAUDE.md", func(t *testing.T) {
		m := newTestManager(t, nil)
		claudeMdPath := filepath.Join(m.configPath, "CLAUDE.md")
		require.NoError(t, os.WriteFile(claudeMdPath, []byte("# Old content"), 0o644))

		err := m.Install(ctx, agentx.HookConfig{
			SystemInstructions: "# Replaced content",
			Merge:              false,
		})
		require.NoError(t, err)

		content, err := os.ReadFile(claudeMdPath)
		require.NoError(t, err)
		assert.Equal(t, "# Replaced content", string(content))
		assert.NotContains(t, string(content), "Old content")
	})
}

// --- TestClaudeCodeHookManager_InstallEventHooks ---

func TestClaudeCodeHookManager_InstallEventHooks(t *testing.T) {
	ctx := context.Background()

	t.Run("fresh creates settings.json with hooks section", func(t *testing.T) {
		m := newTestManager(t, nil)
		err := m.Install(ctx, agentx.HookConfig{
			EventHooks: agentx.EventHooks{
				agentx.HookEventPreToolUse: {
					{
						Matcher: "Bash",
						Hooks:   []agentx.HookAction{{Type: "command", Command: "echo pre"}},
					},
				},
			},
		})
		require.NoError(t, err)

		config := readJSON(t, filepath.Join(m.configPath, "settings.json"))
		hooks := config["hooks"].(map[string]interface{})
		preToolUse := hooks["PreToolUse"].([]interface{})
		assert.Len(t, preToolUse, 1)
		rule := preToolUse[0].(map[string]interface{})
		assert.Equal(t, "Bash", rule["matcher"])
	})

	t.Run("merge adds rules to existing events without duplicates", func(t *testing.T) {
		m := newTestManager(t, nil)
		settingsPath := filepath.Join(m.configPath, "settings.json")

		// pre-populate with an existing hook rule
		writeJSON(t, settingsPath, map[string]interface{}{
			"hooks": map[string]interface{}{
				"PreToolUse": []interface{}{
					map[string]interface{}{
						"matcher": "Bash",
						"hooks": []interface{}{
							map[string]interface{}{"type": "command", "command": "echo existing"},
						},
					},
				},
			},
		})

		// install same event with a different rule
		err := m.Install(ctx, agentx.HookConfig{
			EventHooks: agentx.EventHooks{
				agentx.HookEventPreToolUse: {
					{
						Matcher: "Write",
						Hooks:   []agentx.HookAction{{Type: "command", Command: "echo new"}},
					},
				},
			},
			Merge: true,
		})
		require.NoError(t, err)

		config := readJSON(t, settingsPath)
		hooks := config["hooks"].(map[string]interface{})
		preToolUse := hooks["PreToolUse"].([]interface{})
		assert.Len(t, preToolUse, 2, "should have both existing and new rules")
	})

	t.Run("merge skips duplicate rules", func(t *testing.T) {
		m := newTestManager(t, nil)
		settingsPath := filepath.Join(m.configPath, "settings.json")

		writeJSON(t, settingsPath, map[string]interface{}{
			"hooks": map[string]interface{}{
				"PreToolUse": []interface{}{
					map[string]interface{}{
						"matcher": "Bash",
						"hooks": []interface{}{
							map[string]interface{}{"type": "command", "command": "echo hello"},
						},
					},
				},
			},
		})

		// install identical rule
		err := m.Install(ctx, agentx.HookConfig{
			EventHooks: agentx.EventHooks{
				agentx.HookEventPreToolUse: {
					{
						Matcher: "Bash",
						Hooks:   []agentx.HookAction{{Type: "command", Command: "echo hello"}},
					},
				},
			},
			Merge: true,
		})
		require.NoError(t, err)

		config := readJSON(t, settingsPath)
		hooks := config["hooks"].(map[string]interface{})
		preToolUse := hooks["PreToolUse"].([]interface{})
		assert.Len(t, preToolUse, 1, "duplicate rule should not be added")
	})
}

// --- TestClaudeCodeHookManager_Uninstall ---

func TestClaudeCodeHookManager_Uninstall(t *testing.T) {
	ctx := context.Background()

	t.Run("removes sageox from MCP servers", func(t *testing.T) {
		m := newTestManager(t, nil)
		mcpPath := filepath.Join(m.configPath, "mcp_config.json")

		writeJSON(t, mcpPath, map[string]interface{}{
			"mcpServers": map[string]interface{}{
				"sageox": map[string]interface{}{"command": "ox"},
				"other":  map[string]interface{}{"command": "other-cmd"},
			},
		})

		err := m.Uninstall(ctx)
		require.NoError(t, err)

		config := readJSON(t, mcpPath)
		servers := config["mcpServers"].(map[string]interface{})
		assert.NotContains(t, servers, "sageox")
		assert.Contains(t, servers, "other", "other servers should remain")
	})

	t.Run("preserves other MCP servers", func(t *testing.T) {
		m := newTestManager(t, nil)
		mcpPath := filepath.Join(m.configPath, "mcp_config.json")

		writeJSON(t, mcpPath, map[string]interface{}{
			"mcpServers": map[string]interface{}{
				"sageox":    map[string]interface{}{"command": "ox"},
				"tool-a":    map[string]interface{}{"command": "a"},
				"tool-b":    map[string]interface{}{"command": "b"},
			},
		})

		err := m.Uninstall(ctx)
		require.NoError(t, err)

		config := readJSON(t, mcpPath)
		servers := config["mcpServers"].(map[string]interface{})
		assert.Len(t, servers, 2)
		assert.Contains(t, servers, "tool-a")
		assert.Contains(t, servers, "tool-b")
	})

	t.Run("no mcp_config.json is not an error", func(t *testing.T) {
		m := newTestManager(t, nil)
		err := m.Uninstall(ctx)
		assert.NoError(t, err)
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		m := newTestManager(t, nil)
		mcpPath := filepath.Join(m.configPath, "mcp_config.json")
		require.NoError(t, os.WriteFile(mcpPath, []byte("not json"), 0o644))

		err := m.Uninstall(ctx)
		assert.Error(t, err)
	})
}

// --- TestClaudeCodeHookManager_IsInstalled ---

func TestClaudeCodeHookManager_IsInstalled(t *testing.T) {
	ctx := context.Background()

	t.Run("returns true when sageox MCP server exists", func(t *testing.T) {
		m := newTestManager(t, nil)
		mcpPath := filepath.Join(m.configPath, "mcp_config.json")

		writeJSON(t, mcpPath, map[string]interface{}{
			"mcpServers": map[string]interface{}{
				"sageox": map[string]interface{}{"command": "ox"},
			},
		})

		installed, err := m.IsInstalled(ctx)
		require.NoError(t, err)
		assert.True(t, installed)
	})

	t.Run("returns false when no sageox entry", func(t *testing.T) {
		m := newTestManager(t, nil)
		mcpPath := filepath.Join(m.configPath, "mcp_config.json")

		writeJSON(t, mcpPath, map[string]interface{}{
			"mcpServers": map[string]interface{}{
				"other-tool": map[string]interface{}{"command": "other"},
			},
		})

		installed, err := m.IsInstalled(ctx)
		require.NoError(t, err)
		assert.False(t, installed)
	})

	t.Run("returns false when file does not exist", func(t *testing.T) {
		m := newTestManager(t, nil)
		installed, err := m.IsInstalled(ctx)
		require.NoError(t, err)
		assert.False(t, installed)
	})

	t.Run("returns false on invalid JSON", func(t *testing.T) {
		m := newTestManager(t, nil)
		mcpPath := filepath.Join(m.configPath, "mcp_config.json")
		require.NoError(t, os.WriteFile(mcpPath, []byte("{bad json}"), 0o644))

		installed, err := m.IsInstalled(ctx)
		require.NoError(t, err)
		assert.False(t, installed)
	})
}

// --- TestClaudeCodeHookManager_Validate ---

func TestClaudeCodeHookManager_Validate(t *testing.T) {
	ctx := context.Background()

	t.Run("returns nil when installed", func(t *testing.T) {
		m := newTestManager(t, nil)
		mcpPath := filepath.Join(m.configPath, "mcp_config.json")
		writeJSON(t, mcpPath, map[string]interface{}{
			"mcpServers": map[string]interface{}{
				"sageox": map[string]interface{}{"command": "ox"},
			},
		})

		err := m.Validate(ctx)
		assert.NoError(t, err)
	})

	t.Run("returns error when not installed", func(t *testing.T) {
		m := newTestManager(t, nil)
		err := m.Validate(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not installed")
	})
}

// --- TestIsDuplicateRule ---

func TestIsDuplicateRule(t *testing.T) {
	t.Run("exact match returns true", func(t *testing.T) {
		existing := []interface{}{
			map[string]interface{}{
				"matcher": "Bash",
				"hooks": []interface{}{
					map[string]interface{}{"type": "command", "command": "echo hello"},
				},
			},
		}
		newRule := map[string]interface{}{
			"matcher": "Bash",
			"hooks": []map[string]interface{}{
				{"type": "command", "command": "echo hello"},
			},
		}
		assert.True(t, isDuplicateRule(existing, newRule))
	})

	t.Run("different matcher returns false", func(t *testing.T) {
		existing := []interface{}{
			map[string]interface{}{
				"matcher": "Bash",
				"hooks": []interface{}{
					map[string]interface{}{"type": "command", "command": "echo hello"},
				},
			},
		}
		newRule := map[string]interface{}{
			"matcher": "Write",
			"hooks": []map[string]interface{}{
				{"type": "command", "command": "echo hello"},
			},
		}
		assert.False(t, isDuplicateRule(existing, newRule))
	})

	t.Run("different command returns false", func(t *testing.T) {
		existing := []interface{}{
			map[string]interface{}{
				"matcher": "Bash",
				"hooks": []interface{}{
					map[string]interface{}{"type": "command", "command": "echo hello"},
				},
			},
		}
		newRule := map[string]interface{}{
			"matcher": "Bash",
			"hooks": []map[string]interface{}{
				{"type": "command", "command": "echo different"},
			},
		}
		assert.False(t, isDuplicateRule(existing, newRule))
	})

	t.Run("different hook count returns false", func(t *testing.T) {
		existing := []interface{}{
			map[string]interface{}{
				"matcher": "Bash",
				"hooks": []interface{}{
					map[string]interface{}{"type": "command", "command": "echo one"},
					map[string]interface{}{"type": "command", "command": "echo two"},
				},
			},
		}
		newRule := map[string]interface{}{
			"matcher": "Bash",
			"hooks": []map[string]interface{}{
				{"type": "command", "command": "echo one"},
			},
		}
		assert.False(t, isDuplicateRule(existing, newRule))
	})

	t.Run("empty existing returns false", func(t *testing.T) {
		newRule := map[string]interface{}{
			"matcher": "Bash",
			"hooks": []map[string]interface{}{
				{"type": "command", "command": "echo hello"},
			},
		}
		assert.False(t, isDuplicateRule([]interface{}{}, newRule))
	})

	t.Run("non-map entries in existing are skipped", func(t *testing.T) {
		existing := []interface{}{
			"not a map",
			42,
		}
		newRule := map[string]interface{}{
			"matcher": "Bash",
			"hooks": []map[string]interface{}{
				{"type": "command", "command": "echo hello"},
			},
		}
		assert.False(t, isDuplicateRule(existing, newRule))
	})

	t.Run("order independent match returns true", func(t *testing.T) {
		existing := []interface{}{
			map[string]interface{}{
				"matcher": "test",
				"hooks": []interface{}{
					map[string]interface{}{"type": "command", "command": "hook-a"},
					map[string]interface{}{"type": "command", "command": "hook-b"},
				},
			},
		}
		// Same hooks, different order
		newRule := map[string]interface{}{
			"matcher": "test",
			"hooks": []map[string]interface{}{
				{"type": "command", "command": "hook-b"},
				{"type": "command", "command": "hook-a"},
			},
		}
		assert.True(t, isDuplicateRule(existing, newRule))
	})

	t.Run("entry hooks not a slice continues without panic", func(t *testing.T) {
		existing := []interface{}{
			map[string]interface{}{
				"matcher": "Bash",
				"hooks":   "not-a-slice", // hooks is a string instead of []interface{}
			},
		}
		newRule := map[string]interface{}{
			"matcher": "Bash",
			"hooks": []map[string]interface{}{
				{"type": "command", "command": "echo hello"},
			},
		}
		// should not match because existing hooks isn't a valid []interface{}
		assert.False(t, isDuplicateRule(existing, newRule))
	})
}

// --- Additional coverage tests ---

func TestGetConfigPath_CachedPath(t *testing.T) {
	env := agentx.NewMockEnvironment(nil)
	m := NewClaudeCodeHookManager(env)
	m.configPath = "/cached/path"

	path, err := m.getConfigPath()
	require.NoError(t, err)
	assert.Equal(t, "/cached/path", path)
}

func TestGetConfigPath_FromHomeDir(t *testing.T) {
	env := agentx.NewMockEnvironment(nil)
	m := NewClaudeCodeHookManager(env)
	// configPath is empty, so it should call HomeDir

	path, err := m.getConfigPath()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(env.Home, ".claude"), path)
	// second call should return cached
	assert.Equal(t, filepath.Join(env.Home, ".claude"), m.configPath)
}

func TestInstall_MkdirAllFailure(t *testing.T) {
	ctx := context.Background()
	// use SystemEnvironment so real os.MkdirAll fails on a file path
	m := NewClaudeCodeHookManager(agentx.NewSystemEnvironment())

	// point configPath at a file (not a directory) so MkdirAll fails
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "not-a-dir")
	require.NoError(t, os.WriteFile(filePath, []byte("blocker"), 0o644))
	m.configPath = filePath

	err := m.Install(ctx, agentx.HookConfig{
		MCPServers: map[string]agentx.MCPServerConfig{
			"sageox": {Command: "ox"},
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create claude directory")
}

func TestIsInstalled_ReadError(t *testing.T) {
	ctx := context.Background()
	m := newTestManager(t, nil)
	mcpPath := filepath.Join(m.configPath, "mcp_config.json")

	// create the file but make it unreadable
	require.NoError(t, os.WriteFile(mcpPath, []byte(`{"mcpServers":{}}`), 0o644))
	require.NoError(t, os.Chmod(mcpPath, 0o000))
	t.Cleanup(func() { os.Chmod(mcpPath, 0o644) })

	installed, err := m.IsInstalled(ctx)
	assert.Error(t, err, "should return error when file is unreadable")
	assert.False(t, installed)
	assert.Contains(t, err.Error(), "read mcp config")
}

func TestIsInstalled_McpServersNotMap(t *testing.T) {
	ctx := context.Background()
	m := newTestManager(t, nil)
	mcpPath := filepath.Join(m.configPath, "mcp_config.json")

	// mcpServers is a string instead of a map
	require.NoError(t, os.WriteFile(mcpPath, []byte(`{"mcpServers":"not-a-map"}`), 0o644))

	installed, err := m.IsInstalled(ctx)
	require.NoError(t, err)
	assert.False(t, installed)
}

func TestInstallMCPServers_MergeWithCorruptFile(t *testing.T) {
	ctx := context.Background()
	m := newTestManager(t, nil)
	mcpPath := filepath.Join(m.configPath, "mcp_config.json")

	// write invalid JSON
	require.NoError(t, os.WriteFile(mcpPath, []byte("not valid json{{{"), 0o644))

	err := m.Install(ctx, agentx.HookConfig{
		MCPServers: map[string]agentx.MCPServerConfig{
			"sageox": {Command: "ox"},
		},
		Merge: true,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse mcp config")
}

func TestInstallEventHooks_ReadError(t *testing.T) {
	ctx := context.Background()
	m := newTestManager(t, nil)
	settingsPath := filepath.Join(m.configPath, "settings.json")

	// write invalid JSON to settings.json
	require.NoError(t, os.WriteFile(settingsPath, []byte("broken json!!!"), 0o644))

	err := m.Install(ctx, agentx.HookConfig{
		EventHooks: agentx.EventHooks{
			agentx.HookEventPreToolUse: {
				{Matcher: "Bash", Hooks: []agentx.HookAction{{Type: "command", Command: "echo hi"}}},
			},
		},
		Merge: true,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse settings")
}

func TestInstallEventHooks_ExistingNotSlice(t *testing.T) {
	ctx := context.Background()
	m := newTestManager(t, nil)
	settingsPath := filepath.Join(m.configPath, "settings.json")

	// existing hooks[PreToolUse] is a string, not []interface{}
	writeJSON(t, settingsPath, map[string]interface{}{
		"hooks": map[string]interface{}{
			"PreToolUse": "not-a-slice",
		},
	})

	err := m.Install(ctx, agentx.HookConfig{
		EventHooks: agentx.EventHooks{
			agentx.HookEventPreToolUse: {
				{Matcher: "Bash", Hooks: []agentx.HookAction{{Type: "command", Command: "echo hi"}}},
			},
		},
		Merge: true,
	})
	require.NoError(t, err)

	// should overwrite with the new rules since existing wasn't a valid slice
	config := readJSON(t, settingsPath)
	hooks := config["hooks"].(map[string]interface{})
	preToolUse := hooks["PreToolUse"].([]interface{})
	assert.Len(t, preToolUse, 1)
}

func TestRemoveSageoxMCPServers_NotMap(t *testing.T) {
	m := newTestManager(t, nil)
	mcpPath := filepath.Join(m.configPath, "mcp_config.json")

	// mcpServers is not a map
	require.NoError(t, os.WriteFile(mcpPath, []byte(`{"mcpServers": "just-a-string"}`), 0o644))

	err := m.removeSageoxMCPServers(mcpPath)
	assert.NoError(t, err, "should return nil when mcpServers is not a map")
}

func TestInstallSystemInstructions_MergeNonExistent(t *testing.T) {
	ctx := context.Background()
	m := newTestManager(t, nil)

	// CLAUDE.md doesn't exist yet, merge=true should just write the new content
	err := m.Install(ctx, agentx.HookConfig{
		SystemInstructions: "# Brand new instructions",
		Merge:              true,
	})
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(m.configPath, "CLAUDE.md"))
	require.NoError(t, err)
	assert.Equal(t, "# Brand new instructions", string(content))
}

func TestUninstall_GetConfigPathError(t *testing.T) {
	ctx := context.Background()
	env := &agentx.MockEnvironment{HomeError: os.ErrPermission}
	m := NewClaudeCodeHookManager(env)
	// configPath is empty, so getConfigPath will try HomeDir and fail

	err := m.Uninstall(ctx)
	assert.Error(t, err)
}
