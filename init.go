package agentx

func init() {
	// Register all supported agents with the default registry
	DefaultRegistry.Register(NewClaudeCodeAgent())
	DefaultRegistry.Register(NewCursorAgent())
	DefaultRegistry.Register(NewWindsurfAgent())
	DefaultRegistry.Register(NewCopilotAgent())
	DefaultRegistry.Register(NewAiderAgent())
	DefaultRegistry.Register(NewCodyAgent())
	DefaultRegistry.Register(NewContinueAgent())
	DefaultRegistry.Register(NewCodePuppyAgent())
	DefaultRegistry.Register(NewKiroAgent())
	DefaultRegistry.Register(NewOpenCodeAgent())
	DefaultRegistry.Register(NewGooseAgent())
	DefaultRegistry.Register(NewAmpAgent())
}
