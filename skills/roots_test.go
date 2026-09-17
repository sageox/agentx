package skills

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sageox/agentx"
)

// deliberatelyUnsupported lists every registered agent that has NO Agent Skills
// support, so that absence from rootsByAgent is a recorded decision rather than
// an oversight.
//
// Orchestrators are here because they drive other agents rather than loading
// skills themselves; the rest are agents the Agent Skills client showcase does
// not list as of 2026-09.
var deliberatelyUnsupported = map[agentx.AgentType]string{
	agentx.AgentTypeWindsurf:  "not listed as an Agent Skills client",
	agentx.AgentTypeAider:     "not listed as an Agent Skills client",
	agentx.AgentTypeCody:      "not listed as an Agent Skills client",
	agentx.AgentTypeContinue:  "not listed as an Agent Skills client",
	agentx.AgentTypeCodePuppy: "not listed as an Agent Skills client",
	agentx.AgentTypeCline:     "not listed as an Agent Skills client",
	agentx.AgentTypeOpenClaw:  "orchestrator; drives agents rather than loading skills",
	agentx.AgentTypeConductor: "orchestrator; drives agents rather than loading skills",
	agentx.AgentTypeGasCity:   "orchestrator; drives agents rather than loading skills",
	agentx.AgentTypeBuzz:      "orchestrator; drives agents rather than loading skills",
}

// TestEveryRegisteredAgentIsClassified is the anti-rot gate.
//
// agentx has no invariant tying a capability table back to the registry, which
// is how other per-agent tables silently fell behind the agent list. Adding an
// agent must fail this test until somebody decides where its skills go — or
// records, in deliberatelyUnsupported, that it has none.
func TestEveryRegisteredAgentIsClassified(t *testing.T) {
	for _, agent := range agentx.SupportedAgents {
		_, supported := rootsByAgent[agent]
		_, unsupported := deliberatelyUnsupported[agent]

		assert.Falsef(t, supported && unsupported,
			"%s is both classified as supporting skills and listed as unsupported", agent)
		assert.Truef(t, supported || unsupported,
			"%s is registered but unclassified: add it to rootsByAgent, or to deliberatelyUnsupported with a reason", agent)
	}

	for agent := range rootsByAgent {
		assert.Containsf(t, agentx.SupportedAgents, agent,
			"rootsByAgent has %s, which is not a registered agent", agent)
	}
	for agent := range deliberatelyUnsupported {
		assert.Containsf(t, agentx.SupportedAgents, agent,
			"deliberatelyUnsupported has %s, which is not a registered agent", agent)
	}
}

// TestWritePrefersTheCanonicalRoot encodes the one-root rule: a skill is
// installed once, into .agents/skills, unless the agent genuinely cannot read it.
//
// The holdouts are named explicitly so that an agent gaining canonical support
// upstream shows up here as a failing test rather than as a permanently
// duplicated install.
func TestWritePrefersTheCanonicalRoot(t *testing.T) {
	holdouts := map[agentx.AgentType]struct{ project, personal string }{
		agentx.AgentTypeClaudeCode: {".claude/skills", "~/.claude/skills"},
		agentx.AgentTypeKiro:       {".kiro/skills", "~/.kiro/skills"},
	}

	for agent := range rootsByAgent {
		project, ok := RootsFor(agent, ScopeProject)
		require.Truef(t, ok, "%s has no project roots", agent)
		personal, ok := RootsFor(agent, ScopePersonal)
		require.Truef(t, ok, "%s has no personal roots", agent)

		if want, isHoldout := holdouts[agent]; isHoldout {
			assert.Equal(t, want.project, project.Write, "%s project write root", agent)
			assert.Equal(t, want.personal, personal.Write, "%s personal write root", agent)
			assert.NotContainsf(t, project.Read, CanonicalProject,
				"%s is listed as a holdout but reads %s — it should no longer be one", agent, CanonicalProject)
			continue
		}

		assert.Equalf(t, CanonicalProject, project.Write,
			"%s must install into the canonical project root, not a native one", agent)
		assert.Equalf(t, CanonicalPersonal, personal.Write,
			"%s must install into the canonical personal root, not a native one", agent)
	}
}

// TestWriteRootIsAlwaysReadable guards the contract Read documents: Write first,
// and Write is always one of the places the agent actually looks. A write root
// missing from Read would mean installing into a directory nothing reads.
func TestWriteRootIsAlwaysReadable(t *testing.T) {
	for agent, byScope := range rootsByAgent {
		for scope, roots := range byScope {
			require.NotEmptyf(t, roots.Read, "%s/%s has no read roots", agent, scope)
			assert.Equalf(t, roots.Write, roots.Read[0],
				"%s/%s must list its write root first", agent, scope)
			assert.Lenf(t, dedupe(roots.Read), len(roots.Read),
				"%s/%s lists a duplicate read root", agent, scope)
		}
	}
}

// TestScopesDoNotLeakIntoEachOther catches the copy-paste error this table is
// most prone to: a personal "~/" path pasted into the project list, which would
// send a repo-scoped skill into the user's home directory.
func TestScopesDoNotLeakIntoEachOther(t *testing.T) {
	for agent, byScope := range rootsByAgent {
		for _, root := range byScope[ScopeProject].Read {
			assert.Falsef(t, strings.HasPrefix(root, "~/"),
				"%s project root %q is a personal path", agent, root)
		}
		for _, root := range byScope[ScopePersonal].Read {
			assert.Truef(t, strings.HasPrefix(root, "~/") || strings.HasPrefix(root, "/"),
				"%s personal root %q is not an absolute or home-relative path", agent, root)
		}
	}
}

func TestRootsForUnsupportedAgentReportsNotOk(t *testing.T) {
	for agent := range deliberatelyUnsupported {
		_, ok := RootsFor(agent, ScopeProject)
		assert.Falsef(t, ok, "%s must report no skills support", agent)
		assert.Falsef(t, Supported(agent), "%s must report no skills support", agent)
	}
}

func TestSupportedAgentsIsStableAndComplete(t *testing.T) {
	got := SupportedAgents()
	assert.Len(t, got, len(rootsByAgent))
	assert.Equal(t, got, SupportedAgents(), "order must not vary between calls")

	// Ordered by the canonical registry, so Claude Code precedes Gemini.
	claude := indexOf(got, agentx.AgentTypeClaudeCode)
	gemini := indexOf(got, agentx.AgentTypeGemini)
	require.NotEqual(t, -1, claude)
	require.NotEqual(t, -1, gemini)
	assert.Less(t, claude, gemini, "result must follow agentx.SupportedAgents order")
}

func dedupe(in []string) []string {
	seen := make(map[string]bool, len(in))
	var out []string
	for _, v := range in {
		if seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

func indexOf(list []agentx.AgentType, want agentx.AgentType) int {
	for i, v := range list {
		if v == want {
			return i
		}
	}
	return -1
}
