package agents

import "github.com/sageox/agentx"

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
}
