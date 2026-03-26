package agentx

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// ReadCommandFiles
// ---------------------------------------------------------------------------

func TestReadCommandFiles_MDFiles(t *testing.T) {
	fsys := fstest.MapFS{
		"commands/hello.md":   {Data: []byte("# Hello\n")},
		"commands/goodbye.md": {Data: []byte("# Goodbye\n")},
	}

	cmds, err := ReadCommandFiles(fsys, "commands")
	require.NoError(t, err)
	assert.Len(t, cmds, 2)

	names := make(map[string][]byte, len(cmds))
	for _, c := range cmds {
		names[c.Name] = c.Content
	}
	assert.Equal(t, []byte("# Hello\n"), names["hello.md"])
	assert.Equal(t, []byte("# Goodbye\n"), names["goodbye.md"])
}

func TestReadCommandFiles_MixedFiles(t *testing.T) {
	fsys := fstest.MapFS{
		"dir/readme.md":  {Data: []byte("# Readme\n")},
		"dir/config.json": {Data: []byte("{}")},
		"dir/notes.txt":  {Data: []byte("notes")},
		"dir/plan.md":    {Data: []byte("# Plan\n")},
	}

	cmds, err := ReadCommandFiles(fsys, "dir")
	require.NoError(t, err)
	assert.Len(t, cmds, 2, "only .md files should be returned")

	for _, c := range cmds {
		assert.Contains(t, []string{"readme.md", "plan.md"}, c.Name)
	}
}

func TestReadCommandFiles_SkipsSubdirectories(t *testing.T) {
	fsys := fstest.MapFS{
		"dir/top.md":       {Data: []byte("# Top\n")},
		"dir/sub/nested.md": {Data: []byte("# Nested\n")},
		"dir/sub":           {Mode: 1<<31 | 0o755},
	}

	cmds, err := ReadCommandFiles(fsys, "dir")
	require.NoError(t, err)
	// sub is a directory and should be skipped; nested.md is inside sub
	// and won't appear because ReadCommandFiles only reads entries directly in dir
	assert.Len(t, cmds, 1)
	assert.Equal(t, "top.md", cmds[0].Name)
}

func TestReadCommandFiles_EmptyDirectory(t *testing.T) {
	fsys := fstest.MapFS{
		"empty": {Mode: 1<<31 | 0o755},
	}

	cmds, err := ReadCommandFiles(fsys, "empty")
	require.NoError(t, err)
	assert.Empty(t, cmds)
}

func TestReadCommandFiles_NonExistentDirectory(t *testing.T) {
	fsys := fstest.MapFS{}

	_, err := ReadCommandFiles(fsys, "nonexistent")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ContentHash
// ---------------------------------------------------------------------------

func TestContentHash_KnownLength(t *testing.T) {
	hash := ContentHash([]byte("hello world"))
	assert.Len(t, hash, 12, "hash should be 12 hex characters")
}

func TestContentHash_EmptyContent(t *testing.T) {
	hash := ContentHash([]byte{})
	assert.Len(t, hash, 12)
	assert.NotEmpty(t, hash)
}

func TestContentHash_Deterministic(t *testing.T) {
	content := []byte("deterministic content")
	h1 := ContentHash(content)
	h2 := ContentHash(content)
	assert.Equal(t, h1, h2)
}

func TestContentHash_DifferentContent(t *testing.T) {
	h1 := ContentHash([]byte("content A"))
	h2 := ContentHash([]byte("content B"))
	assert.NotEqual(t, h1, h2)
}

// ---------------------------------------------------------------------------
// StampedContent
// ---------------------------------------------------------------------------

func TestStampedContent_Format(t *testing.T) {
	content := []byte("# My Command\nDo stuff\n")
	stamped := StampedContent(content, "1.2.3", DefaultStampPrefix)

	hash := ContentHash(content)
	expectedPrefix := "<!-- agentx-hash: " + hash + " ver: 1.2.3 -->\n"

	assert.True(t, len(stamped) > len(content))
	assert.Equal(t, expectedPrefix+"# My Command\nDo stuff\n", string(stamped))
}

func TestStampedContent_HashMatchesContentHash(t *testing.T) {
	content := []byte("body text")
	stamped := StampedContent(content, "0.1.0", DefaultStampPrefix)

	extracted := ExtractCommandHash(stamped, DefaultStampPrefix)
	assert.Equal(t, ContentHash(content), extracted)
}

// ---------------------------------------------------------------------------
// ExtractCommandHash
// ---------------------------------------------------------------------------

func TestExtractCommandHash_Valid(t *testing.T) {
	stamped := []byte("<!-- agentx-hash: aabbccdd1122 ver: 1.0.0 -->\n# Body\n")
	assert.Equal(t, "aabbccdd1122", ExtractCommandHash(stamped, DefaultStampPrefix))
}

func TestExtractCommandHash_NoStamp(t *testing.T) {
	assert.Equal(t, "", ExtractCommandHash([]byte("# Just content\n"), DefaultStampPrefix))
}

func TestExtractCommandHash_ShortStamp(t *testing.T) {
	assert.Equal(t, "", ExtractCommandHash([]byte("<!-- agentx-hash: abc -->"), DefaultStampPrefix))
}

func TestExtractCommandHash_DifferentPrefix(t *testing.T) {
	stamped := []byte("<!-- ox-hash: aabbccdd1122 ver: 1.0.0 -->\n# Body\n")
	assert.Equal(t, "", ExtractCommandHash(stamped, DefaultStampPrefix))
	assert.Equal(t, "aabbccdd1122", ExtractCommandHash(stamped, "ox"))
}

// ---------------------------------------------------------------------------
// ExtractStampVersion
// ---------------------------------------------------------------------------

func TestExtractStampVersion_Valid(t *testing.T) {
	stamped := []byte("<!-- agentx-hash: aabbccdd1122 ver: 2.5.1 -->\n# Body\n")
	assert.Equal(t, "2.5.1", ExtractStampVersion(stamped, DefaultStampPrefix))
}

func TestExtractStampVersion_NoVersionMarker(t *testing.T) {
	stamped := []byte("<!-- agentx-hash: aabbccdd1122 -->\n# Body\n")
	assert.Equal(t, "", ExtractStampVersion(stamped, DefaultStampPrefix))
}

func TestExtractStampVersion_NoStamp(t *testing.T) {
	assert.Equal(t, "", ExtractStampVersion([]byte("# Just content\n"), DefaultStampPrefix))
}

// ---------------------------------------------------------------------------
// CompareVersions
// ---------------------------------------------------------------------------

func TestCompareVersions_Table(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"1.0.0", "2.0.0", true},
		{"2.0.0", "1.0.0", false},
		{"1.0.0", "1.0.0", false},
		{"1.0.0", "1.0.1", true},
		{"1.0.0", "1.1.0", true},
		{"1.0", "1.0.1", true},
		{"1.0.1", "1.0", false},
		{"0.9.0", "0.10.0", true},
		{"0.10.0", "0.9.0", false},
		{"3.2.1", "3.2.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			assert.Equal(t, tt.want, CompareVersions(tt.a, tt.b),
				"%s < %s should be %v", tt.a, tt.b, tt.want)
		})
	}
}

// ---------------------------------------------------------------------------
// ShouldWriteCommand
// ---------------------------------------------------------------------------

func TestShouldWriteCommand_FileDoesNotExist(t *testing.T) {
	cmd := CommandFile{Name: "test.md", Content: []byte("# Test\n"), Version: "1.0.0"}
	assert.True(t, ShouldWriteCommand(nil, cmd, false, DefaultStampPrefix))
	assert.True(t, ShouldWriteCommand(nil, cmd, true, DefaultStampPrefix))
}

func TestShouldWriteCommand_ExistsNoOverwrite(t *testing.T) {
	cmd := CommandFile{Name: "test.md", Content: []byte("# Test\n"), Version: "1.0.0"}
	existing := StampedContent(cmd.Content, "1.0.0", DefaultStampPrefix)
	assert.False(t, ShouldWriteCommand(existing, cmd, false, DefaultStampPrefix))
}

func TestShouldWriteCommand_UserManaged(t *testing.T) {
	cmd := CommandFile{Name: "test.md", Content: []byte("# Test\n"), Version: "1.0.0"}
	// no stamp at all
	assert.False(t, ShouldWriteCommand([]byte("# User wrote this\n"), cmd, true, DefaultStampPrefix))
}

func TestShouldWriteCommand_IdenticalHash(t *testing.T) {
	cmd := CommandFile{Name: "test.md", Content: []byte("# Test\n"), Version: "1.0.0"}
	existing := StampedContent(cmd.Content, "1.0.0", DefaultStampPrefix)
	assert.False(t, ShouldWriteCommand(existing, cmd, true, DefaultStampPrefix))
}

func TestShouldWriteCommand_DowngradeGuard(t *testing.T) {
	// installed version is 2.0.0, incoming is 1.0.0 — block downgrade
	existing := StampedContent([]byte("# Newer\n"), "2.0.0", DefaultStampPrefix)
	cmd := CommandFile{Name: "test.md", Content: []byte("# Older\n"), Version: "1.0.0"}
	assert.False(t, ShouldWriteCommand(existing, cmd, true, DefaultStampPrefix))
}

func TestShouldWriteCommand_OlderInstalled(t *testing.T) {
	// installed version is 1.0.0, incoming is 2.0.0 — allow upgrade
	existing := StampedContent([]byte("# Old\n"), "1.0.0", DefaultStampPrefix)
	cmd := CommandFile{Name: "test.md", Content: []byte("# New\n"), Version: "2.0.0"}
	assert.True(t, ShouldWriteCommand(existing, cmd, true, DefaultStampPrefix))
}

func TestShouldWriteCommand_NoVersions(t *testing.T) {
	// hash differs but neither has a version
	existing := StampedContent([]byte("# Old\n"), "", DefaultStampPrefix)
	cmd := CommandFile{Name: "test.md", Content: []byte("# New\n"), Version: ""}
	assert.True(t, ShouldWriteCommand(existing, cmd, true, DefaultStampPrefix))
}

// ---------------------------------------------------------------------------
// IsCommandStale
// ---------------------------------------------------------------------------

func TestIsCommandStale_NoStamp(t *testing.T) {
	cmd := CommandFile{Name: "test.md", Content: []byte("# Test\n"), Version: "1.0.0"}
	assert.False(t, IsCommandStale([]byte("# User content\n"), cmd, DefaultStampPrefix))
}

func TestIsCommandStale_HashMatches(t *testing.T) {
	cmd := CommandFile{Name: "test.md", Content: []byte("# Test\n"), Version: "1.0.0"}
	existing := StampedContent(cmd.Content, "1.0.0", DefaultStampPrefix)
	assert.False(t, IsCommandStale(existing, cmd, DefaultStampPrefix))
}

func TestIsCommandStale_InstalledNewer(t *testing.T) {
	existing := StampedContent([]byte("# Newer\n"), "2.0.0", DefaultStampPrefix)
	cmd := CommandFile{Name: "test.md", Content: []byte("# Older\n"), Version: "1.0.0"}
	assert.False(t, IsCommandStale(existing, cmd, DefaultStampPrefix))
}

func TestIsCommandStale_InstalledOlder(t *testing.T) {
	existing := StampedContent([]byte("# Old\n"), "1.0.0", DefaultStampPrefix)
	cmd := CommandFile{Name: "test.md", Content: []byte("# Updated\n"), Version: "2.0.0"}
	assert.True(t, IsCommandStale(existing, cmd, DefaultStampPrefix))
}

// ---------------------------------------------------------------------------
// StampComment
// ---------------------------------------------------------------------------

func TestStampComment_DefaultPrefix(t *testing.T) {
	assert.Equal(t, "<!-- agentx-hash: ", StampComment("agentx"))
}

func TestStampComment_CustomPrefix(t *testing.T) {
	assert.Equal(t, "<!-- myapp-hash: ", StampComment("myapp"))
}

// ---------------------------------------------------------------------------
// DetectOriginFromOS
// ---------------------------------------------------------------------------

func TestDetectOriginFromOS_NoPanic(t *testing.T) {
	// just verify it returns a valid value without panicking
	origin := DetectOriginFromOS("")
	assert.True(t, origin.IsValid())
}

// ---------------------------------------------------------------------------
// SupportedAgents
// ---------------------------------------------------------------------------

func TestSupportedAgents_NoDuplicates(t *testing.T) {
	seen := make(map[AgentType]bool, len(SupportedAgents))
	for _, at := range SupportedAgents {
		assert.False(t, seen[at], "duplicate agent type: %s", at)
		seen[at] = true
	}
}

func TestSupportedAgents_ContainsExpected(t *testing.T) {
	expected := []AgentType{
		AgentTypeClaudeCode, AgentTypeCursor, AgentTypeWindsurf,
		AgentTypeCopilot, AgentTypeAider, AgentTypeOpenClaw,
	}
	for _, e := range expected {
		found := false
		for _, s := range SupportedAgents {
			if s == e {
				found = true
				break
			}
		}
		assert.True(t, found, "expected %s in SupportedAgents", e)
	}
}

// ---------------------------------------------------------------------------
// AllPhases
// ---------------------------------------------------------------------------

func TestAllPhases_ContainsExpected(t *testing.T) {
	expected := []Phase{
		PhaseStart, PhaseEnd, PhaseBeforeTool, PhaseAfterTool,
		PhasePrompt, PhaseStop, PhaseCompact,
	}
	assert.Equal(t, expected, AllPhases)
}

// ---------------------------------------------------------------------------
// firstLine (tested indirectly via public API)
// ---------------------------------------------------------------------------

func TestFirstLine_ViaExtractCommandHash(t *testing.T) {
	// content with multiple lines, stamp on first line only
	content := []byte("<!-- agentx-hash: aabbccdd1122 ver: 1.0.0 -->\n<!-- agentx-hash: zzzzzzzzzzz ver: 9.0.0 -->\n")
	assert.Equal(t, "aabbccdd1122", ExtractCommandHash(content, DefaultStampPrefix))
	assert.Equal(t, "1.0.0", ExtractStampVersion(content, DefaultStampPrefix))
}

func TestFirstLine_NoNewline(t *testing.T) {
	// single line, no trailing newline
	content := []byte("<!-- agentx-hash: aabbccdd1122 ver: 1.0.0 -->")
	assert.Equal(t, "aabbccdd1122", ExtractCommandHash(content, DefaultStampPrefix))
}
