package agentx

func init() {
	RegisterDefaultAgents()
}

// RegisterDefaultAgents registers all supported agents with the default registry.
// This is called automatically when this package is imported.
func RegisterDefaultAgents() {
	// Register all supported agents
	_ = DefaultRegistry.Register(NewClaudeCodeAgent())
	_ = DefaultRegistry.Register(NewCursorAgent())
	_ = DefaultRegistry.Register(NewWindsurfAgent())
	_ = DefaultRegistry.Register(NewCopilotAgent())
	_ = DefaultRegistry.Register(NewAiderAgent())
	_ = DefaultRegistry.Register(NewCodyAgent())
	_ = DefaultRegistry.Register(NewContinueAgent())
	_ = DefaultRegistry.Register(NewCodePuppyAgent())
	_ = DefaultRegistry.Register(NewKiroAgent())
	_ = DefaultRegistry.Register(NewOpenCodeAgent())
	_ = DefaultRegistry.Register(NewGooseAgent())
	_ = DefaultRegistry.Register(NewAmpAgent())
}
