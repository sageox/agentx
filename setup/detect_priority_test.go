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
		{"Claude Code inside a real orb (inherits AMP_THREAD_ID)", map[string]string{"AMP_ORB": "1", "AMP_THREAD_ID": "T-1", "AMP_URL": "https://ampcode.com", "CLAUDECODE": "1"}, codexDir, agentx.AgentTypeClaudeCode},
		{"AGENT_ENV=codex inside an orb", map[string]string{"AMP_ORB": "1", "AGENT_ENV": "codex"}, nil, agentx.AgentTypeCodex},
		// agents with no runtime check have only Detect(); their own marker
		// still outranks the orb they run in
		{"Goose inside a real orb", map[string]string{"AMP_ORB": "1", "AMP_THREAD_ID": "T-1", "GOOSE": "1"}, codexDir, agentx.AgentTypeGoose},
		{"Cursor's agent inside an orb", map[string]string{"AMP_ORB": "1", "CURSOR_AGENT": "1"}, codexDir, agentx.AgentTypeCursor},
		{"an agent's own marker beats a .codex/ hint", map[string]string{"CURSOR_AGENT": "1"}, codexDir, agentx.AgentTypeCursor},
		// an alias counts even for an agent that maps no lifecycle events
		{"AGENT_ENV=oh-my-pi (an alias) inside an orb", map[string]string{"AMP_ORB": "1", "AGENT_ENV": "oh-my-pi"}, codexDir, agentx.AgentTypeOMP},
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

// everythingExists reports every path as present and every binary as
// installed, so a Detect() that reads the filesystem or PATH shows it.
type everythingExists struct{ *agentx.MockEnvironment }

func (everythingExists) FileExists(string) bool            { return true }
func (everythingExists) IsDir(string) bool                 { return true }
func (everythingExists) LookPath(n string) (string, error) { return "/usr/local/bin/" + n, nil }
func (everythingExists) ReadFile(string) ([]byte, error)   { return []byte("{}"), nil }

// TestDetect_AgentsWithoutRuntimeChecksUseNoFilesystemHints pins what phase 3
// of detection relies on: an agent without DetectRuntime ranks its Detect()
// ahead of an orb because that Detect() only reads the agent's own markers.
// One that read a folder or PATH would let any repo outrank the orb, which
// is sageox/agentx#17 again. Such an agent needs DetectRuntime first.
func TestDetect_AgentsWithoutRuntimeChecksUseNoFilesystemHints(t *testing.T) {
	orig := agentx.DefaultRegistry
	defer func() { agentx.DefaultRegistry = orig }()
	agentx.DefaultRegistry = agentx.NewRegistry()
	RegisterDefaultAgents()

	env := everythingExists{agentx.NewMockEnvironment(nil)}
	checked, hinted := 0, 0
	for _, agent := range agentx.DefaultRegistry.List() {
		if _, ok := agent.(agentx.RuntimeDetector); ok {
			if detected, _ := agent.Detect(context.Background(), env); detected {
				hinted++
			}
			continue
		}
		checked++
		detected, err := agent.Detect(context.Background(), env)
		require.NoError(t, err)
		require.False(t, detected, "%s: Detect() matched on the filesystem or PATH alone", agent.Type())
	}
	require.NotZero(t, checked, "no agent without DetectRuntime was checked")
	// Codex reads .codex/, so the probe must make at least one agent match
	require.NotZero(t, hinted, "the probe environment shows no filesystem hint")
}
