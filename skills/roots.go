// Package skills describes where coding agents discover Agent Skills on disk,
// and validates the SKILL.md format those skills are written in.
//
// It is deliberately DATA AND VALIDATION ONLY — there is no installer here.
// Callers that materialize skills (ox, for one) already own digest tracking,
// retirement, and conflict handling for their whole managed inventory, and a
// second installer living in this library would be a parallel mechanism that
// drifts from the first. What a caller cannot reasonably derive on its own is
// which directories each vendor actually reads, so that is what this package
// supplies.
//
// # The one-root rule
//
// The ecosystem converged on .agents/skills as a shared discovery root. Most
// agents read it; two holdouts do not — Claude Code discovers only from
// .claude/skills, and Kiro only from .kiro/skills. Callers must therefore write
// to exactly one root per agent — Write — and never fan a skill out across every
// path in Read. Two copies of one skill on disk is two things to keep in sync,
// two things in a diff, and two answers when they disagree.
//
// Read exists for discovery and cleanup: finding a skill a user placed by hand,
// or noticing a copy left behind by an older layout. Never write to it.
package skills

import "github.com/sageox/agentx"

// Scope distinguishes a skill installed for one project from one installed for
// the user across every project.
type Scope string

const (
	// ScopeProject is repository-local: paths are relative to the project root.
	ScopeProject Scope = "project"

	// ScopePersonal is user-global: paths begin with "~/".
	ScopePersonal Scope = "personal"
)

// Canonical is the cross-agent discovery root the ecosystem settled on.
const (
	CanonicalProject  = ".agents/skills"
	CanonicalPersonal = "~/.agents/skills"
)

// Roots describes where one agent finds skills at one scope.
type Roots struct {
	// Write is the single directory to install into. Always the canonical root
	// when the agent reads it.
	Write string

	// Read lists every directory the agent discovers skills from, Write first.
	// Consult it to FIND skills; never write to anything but Write.
	Read []string
}

// rootsByAgent is the whole model. A registry-driven table rather than a method
// on each of the ~20 agent implementations: one boolean per agent file would be
// ~25 files touched per capability, and the resulting column rots silently when
// the next agent is added. TestEveryRegisteredAgentIsClassified holds this table
// to the registry instead.
//
// An agent absent from this table has no Agent Skills support and MUST NOT be
// given a skills directory. Absence is a claim, not an oversight — see the test.
var rootsByAgent = map[agentx.AgentType]map[Scope]Roots{
	// Claude Code does not read the canonical root; it discovers only from
	// .claude/skills. Upstream request:
	// https://github.com/anthropics/claude-code/issues/56193
	agentx.AgentTypeClaudeCode: {
		ScopeProject:  {Write: ".claude/skills", Read: []string{".claude/skills"}},
		ScopePersonal: {Write: "~/.claude/skills", Read: []string{"~/.claude/skills"}},
	},
	agentx.AgentTypeCodex: {
		ScopeProject: {Write: CanonicalProject, Read: []string{CanonicalProject}},
		// /etc/codex/skills is a machine-wide root an administrator populates. It
		// is readable but never ours to write.
		ScopePersonal: {Write: CanonicalPersonal, Read: []string{CanonicalPersonal, "/etc/codex/skills"}},
	},
	agentx.AgentTypeGemini: {
		ScopeProject:  {Write: CanonicalProject, Read: []string{CanonicalProject, ".gemini/skills"}},
		ScopePersonal: {Write: CanonicalPersonal, Read: []string{CanonicalPersonal, "~/.gemini/skills"}},
	},
	agentx.AgentTypeCursor: {
		ScopeProject:  {Write: CanonicalProject, Read: []string{CanonicalProject, ".cursor/skills", ".claude/skills", ".codex/skills"}},
		ScopePersonal: {Write: CanonicalPersonal, Read: []string{CanonicalPersonal, "~/.cursor/skills", "~/.claude/skills", "~/.codex/skills"}},
	},
	agentx.AgentTypeCopilot: {
		ScopeProject:  {Write: CanonicalProject, Read: []string{CanonicalProject, ".github/skills", ".claude/skills"}},
		ScopePersonal: {Write: CanonicalPersonal, Read: []string{CanonicalPersonal, "~/.copilot/skills"}},
	},
	agentx.AgentTypeDroid: {
		ScopeProject:  {Write: CanonicalProject, Read: []string{CanonicalProject, ".factory/skills", ".agent/skills"}},
		ScopePersonal: {Write: CanonicalPersonal, Read: []string{CanonicalPersonal, "~/.factory/skills", "~/.agent/skills"}},
	},
	agentx.AgentTypeOpenCode: {
		ScopeProject:  {Write: CanonicalProject, Read: []string{CanonicalProject, ".opencode/skills", ".claude/skills"}},
		ScopePersonal: {Write: CanonicalPersonal, Read: []string{CanonicalPersonal, "~/.config/opencode/skills", "~/.claude/skills"}},
	},
	agentx.AgentTypeAmp: {
		ScopeProject:  {Write: CanonicalProject, Read: []string{CanonicalProject, ".claude/skills"}},
		ScopePersonal: {Write: CanonicalPersonal, Read: []string{CanonicalPersonal, "~/.config/agents/skills", "~/.config/amp/skills", "~/.claude/skills"}},
	},
	agentx.AgentTypeGoose: {
		ScopeProject:  {Write: CanonicalProject, Read: []string{CanonicalProject, ".goose/skills", ".claude/skills"}},
		ScopePersonal: {Write: CanonicalPersonal, Read: []string{CanonicalPersonal, "~/.claude/skills"}},
	},
	agentx.AgentTypePi: {
		ScopeProject:  {Write: CanonicalProject, Read: []string{CanonicalProject, ".pi/skills"}},
		ScopePersonal: {Write: CanonicalPersonal, Read: []string{CanonicalPersonal, "~/.pi/agent/skills"}},
	},
	// Kiro is the other holdout: its docs list .kiro/skills and ~/.kiro/skills and
	// do not mention the canonical root, so writing there would install a skill
	// Kiro never loads.
	agentx.AgentTypeKiro: {
		ScopeProject:  {Write: ".kiro/skills", Read: []string{".kiro/skills"}},
		ScopePersonal: {Write: "~/.kiro/skills", Read: []string{"~/.kiro/skills"}},
	},
	agentx.AgentTypeOMP: {
		ScopeProject:  {Write: CanonicalProject, Read: []string{CanonicalProject}},
		ScopePersonal: {Write: CanonicalPersonal, Read: []string{CanonicalPersonal}},
	},
}

// RootsFor returns where the agent discovers skills at the given scope.
// ok is false when the agent has no Agent Skills support, in which case the
// caller must install nothing rather than guessing a directory.
func RootsFor(agent agentx.AgentType, scope Scope) (roots Roots, ok bool) {
	byScope, found := rootsByAgent[agent]
	if !found {
		return Roots{}, false
	}
	roots, ok = byScope[scope]
	return roots, ok
}

// Supported reports whether the agent supports Agent Skills at all.
func Supported(agent agentx.AgentType) bool {
	_, found := rootsByAgent[agent]
	return found
}

// SupportedAgents lists every agent with Agent Skills support.
func SupportedAgents() []agentx.AgentType {
	// Built from the canonical agent order, not from map iteration, so the result
	// is stable across runs and diffs cleanly in callers' output.
	out := make([]agentx.AgentType, 0, len(rootsByAgent))
	for _, agent := range agentx.SupportedAgents {
		if _, found := rootsByAgent[agent]; found {
			out = append(out, agent)
		}
	}
	return out
}
