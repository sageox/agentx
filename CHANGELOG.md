# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.13] - 2026-07-30

### Added

- **`HookEventPostToolUseFailure`**: canonical constant for the failed-tool-call event. Agents that define it (Goose, Cursor) fire `PostToolUse` only on **success**, so without a mapping for the failure event a failed turn is invisible to a consumer until the next successful tool call or `Stop`. Cursor spells it camelCase (`CursorEventPostToolUseFailure`); the canonical constant uses Goose's PascalCase.
- **Goose `PostToolUseFailure` → `PhaseAfterTool`**: `GooseAgent.EventPhases()` now covers seven events. Goose's four remaining events (`BeforeReadFile`, `AfterFileEdit`, `BeforeShellExecution`, `AfterShellExecution`) stay deliberately unmapped — each is a strict subset of `PreToolUse`/`PostToolUse`, since reading a file and running a shell command are both tool calls, so mapping them would fire a consumer twice per tool call for no additional signal. That reasoning is now recorded on the method.

## [0.1.12] - 2026-07-30

### Fixed

- **Goose lifecycle hooks**: Goose shipped a full hooks system in v1.38 (Open Plugins spec), but `GooseAgent` still reported `Hooks: false` and did not implement `LifecycleEventMapper` — so `AGENT_ENV=goose` was absent from `BuildEventPhaseMap()` and callers had to fall back to scanning every other agent's map to resolve a Goose event. Add `EventPhases()` covering `SessionStart`, `SessionEnd`, `PreToolUse`, `PostToolUse`, `UserPromptSubmit`, and `Stop`, plus `AgentENVAliases() == ["goose"]`. `PhaseCompact` is deliberately unmapped: Goose fires no compaction event, so mapping one would promise a re-prime that never happens.
- **Goose context files**: `ContextFiles()` returned `[".goose/config.yaml", ".goosehints"]`. `.goose/config.yaml` is configuration, not an instruction file, and the list omitted `AGENTS.md` entirely — which Goose loads *first*, ahead of `.goosehints`. Now returns `["AGENTS.md", ".goosehints"]` in Goose's own load order, so a caller injecting a context marker writes to the file Goose actually reads first.
- **Goose session support**: `SupportsSession()` was `false`. Goose supplies `session_id` on every hook payload, so it now returns `true`, following the OpenCode precedent. `SessionID(env)` still returns empty — Goose exposes no session-ID environment variable, and callers must read it from the hook payload.
- **Goose capabilities**: `Hooks: true`, `MinVersion: "1.38.0"` (the release hooks landed in), and the `ProjectContext` comment now names `AGENTS.md` alongside `.goosehints`.

## [0.1.11] - 2026-07-23

### Added

- **Buzz orchestrator**: Detection for Block's Buzz (`AgentTypeBuzz`), a chat/agent platform whose `buzz-acp` harness spawns Claude Code / Codex / Goose over ACP. Detects via `ORCHESTRATOR_ENV=buzz` (generic hatch), `BUZZ_ACP_AGENT_COMMAND` (best-effort env secondary), or — the reliable, config-independent signal — the `buzz-acp` process appearing in the ancestry chain. Deliberately does not sniff the workstation-scoped `BUZZ_RELAY_URL` / `BUZZ_PRIVATE_KEY`, which would false-tag a plain agent session in a Buzz user's shell.
- **`Environment.ProcessAncestry()`**: New interface method returning ancestor process executable base names (parent → PID 1). Unix walks a single `ps` snapshot in memory (no new dependency, works on macOS/BSD/Linux); Windows is a stub (falls back to env-var signals), matching ox's own `internal/proc` approach. Enables detecting orchestrators that spawn agents without setting a marker env var.

## [0.1.10] - 2026-05-08

### Added

- **Pi `PI_CODING_AGENT` detection**: `pi-mono`'s CLI sets `process.env.PI_CODING_AGENT="true"` at startup, which Node propagates to every subprocess pi spawns (including the bash tool). Add a truthy check (`"true"`/`"1"`) as the first, cheapest signal in `DetectRuntime`, ahead of `PI_CODING_AGENT_DIR` and the `$_` heuristic.

### Fixed

- Refresh stale Pi capability assertions to reflect prior changes: `Hooks` and `CustomCommands` are now `true` (Pi has a TypeScript extension bridge); `SupportsSession` is `true` and reads `PI_SESSION_ID`; remove Pi from `TestNonLifecycleAgents_DoNotImplementMapper`. Update the Pi spec in `agents_comprehensive_test.go` (capabilities, session env var, lifecycle metadata: `envAliases`, `eventPhaseCount=5`).

## [0.1.8] - 2026-04-21

### Added

- **Gas City orchestrator**: Detection for Gas City multi-agent orchestration framework (`AgentTypeGasCity`). Detects via `GASCITY=1`, `GC_VERSION`, `GC_RIG`, `GC_PACK`, `GC_RUN_ID`, or `ORCHESTRATOR_ENV=gascity`. Supports session ID via `GC_RUN_ID`.

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
