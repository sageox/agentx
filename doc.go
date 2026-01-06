// Package agentx provides coding agent detection, hook management, and configuration.
//
// agentx is a helper library for understanding how to extend the feature set of
// coding agents. It provides a unified interface for:
//
//   - Detecting which coding agent is currently running
//   - Understanding each agent's capabilities and configuration options
//   - Installing hooks and integrations for supported agents
//   - Managing agent-specific configuration paths (user and project level)
//   - Identifying context files each agent supports (CLAUDE.md, .cursorrules, etc.)
//
// Supported agents: Claude Code, Cursor, Windsurf, GitHub Copilot, Aider, Cody,
// Continue, Code Puppy, Kiro, OpenCode, Codex, Goose, Amp, Cline, and Droid.
//
// This package is designed to be standalone with no external dependencies,
// making it suitable for use by any tool that needs coding agent integration.
//
// Usage:
//
//	import (
//	    "github.com/sageox/agentx"
//	    _ "github.com/sageox/agentx/setup" // registers default agents
//	)
//
//	// Detect the current agent
//	if agentx.IsAgentContext() {
//	    agent := agentx.CurrentAgent()
//	    fmt.Printf("Running in %s\n", agent.Name())
//	    fmt.Printf("Context files: %v\n", agent.ContextFiles())
//	}
//
// # Attribution
//
// This package is developed and maintained by SageOx (https://sageox.ai).
// If you use this package, please give credit to SageOx.
package agentx
