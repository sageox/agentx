package sessions

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// geminiProvider implements SessionProvider for Gemini CLI sessions.
type geminiProvider struct{}

// gemConversationRecord represents a Gemini session file.
type gemConversationRecord struct {
	SessionID   string               `json:"sessionId"`
	ProjectHash string               `json:"projectHash"`
	StartTime   int64                `json:"startTime"`
	LastUpdated int64                `json:"lastUpdated"`
	Messages    []gemMessageRecord   `json:"messages"`
	Summary     string               `json:"summary"`
}

// gemMessageRecord represents a single message in a Gemini session.
type gemMessageRecord struct {
	Author    string          `json:"author"`
	Type      string          `json:"type"`
	Timestamp int64           `json:"timestamp"`
	Content   json.RawMessage `json:"content"`
	ToolCalls []gemToolCall   `json:"toolCalls,omitempty"`
	Tokens    gemTokens       `json:"tokens,omitempty"`
}

// gemToolCall represents a tool call in Gemini.
type gemToolCall struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// gemTokens represents token usage in Gemini format.
type gemTokens struct {
	Input  int64 `json:"input"`
	Output int64 `json:"output"`
	Cached int64 `json:"cached"`
}

// gemPart represents a content part in Gemini.
type gemPart struct {
	Text string `json:"text,omitempty"`
}

func (p *geminiProvider) AgentType() string {
	return "gemini"
}

func (p *geminiProvider) ListProjects() ([]string, error) {
	chatDir := filepath.Join(os.Getenv("HOME"), ".gemini", "tmp")
	entries, err := os.ReadDir(chatDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	projectMap := make(map[string]bool)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		chatsSubdir := filepath.Join(chatDir, entry.Name(), "chats")
		if _, err := os.Stat(chatsSubdir); err == nil {
			// This is a project hash directory
			// We could reverse-map the hash to a project path, but for now
			// we'll use the hash as the project identifier
			projectMap[entry.Name()] = true
		}
	}

	var projects []string
	for project := range projectMap {
		projects = append(projects, project)
	}
	sort.Strings(projects)

	return projects, nil
}

func (p *geminiProvider) ListSessions(projectPath string) ([]SessionSummary, error) {
	chatDir := filepath.Join(os.Getenv("HOME"), ".gemini", "tmp", projectPath, "chats")
	entries, err := os.ReadDir(chatDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	var summaries []SessionSummary
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		sessionFile := filepath.Join(chatDir, entry.Name(), "session.json")
		data, err := os.ReadFile(sessionFile)
		if err != nil {
			continue
		}

		var record gemConversationRecord
		if err := json.Unmarshal(data, &record); err != nil {
			continue
		}

		messageCount := len(record.Messages)
		summary := SessionSummary{
			ID:           record.SessionID,
			AgentType:    "gemini",
			ProjectPath:  projectPath,
			Summary:      record.Summary,
			MessageCount: messageCount,
			Created:      time.UnixMilli(record.StartTime),
			Modified:     time.UnixMilli(record.LastUpdated),
		}

		// Extract first prompt from messages
		if len(record.Messages) > 0 {
			firstContent := p.extractContent(record.Messages[0].Content)
			if len(firstContent) > 200 {
				summary.FirstPrompt = firstContent[:200] + "..."
			} else {
				summary.FirstPrompt = firstContent
			}
		}

		summaries = append(summaries, summary)
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Modified.After(summaries[j].Modified)
	})

	return summaries, nil
}

func (p *geminiProvider) ReadSession(sessionID string) (*Session, error) {
	// Find the session file by searching through all projects
	projects, err := p.ListProjects()
	if err != nil {
		return nil, err
	}

	for _, projectPath := range projects {
		chatDir := filepath.Join(os.Getenv("HOME"), ".gemini", "tmp", projectPath, "chats")
		entries, err := os.ReadDir(chatDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			sessionFile := filepath.Join(chatDir, entry.Name(), "session.json")
			data, err := os.ReadFile(sessionFile)
			if err != nil {
				continue
			}

			var record gemConversationRecord
			if err := json.Unmarshal(data, &record); err != nil {
				continue
			}

			if record.SessionID == sessionID {
				return p.parseSession(&record, projectPath), nil
			}
		}
	}

	return nil, ErrSessionNotFound
}

func (p *geminiProvider) ScanMessages(sessionID string, fn func(*Message) error) error {
	projects, err := p.ListProjects()
	if err != nil {
		return err
	}

	for _, projectPath := range projects {
		chatDir := filepath.Join(os.Getenv("HOME"), ".gemini", "tmp", projectPath, "chats")
		entries, err := os.ReadDir(chatDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			sessionFile := filepath.Join(chatDir, entry.Name(), "session.json")
			data, err := os.ReadFile(sessionFile)
			if err != nil {
				continue
			}

			var record gemConversationRecord
			if err := json.Unmarshal(data, &record); err != nil {
				continue
			}

			if record.SessionID == sessionID {
				for i, msgRecord := range record.Messages {
					msg := &Message{
						ID:        fmt.Sprintf("%s-%d", sessionID, i),
						SessionID: sessionID,
						Role:      p.roleFromAuthor(msgRecord.Author),
						Type:      p.messageType(msgRecord.ToolCalls),
						Content:   p.extractContent(msgRecord.Content),
						Timestamp: time.UnixMilli(msgRecord.Timestamp),
						RawJSON:   msgRecord.Content,
					}

					// Add token usage if available
					if msgRecord.Tokens.Input > 0 || msgRecord.Tokens.Output > 0 {
						msg.Usage = &TokenUsage{
							InputTokens:  msgRecord.Tokens.Input,
							OutputTokens: msgRecord.Tokens.Output,
							CacheRead:    msgRecord.Tokens.Cached,
						}
					}

					if err := fn(msg); err != nil {
						return err
					}
				}
				return nil
			}
		}
	}

	return ErrSessionNotFound
}

func (p *geminiProvider) parseSession(record *gemConversationRecord, projectPath string) *Session {
	session := &Session{
		SessionSummary: SessionSummary{
			ID:           record.SessionID,
			AgentType:    "gemini",
			ProjectPath:  projectPath,
			Summary:      record.Summary,
			MessageCount: len(record.Messages),
			Created:      time.UnixMilli(record.StartTime),
			Modified:     time.UnixMilli(record.LastUpdated),
		},
	}

	for i, msgRecord := range record.Messages {
		msg := Message{
			ID:        fmt.Sprintf("%s-%d", record.SessionID, i),
			SessionID: record.SessionID,
			Role:      p.roleFromAuthor(msgRecord.Author),
			Type:      p.messageType(msgRecord.ToolCalls),
			Content:   p.extractContent(msgRecord.Content),
			Timestamp: time.UnixMilli(msgRecord.Timestamp),
			RawJSON:   msgRecord.Content,
		}

		// Add token usage if available
		if msgRecord.Tokens.Input > 0 || msgRecord.Tokens.Output > 0 {
			msg.Usage = &TokenUsage{
				InputTokens:  msgRecord.Tokens.Input,
				OutputTokens: msgRecord.Tokens.Output,
				CacheRead:    msgRecord.Tokens.Cached,
			}
		}

		session.Messages = append(session.Messages, msg)
	}

	// Extract first prompt
	if len(record.Messages) > 0 {
		firstContent := p.extractContent(record.Messages[0].Content)
		if len(firstContent) > 200 {
			session.FirstPrompt = firstContent[:200] + "..."
		} else {
			session.FirstPrompt = firstContent
		}
	}

	return session
}

func (p *geminiProvider) extractContent(raw json.RawMessage) string {
	// Try to parse as string first
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return str
	}

	// Try to parse as Part object
	var part gemPart
	if err := json.Unmarshal(raw, &part); err == nil {
		return part.Text
	}

	// Try to parse as array of Parts
	var parts []gemPart
	if err := json.Unmarshal(raw, &parts); err == nil {
		var result []string
		for _, p := range parts {
			if p.Text != "" {
				result = append(result, p.Text)
			}
		}
		return strings.Join(result, "\n")
	}

	return ""
}

func (p *geminiProvider) roleFromAuthor(author string) MessageRole {
	switch strings.ToLower(author) {
	case "user":
		return RoleUser
	case "gemini":
		return RoleAssistant
	case "system":
		return RoleSystem
	default:
		return RoleUser
	}
}

func (p *geminiProvider) messageType(toolCalls []gemToolCall) MessageType {
	if len(toolCalls) > 0 {
		return TypeToolUse
	}
	return TypeChat
}

// hashProjectPath computes the hash of a project path for directory naming.
func hashProjectPath(projectPath string) string {
	hash := md5.Sum([]byte(projectPath))
	return fmt.Sprintf("%x", hash)
}
