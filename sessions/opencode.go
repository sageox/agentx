package sessions

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
)

// opencodeProvider is a stub implementation for OpenCode.
// OpenCode uses a SQLite database, so consumers with SQLite access
// can use DatabasePath() to access the database directly.
type opencodeProvider struct{}

func (p *opencodeProvider) AgentType() string {
	return "opencode"
}

func (p *opencodeProvider) ListProjects() ([]string, error) {
	// OpenCode uses SQLite, which we don't parse directly
	// Consumers with SQLite access should use DatabasePath()
	return nil, errors.New("opencode uses SQLite; use DatabasePath() to access the database directly")
}

func (p *opencodeProvider) ListSessions(projectPath string) ([]SessionSummary, error) {
	return nil, errors.New("opencode uses SQLite; use DatabasePath() to access the database directly")
}

func (p *opencodeProvider) ReadSession(sessionID string) (*Session, error) {
	return nil, errors.New("opencode uses SQLite; use DatabasePath() to access the database directly")
}

func (p *opencodeProvider) ScanMessages(sessionID string, fn func(*Message) error) error {
	return errors.New("opencode uses SQLite; use DatabasePath() to access the database directly")
}

// DatabasePath returns the path to the OpenCode SQLite database if it exists.
// Returns empty string if the database is not found.
func DatabasePath() string {
	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	var dbPath string

	if xdgDataHome != "" {
		dbPath = filepath.Join(xdgDataHome, "opencode", "opencode.db")
	} else {
		usr, err := user.Current()
		if err != nil {
			return ""
		}
		dbPath = filepath.Join(usr.HomeDir, ".local", "share", "opencode", "opencode.db")
	}

	if _, err := os.Stat(dbPath); err == nil {
		return dbPath
	}

	return ""
}

func init() {
	Register("opencode", &opencodeProvider{})
}
