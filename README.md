# agentx

Detect which coding agent (Claude Code, Cursor, Aider, etc.) is calling your CLI tool via environment variables and heuristics.

## Installation

```bash
go get github.com/sageox/agentx
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/sageox/agentx"
)

func main() {
    agent := agentx.Detect()

    if agent != nil {
        fmt.Printf("Running in: %s\n", agent.Name())
        fmt.Printf("Agent type: %s\n", agent.Type())
        fmt.Printf("Context files: %v\n", agent.ContextFiles())
    } else {
        fmt.Println("No coding agent detected")
    }
}
```

## Supported Agents

| Agent | Type | AGENT_ENV Values | Native Env Vars |
|-------|------|------------------|-----------------|
| Claude Code | `claude-code` | `claude-code`, `claudecode`, `claude` | `CLAUDECODE=1` |
| Cursor | `cursor` | `cursor` | `CURSOR_AGENT=1` |
| Windsurf | `windsurf` | `windsurf`, `codeium` | `WINDSURF_AGENT=1`, `CODEIUM_AGENT=1` |
| GitHub Copilot | `copilot` | `copilot`, `github-copilot` | `COPILOT_AGENT=1` |
| Aider | `aider` | `aider` | `AIDER=1`, `AIDER_AGENT=1` |
| Cody | `cody` | `cody` | `CODY_AGENT=1` |
| Continue | `continue` | `continue` | `CONTINUE_AGENT=1` |
| Code Puppy | `code-puppy` | `code-puppy`, `codepuppy` | `CODE_PUPPY=1`, `CODE_PUPPY_AGENT=1` |
| Kiro | `kiro` | `kiro` | `KIRO=1`, `KIRO_AGENT=1` |
| OpenCode | `opencode` | `opencode` | `OPENCODE=1`, `OPENCODE_AGENT=1` |
| Goose | `goose` | `goose` | `GOOSE=1`, `GOOSE_AGENT=1` |
| Amp | `amp` | `amp` | `AMP=1`, `AMP_AGENT=1` |

## AGENT_ENV Standard

The `AGENT_ENV` environment variable provides explicit agent identification:

```bash
export AGENT_ENV=claude-code
export AGENT_ENV=cursor
export AGENT_ENV=aider
```

This is the recommended way to identify coding agents when native detection isn't available.

## Detection Priority

1. **AGENT_ENV** - Explicit override via environment variable
2. **Native env vars** - Agent-specific environment variables (e.g., `CLAUDECODE=1`)
3. **Binary heuristics** - Checking `$_` for agent name in the path

## API Reference

### Detection Functions

```go
// Detect returns the currently active coding agent, or nil if none detected.
func Detect() Agent

// DetectWithEnv detects using a custom environment (useful for testing).
func DetectWithEnv(env Environment) Agent

// IsAgentContext returns true if running inside any coding agent.
func IsAgentContext() bool

// CurrentAgent is an alias for Detect().
func CurrentAgent() Agent

// RequireAgent returns an error message if not running in an agent context.
func RequireAgent(commandName string) string
```

### Agent Interface

```go
type Agent interface {
    Type() AgentType              // e.g., "claude-code"
    Name() string                 // e.g., "Claude Code"
    URL() string                  // Official project URL
    Detect(ctx, env) (bool, error)
    IsInstalled(ctx, env) (bool, error)
    UserConfigPath(env) (string, error)  // e.g., ~/.claude
    ProjectConfigPath() string           // e.g., .claude
    ContextFiles() []string              // e.g., ["CLAUDE.md"]
}
```

### Registry

```go
// DefaultRegistry contains all supported agents
var DefaultRegistry Registry

// Get an agent by type
agent, ok := agentx.DefaultRegistry.Get(agentx.AgentTypeClaudeCode)

// List all registered agents
agents := agentx.DefaultRegistry.List()
```

## Testing

The package provides `MockEnvironment` for testing:

```go
func TestMyFunction(t *testing.T) {
    env := agentx.NewMockEnvironment(map[string]string{
        "AGENT_ENV": "claude-code",
    })

    agent := agentx.DetectWithEnv(env)
    if agent.Type() != agentx.AgentTypeClaudeCode {
        t.Error("expected Claude Code")
    }
}
```

## License

MIT License - see [LICENSE](LICENSE) for details.
