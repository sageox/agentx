package agentx

import (
	
	"context"
	"path/filepath"
	"strings"

)

// CursorAgent implements Agent for Cursor.
type CursorAgent struct{}

// NewCursorAgent creates a new Cursor agent.
func NewCursorAgent() *CursorAgent {
	return &CursorAgent{}
}

func (a *CursorAgent) Type() AgentType {
	return AgentTypeCursor
}

func (a *CursorAgent) Name() string {
	return "Cursor"
}

func (a *CursorAgent) URL() string {
	return "https://github.com/getcursor/cursor"
}

// Detect checks if Cursor is the active agent.
//
// Detection methods:
//   - CURSOR_AGENT=1 (future standard)
//   - AGENT_ENV=cursor
//   - Running from cursor CLI (heuristic)
func (a *CursorAgent) Detect(ctx context.Context, env Environment) (bool, error) {
	// Check explicit CURSOR_AGENT env var (future standard)
	if env.GetEnv("CURSOR_AGENT") == "1" {
		return true, nil
	}

	// Check AGENT_ENV
	if env.GetEnv("AGENT_ENV") == "cursor" {
		return true, nil
	}

	// Heuristic: check if running from cursor CLI
	if execPath := env.GetEnv("_"); strings.Contains(strings.ToLower(execPath), "cursor") {
		return true, nil
	}

	return false, nil
}

// UserConfigPath returns the Cursor user configuration directory (~/.cursor).
func (a *CursorAgent) UserConfigPath(env Environment) (string, error) {
	home, err := env.HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cursor"), nil
}

// ProjectConfigPath returns the Cursor project configuration directory.
func (a *CursorAgent) ProjectConfigPath() string {
	return ".cursor"
}

// ContextFiles returns the context/instruction files Cursor supports.
func (a *CursorAgent) ContextFiles() []string {
	return []string{".cursorrules"}
}

// IsInstalled checks if Cursor is installed on the system.
func (a *CursorAgent) IsInstalled(ctx context.Context, env Environment) (bool, error) {
	// Check if cursor CLI is in PATH
	if _, err := env.LookPath("cursor"); err == nil {
		return true, nil
	}

	// Check for macOS application bundle
	if env.GOOS() == "darwin" {
		if env.IsDir("/Applications/Cursor.app") {
			return true, nil
		}
	}

	// Fallback: check if config directory exists
	configPath, err := a.UserConfigPath(env)
	if err != nil {
		return false, nil
	}
	if env.IsDir(configPath) {
		return true, nil
	}

	return false, nil
}

var _ Agent = (*CursorAgent)(nil)
