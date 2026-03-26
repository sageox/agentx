package sessions

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// claudeCodeProvider implements SessionProvider for Claude Code sessions.
type claudeCodeProvider struct{}

// ccSessionsIndex represents the structure of sessions-index.json.
type ccSessionsIndex struct {
	Sessions map[string]ccIndexEntry `json:"sessions"`
}

// ccIndexEntry represents a single entry in the sessions index.
type ccIndexEntry struct {
	ProjectPath string `json:"projectPath"`
	GitBranch   string `json:"gitBranch"`
	FirstPrompt string `json:"firstPrompt"`
	Summary     string `json:"summary"`
	Created     int64  `json:"created"`
	Modified    int64  `json:"modified"`
}

// ccRawEntry is a raw entry from a session JSONL file.
type ccRawEntry struct {
	Type      string          `json:"type"`
	Role      string          `json:"role,omitempty"`
	Model     string          `json:"model,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
	Timestamp string          `json:"timestamp,omitempty"`
	Usage     *ccUsage        `json:"usage,omitempty"`
	ParentID  string          `json:"parentId,omitempty"`
}

// ccMessage is a parsed message from Claude Code.
type ccMessage struct {
	Type      string
	Role      string
	Model     string
	Content   string
	Timestamp time.Time
	Usage     *TokenUsage
	ParentID  string
}

// ccUsage represents token usage in Claude Code format.
type ccUsage struct {
	Input       int64 `json:"input_tokens"`
	Output      int64 `json:"output_tokens"`
	CacheRead   int64 `json:"cache_read_tokens,omitempty"`
	CacheWrite  int64 `json:"cache_write_tokens,omitempty"`
}

// ccContentBlock represents a content block (for polymorphic content handling).
type ccContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

func (p *claudeCodeProvider) AgentType() string {
	return "claude-code"
}

func (p *claudeCodeProvider) ListProjects() ([]string, error) {
	projectsDir := filepath.Join(os.Getenv("HOME"), ".claude", "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var projects []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if this directory has Claude Code sessions
		indexPath := filepath.Join(projectsDir, entry.Name(), "sessions-index.json")
		if _, err := os.Stat(indexPath); err == nil {
			// Decode the project path from the directory name
			projects = append(projects, decodeProjectPath(entry.Name()))
		}
	}

	sort.Strings(projects)
	return projects, nil
}

func (p *claudeCodeProvider) ListSessions(projectPath string) ([]SessionSummary, error) {
	encodedPath := encodeProjectPath(projectPath)
	indexPath := filepath.Join(os.Getenv("HOME"), ".claude", "projects", encodedPath, "sessions-index.json")

	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	var index ccSessionsIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse sessions index: %w", err)
	}

	var summaries []SessionSummary
	for sessionID, entry := range index.Sessions {
		summaries = append(summaries, SessionSummary{
			ID:           sessionID,
			AgentType:    "claude-code",
			ProjectPath:  entry.ProjectPath,
			GitBranch:    entry.GitBranch,
			FirstPrompt:  entry.FirstPrompt,
			Summary:      entry.Summary,
			Created:      time.Unix(0, entry.Created*1000000),
			Modified:     time.Unix(0, entry.Modified*1000000),
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Modified.After(summaries[j].Modified)
	})

	return summaries, nil
}

func (p *claudeCodeProvider) ReadSession(sessionID string) (*Session, error) {
	// Find the session file
	projectsDir := filepath.Join(os.Getenv("HOME"), ".claude", "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		sessionFile := filepath.Join(projectsDir, entry.Name(), sessionID+".jsonl")
		if _, err := os.Stat(sessionFile); err == nil {
			return p.parseSessionFile(sessionFile, sessionID, entry.Name())
		}
	}

	return nil, ErrSessionNotFound
}

func (p *claudeCodeProvider) ScanMessages(sessionID string, fn func(*Message) error) error {
	projectsDir := filepath.Join(os.Getenv("HOME"), ".claude", "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		sessionFile := filepath.Join(projectsDir, entry.Name(), sessionID+".jsonl")
		if _, err := os.Stat(sessionFile); err == nil {
			file, err := os.Open(sessionFile)
			if err != nil {
				return err
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			messageIndex := 0
			for scanner.Scan() {
				var raw ccRawEntry
				if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
					continue
				}

				// Skip the first entry if it's version info
				if messageIndex == 0 && raw.Type == "version" {
					messageIndex++
					continue
				}

				parsed := p.parseMessage(raw, sessionID)
				if parsed == nil {
					continue
				}

				msg := &Message{
					ID:        fmt.Sprintf("%s-%d", sessionID, messageIndex),
					ParentID:  parsed.ParentID,
					SessionID: sessionID,
					Role:      MessageRole(parsed.Role),
					Type:      p.messageType(raw.Type),
					Content:   parsed.Content,
					Model:     parsed.Model,
					Timestamp: parsed.Timestamp,
					Usage:     parsed.Usage,
					RawJSON:   scanner.Bytes(),
				}

				if err := fn(msg); err != nil {
					return err
				}
				messageIndex++
			}

			return scanner.Err()
		}
	}

	return ErrSessionNotFound
}

func (p *claudeCodeProvider) parseSessionFile(filePath string, sessionID string, encodedProjectPath string) (*Session, error) {
	projectPath := decodeProjectPath(encodedProjectPath)

	// Get summary from index
	indexPath := filepath.Join(filepath.Dir(filePath), "sessions-index.json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}

	var index ccSessionsIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse sessions index: %w", err)
	}

	indexEntry, ok := index.Sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}

	session := &Session{
		SessionSummary: SessionSummary{
			ID:           sessionID,
			AgentType:    "claude-code",
			ProjectPath:  projectPath,
			GitBranch:    indexEntry.GitBranch,
			FirstPrompt:  indexEntry.FirstPrompt,
			Summary:      indexEntry.Summary,
			Created:      time.Unix(0, indexEntry.Created*1000000),
			Modified:     time.Unix(0, indexEntry.Modified*1000000),
		},
	}

	// Parse messages
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	messageIndex := 0
	var version string

	for scanner.Scan() {
		var raw ccRawEntry
		if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			continue
		}

		// Extract version from first entry
		if messageIndex == 0 && raw.Type == "version" {
			// Version info is in the type field or elsewhere
			messageIndex++
			continue
		}

		parsed := p.parseMessage(raw, sessionID)
		if parsed == nil {
			continue
		}

		msg := Message{
			ID:        fmt.Sprintf("%s-%d", sessionID, messageIndex),
			ParentID:  parsed.ParentID,
			SessionID: sessionID,
			Role:      MessageRole(parsed.Role),
			Type:      p.messageType(raw.Type),
			Content:   parsed.Content,
			Model:     parsed.Model,
			Timestamp: parsed.Timestamp,
			Usage:     parsed.Usage,
			RawJSON:   json.RawMessage(scanner.Bytes()),
		}

		session.Messages = append(session.Messages, msg)
		session.MessageCount++
		messageIndex++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if version != "" {
		session.Version = version
	}

	return session, nil
}

func (p *claudeCodeProvider) parseMessage(raw ccRawEntry, sessionID string) *ccMessage {
	if raw.Type == "version" || raw.Type == "" {
		return nil
	}

	// Parse content (can be string or []ContentBlock)
	var content string
	if raw.Content != nil {
		// Try to parse as string first
		if err := json.Unmarshal(raw.Content, &content); err == nil {
			// Filter out thinking blocks
			if !strings.Contains(content, "<thinking>") {
				// It's a string content
			} else {
				// Has thinking, extract just the text
				content = p.extractContentWithoutThinking(content)
			}
		} else {
			// Try to parse as ContentBlock array
			var blocks []ccContentBlock
			if err := json.Unmarshal(raw.Content, &blocks); err == nil {
				content = p.contentBlocksToString(blocks)
			} else {
				// Try single ContentBlock
				var block ccContentBlock
				if err := json.Unmarshal(raw.Content, &block); err == nil {
					content = block.Text
				}
			}
		}
	}

	ts := time.Now()
	if raw.Timestamp != "" {
		if parsed, err := time.Parse(time.RFC3339, raw.Timestamp); err == nil {
			ts = parsed
		}
	}

	var usage *TokenUsage
	if raw.Usage != nil {
		usage = &TokenUsage{
			InputTokens:  raw.Usage.Input,
			OutputTokens: raw.Usage.Output,
			CacheRead:    raw.Usage.CacheRead,
			CacheWrite:   raw.Usage.CacheWrite,
		}
	}

	return &ccMessage{
		Type:      raw.Type,
		Role:      raw.Role,
		Model:     raw.Model,
		Content:   content,
		Timestamp: ts,
		Usage:     usage,
		ParentID:  raw.ParentID,
	}
}

func (p *claudeCodeProvider) extractContentWithoutThinking(content string) string {
	parts := strings.Split(content, "</thinking>")
	if len(parts) > 1 {
		return strings.TrimSpace(parts[1])
	}
	return content
}

func (p *claudeCodeProvider) contentBlocksToString(blocks []ccContentBlock) string {
	var result []string
	for _, block := range blocks {
		if block.Type != "thinking" && block.Text != "" {
			result = append(result, block.Text)
		}
	}
	return strings.Join(result, "\n")
}

func (p *claudeCodeProvider) messageType(rawType string) MessageType {
	switch rawType {
	case "tool-use":
		return TypeToolUse
	case "tool-result":
		return TypeToolResult
	case "progress":
		return TypeProgress
	case "snapshot":
		return TypeSnapshot
	default:
		return TypeChat
	}
}

// encodeProjectPath encodes a project path for use as a directory name.
// This matches Claude Code's directory naming in ~/.claude/projects/.
// WARNING: This encoding is lossy — hyphens in path segments are
// indistinguishable from path separators after encoding.
// Example: "/Users/my-name/code" and "/Users/my/name/code" both
// encode to "-Users-my-name-code".
func encodeProjectPath(projectPath string) string {
	return strings.ReplaceAll(projectPath, "/", "-")
}

// decodeProjectPath decodes a project path from an encoded directory name.
// WARNING: This decoding is lossy for paths containing hyphens.
// All dashes are converted to slashes, so paths with original hyphens
// will be decoded incorrectly. This matches Claude Code's behavior.
func decodeProjectPath(encoded string) string {
	if strings.HasPrefix(encoded, "-") {
		parts := strings.Split(encoded[1:], "-")
		return "/" + strings.Join(parts, "/")
	}
	return encoded
}
