package setup

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sageox/agentx"
)

// TestDetect_PriorityAcrossRealAgents runs detection against the full
// registered agent set, because the bugs it covers come from how agents
// compete, not from any one agent (sageox/agentx#17). An Amp orb's agent
// shell carries AMP_ORB=1 and none of Amp's runtime variables, and a repo
// with a checked-in .codex/ made Codex win at random.
func TestDetect_PriorityAcrossRealAgents(t *testing.T) {
	orig := agentx.DefaultRegistry
	defer func() { agentx.DefaultRegistry = orig }()
	agentx.DefaultRegistry = agentx.NewRegistry()
	RegisterDefaultAgents()

	codexDir := map[string]bool{".codex": true}
	tests := []struct {
		name string
		env  map[string]string
		dirs map[string]bool
		want agentx.AgentType
	}{
		{"Amp orb shell in a repo with .codex/", map[string]string{"AMP_ORB": "1"}, codexDir, agentx.AgentTypeAmp},
		{"real orb agent shell (thread id, no URL)", map[string]string{"AMP_ORB": "1", "AMP_THREAD_ID": "T-1", "AMP_URL": "https://ampcode.com"}, codexDir, agentx.AgentTypeAmp},
		{"Amp orb shell, nothing else", map[string]string{"AMP_ORB": "1"}, nil, agentx.AgentTypeAmp},
		{"Claude Code running inside an orb", map[string]string{"AMP_ORB": "1", "CLAUDECODE": "1"}, nil, agentx.AgentTypeClaudeCode},
		{"AGENT_ENV=codex inside an orb", map[string]string{"AMP_ORB": "1", "AGENT_ENV": "codex"}, nil, agentx.AgentTypeCodex},
		{"AGENT_ENV=amp beats a .codex/ hint", map[string]string{"AGENT_ENV": "amp"}, codexDir, agentx.AgentTypeAmp},
		{"AGENT_ENV is trimmed and case-insensitive", map[string]string{"AGENT_ENV": " Amp "}, codexDir, agentx.AgentTypeAmp},
		{"no signals but .codex/ still means Codex", map[string]string{}, codexDir, agentx.AgentTypeCodex},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := agentx.NewMockEnvironment(tt.env)
			env.ExistingDirs = tt.dirs
			// Map iteration order changes from run to run, so one pass can
			// pass by luck; repeat until an order-dependent winner would show.
			for i := 0; i < 200; i++ {
				got, err := agentx.NewDetectorWithEnv(env).Detect(context.Background())
				require.NoError(t, err)
				require.NotNil(t, got, "no agent detected on pass %d", i)
				require.Equal(t, tt.want, got.Type(), "wrong agent on pass %d", i)
			}
		})
	}
}

// TestRegistryList_IsSortedByType pins the order every detection phase
// walks, so "first match wins" names the same agent on every run.
func TestRegistryList_IsSortedByType(t *testing.T) {
	orig := agentx.DefaultRegistry
	defer func() { agentx.DefaultRegistry = orig }()
	agentx.DefaultRegistry = agentx.NewRegistry()
	RegisterDefaultAgents()

	for i := 0; i < 20; i++ {
		list := agentx.DefaultRegistry.List()
		for j := 1; j < len(list); j++ {
			require.Less(t, string(list[j-1].Type()), string(list[j].Type()), "List() out of order on pass %d", i)
		}
	}
}
