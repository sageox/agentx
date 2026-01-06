# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-03-10

### Added

- **Orchestrator support**: OpenClaw and Conductor orchestrator detection
- **Session support**: `SupportsSession()` and `SessionID()` for agents that track sessions
- **Hook input parsing**: `ReadHookInput()` for parsing JSON hook payloads from stdin
- **Configurable stamp prefix**: `StampComment()`, `StampedContent()`, and related functions accept a custom prefix parameter, allowing multiple tools to stamp their own content
- **Codex agent**: Support for OpenAI Codex CLI agent
- **Kiro agent**: Support for AWS Kiro agent
- **OpenCode agent**: Support for OpenCode agent
- **Config paths**: XDG-compliant path resolution for agent configuration

### Changed

- **Package restructure**: Moved from flat `pkg/` layout to root package with sub-packages (`agents/`, `orchestrators/`, `commands/`, `hooks/`, `config/`, `setup/`)
- **Import path**: Use `github.com/sageox/agentx` directly instead of `github.com/sageox/agentx/pkg`
- **Default stamp prefix**: Changed from `ox` to `agentx` for library independence
- **Go version**: Minimum Go version bumped to 1.24

## [0.2.0] - 2026-02-01

### Added

- **Interface segregation**: Split `Agent` into focused interfaces (`AgentIdentity`, `AgentDetector`, `AgentConfig`, `AgentExtensions`)
- **Capabilities**: `Capabilities` struct describes what features each agent supports (hooks, MCP servers, custom commands)
- **Hook management**: `HookManager` interface with full Claude Code implementation (`ClaudeCodeHookManager`)
- **Command management**: `CommandManager` interface with Claude Code implementation (`ClaudeCodeCommandManager`)
- **Hook events**: Support for all Claude Code lifecycle events (`PreToolUse`, `PostToolUse`, `UserPromptSubmit`, `PermissionRequest`, `Stop`, `SubagentStop`, `SessionEnd`)
- **Command versioning**: Content-hash stamping for slash commands with downgrade protection
- **XDG compliance**: `DataDir()` and `CacheDir()` methods in `Environment` interface
- **Version utilities**: `ContentHash`, `StampedContent`, `ExtractCommandHash`, `ExtractStampVersion`, `CompareVersions`, `ShouldWriteCommand`, `IsCommandStale`
- **Agent detection improvements**: Better Claude Code detection using `CLAUDE_CODE_ENTRYPOINT` and `CLAUDE_CODE_SESSION_ID`
- **Cline agent**: Support for Cline AI coding agent
- **Droid agent**: Support for Factory Droid agent
- Comprehensive test suite for version stamping and command management

### Changed

- `Environment` interface now includes `DataDir()` and `CacheDir()` methods
- `Registry` interface now includes `Detector()` method
- All agent implementations now support `Capabilities()`, `SupportsXDGConfig()`, `HookManager()`, and `CommandManager()` methods
- MockEnvironment updated with `Data` and `Cache` fields

## [0.1.0] - 2026-01-05

### Added

- Initial release
- Agent detection for 12 coding agents: Claude Code, Cursor, Windsurf, GitHub Copilot, Aider, Cody, Continue, Code Puppy, Kiro, OpenCode, Goose, Amp
- `Init()` function to detect agent and set `AGENT_ENV` for child processes
- `Detect()` and `DetectWithEnv()` for agent detection
- `IsAgentContext()` and `RequireAgent()` helper functions
- Agent configuration paths (`UserConfigPath`, `ProjectConfigPath`)
- Context file discovery (`ContextFiles`)
- Installation detection (`IsInstalled`)
- `MockEnvironment` for testing
