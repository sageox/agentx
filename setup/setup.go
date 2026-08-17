// Package setup provides initialization for the agentx package.
// Import this package to register all default agents with the registry.
//
// Usage:
//
//	import _ "github.com/sageox/agentx/setup"
package setup

import (
	"github.com/sageox/agentx"
	"github.com/sageox/agentx/agents"
	"github.com/sageox/agentx/commands"
	"github.com/sageox/agentx/hooks"
	"github.com/sageox/agentx/orchestrators"
	"github.com/sageox/agentx/rules"
)

func init() {
	RegisterDefaultAgents()
}

// RegisterDefaultAgents registers all supported agents with the default registry.
// This is called automatically when this package is imported.
func RegisterDefaultAgents() {
	env := agentx.NewSystemEnvironment()

	// Claude Code with hook manager and command manager
	claudeCode := agents.NewClaudeCodeAgent()
	claudeCode.SetHookManager(hooks.NewClaudeCodeHookManager(env))
	claudeCode.SetCommandManager(commands.NewClaudeCodeCommandManager())
	claudeCode.SetRulesManager(rules.NewClaudeCodeRulesManager())
	agentx.DefaultRegistry.Register(claudeCode)

	// Cursor
	cursor := agents.NewCursorAgent()
	cursor.SetRulesManager(rules.NewCursorRulesManager())
	agentx.DefaultRegistry.Register(cursor)

	// Windsurf
	windsurf := agents.NewWindsurfAgent()
	windsurf.SetRulesManager(rules.NewWindsurfRulesManager())
	agentx.DefaultRegistry.Register(windsurf)

	// Copilot
	copilot := agents.NewCopilotAgent()
	copilot.SetRulesManager(rules.NewCopilotRulesManager())
	agentx.DefaultRegistry.Register(copilot)

	// Aider
	agentx.DefaultRegistry.Register(agents.NewAiderAgent())

	// Cody
	agentx.DefaultRegistry.Register(agents.NewCodyAgent())

	// Continue
	agentx.DefaultRegistry.Register(agents.NewContinueAgent())

	// Code Puppy
	agentx.DefaultRegistry.Register(agents.NewCodePuppyAgent())

	// Kiro
	kiro := agents.NewKiroAgent()
	kiro.SetRulesManager(rules.NewKiroRulesManager())
	agentx.DefaultRegistry.Register(kiro)

	// OpenCode
	agentx.DefaultRegistry.Register(agents.NewOpenCodeAgent())

	// Codex
	agentx.DefaultRegistry.Register(agents.NewCodexAgent())

	// Goose
	agentx.DefaultRegistry.Register(agents.NewGooseAgent())

	// Amp
	agentx.DefaultRegistry.Register(agents.NewAmpAgent())

	// Cline
	cline := agents.NewClineAgent()
	cline.SetRulesManager(rules.NewClineRulesManager())
	agentx.DefaultRegistry.Register(cline)

	// Droid (Factory.ai)
	droid := agents.NewDroidAgent()
	droid.SetRulesManager(rules.NewDroidRulesManager())
	agentx.DefaultRegistry.Register(droid)

	// OMP / Oh My Pi (github.com/can1357/oh-my-pi) — a Pi fork. Its runtime
	// detector is mutually exclusive with Pi's (see OMPAgent/PiAgent
	// DetectRuntime), so registration order does not affect detection.
	agentx.DefaultRegistry.Register(agents.NewOMPAgent())

	// Pi (shittycodingagent.ai)
	agentx.DefaultRegistry.Register(agents.NewPiAgent())

	// Gemini CLI
	agentx.DefaultRegistry.Register(agents.NewGeminiAgent())

	// Orchestrators
	agentx.DefaultRegistry.Register(orchestrators.NewOpenClawAgent())
	agentx.DefaultRegistry.Register(orchestrators.NewConductorAgent())
	agentx.DefaultRegistry.Register(orchestrators.NewGasCityAgent())
	agentx.DefaultRegistry.Register(orchestrators.NewBuzzAgent())
}
