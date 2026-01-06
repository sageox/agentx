# AgentX

Part of the **Agent Experience (AX)** toolchain by [SageOx](https://sageox.ai) — a new category of developer tooling that optimizes CLIs for AI coding agents, not just humans.

**Build agent-aware CLI tools that know which AI coding assistant is calling them.**

AgentX detects whether your tool is running inside Claude Code, Cursor, Aider, or 14 other coding agents - and gives you access to their configuration, context files, capabilities, and extension management.

## Why AX?

AI coding agents are the new users of your CLI. **Agent Experience (AX)** is the practice of designing tools that work well for both humans and agents — not just tolerating agent usage, but actively optimizing for it.

AgentX gives your CLI the primitives it needs:

- **Which agent is calling**: Tailor output format, verbosity, or behavior per agent
- **Where agent config lives**: Read/write to `~/.claude`, `~/.cursor`, etc.
- **What context files exist**: Find `CLAUDE.md`, `.cursorrules`, and similar files
- **How to propagate context**: Set `AGENT_ENV` so downstream tools know too

```go
package main

import (
    "fmt"

    "github.com/sageox/agentx/pkg"
)

func main() {
    // Detect agent and set AGENT_ENV for child processes
    agent := agentx.Init()

    if agent != nil {
        fmt.Printf("Running in %s\n", agent.Name())
        fmt.Printf("Config: %s\n", must(agent.UserConfigPath(agentx.NewSystemEnvironment())))
        fmt.Printf("Context files: %v\n", agent.ContextFiles())
    }
}
```

## Key Features

### 1. Automatic Agent Detection

Detects 14 coding agents via environment variables and heuristics:

```go
agent := agentx.Detect()
if agent != nil {
    switch agent.Type() {
    case agentx.AgentTypeClaudeCode:
        // Claude Code specific behavior
    case agentx.AgentTypeCursor:
        // Cursor specific behavior
    }
}
```

### 2. AGENT_ENV Propagation

`Init()` detects the agent and sets `AGENT_ENV` in the process environment. This lets downstream code and child processes simply check `AGENT_ENV` instead of implementing their own detection:

```go
func main() {
    agentx.Init() // Detects "cursor" from heuristics, sets AGENT_ENV=cursor

    // Now any library or subprocess can do:
    // os.Getenv("AGENT_ENV") → "cursor"
}
```

### 3. Agent Configuration Paths

Access each agent's user and project configuration directories:

```go
agent, _ := agentx.DefaultRegistry.Get(agentx.AgentTypeClaudeCode)
env := agentx.NewSystemEnvironment()

userConfig, _ := agent.UserConfigPath(env)   // ~/.claude
projectConfig := agent.ProjectConfigPath()   // .claude
```

| Agent | User Config | Project Config |
|-------|-------------|----------------|
| Claude Code | `~/.claude` | `.claude` |
| Cursor | `~/.cursor` | `.cursor` |
| Aider | `~/.aider` | `.aider` |
| Windsurf | `~/.codeium` | `.windsurf` |
| Cody | `~/.config/cody` | `.cody` |

### 4. Context Files

Know which instruction files each agent uses:

```go
agent, _ := agentx.DefaultRegistry.Get(agentx.AgentTypeClaudeCode)
files := agent.ContextFiles() // ["CLAUDE.md", "AGENTS.md"]
```

| Agent | Context Files |
|-------|---------------|
| Claude Code | `CLAUDE.md`, `AGENTS.md` |
| Cursor | `.cursorrules` |
| Windsurf | `.windsurfrules` |
| GitHub Copilot | `.github/copilot-instructions.md` |

### 5. Installation Detection

Check if an agent is installed on the system:

```go
ctx := context.Background()
env := agentx.NewSystemEnvironment()

agent, _ := agentx.DefaultRegistry.Get(agentx.AgentTypeCursor)
installed, _ := agent.IsInstalled(ctx, env)
```

### 6. Agent Capabilities

Query what features each agent supports:

```go
agent, _ := agentx.DefaultRegistry.Get(agentx.AgentTypeClaudeCode)
caps := agent.Capabilities()

if caps.Hooks {
    // Agent supports lifecycle hooks
}
if caps.CustomCommands {
    // Agent supports slash commands
}
if caps.MCPServers {
    // Agent supports MCP server configuration
}
```

### 7. Hook and Command Management

For agents that support extensions, manage hooks and custom commands:

```go
agent, _ := agentx.DefaultRegistry.Get(agentx.AgentTypeClaudeCode)

// Install custom slash commands
if cmdMgr := agent.CommandManager(); cmdMgr != nil {
    commands := []agentx.CommandFile{
        {Name: "my-tool.md", Content: []byte("# My Tool\n..."), Version: "1.0.0"},
    }
    written, _ := cmdMgr.Install(ctx, projectRoot, commands, false)
}

// Install hooks and MCP servers
if hookMgr := agent.HookManager(); hookMgr != nil {
    config := agentx.HookConfig{
        MCPServers: map[string]agentx.MCPServerConfig{
            "my-server": {Command: "my-mcp-server", Args: []string{"--port", "8080"}},
        },
        Merge: true,
    }
    hookMgr.Install(ctx, config)
}
```

## Installation

```bash
go get github.com/sageox/agentx/pkg
```

## Quick Start

```go
package main

import (
    "fmt"
    "os"

    "github.com/sageox/agentx/pkg"
)

func main() {
    // Initialize - detects agent and sets AGENT_ENV
    agent := agentx.Init()

    if agent == nil {
        // Not running in a coding agent
        fmt.Println("Run this from within a coding agent")
        os.Exit(1)
    }

    fmt.Printf("Agent: %s (%s)\n", agent.Name(), agent.Type())
    fmt.Printf("URL: %s\n", agent.URL())
    fmt.Printf("Context files: %v\n", agent.ContextFiles())
}
```

## Supported Agents

| Agent | Type | AGENT_ENV | Native Detection |
|-------|------|-----------|------------------|
| Claude Code | `claude` | `claude` | `CLAUDECODE=1`, `CLAUDE_CODE_ENTRYPOINT`, `CLAUDE_CODE_SESSION_ID` |
| Cursor | `cursor` | `cursor` | `CURSOR_AGENT=1`, `$_` path |
| Windsurf | `windsurf` | `windsurf` | `WINDSURF_AGENT=1`, `CODEIUM_AGENT=1` |
| GitHub Copilot | `copilot` | `copilot` | `COPILOT_AGENT=1` |
| Aider | `aider` | `aider` | `AIDER=1` |
| Cody | `cody` | `cody` | `CODY_AGENT=1` |
| Continue | `continue` | `continue` | `CONTINUE_AGENT=1` |
| Code Puppy | `code-puppy` | `code-puppy` | `CODE_PUPPY=1` |
| Kiro | `kiro` | `kiro` | `KIRO=1` |
| OpenCode | `opencode` | `opencode` | `OPENCODE=1` |
| Goose | `goose` | `goose` | `GOOSE=1` |
| Amp | `amp` | `amp` | `AMP=1` |
| Cline | `cline` | `cline` | `CLINE=1`, `CLINE_AGENT=1` |
| Droid | `droid` | `droid` | `DROID=1`, `FACTORY_DROID=1` |

## Detection Priority

1. **AGENT_ENV** - If set, this is the definitive answer (no fallback)
2. **Native env vars** - Agent-specific variables like `CLAUDECODE=1`
3. **Binary heuristics** - Checking `$_` for agent name in the executable path

## API Reference

### Initialization

```go
// Init detects the agent and sets AGENT_ENV for child processes.
// Call early in main() to propagate agent context.
func Init() Agent

// InitWithEnv is like Init but uses a custom environment (for testing).
func InitWithEnv(env Environment) Agent
```

### Detection

```go
// Detect returns the active agent without setting AGENT_ENV.
func Detect() Agent

// DetectWithEnv detects using a custom environment.
func DetectWithEnv(env Environment) Agent

// IsAgentContext returns true if running in any coding agent.
func IsAgentContext() bool

// RequireAgent returns an error message if not in agent context.
func RequireAgent(commandName string) string
```

### Agent Interface

The `Agent` interface is composed of focused sub-interfaces:

```go
// AgentIdentity - basic identification
type AgentIdentity interface {
    Type() AgentType  // "claude", "cursor", etc.
    Name() string     // "Claude Code", "Cursor", etc.
    URL() string      // Official project URL
}

// AgentDetector - detection capabilities
type AgentDetector interface {
    Detect(ctx, env) (bool, error)      // Check if this agent is active
    IsInstalled(ctx, env) (bool, error) // Check if installed on system
}

// AgentConfig - configuration paths
type AgentConfig interface {
    UserConfigPath(env) (string, error)  // ~/.claude, ~/.cursor, etc.
    ProjectConfigPath() string           // .claude, .cursor, etc.
    ContextFiles() []string              // ["CLAUDE.md"], [".cursorrules"], etc.
    SupportsXDGConfig() bool             // true if uses ~/.config/app/
}

// AgentExtensions - hooks and commands
type AgentExtensions interface {
    Capabilities() Capabilities    // What features this agent supports
    HookManager() HookManager      // Hook management (or nil)
    CommandManager() CommandManager // Command management (or nil)
}

// Agent - full interface (embeds all above)
type Agent interface {
    AgentIdentity
    AgentDetector
    AgentConfig
    AgentExtensions
}
```

### Registry

```go
// Get agent by type
agent, ok := agentx.DefaultRegistry.Get(agentx.AgentTypeClaudeCode)

// List all registered agents
agents := agentx.DefaultRegistry.List()
```

## Testing

Use `MockEnvironment` for deterministic testing:

```go
func TestMyTool(t *testing.T) {
    env := agentx.NewMockEnvironment(map[string]string{
        "AGENT_ENV": "claude",
    })

    agent := agentx.DetectWithEnv(env)
    if agent.Type() != agentx.AgentTypeClaudeCode {
        t.Error("expected Claude Code")
    }
}
```

Configure mock paths and binaries:

```go
env := agentx.NewMockEnvironment(nil)
env.Home = "/home/testuser"
env.ExistingDirs = map[string]bool{"/home/testuser/.claude": true}
env.PathBinaries = map[string]string{"claude": "/usr/bin/claude"}
```

## Package Structure

Single package - just import and use:

```go
import "github.com/sageox/agentx/pkg"
```

All 14 agents are automatically registered.

## AX Patterns

### Agent-aware output

```go
agent := agentx.Detect()
if agent != nil && agent.Type() == agentx.AgentTypeClaudeCode {
    // Claude Code prefers markdown
    fmt.Println("## Results\n")
} else {
    // Default terminal output
    fmt.Println("Results:")
}
```

### Find agent context files

```go
agent := agentx.Detect()
if agent != nil {
    for _, file := range agent.ContextFiles() {
        if content, err := os.ReadFile(file); err == nil {
            fmt.Printf("Found %s:\n%s\n", file, content)
        }
    }
}
```

### Require agent context

```go
func main() {
    if msg := agentx.RequireAgent("my-tool"); msg != "" {
        fmt.Fprintln(os.Stderr, msg)
        os.Exit(1)
    }
    // Tool requires agent context to function
}
```

## AX Toolchain

AgentX is part of the **Agent Experience (AX)** toolchain — open-source Go libraries for building CLIs that work well for AI coding agents, not just humans:

- **[AgentX](https://github.com/sageox/agentx/)** *(this repo)* — Detect which coding agent is calling and adapt behavior accordingly
- **[FrictionAX](https://github.com/sageox/frictionax/)** — CLI error detection, smart correction, and telemetry for surfacing agent desire paths
- **[ox](https://github.com/sageox/ox/)** — The CLI where the AX toolchain was born

> **Read more:** [The Hive is Buzzing](https://sageox.ai/blog/the-hive-is-buzzing) — how we use the AX toolchain to fine-tune CLIs for coding agents.

## License

MIT License - see [LICENSE](LICENSE) for details.
