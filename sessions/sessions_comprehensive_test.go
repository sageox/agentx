package sessions

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock Provider ---

type mockProvider struct {
	agentType    string
	projects     []string
	projectsErr  error
	sessions     []SessionSummary
	sessionsErr  error
	session      *Session
	sessionErr   error
	scanMessages []*Message
	scanErr      error
}

func (m *mockProvider) AgentType() string { return m.agentType }

func (m *mockProvider) ListProjects() ([]string, error) {
	return m.projects, m.projectsErr
}

func (m *mockProvider) ListSessions(projectPath string) ([]SessionSummary, error) {
	return m.sessions, m.sessionsErr
}

func (m *mockProvider) ReadSession(sessionID string) (*Session, error) {
	return m.session, m.sessionErr
}

func (m *mockProvider) ScanMessages(sessionID string, fn func(*Message) error) error {
	if m.scanErr != nil {
		return m.scanErr
	}
	for _, msg := range m.scanMessages {
		if err := fn(msg); err != nil {
			return err
		}
	}
	return nil
}

// --- Registry Tests ---

func TestRegistry_RegisterAndRetrieve(t *testing.T) {
	mock := &mockProvider{agentType: "test-registry-1"}
	Register("test-registry-1", mock)
	defer func() {
		registryMu.Lock()
		delete(registry, "test-registry-1")
		registryMu.Unlock()
	}()

	got, err := ProviderFor("test-registry-1")
	require.NoError(t, err)
	assert.Equal(t, mock, got)
}

func TestRegistry_ProviderNotFound(t *testing.T) {
	_, err := ProviderFor("nonexistent-provider-abc-999")
	assert.ErrorIs(t, err, ErrProviderNotFound)
}

func TestRegistry_DefaultProvidersReturnsCopy(t *testing.T) {
	providers := DefaultProviders()
	require.NotNil(t, providers)

	// modifying returned map should not affect the registry
	providers["injected-fake"] = &mockProvider{agentType: "injected-fake"}

	providers2 := DefaultProviders()
	_, exists := providers2["injected-fake"]
	assert.False(t, exists, "modification to returned map must not leak into registry")
}

func TestRegistry_DefaultProvidersContainsBuiltins(t *testing.T) {
	providers := DefaultProviders()
	assert.Contains(t, providers, "claude-code")
	assert.Contains(t, providers, "gemini")
	assert.Contains(t, providers, "opencode")
}

func TestListAllSessions_WithMockProviders(t *testing.T) {
	now := time.Now()
	mock := &mockProvider{
		agentType: "test-list-all",
		projects:  []string{"/proj/a"},
		sessions: []SessionSummary{
			{ID: "s1", AgentType: "test-list-all", ProjectPath: "/proj/a", Created: now, Modified: now},
			{ID: "s2", AgentType: "test-list-all", ProjectPath: "/proj/a", Created: now, Modified: now},
		},
	}
	Register("test-list-all", mock)
	defer func() {
		registryMu.Lock()
		delete(registry, "test-list-all")
		registryMu.Unlock()
	}()

	result, err := ListAllSessions()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result["test-list-all"]), 2)
}

func TestListAllSessions_SkipsProviderWithProjectError(t *testing.T) {
	mock := &mockProvider{
		agentType:   "test-skip-proj",
		projectsErr: errors.New("disk on fire"),
	}
	Register("test-skip-proj", mock)
	defer func() {
		registryMu.Lock()
		delete(registry, "test-skip-proj")
		registryMu.Unlock()
	}()

	result, err := ListAllSessions()
	require.NoError(t, err)
	// should not have entries for the failing provider
	assert.Empty(t, result["test-skip-proj"])
}

func TestListAllSessions_SkipsProviderWithSessionError(t *testing.T) {
	mock := &mockProvider{
		agentType:   "test-skip-sess",
		projects:    []string{"/proj/x"},
		sessionsErr: errors.New("corrupted index"),
	}
	Register("test-skip-sess", mock)
	defer func() {
		registryMu.Lock()
		delete(registry, "test-skip-sess")
		registryMu.Unlock()
	}()

	result, err := ListAllSessions()
	require.NoError(t, err)
	assert.Empty(t, result["test-skip-sess"])
}

// --- Claude Code: messageType ---

func TestClaudeCode_MessageType(t *testing.T) {
	p := &claudeCodeProvider{}

	tests := []struct {
		input    string
		expected MessageType
	}{
		{"tool-use", TypeToolUse},
		{"tool-result", TypeToolResult},
		{"progress", TypeProgress},
		{"snapshot", TypeSnapshot},
		{"", TypeChat},
		{"unknown", TypeChat},
		{"message", TypeChat},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, p.messageType(tc.input))
		})
	}
}

// --- Claude Code: extractContentWithoutThinking ---

func TestClaudeCode_ExtractContentWithoutThinking(t *testing.T) {
	p := &claudeCodeProvider{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with thinking block",
			input:    "<thinking>internal reasoning</thinking>visible text",
			expected: "visible text",
		},
		{
			name:     "no thinking block",
			input:    "just regular content",
			expected: "just regular content",
		},
		{
			name:     "thinking at end with no trailing text",
			input:    "<thinking>stuff</thinking>",
			expected: "",
		},
		{
			name:     "multiple closing tags uses text between first pair",
			input:    "<thinking>a</thinking>b</thinking>c",
			expected: "b",
		},
		{
			name:     "whitespace after thinking",
			input:    "<thinking>blah</thinking>  trimmed  ",
			expected: "trimmed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, p.extractContentWithoutThinking(tc.input))
		})
	}
}

// --- Claude Code: contentBlocksToString ---

func TestClaudeCode_ContentBlocksToString(t *testing.T) {
	p := &claudeCodeProvider{}

	tests := []struct {
		name     string
		blocks   []ccContentBlock
		expected string
	}{
		{
			name: "multiple text blocks joined",
			blocks: []ccContentBlock{
				{Type: "text", Text: "hello"},
				{Type: "text", Text: "world"},
			},
			expected: "hello\nworld",
		},
		{
			name: "thinking blocks filtered",
			blocks: []ccContentBlock{
				{Type: "thinking", Text: "internal"},
				{Type: "text", Text: "visible"},
			},
			expected: "visible",
		},
		{
			name: "empty text blocks filtered",
			blocks: []ccContentBlock{
				{Type: "text", Text: ""},
				{Type: "text", Text: "content"},
			},
			expected: "content",
		},
		{
			name:     "empty input",
			blocks:   []ccContentBlock{},
			expected: "",
		},
		{
			name:     "nil input",
			blocks:   nil,
			expected: "",
		},
		{
			name: "all thinking blocks",
			blocks: []ccContentBlock{
				{Type: "thinking", Text: "a"},
				{Type: "thinking", Text: "b"},
			},
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, p.contentBlocksToString(tc.blocks))
		})
	}
}

// --- Claude Code: parseMessage ---

func TestClaudeCode_ParseMessage(t *testing.T) {
	p := &claudeCodeProvider{}

	t.Run("version type returns nil", func(t *testing.T) {
		raw := ccRawEntry{Type: "version"}
		assert.Nil(t, p.parseMessage(raw, "sess-1"))
	})

	t.Run("empty type returns nil", func(t *testing.T) {
		raw := ccRawEntry{Type: ""}
		assert.Nil(t, p.parseMessage(raw, "sess-1"))
	})

	t.Run("string content parsed", func(t *testing.T) {
		raw := ccRawEntry{
			Type:    "message",
			Role:    "user",
			Content: json.RawMessage(`"hello world"`),
		}
		msg := p.parseMessage(raw, "sess-1")
		require.NotNil(t, msg)
		assert.Equal(t, "hello world", msg.Content)
		assert.Equal(t, "user", msg.Role)
	})

	t.Run("content block array parsed", func(t *testing.T) {
		raw := ccRawEntry{
			Type:    "message",
			Role:    "assistant",
			Content: json.RawMessage(`[{"type":"text","text":"hi"},{"type":"text","text":"there"}]`),
		}
		msg := p.parseMessage(raw, "sess-1")
		require.NotNil(t, msg)
		assert.Equal(t, "hi\nthere", msg.Content)
	})

	t.Run("single content block parsed", func(t *testing.T) {
		raw := ccRawEntry{
			Type:    "message",
			Role:    "assistant",
			Content: json.RawMessage(`{"type":"text","text":"single block"}`),
		}
		msg := p.parseMessage(raw, "sess-1")
		require.NotNil(t, msg)
		assert.Equal(t, "single block", msg.Content)
	})

	t.Run("with timestamp parsed correctly", func(t *testing.T) {
		ts := "2024-06-15T10:30:00Z"
		raw := ccRawEntry{
			Type:      "message",
			Role:      "user",
			Content:   json.RawMessage(`"test"`),
			Timestamp: ts,
		}
		msg := p.parseMessage(raw, "sess-1")
		require.NotNil(t, msg)
		expected, _ := time.Parse(time.RFC3339, ts)
		assert.Equal(t, expected, msg.Timestamp)
	})

	t.Run("without timestamp uses current time", func(t *testing.T) {
		before := time.Now().Add(-time.Second)
		raw := ccRawEntry{
			Type:    "message",
			Role:    "user",
			Content: json.RawMessage(`"test"`),
		}
		msg := p.parseMessage(raw, "sess-1")
		require.NotNil(t, msg)
		assert.True(t, msg.Timestamp.After(before), "timestamp should be recent")
	})

	t.Run("with usage stats populated", func(t *testing.T) {
		raw := ccRawEntry{
			Type:    "message",
			Role:    "assistant",
			Content: json.RawMessage(`"response"`),
			Usage: &ccUsage{
				Input:      100,
				Output:     50,
				CacheRead:  20,
				CacheWrite: 10,
			},
		}
		msg := p.parseMessage(raw, "sess-1")
		require.NotNil(t, msg)
		require.NotNil(t, msg.Usage)
		assert.Equal(t, int64(100), msg.Usage.InputTokens)
		assert.Equal(t, int64(50), msg.Usage.OutputTokens)
		assert.Equal(t, int64(20), msg.Usage.CacheRead)
		assert.Equal(t, int64(10), msg.Usage.CacheWrite)
	})

	t.Run("without usage is nil", func(t *testing.T) {
		raw := ccRawEntry{
			Type:    "message",
			Role:    "user",
			Content: json.RawMessage(`"test"`),
		}
		msg := p.parseMessage(raw, "sess-1")
		require.NotNil(t, msg)
		assert.Nil(t, msg.Usage)
	})

	t.Run("with parentId", func(t *testing.T) {
		raw := ccRawEntry{
			Type:     "message",
			Role:     "assistant",
			Content:  json.RawMessage(`"reply"`),
			ParentID: "parent-123",
		}
		msg := p.parseMessage(raw, "sess-1")
		require.NotNil(t, msg)
		assert.Equal(t, "parent-123", msg.ParentID)
	})

	t.Run("with model", func(t *testing.T) {
		raw := ccRawEntry{
			Type:    "message",
			Role:    "assistant",
			Model:   "claude-3-5-sonnet",
			Content: json.RawMessage(`"answer"`),
		}
		msg := p.parseMessage(raw, "sess-1")
		require.NotNil(t, msg)
		assert.Equal(t, "claude-3-5-sonnet", msg.Model)
	})

	t.Run("string content with thinking block", func(t *testing.T) {
		raw := ccRawEntry{
			Type:    "message",
			Role:    "assistant",
			Content: json.RawMessage(`"<thinking>reasoning</thinking>the answer"`),
		}
		msg := p.parseMessage(raw, "sess-1")
		require.NotNil(t, msg)
		assert.Equal(t, "the answer", msg.Content)
	})

	t.Run("nil content", func(t *testing.T) {
		raw := ccRawEntry{
			Type: "message",
			Role: "user",
		}
		msg := p.parseMessage(raw, "sess-1")
		require.NotNil(t, msg)
		assert.Equal(t, "", msg.Content)
	})
}

// --- Claude Code: encodeProjectPath / decodeProjectPath ---

func TestClaudeCode_EncodeDecodePath(t *testing.T) {
	t.Run("encode replaces slashes with dashes", func(t *testing.T) {
		assert.Equal(t, "-Users-person-code", encodeProjectPath("/Users/person/code"))
	})

	t.Run("encode root path", func(t *testing.T) {
		assert.Equal(t, "-", encodeProjectPath("/"))
	})

	t.Run("decode starting with dash", func(t *testing.T) {
		assert.Equal(t, "/Users/person/code", decodeProjectPath("-Users-person-code"))
	})

	t.Run("decode not starting with dash returns as-is", func(t *testing.T) {
		assert.Equal(t, "relative-path", decodeProjectPath("relative-path"))
	})

	t.Run("round-trip encode then decode", func(t *testing.T) {
		original := "/Users/person/projects/myapp"
		encoded := encodeProjectPath(original)
		decoded := decodeProjectPath(encoded)
		assert.Equal(t, original, decoded)
	})
}

// --- Claude Code: file-based ListProjects ---

func TestClaudeCode_ListProjects_MultipleProjects(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	projectsDir := filepath.Join(tmpDir, ".claude", "projects")
	require.NoError(t, os.MkdirAll(projectsDir, 0755))

	// create two valid project dirs with sessions-index.json
	for _, path := range []string{"/Users/person/project1", "/Users/person/project2"} {
		encoded := encodeProjectPath(path)
		dir := filepath.Join(projectsDir, encoded)
		require.NoError(t, os.MkdirAll(dir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "sessions-index.json"), []byte(`{"sessions":{}}`), 0644))
	}

	// create a file (not a dir) that should be skipped
	require.NoError(t, os.WriteFile(filepath.Join(projectsDir, "not-a-dir"), []byte("nope"), 0644))

	// create a dir without sessions-index.json (should be skipped)
	require.NoError(t, os.MkdirAll(filepath.Join(projectsDir, "-Users-person-empty"), 0755))

	p := &claudeCodeProvider{}
	projects, err := p.ListProjects()
	require.NoError(t, err)
	assert.Len(t, projects, 2)
	// should be sorted
	assert.Equal(t, "/Users/person/project1", projects[0])
	assert.Equal(t, "/Users/person/project2", projects[1])
}

func TestClaudeCode_ListProjects_NoDir(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	// don't create .claude/projects at all

	p := &claudeCodeProvider{}
	projects, err := p.ListProjects()
	assert.NoError(t, err)
	assert.Nil(t, projects)
}

// --- Claude Code: file-based ListSessions ---

func TestClaudeCode_ListSessions_FieldsPopulated(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	projectsDir := filepath.Join(tmpDir, ".claude", "projects")
	projectPath := "/Users/person/myproject"
	encoded := encodeProjectPath(projectPath)
	dir := filepath.Join(projectsDir, encoded)
	require.NoError(t, os.MkdirAll(dir, 0755))

	now := time.Now()
	earlier := now.Add(-time.Hour)

	index := ccSessionsIndex{
		Sessions: map[string]ccIndexEntry{
			"sess-a": {
				ProjectPath: projectPath,
				GitBranch:   "main",
				FirstPrompt: "do something",
				Summary:     "did something",
				Created:     earlier.UnixMilli(),
				Modified:    earlier.UnixMilli(),
			},
			"sess-b": {
				ProjectPath: projectPath,
				GitBranch:   "feature",
				FirstPrompt: "fix bug",
				Summary:     "fixed bug",
				Created:     now.UnixMilli(),
				Modified:    now.UnixMilli(),
			},
		},
	}
	data, err := json.Marshal(index)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sessions-index.json"), data, 0644))

	p := &claudeCodeProvider{}
	summaries, err := p.ListSessions(projectPath)
	require.NoError(t, err)
	assert.Len(t, summaries, 2)

	// sorted by Modified desc, so sess-b first
	assert.Equal(t, "sess-b", summaries[0].ID)
	assert.Equal(t, "sess-a", summaries[1].ID)

	// verify fields
	assert.Equal(t, "claude-code", summaries[0].AgentType)
	assert.Equal(t, "feature", summaries[0].GitBranch)
	assert.Equal(t, "fix bug", summaries[0].FirstPrompt)
	assert.Equal(t, "fixed bug", summaries[0].Summary)
}

func TestClaudeCode_ListSessions_ProjectNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	p := &claudeCodeProvider{}
	_, err := p.ListSessions("/no/such/project")
	assert.ErrorIs(t, err, ErrProjectNotFound)
}

// --- Claude Code: file-based ReadSession ---

func TestClaudeCode_ReadSession_WithVersionAndMessages(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	projectsDir := filepath.Join(tmpDir, ".claude", "projects")
	projectPath := "/Users/person/proj"
	encoded := encodeProjectPath(projectPath)
	dir := filepath.Join(projectsDir, encoded)
	require.NoError(t, os.MkdirAll(dir, 0755))

	sessionID := "test-read-sess"

	// write sessions-index.json
	index := ccSessionsIndex{
		Sessions: map[string]ccIndexEntry{
			sessionID: {
				ProjectPath: projectPath,
				GitBranch:   "main",
				FirstPrompt: "hello",
				Summary:     "greeting",
				Created:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli(),
				Modified:    time.Date(2024, 1, 1, 0, 1, 0, 0, time.UTC).UnixMilli(),
			},
		},
	}
	indexData, _ := json.Marshal(index)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sessions-index.json"), indexData, 0644))

	// write JSONL session file
	lines := []string{
		`{"type":"version","version":"1.0"}`,
		`{"type":"human","role":"user","content":"hello","timestamp":"2024-01-01T00:00:00Z"}`,
		`{"type":"assistant","role":"assistant","content":[{"type":"text","text":"hi"}],"timestamp":"2024-01-01T00:01:00Z"}`,
	}
	sessionFile := filepath.Join(dir, sessionID+".jsonl")
	f, err := os.Create(sessionFile)
	require.NoError(t, err)
	for _, line := range lines {
		_, _ = f.WriteString(line + "\n")
	}
	f.Close()

	p := &claudeCodeProvider{}
	session, err := p.ReadSession(sessionID)
	require.NoError(t, err)
	assert.Equal(t, sessionID, session.ID)
	assert.Equal(t, "claude-code", session.AgentType)
	assert.Len(t, session.Messages, 2)

	assert.Equal(t, MessageRole("user"), session.Messages[0].Role)
	assert.Equal(t, "hello", session.Messages[0].Content)

	assert.Equal(t, MessageRole("assistant"), session.Messages[1].Role)
	assert.Equal(t, "hi", session.Messages[1].Content)
}

func TestClaudeCode_ReadSession_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	projectsDir := filepath.Join(tmpDir, ".claude", "projects")
	require.NoError(t, os.MkdirAll(projectsDir, 0755))

	p := &claudeCodeProvider{}
	_, err := p.ReadSession("nonexistent-session")
	assert.ErrorIs(t, err, ErrSessionNotFound)
}

func TestClaudeCode_ReadSession_NoProjectsDir(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	// no .claude/projects at all

	p := &claudeCodeProvider{}
	_, err := p.ReadSession("any-session")
	assert.Error(t, err)
}

// --- Claude Code: file-based ScanMessages ---

func TestClaudeCode_ScanMessages_WithVersionSkipped(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	projectsDir := filepath.Join(tmpDir, ".claude", "projects")
	projectPath := "/Users/person/scan"
	encoded := encodeProjectPath(projectPath)
	dir := filepath.Join(projectsDir, encoded)
	require.NoError(t, os.MkdirAll(dir, 0755))

	sessionID := "scan-test"

	lines := []string{
		`{"type":"version","version":"1.0"}`,
		`{"type":"message","role":"user","content":"question","timestamp":"2024-01-01T00:00:00Z"}`,
		`{"type":"message","role":"assistant","content":"answer","timestamp":"2024-01-01T00:01:00Z"}`,
	}
	sessionFile := filepath.Join(dir, sessionID+".jsonl")
	f, _ := os.Create(sessionFile)
	for _, line := range lines {
		_, _ = f.WriteString(line + "\n")
	}
	f.Close()

	p := &claudeCodeProvider{}
	var messages []*Message
	err := p.ScanMessages(sessionID, func(msg *Message) error {
		messages = append(messages, msg)
		return nil
	})
	require.NoError(t, err)
	assert.Len(t, messages, 2) // version skipped
	assert.Equal(t, "question", messages[0].Content)
	assert.Equal(t, "answer", messages[1].Content)
}

func TestClaudeCode_ScanMessages_InvalidJSONSkipped(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	projectsDir := filepath.Join(tmpDir, ".claude", "projects")
	dir := filepath.Join(projectsDir, "-Users-person-invalid")
	require.NoError(t, os.MkdirAll(dir, 0755))

	sessionID := "invalid-json-test"
	lines := []string{
		`not valid json`,
		`{"type":"message","role":"user","content":"valid","timestamp":"2024-01-01T00:00:00Z"}`,
		`{broken`,
	}
	sessionFile := filepath.Join(dir, sessionID+".jsonl")
	f, _ := os.Create(sessionFile)
	for _, line := range lines {
		_, _ = f.WriteString(line + "\n")
	}
	f.Close()

	p := &claudeCodeProvider{}
	var count int
	err := p.ScanMessages(sessionID, func(msg *Message) error {
		count++
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, count) // only the valid message
}

func TestClaudeCode_ScanMessages_CallbackErrorStopsScanning(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	projectsDir := filepath.Join(tmpDir, ".claude", "projects")
	dir := filepath.Join(projectsDir, "-Users-person-stop")
	require.NoError(t, os.MkdirAll(dir, 0755))

	sessionID := "stop-test"
	lines := []string{
		`{"type":"message","role":"user","content":"first","timestamp":"2024-01-01T00:00:00Z"}`,
		`{"type":"message","role":"user","content":"second","timestamp":"2024-01-01T00:01:00Z"}`,
		`{"type":"message","role":"user","content":"third","timestamp":"2024-01-01T00:02:00Z"}`,
	}
	sessionFile := filepath.Join(dir, sessionID+".jsonl")
	f, _ := os.Create(sessionFile)
	for _, line := range lines {
		_, _ = f.WriteString(line + "\n")
	}
	f.Close()

	stopErr := errors.New("stop now")
	p := &claudeCodeProvider{}
	var count int
	err := p.ScanMessages(sessionID, func(msg *Message) error {
		count++
		if count == 2 {
			return stopErr
		}
		return nil
	})
	assert.ErrorIs(t, err, stopErr)
	assert.Equal(t, 2, count)
}

func TestClaudeCode_ScanMessages_SessionNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	projectsDir := filepath.Join(tmpDir, ".claude", "projects")
	require.NoError(t, os.MkdirAll(projectsDir, 0755))

	p := &claudeCodeProvider{}
	err := p.ScanMessages("missing-session", func(msg *Message) error {
		return nil
	})
	assert.ErrorIs(t, err, ErrSessionNotFound)
}

func TestClaudeCode_ScanMessages_SkipsNonDirEntries(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	projectsDir := filepath.Join(tmpDir, ".claude", "projects")
	require.NoError(t, os.MkdirAll(projectsDir, 0755))

	// create a file in projects dir (not a directory)
	require.NoError(t, os.WriteFile(filepath.Join(projectsDir, "stray-file"), []byte("x"), 0644))

	// create an actual project with session
	dir := filepath.Join(projectsDir, "-Users-person-scanskip")
	require.NoError(t, os.MkdirAll(dir, 0755))
	sessionID := "scanskip-test"
	line := `{"type":"message","role":"user","content":"found","timestamp":"2024-01-01T00:00:00Z"}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, sessionID+".jsonl"), []byte(line+"\n"), 0644))

	p := &claudeCodeProvider{}
	var count int
	err := p.ScanMessages(sessionID, func(msg *Message) error {
		count++
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// --- Gemini: extractContent edge cases ---

func TestGemini_ExtractContent(t *testing.T) {
	p := &geminiProvider{}

	tests := []struct {
		name     string
		input    json.RawMessage
		expected string
	}{
		{"string content", json.RawMessage(`"hello world"`), "hello world"},
		{"part object", json.RawMessage(`{"text":"hello part"}`), "hello part"},
		{"parts array", json.RawMessage(`[{"text":"a"},{"text":"b"}]`), "a\nb"},
		{"parts with empty text filtered", json.RawMessage(`[{"text":""},{"text":"ok"}]`), "ok"},
		{"invalid json", json.RawMessage(`{broken`), ""},
		{"null", json.RawMessage(`null`), ""},
		{"number fallthrough", json.RawMessage(`42`), ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, p.extractContent(tc.input))
		})
	}
}

// --- Gemini: roleFromAuthor ---

func TestGemini_RoleFromAuthor(t *testing.T) {
	p := &geminiProvider{}

	tests := []struct {
		author   string
		expected MessageRole
	}{
		{"user", RoleUser},
		{"gemini", RoleAssistant},
		{"system", RoleSystem},
		{"unknown", RoleUser},
		{"USER", RoleUser},     // case-insensitive
		{"Gemini", RoleAssistant}, // case-insensitive
		{"SYSTEM", RoleSystem},
		{"", RoleUser},
	}

	for _, tc := range tests {
		t.Run(tc.author, func(t *testing.T) {
			assert.Equal(t, tc.expected, p.roleFromAuthor(tc.author))
		})
	}
}

// --- Gemini: messageType ---

func TestGemini_MessageType(t *testing.T) {
	p := &geminiProvider{}

	t.Run("with tool calls", func(t *testing.T) {
		calls := []gemToolCall{{ID: "1", Name: "read_file"}}
		assert.Equal(t, TypeToolUse, p.messageType(calls))
	})

	t.Run("without tool calls", func(t *testing.T) {
		assert.Equal(t, TypeChat, p.messageType(nil))
	})

	t.Run("empty tool calls", func(t *testing.T) {
		assert.Equal(t, TypeChat, p.messageType([]gemToolCall{}))
	})
}

// --- Gemini: hashProjectPath ---

func TestGemini_HashProjectPath(t *testing.T) {
	t.Run("produces md5 hex", func(t *testing.T) {
		input := "/Users/person/project"
		expected := fmt.Sprintf("%x", md5.Sum([]byte(input)))
		assert.Equal(t, expected, hashProjectPath(input))
	})

	t.Run("different inputs produce different hashes", func(t *testing.T) {
		h1 := hashProjectPath("/path/a")
		h2 := hashProjectPath("/path/b")
		assert.NotEqual(t, h1, h2)
	})

	t.Run("same input produces same hash", func(t *testing.T) {
		assert.Equal(t, hashProjectPath("/foo"), hashProjectPath("/foo"))
	})
}

// --- Gemini: file-based tests ---

func TestGemini_ListProjects_NoDir(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	p := &geminiProvider{}
	projects, err := p.ListProjects()
	assert.NoError(t, err)
	assert.Nil(t, projects)
}

func TestGemini_ListSessions_ProjectNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	p := &geminiProvider{}
	_, err := p.ListSessions("nonexistent-hash")
	assert.ErrorIs(t, err, ErrProjectNotFound)
}

func TestGemini_ReadSession_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// create the gemini tmp dir but no sessions
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".gemini", "tmp"), 0755))

	p := &geminiProvider{}
	_, err := p.ReadSession("nonexistent-session-id")
	assert.ErrorIs(t, err, ErrSessionNotFound)
}

func TestGemini_ScanMessages_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".gemini", "tmp"), 0755))

	p := &geminiProvider{}
	err := p.ScanMessages("nonexistent-session", func(msg *Message) error { return nil })
	assert.ErrorIs(t, err, ErrSessionNotFound)
}

func TestGemini_ScanMessages_CallbackError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	geminiDir := filepath.Join(tmpDir, ".gemini")
	require.NoError(t, os.MkdirAll(geminiDir, 0755))

	projectHash := "proj-cb-err"
	sessionID := "sess-cb-err"
	createTestGeminiSession(t, geminiDir, projectHash, sessionID)

	p := &geminiProvider{}
	stopErr := errors.New("halt")
	err := p.ScanMessages(sessionID, func(msg *Message) error {
		return stopErr
	})
	assert.ErrorIs(t, err, stopErr)
}

func TestGemini_ScanMessages_TokenUsage(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	geminiDir := filepath.Join(tmpDir, ".gemini")
	require.NoError(t, os.MkdirAll(geminiDir, 0755))

	projectHash := "proj-tokens"
	sessionID := "sess-tokens"
	createTestGeminiSession(t, geminiDir, projectHash, sessionID)

	p := &geminiProvider{}
	var messages []*Message
	err := p.ScanMessages(sessionID, func(msg *Message) error {
		messages = append(messages, msg)
		return nil
	})
	require.NoError(t, err)
	assert.Len(t, messages, 2)

	// first message (user) has no tokens
	assert.Nil(t, messages[0].Usage)

	// second message (gemini) has tokens
	require.NotNil(t, messages[1].Usage)
	assert.Equal(t, int64(10), messages[1].Usage.InputTokens)
	assert.Equal(t, int64(5), messages[1].Usage.OutputTokens)
}

func TestGemini_ListSessions_LongFirstPromptTruncated(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	geminiDir := filepath.Join(tmpDir, ".gemini")
	projectHash := "proj-trunc"
	sessionID := "sess-trunc"

	chatsDir := filepath.Join(geminiDir, "tmp", projectHash, "chats", sessionID)
	require.NoError(t, os.MkdirAll(chatsDir, 0755))

	// create a long first prompt (> 200 chars)
	longPrompt := ""
	for i := 0; i < 250; i++ {
		longPrompt += "x"
	}

	record := gemConversationRecord{
		SessionID:   sessionID,
		ProjectHash: projectHash,
		StartTime:   time.Now().UnixMilli(),
		LastUpdated: time.Now().UnixMilli(),
		Messages: []gemMessageRecord{
			{
				Author:    "user",
				Timestamp: time.Now().UnixMilli(),
				Content:   json.RawMessage(fmt.Sprintf(`"%s"`, longPrompt)),
			},
		},
	}

	data, _ := json.Marshal(record)
	require.NoError(t, os.WriteFile(filepath.Join(chatsDir, "session.json"), data, 0644))

	p := &geminiProvider{}
	summaries, err := p.ListSessions(projectHash)
	require.NoError(t, err)
	require.Len(t, summaries, 1)
	assert.Len(t, summaries[0].FirstPrompt, 203) // 200 + "..."
	assert.True(t, len(summaries[0].FirstPrompt) <= 203)
}

func TestGemini_ReadSession_WithToolCalls(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	geminiDir := filepath.Join(tmpDir, ".gemini")
	projectHash := "proj-tools"
	sessionID := "sess-tools"

	chatsDir := filepath.Join(geminiDir, "tmp", projectHash, "chats", sessionID)
	require.NoError(t, os.MkdirAll(chatsDir, 0755))

	record := gemConversationRecord{
		SessionID:   sessionID,
		ProjectHash: projectHash,
		StartTime:   time.Now().UnixMilli(),
		LastUpdated: time.Now().UnixMilli(),
		Messages: []gemMessageRecord{
			{
				Author:    "user",
				Timestamp: time.Now().UnixMilli(),
				Content:   json.RawMessage(`"read this file"`),
			},
			{
				Author:    "gemini",
				Timestamp: time.Now().UnixMilli(),
				Content:   json.RawMessage(`"reading file"`),
				ToolCalls: []gemToolCall{{ID: "tc-1", Name: "read_file"}},
				Tokens:    gemTokens{Input: 50, Output: 25, Cached: 10},
			},
		},
	}

	data, _ := json.Marshal(record)
	require.NoError(t, os.WriteFile(filepath.Join(chatsDir, "session.json"), data, 0644))

	p := &geminiProvider{}
	session, err := p.ReadSession(sessionID)
	require.NoError(t, err)
	assert.Len(t, session.Messages, 2)

	// user message is TypeChat
	assert.Equal(t, TypeChat, session.Messages[0].Type)

	// assistant message with tool calls is TypeToolUse
	assert.Equal(t, TypeToolUse, session.Messages[1].Type)
	require.NotNil(t, session.Messages[1].Usage)
	assert.Equal(t, int64(50), session.Messages[1].Usage.InputTokens)
	assert.Equal(t, int64(10), session.Messages[1].Usage.CacheRead)
}

func TestGemini_ListSessions_SkipsNonDirEntries(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	projectHash := "proj-skipfile"
	chatsDir := filepath.Join(tmpDir, ".gemini", "tmp", projectHash, "chats")
	require.NoError(t, os.MkdirAll(chatsDir, 0755))

	// create a file in chats dir (not a subdir)
	require.NoError(t, os.WriteFile(filepath.Join(chatsDir, "stray.txt"), []byte("x"), 0644))

	p := &geminiProvider{}
	summaries, err := p.ListSessions(projectHash)
	require.NoError(t, err)
	assert.Empty(t, summaries)
}

func TestGemini_ListProjects_SkipsNonDirEntries(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	tmpGemini := filepath.Join(tmpDir, ".gemini", "tmp")
	require.NoError(t, os.MkdirAll(tmpGemini, 0755))

	// create a file (not a dir) in tmp
	require.NoError(t, os.WriteFile(filepath.Join(tmpGemini, "stray-file"), []byte("x"), 0644))

	// create a dir without chats subdir (should be skipped)
	require.NoError(t, os.MkdirAll(filepath.Join(tmpGemini, "no-chats"), 0755))

	// create a valid project
	validHash := "valid-project"
	require.NoError(t, os.MkdirAll(filepath.Join(tmpGemini, validHash, "chats"), 0755))

	p := &geminiProvider{}
	projects, err := p.ListProjects()
	require.NoError(t, err)
	assert.Len(t, projects, 1)
	assert.Equal(t, validHash, projects[0])
}

// --- OpenCode Provider ---

func TestOpenCode_AgentType(t *testing.T) {
	p := &opencodeProvider{}
	assert.Equal(t, "opencode", p.AgentType())
}

func TestOpenCode_AllMethodsReturnErrors(t *testing.T) {
	p := &opencodeProvider{}

	t.Run("ListProjects", func(t *testing.T) {
		_, err := p.ListProjects()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SQLite")
	})

	t.Run("ListSessions", func(t *testing.T) {
		_, err := p.ListSessions("/any")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SQLite")
	})

	t.Run("ReadSession", func(t *testing.T) {
		_, err := p.ReadSession("any-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SQLite")
	})

	t.Run("ScanMessages", func(t *testing.T) {
		err := p.ScanMessages("any-id", func(msg *Message) error { return nil })
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SQLite")
	})
}

func TestDatabasePath_WithXDGDataHome(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// database doesn't exist yet
	assert.Equal(t, "", DatabasePath())

	// create the db file
	dbDir := filepath.Join(tmpDir, "opencode")
	require.NoError(t, os.MkdirAll(dbDir, 0755))
	dbFile := filepath.Join(dbDir, "opencode.db")
	require.NoError(t, os.WriteFile(dbFile, []byte("sqlite"), 0644))

	assert.Equal(t, dbFile, DatabasePath())
}

func TestDatabasePath_WithoutXDGDataHome(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")

	// without the db file existing, should return empty
	result := DatabasePath()
	// it returns either the path (if it exists) or empty
	// we can't guarantee the file exists on CI, so just verify no panic
	_ = result
}

// --- Claude Code: ReadSession skips non-dir entries in projects ---

func TestClaudeCode_ReadSession_SkipsNonDirEntries(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	projectsDir := filepath.Join(tmpDir, ".claude", "projects")
	require.NoError(t, os.MkdirAll(projectsDir, 0755))

	// put a stray file in projects dir
	require.NoError(t, os.WriteFile(filepath.Join(projectsDir, "stray.txt"), []byte("x"), 0644))

	// create actual project with session
	projectPath := "/Users/person/skiptest"
	encoded := encodeProjectPath(projectPath)
	dir := filepath.Join(projectsDir, encoded)
	require.NoError(t, os.MkdirAll(dir, 0755))

	sessionID := "skip-nondir-sess"
	index := ccSessionsIndex{
		Sessions: map[string]ccIndexEntry{
			sessionID: {
				ProjectPath: projectPath,
				Created:     time.Now().UnixMilli(),
				Modified:    time.Now().UnixMilli(),
			},
		},
	}
	indexData, _ := json.Marshal(index)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sessions-index.json"), indexData, 0644))

	line := `{"type":"message","role":"user","content":"hi","timestamp":"2024-01-01T00:00:00Z"}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, sessionID+".jsonl"), []byte(line+"\n"), 0644))

	p := &claudeCodeProvider{}
	session, err := p.ReadSession(sessionID)
	require.NoError(t, err)
	assert.Equal(t, sessionID, session.ID)
	assert.Len(t, session.Messages, 1)
}

// --- Claude Code: ReadSession with usage stats ---

func TestClaudeCode_ReadSession_WithUsageStats(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	projectsDir := filepath.Join(tmpDir, ".claude", "projects")
	projectPath := "/Users/person/usage"
	encoded := encodeProjectPath(projectPath)
	dir := filepath.Join(projectsDir, encoded)
	require.NoError(t, os.MkdirAll(dir, 0755))

	sessionID := "usage-test"
	index := ccSessionsIndex{
		Sessions: map[string]ccIndexEntry{
			sessionID: {
				ProjectPath: projectPath,
				Created:     time.Now().UnixMilli(),
				Modified:    time.Now().UnixMilli(),
			},
		},
	}
	indexData, _ := json.Marshal(index)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sessions-index.json"), indexData, 0644))

	lines := []string{
		`{"type":"message","role":"assistant","content":"response","timestamp":"2024-01-01T00:00:00Z","usage":{"input_tokens":100,"output_tokens":50,"cache_read_tokens":20,"cache_write_tokens":10}}`,
	}
	sessionFile := filepath.Join(dir, sessionID+".jsonl")
	f, _ := os.Create(sessionFile)
	for _, line := range lines {
		_, _ = f.WriteString(line + "\n")
	}
	f.Close()

	p := &claudeCodeProvider{}
	session, err := p.ReadSession(sessionID)
	require.NoError(t, err)
	require.Len(t, session.Messages, 1)
	require.NotNil(t, session.Messages[0].Usage)
	assert.Equal(t, int64(100), session.Messages[0].Usage.InputTokens)
	assert.Equal(t, int64(50), session.Messages[0].Usage.OutputTokens)
	assert.Equal(t, int64(20), session.Messages[0].Usage.CacheRead)
	assert.Equal(t, int64(10), session.Messages[0].Usage.CacheWrite)
}

// --- Gemini: ReadSession first prompt truncation ---

func TestGemini_ReadSession_FirstPromptTruncation(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	geminiDir := filepath.Join(tmpDir, ".gemini")
	projectHash := "proj-fp-trunc"
	sessionID := "sess-fp-trunc"

	chatsDir := filepath.Join(geminiDir, "tmp", projectHash, "chats", sessionID)
	require.NoError(t, os.MkdirAll(chatsDir, 0755))

	longPrompt := ""
	for i := 0; i < 300; i++ {
		longPrompt += "y"
	}

	record := gemConversationRecord{
		SessionID:   sessionID,
		ProjectHash: projectHash,
		StartTime:   time.Now().UnixMilli(),
		LastUpdated: time.Now().UnixMilli(),
		Messages: []gemMessageRecord{
			{
				Author:    "user",
				Timestamp: time.Now().UnixMilli(),
				Content:   json.RawMessage(fmt.Sprintf(`"%s"`, longPrompt)),
			},
		},
	}

	data, _ := json.Marshal(record)
	require.NoError(t, os.WriteFile(filepath.Join(chatsDir, "session.json"), data, 0644))

	p := &geminiProvider{}
	session, err := p.ReadSession(sessionID)
	require.NoError(t, err)
	assert.Len(t, session.FirstPrompt, 203) // 200 + "..."
}

// --- Gemini: ReadSession no messages ---

func TestGemini_ReadSession_NoMessages(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	geminiDir := filepath.Join(tmpDir, ".gemini")
	projectHash := "proj-empty"
	sessionID := "sess-empty"

	chatsDir := filepath.Join(geminiDir, "tmp", projectHash, "chats", sessionID)
	require.NoError(t, os.MkdirAll(chatsDir, 0755))

	record := gemConversationRecord{
		SessionID:   sessionID,
		ProjectHash: projectHash,
		StartTime:   time.Now().UnixMilli(),
		LastUpdated: time.Now().UnixMilli(),
		Messages:    []gemMessageRecord{},
	}

	data, _ := json.Marshal(record)
	require.NoError(t, os.WriteFile(filepath.Join(chatsDir, "session.json"), data, 0644))

	p := &geminiProvider{}
	session, err := p.ReadSession(sessionID)
	require.NoError(t, err)
	assert.Empty(t, session.Messages)
	assert.Empty(t, session.FirstPrompt)
}

// --- Claude Code: ScanMessages message type mapping ---

func TestClaudeCode_ScanMessages_MessageTypeMapping(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	projectsDir := filepath.Join(tmpDir, ".claude", "projects")
	dir := filepath.Join(projectsDir, "-Users-person-types")
	require.NoError(t, os.MkdirAll(dir, 0755))

	sessionID := "type-mapping-test"
	lines := []string{
		`{"type":"tool-use","role":"assistant","content":"calling tool","timestamp":"2024-01-01T00:00:00Z"}`,
		`{"type":"tool-result","role":"tool","content":"tool output","timestamp":"2024-01-01T00:01:00Z"}`,
		`{"type":"progress","role":"assistant","content":"working...","timestamp":"2024-01-01T00:02:00Z"}`,
	}
	sessionFile := filepath.Join(dir, sessionID+".jsonl")
	f, _ := os.Create(sessionFile)
	for _, line := range lines {
		_, _ = f.WriteString(line + "\n")
	}
	f.Close()

	p := &claudeCodeProvider{}
	var messages []*Message
	err := p.ScanMessages(sessionID, func(msg *Message) error {
		messages = append(messages, msg)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, messages, 3)
	assert.Equal(t, TypeToolUse, messages[0].Type)
	assert.Equal(t, TypeToolResult, messages[1].Type)
	assert.Equal(t, TypeProgress, messages[2].Type)
}
