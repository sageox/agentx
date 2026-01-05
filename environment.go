package agentx

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Environment provides access to system environment for agent detection.
// This abstraction enables testing without real file system access.
type Environment interface {
	// GetEnv retrieves an environment variable value.
	GetEnv(key string) string

	// LookupEnv retrieves an environment variable and reports if it exists.
	LookupEnv(key string) (string, bool)

	// HomeDir returns the user's home directory.
	HomeDir() (string, error)

	// ConfigDir returns the XDG config directory.
	ConfigDir() (string, error)

	// GOOS returns the operating system name.
	GOOS() string

	// LookPath searches for an executable in PATH.
	LookPath(name string) (string, error)

	// FileExists checks if a file or directory exists.
	FileExists(path string) bool

	// IsDir checks if a path is a directory.
	IsDir(path string) bool
}

// SystemEnvironment implements Environment using the real system.
type SystemEnvironment struct{}

// NewSystemEnvironment creates a new system environment.
func NewSystemEnvironment() Environment {
	return &SystemEnvironment{}
}

func (e *SystemEnvironment) GetEnv(key string) string {
	return os.Getenv(key)
}

func (e *SystemEnvironment) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

func (e *SystemEnvironment) HomeDir() (string, error) {
	return os.UserHomeDir()
}

func (e *SystemEnvironment) ConfigDir() (string, error) {
	// XDG_CONFIG_HOME or default
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		return configHome, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "darwin":
		// macOS: prefer ~/.config for CLI tools (XDG-style)
		return filepath.Join(home, ".config"), nil
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return appData, nil
		}
		return filepath.Join(home, "AppData", "Roaming"), nil
	default:
		return filepath.Join(home, ".config"), nil
	}
}

func (e *SystemEnvironment) GOOS() string {
	return runtime.GOOS
}

func (e *SystemEnvironment) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

func (e *SystemEnvironment) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (e *SystemEnvironment) IsDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// MockEnvironment is a test implementation of Environment.
type MockEnvironment struct {
	EnvVars       map[string]string
	Home          string
	Config        string
	OS            string
	HomeError     error
	ExistingPaths map[string]bool // paths that exist
	ExistingDirs  map[string]bool // paths that are directories
	PathBinaries  map[string]string // binaries in PATH
}

// NewMockEnvironment creates a mock environment for testing.
func NewMockEnvironment(envVars map[string]string) *MockEnvironment {
	return &MockEnvironment{
		EnvVars:       envVars,
		Home:          "/home/test",
		Config:        "/home/test/.config",
		OS:            "linux",
		ExistingPaths: make(map[string]bool),
		ExistingDirs:  make(map[string]bool),
		PathBinaries:  make(map[string]string),
	}
}

func (e *MockEnvironment) GetEnv(key string) string {
	if e.EnvVars == nil {
		return ""
	}
	return e.EnvVars[key]
}

func (e *MockEnvironment) LookupEnv(key string) (string, bool) {
	if e.EnvVars == nil {
		return "", false
	}
	val, ok := e.EnvVars[key]
	return val, ok
}

func (e *MockEnvironment) HomeDir() (string, error) {
	if e.HomeError != nil {
		return "", e.HomeError
	}
	return e.Home, nil
}

func (e *MockEnvironment) ConfigDir() (string, error) {
	return e.Config, nil
}

func (e *MockEnvironment) GOOS() string {
	if e.OS == "" {
		return "linux"
	}
	return e.OS
}

func (e *MockEnvironment) LookPath(name string) (string, error) {
	if e.PathBinaries != nil {
		if path, ok := e.PathBinaries[name]; ok {
			return path, nil
		}
	}
	return "", exec.ErrNotFound
}

func (e *MockEnvironment) FileExists(path string) bool {
	if e.ExistingPaths != nil {
		return e.ExistingPaths[path]
	}
	return false
}

func (e *MockEnvironment) IsDir(path string) bool {
	if e.ExistingDirs != nil {
		return e.ExistingDirs[path]
	}
	return false
}
