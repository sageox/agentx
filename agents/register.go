package agents

import "github.com/sageox/agentx"

// init registers all agent implementations with the DefaultRegistry.
// These agents are registered without HookManager or CommandManager.
// For fully-wired agents with hook and command management,
// import github.com/sageox/agentx/setup instead.
func init() {
	r := agentx.DefaultRegistry
	r.Register(NewClaudeCodeAgent())
	r.Register(NewCursorAgent())
	r.Register(NewWindsurfAgent())
	r.Register(NewCopilotAgent())
	r.Register(NewAiderAgent())
	r.Register(NewCodyAgent())
	r.Register(NewContinueAgent())
	r.Register(NewCodePuppyAgent())
	r.Register(NewKiroAgent())
	r.Register(NewOpenCodeAgent())
	r.Register(NewGooseAgent())
	r.Register(NewAmpAgent())
	r.Register(NewPiAgent())
}
