package sessions

import (
	"encoding/json"
	"time"
)

// MessageRole represents the role of a message sender.
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleSystem    MessageRole = "system"
	RoleTool      MessageRole = "tool"
)

// MessageType represents the type of message.
type MessageType string

const (
	TypeChat       MessageType = "chat"
	TypeToolUse    MessageType = "tool-use"
	TypeToolResult MessageType = "tool-result"
	TypeProgress   MessageType = "progress"
	TypeSnapshot   MessageType = "snapshot"
)

// TokenUsage represents token consumption metrics.
type TokenUsage struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
	CacheRead    int64 `json:"cache_read,omitempty"`
	CacheWrite   int64 `json:"cache_write,omitempty"`
}

// Message represents a single message in a session.
type Message struct {
	ID        string          `json:"id"`
	ParentID  string          `json:"parent_id,omitempty"`
	SessionID string          `json:"session_id"`
	Role      MessageRole     `json:"role"`
	Type      MessageType     `json:"type"`
	Content   string          `json:"content"`
	Model     string          `json:"model,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	Usage     *TokenUsage     `json:"usage,omitempty"`
	RawJSON   json.RawMessage `json:"raw_json,omitempty"`
}

// SessionSummary contains metadata about a session.
type SessionSummary struct {
	ID           string    `json:"id"`
	AgentType    string    `json:"agent_type"`
	ProjectPath  string    `json:"project_path"`
	GitBranch    string    `json:"git_branch,omitempty"`
	FirstPrompt  string    `json:"first_prompt,omitempty"`
	Summary      string    `json:"summary,omitempty"`
	MessageCount int       `json:"message_count"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
}

// Session represents a complete conversation session.
type Session struct {
	SessionSummary
	Messages []Message `json:"messages"`
	Version  string    `json:"version,omitempty"`
}

// SessionProvider is the interface that must be implemented by session data providers.
type SessionProvider interface {
	// AgentType returns the agent type this provider handles.
	AgentType() string

	// ListProjects returns a list of project paths that have sessions for this agent.
	ListProjects() ([]string, error)

	// ListSessions returns session metadata for a given project path.
	ListSessions(projectPath string) ([]SessionSummary, error)

	// ReadSession returns a complete session with all messages.
	ReadSession(sessionID string) (*Session, error)

	// ScanMessages iterates through messages in a session, calling fn for each.
	// Useful for streaming large sessions without loading all into memory.
	ScanMessages(sessionID string, fn func(*Message) error) error
}
