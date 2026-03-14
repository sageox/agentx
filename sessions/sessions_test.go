package sessions

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Test helpers

func createTestCCProjectDir(t *testing.T, dir string, projectPath string) {
	encodedPath := encodeProjectPath(projectPath)
	projectDir := filepath.Join(dir, encodedPath)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project directory: %v", err)
	}
}

func createTestCCSession(t *testing.T, dir string, projectPath string, sessionID string) {
	encodedPath := encodeProjectPath(projectPath)
	projectDir := filepath.Join(dir, encodedPath)

	// Create sessions-index.json
	index := ccSessionsIndex{
		Sessions: map[string]ccIndexEntry{
			sessionID: {
				ProjectPath: projectPath,
				GitBranch:   "main",
				FirstPrompt: "test prompt",
				Summary:     "test summary",
				Created:     time.Now().UnixMilli(),
				Modified:    time.Now().UnixMilli(),
			},
		},
	}

	indexData, err := json.Marshal(index)
	if err != nil {
		t.Fatalf("failed to marshal index: %v", err)
	}

	indexPath := filepath.Join(projectDir, "sessions-index.json")
	if err := os.WriteFile(indexPath, indexData, 0644); err != nil {
		t.Fatalf("failed to write index: %v", err)
	}

	// Create session JSONL file
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	f, err := os.Create(sessionFile)
	if err != nil {
		t.Fatalf("failed to create session file: %v", err)
	}
	defer f.Close()

	// Write test messages
	messages := []map[string]interface{}{
		{
			"type":      "message",
			"role":      "user",
			"content":   "hello",
			"timestamp": time.Now().Format(time.RFC3339),
		},
		{
			"type":      "message",
			"role":      "assistant",
			"model":     "claude-3-5-sonnet",
			"content":   "hi there",
			"timestamp": time.Now().Format(time.RFC3339),
			"usage": map[string]int64{
				"input_tokens":  10,
				"output_tokens": 5,
			},
		},
	}

	for _, msg := range messages {
		data, _ := json.Marshal(msg)
		f.Write(data)
		f.WriteString("\n")
	}
}

func createTestGeminiSession(t *testing.T, dir string, projectHash string, sessionID string) {
	chatsDir := filepath.Join(dir, "tmp", projectHash, "chats", sessionID)
	if err := os.MkdirAll(chatsDir, 0755); err != nil {
		t.Fatalf("failed to create chats directory: %v", err)
	}

	record := gemConversationRecord{
		SessionID:   sessionID,
		ProjectHash: projectHash,
		StartTime:   time.Now().UnixMilli(),
		LastUpdated: time.Now().UnixMilli(),
		Summary:     "test summary",
		Messages: []gemMessageRecord{
			{
				Author:    "user",
				Type:      "text",
				Timestamp: time.Now().UnixMilli(),
				Content:   json.RawMessage(`"hello"`),
			},
			{
				Author:    "gemini",
				Type:      "text",
				Timestamp: time.Now().UnixMilli(),
				Content:   json.RawMessage(`"hi"`),
				Tokens: gemTokens{
					Input:  10,
					Output: 5,
				},
			},
		},
	}

	sessionFile := filepath.Join(chatsDir, "session.json")
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("failed to marshal session: %v", err)
	}

	if err := os.WriteFile(sessionFile, data, 0644); err != nil {
		t.Fatalf("failed to write session: %v", err)
	}
}

// Registry Tests

func TestRegisterProvider(t *testing.T) {
	provider := &testProvider{}
	Register("test-agent", provider)

	retrieved, err := ProviderFor("test-agent")
	if err != nil {
		t.Fatalf("ProviderFor failed: %v", err)
	}

	if retrieved != provider {
		t.Errorf("retrieved provider is not the same as registered")
	}
}

func TestProviderNotFound(t *testing.T) {
	_, err := ProviderFor("nonexistent-agent-xyz-123")
	if err != ErrProviderNotFound {
		t.Errorf("expected ErrProviderNotFound, got %v", err)
	}
}

func TestDefaultProviders(t *testing.T) {
	providers := DefaultProviders()
	if providers == nil {
		t.Errorf("DefaultProviders returned nil")
	}

	if len(providers) < 2 {
		t.Errorf("expected at least 2 providers (claude-code, gemini), got %d", len(providers))
	}

	if _, ok := providers["claude-code"]; !ok {
		t.Errorf("claude-code provider not found")
	}

	if _, ok := providers["gemini"]; !ok {
		t.Errorf("gemini provider not found")
	}
}

func TestListAllSessions(t *testing.T) {
	// This test is basic since it depends on actual file system
	sessions, err := ListAllSessions()
	if err != nil {
		t.Fatalf("ListAllSessions failed: %v", err)
	}

	if sessions == nil {
		t.Errorf("expected non-nil sessions map")
	}
}

// Claude Code Provider Tests

func TestClaudeCodeAgentType(t *testing.T) {
	provider := &claudeCodeProvider{}
	if provider.AgentType() != "claude-code" {
		t.Errorf("expected agent type 'claude-code', got '%s'", provider.AgentType())
	}
}

func TestClaudeCodeListProjects(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	projectsDir := filepath.Join(tmpDir, ".claude", "projects")
	os.MkdirAll(projectsDir, 0755)

	createTestCCProjectDir(t, projectsDir, "/test/project/1")
	createTestCCSession(t, projectsDir, "/test/project/1", "session-1")

	provider := &claudeCodeProvider{}
	projects, err := provider.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects failed: %v", err)
	}

	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}
}

func TestClaudeCodeListSessions(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	projectsDir := filepath.Join(tmpDir, ".claude", "projects")
	os.MkdirAll(projectsDir, 0755)

	projectPath := "/test/project/1"
	sessionID := "session-1"

	createTestCCProjectDir(t, projectsDir, projectPath)
	createTestCCSession(t, projectsDir, projectPath, sessionID)

	provider := &claudeCodeProvider{}
	summaries, err := provider.ListSessions(projectPath)
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}

	if len(summaries) != 1 {
		t.Errorf("expected 1 session, got %d", len(summaries))
	}

	if summaries[0].ID != sessionID {
		t.Errorf("expected session ID '%s', got '%s'", sessionID, summaries[0].ID)
	}

	if summaries[0].AgentType != "claude-code" {
		t.Errorf("expected agent type 'claude-code', got '%s'", summaries[0].AgentType)
	}
}

func TestClaudeCodeReadSession(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	projectsDir := filepath.Join(tmpDir, ".claude", "projects")
	os.MkdirAll(projectsDir, 0755)

	projectPath := "/test/project/1"
	sessionID := "session-1"

	createTestCCProjectDir(t, projectsDir, projectPath)
	createTestCCSession(t, projectsDir, projectPath, sessionID)

	provider := &claudeCodeProvider{}
	session, err := provider.ReadSession(sessionID)
	if err != nil {
		t.Fatalf("ReadSession failed: %v", err)
	}

	if session.ID != sessionID {
		t.Errorf("expected session ID '%s', got '%s'", sessionID, session.ID)
	}

	if len(session.Messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(session.Messages))
	}

	if session.Messages[0].Role != RoleUser {
		t.Errorf("expected first message role 'user', got '%s'", session.Messages[0].Role)
	}

	if session.Messages[1].Role != RoleAssistant {
		t.Errorf("expected second message role 'assistant', got '%s'", session.Messages[1].Role)
	}
}

func TestClaudeCodeScanMessages(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	projectsDir := filepath.Join(tmpDir, ".claude", "projects")
	os.MkdirAll(projectsDir, 0755)

	projectPath := "/test/project/1"
	sessionID := "session-1"

	createTestCCProjectDir(t, projectsDir, projectPath)
	createTestCCSession(t, projectsDir, projectPath, sessionID)

	provider := &claudeCodeProvider{}
	messageCount := 0
	err := provider.ScanMessages(sessionID, func(msg *Message) error {
		messageCount++
		return nil
	})

	if err != nil {
		t.Fatalf("ScanMessages failed: %v", err)
	}

	if messageCount != 2 {
		t.Errorf("expected 2 messages, got %d", messageCount)
	}
}

func TestClaudeCodeMessageTyping(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	projectsDir := filepath.Join(tmpDir, ".claude", "projects")
	os.MkdirAll(projectsDir, 0755)

	provider := &claudeCodeProvider{}

	if provider.messageType("tool-use") != TypeToolUse {
		t.Errorf("expected TypeToolUse")
	}

	if provider.messageType("tool-result") != TypeToolResult {
		t.Errorf("expected TypeToolResult")
	}

	if provider.messageType("progress") != TypeProgress {
		t.Errorf("expected TypeProgress")
	}

	if provider.messageType("message") != TypeChat {
		t.Errorf("expected TypeChat")
	}
}

func TestClaudeCodeEncoding(t *testing.T) {
	projectPath := "/Users/ryan/conductor/workspaces/test"
	encoded := encodeProjectPath(projectPath)
	decoded := decodeProjectPath(encoded)

	if decoded != projectPath {
		t.Errorf("encode/decode mismatch: %s -> %s -> %s", projectPath, encoded, decoded)
	}
}

// Gemini Provider Tests

func TestGeminiAgentType(t *testing.T) {
	provider := &geminiProvider{}
	if provider.AgentType() != "gemini" {
		t.Errorf("expected agent type 'gemini', got '%s'", provider.AgentType())
	}
}

func TestGeminiListProjects(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	geminiDir := filepath.Join(tmpDir, ".gemini")
	os.MkdirAll(geminiDir, 0755)

	createTestGeminiSession(t, geminiDir, "project-hash-1", "session-1")

	provider := &geminiProvider{}
	projects, err := provider.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects failed: %v", err)
	}

	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}
}

func TestGeminiListSessions(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	geminiDir := filepath.Join(tmpDir, ".gemini")
	os.MkdirAll(geminiDir, 0755)

	projectHash := "project-hash-1"
	sessionID := "session-1"

	createTestGeminiSession(t, geminiDir, projectHash, sessionID)

	provider := &geminiProvider{}
	summaries, err := provider.ListSessions(projectHash)
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}

	if len(summaries) != 1 {
		t.Errorf("expected 1 session, got %d", len(summaries))
	}

	if summaries[0].ID != sessionID {
		t.Errorf("expected session ID '%s', got '%s'", sessionID, summaries[0].ID)
	}

	if summaries[0].AgentType != "gemini" {
		t.Errorf("expected agent type 'gemini', got '%s'", summaries[0].AgentType)
	}
}

func TestGeminiReadSession(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	geminiDir := filepath.Join(tmpDir, ".gemini")
	os.MkdirAll(geminiDir, 0755)

	projectHash := "project-hash-1"
	sessionID := "session-1"

	createTestGeminiSession(t, geminiDir, projectHash, sessionID)

	provider := &geminiProvider{}
	session, err := provider.ReadSession(sessionID)
	if err != nil {
		t.Fatalf("ReadSession failed: %v", err)
	}

	if session.ID != sessionID {
		t.Errorf("expected session ID '%s', got '%s'", sessionID, session.ID)
	}

	if len(session.Messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(session.Messages))
	}

	if session.Messages[0].Role != RoleUser {
		t.Errorf("expected first message role 'user', got '%s'", session.Messages[0].Role)
	}

	if session.Messages[1].Role != RoleAssistant {
		t.Errorf("expected second message role 'assistant', got '%s'", session.Messages[1].Role)
	}
}

func TestGeminiScanMessages(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	geminiDir := filepath.Join(tmpDir, ".gemini")
	os.MkdirAll(geminiDir, 0755)

	projectHash := "project-hash-1"
	sessionID := "session-1"

	createTestGeminiSession(t, geminiDir, projectHash, sessionID)

	provider := &geminiProvider{}
	messageCount := 0
	err := provider.ScanMessages(sessionID, func(msg *Message) error {
		messageCount++
		return nil
	})

	if err != nil {
		t.Fatalf("ScanMessages failed: %v", err)
	}

	if messageCount != 2 {
		t.Errorf("expected 2 messages, got %d", messageCount)
	}
}

func TestGeminiContentExtraction(t *testing.T) {
	provider := &geminiProvider{}

	// Test string content
	stringContent := json.RawMessage(`"hello world"`)
	if provider.extractContent(stringContent) != "hello world" {
		t.Errorf("failed to extract string content")
	}

	// Test part object
	partContent := json.RawMessage(`{"text":"hello part"}`)
	if provider.extractContent(partContent) != "hello part" {
		t.Errorf("failed to extract part content")
	}

	// Test parts array
	partsContent := json.RawMessage(`[{"text":"part1"},{"text":"part2"}]`)
	extracted := provider.extractContent(partsContent)
	if extracted != "part1\npart2" {
		t.Errorf("failed to extract parts content, got: %s", extracted)
	}
}

func TestGeminiRoleMapping(t *testing.T) {
	provider := &geminiProvider{}

	tests := []struct {
		author   string
		expected MessageRole
	}{
		{"user", RoleUser},
		{"gemini", RoleAssistant},
		{"system", RoleSystem},
		{"unknown", RoleUser},
	}

	for _, test := range tests {
		if provider.roleFromAuthor(test.author) != test.expected {
			t.Errorf("failed to map role for author '%s'", test.author)
		}
	}
}

// Test provider for mocking

type testProvider struct{}

func (p *testProvider) AgentType() string {
	return "test"
}

func (p *testProvider) ListProjects() ([]string, error) {
	return nil, nil
}

func (p *testProvider) ListSessions(projectPath string) ([]SessionSummary, error) {
	return nil, nil
}

func (p *testProvider) ReadSession(sessionID string) (*Session, error) {
	return nil, nil
}

func (p *testProvider) ScanMessages(sessionID string, fn func(*Message) error) error {
	return nil
}
