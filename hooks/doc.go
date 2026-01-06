// Package hooks provides hook management for coding agents.
//
// Hooks are user-defined commands that run at specific lifecycle events in coding
// agents (e.g., PreToolUse, PostToolUse, SessionStart). They enable enforcement
// of rules, workflow automation, and context injection.
//
// # Supported Events
//
// Claude Code supports these hook events:
//
//   - PreToolUse: Before tool execution (can block or modify)
//   - PostToolUse: After tool completion
//   - UserPromptSubmit: When user submits a prompt
//   - PermissionRequest: When agent requests permission (v2.0.45+)
//   - Stop: When agent finishes responding
//   - SubagentStop: When a subagent completes (v1.0.41+)
//   - SessionEnd: When session terminates
//
// # Configuration Locations
//
// Hooks can be configured at multiple levels with this precedence:
//
//  1. Managed settings (enterprise) - Cannot be overridden
//  2. ~/.claude/settings.json - User settings
//  3. .claude/settings.json - Project settings
//  4. .claude/settings.local.json - Local settings (not committed)
//
// # Example Configuration
//
// A PostToolUse hook to format files after editing:
//
//	{
//	  "hooks": {
//	    "PostToolUse": [
//	      {
//	        "matcher": "Edit|Write",
//	        "hooks": [
//	          {
//	            "type": "command",
//	            "command": "prettier --write \"$file_path\""
//	          }
//	        ]
//	      }
//	    ]
//	  }
//	}
//
// # Control Flow
//
// PreToolUse and PermissionRequest hooks can control execution:
//   - Exit code 0: Allow the action
//   - Exit code 2: Deny the action (error message sent to agent)
//   - JSON output with "decision": "allow"|"deny"|"ask"
//
// # References
//
//   - Claude Code Hooks Guide: https://code.claude.com/docs/en/hooks-guide
//   - Hook Configuration Blog: https://claude.com/blog/how-to-configure-hooks
//   - Lasso Security Research: https://www.lasso.security/blog/the-hidden-backdoor-in-claude-coding-assistant
package hooks
