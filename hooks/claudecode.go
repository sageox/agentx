package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sageox/agentx"
)

// ClaudeCodeHookManager implements HookManager for Claude Code.
type ClaudeCodeHookManager struct {
	env        agentx.Environment
	configPath string
}

// NewClaudeCodeHookManager creates a new Claude Code hook manager.
func NewClaudeCodeHookManager(env agentx.Environment) *ClaudeCodeHookManager {
	if env == nil {
		env = agentx.NewSystemEnvironment()
	}
	return &ClaudeCodeHookManager{env: env}
}

func (m *ClaudeCodeHookManager) getConfigPath() (string, error) {
	if m.configPath != "" {
		return m.configPath, nil
	}

	home, err := m.env.HomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}

	m.configPath = filepath.Join(home, ".claude")
	return m.configPath, nil
}

func (m *ClaudeCodeHookManager) Install(ctx context.Context, config agentx.HookConfig) error {
	configPath, err := m.getConfigPath()
	if err != nil {
		return err
	}

	// Ensure .claude directory exists
	if err := m.env.MkdirAll(configPath, 0o755); err != nil {
		return fmt.Errorf("create claude directory: %w", err)
	}

	// Install MCP servers if configured
	if len(config.MCPServers) > 0 {
		if err := m.installMCPServers(ctx, config.MCPServers, config.Merge); err != nil {
			return fmt.Errorf("install mcp servers: %w", err)
		}
	}

	// Install system instructions if configured
	if config.SystemInstructions != "" {
		if err := m.installSystemInstructions(ctx, config.SystemInstructions, config.Merge); err != nil {
			return fmt.Errorf("install system instructions: %w", err)
		}
	}

	// Install event hooks if configured
	if len(config.EventHooks) > 0 {
		if err := m.installEventHooks(ctx, config.EventHooks, config.Merge); err != nil {
			return fmt.Errorf("install event hooks: %w", err)
		}
	}

	return nil
}

func (m *ClaudeCodeHookManager) Uninstall(ctx context.Context) error {
	configPath, err := m.getConfigPath()
	if err != nil {
		return err
	}

	// Remove MCP config (or just sageox entries)
	mcpPath := filepath.Join(configPath, "mcp_config.json")
	if err := m.removeSageoxMCPServers(mcpPath); err != nil {
		return fmt.Errorf("remove mcp servers: %w", err)
	}

	return nil
}

func (m *ClaudeCodeHookManager) IsInstalled(ctx context.Context) (bool, error) {
	configPath, err := m.getConfigPath()
	if err != nil {
		return false, err
	}

	// Check if MCP config exists with sageox entry
	mcpPath := filepath.Join(configPath, "mcp_config.json")
	data, err := m.env.ReadFile(mcpPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("read mcp config: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return false, nil
	}

	mcpServers, ok := config["mcpServers"].(map[string]interface{})
	if !ok {
		return false, nil
	}

	_, hasSageox := mcpServers["sageox"]
	return hasSageox, nil
}

func (m *ClaudeCodeHookManager) Validate(ctx context.Context) error {
	installed, err := m.IsInstalled(ctx)
	if err != nil {
		return fmt.Errorf("check installation: %w", err)
	}

	if !installed {
		return fmt.Errorf("hooks not installed")
	}

	return nil
}

func (m *ClaudeCodeHookManager) installMCPServers(ctx context.Context, servers map[string]agentx.MCPServerConfig, merge bool) error {
	configPath, err := m.getConfigPath()
	if err != nil {
		return err
	}

	mcpPath := filepath.Join(configPath, "mcp_config.json")

	var existingConfig map[string]interface{}

	if merge {
		data, err := m.env.ReadFile(mcpPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read mcp config: %w", err)
		}
		if len(data) > 0 {
			if err := json.Unmarshal(data, &existingConfig); err != nil {
				return fmt.Errorf("parse mcp config: %w", err)
			}
		}
	}

	if existingConfig == nil {
		existingConfig = make(map[string]interface{})
	}

	mcpServers, ok := existingConfig["mcpServers"].(map[string]interface{})
	if !ok {
		mcpServers = make(map[string]interface{})
		existingConfig["mcpServers"] = mcpServers
	}

	for name, server := range servers {
		serverConfig := map[string]interface{}{
			"command": server.Command,
		}
		if len(server.Args) > 0 {
			serverConfig["args"] = server.Args
		}
		if len(server.Env) > 0 {
			serverConfig["env"] = server.Env
		}
		mcpServers[name] = serverConfig
	}

	data, err := json.MarshalIndent(existingConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal mcp config: %w", err)
	}

	if err := m.env.WriteFile(mcpPath, data, 0o644); err != nil {
		return fmt.Errorf("write mcp config: %w", err)
	}

	return nil
}

func (m *ClaudeCodeHookManager) installSystemInstructions(ctx context.Context, instructions string, merge bool) error {
	configPath, err := m.getConfigPath()
	if err != nil {
		return err
	}

	claudeMdPath := filepath.Join(configPath, "CLAUDE.md")

	if merge {
		existing, err := m.env.ReadFile(claudeMdPath)
		if err == nil && len(existing) > 0 {
			// Simple append with separator
			instructions = string(existing) + "\n\n" + instructions
		}
	}

	if err := m.env.WriteFile(claudeMdPath, []byte(instructions), 0o644); err != nil {
		return fmt.Errorf("write CLAUDE.md: %w", err)
	}

	return nil
}

// installEventHooks installs lifecycle event hooks to settings.json.
// Event hooks are configured under the "hooks" key and trigger on events like
// PreToolUse, PostToolUse, etc.
//
// Reference: https://code.claude.com/docs/en/hooks-guide
func (m *ClaudeCodeHookManager) installEventHooks(_ context.Context, hooks agentx.EventHooks, merge bool) error {
	configPath, err := m.getConfigPath()
	if err != nil {
		return err
	}

	settingsPath := filepath.Join(configPath, "settings.json")

	var existingConfig map[string]interface{}

	if merge {
		data, err := m.env.ReadFile(settingsPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read settings: %w", err)
		}
		if len(data) > 0 {
			if err := json.Unmarshal(data, &existingConfig); err != nil {
				return fmt.Errorf("parse settings: %w", err)
			}
		}
	}

	if existingConfig == nil {
		existingConfig = make(map[string]interface{})
	}

	// Get or create the hooks section
	hooksSection, ok := existingConfig["hooks"].(map[string]interface{})
	if !ok {
		hooksSection = make(map[string]interface{})
		existingConfig["hooks"] = hooksSection
	}

	// Add each event's hook rules
	for event, rules := range hooks {
		// Convert rules to JSON-compatible format
		var jsonRules []map[string]interface{}
		for _, rule := range rules {
			jsonRule := make(map[string]interface{})
			if rule.Matcher != "" {
				jsonRule["matcher"] = rule.Matcher
			}

			var jsonHooks []map[string]interface{}
			for _, action := range rule.Hooks {
				jsonHooks = append(jsonHooks, map[string]interface{}{
					"type":    action.Type,
					"command": action.Command,
				})
			}
			jsonRule["hooks"] = jsonHooks
			jsonRules = append(jsonRules, jsonRule)
		}

		if merge {
			// Merge with existing rules, deduplicating by matcher+command
			if existing, ok := hooksSection[string(event)].([]interface{}); ok {
				for _, r := range jsonRules {
					if !isDuplicateRule(existing, r) {
						existing = append(existing, r)
					}
				}
				hooksSection[string(event)] = existing
			} else {
				hooksSection[string(event)] = jsonRules
			}
		} else {
			hooksSection[string(event)] = jsonRules
		}
	}

	data, err := json.MarshalIndent(existingConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	if err := m.env.WriteFile(settingsPath, data, 0o644); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}

	return nil
}

// isDuplicateRule checks if a rule already exists in the entries list by comparing
// matcher and hook commands. Prevents duplicate hooks when merge is called multiple times.
func isDuplicateRule(existing []interface{}, newRule map[string]interface{}) bool {
	newMatcher, _ := newRule["matcher"].(string)
	newHooks, _ := newRule["hooks"].([]map[string]interface{})

	for _, entry := range existing {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}

		existingMatcher, _ := entryMap["matcher"].(string)
		if existingMatcher != newMatcher {
			continue
		}

		// same matcher — check if hooks match
		existingHooks, ok := entryMap["hooks"].([]interface{})
		if !ok {
			continue
		}

		if len(existingHooks) != len(newHooks) {
			continue
		}

		// Build set of (type, command) pairs from existing hooks
		existingSet := make(map[string]bool, len(existingHooks))
		for _, eh := range existingHooks {
			ehMap, ok := eh.(map[string]interface{})
			if !ok {
				continue
			}
			key := fmt.Sprintf("%v|%v", ehMap["type"], ehMap["command"])
			existingSet[key] = true
		}

		// Check all new hooks exist in the set
		allMatch := true
		for _, nh := range newHooks {
			key := fmt.Sprintf("%v|%v", nh["type"], nh["command"])
			if !existingSet[key] {
				allMatch = false
				break
			}
		}

		if allMatch {
			return true
		}
	}

	return false
}

func (m *ClaudeCodeHookManager) removeSageoxMCPServers(mcpPath string) error {
	data, err := m.env.ReadFile(mcpPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	mcpServers, ok := config["mcpServers"].(map[string]interface{})
	if !ok {
		return nil
	}

	// Remove sageox entry
	delete(mcpServers, "sageox")

	newData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return m.env.WriteFile(mcpPath, newData, 0o644)
}

// Ensure ClaudeCodeHookManager implements HookManager.
var _ agentx.HookManager = (*ClaudeCodeHookManager)(nil)
