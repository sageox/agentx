# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
