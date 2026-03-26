package agentx

import (
	"context"
	"errors"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// MockEnvironment - GetEnv
// ---------------------------------------------------------------------------

func TestMockEnvironment_GetEnv_Present(t *testing.T) {
	env := NewMockEnvironment(map[string]string{"FOO": "bar"})
	assert.Equal(t, "bar", env.GetEnv("FOO"))
}

func TestMockEnvironment_GetEnv_Missing(t *testing.T) {
	env := NewMockEnvironment(map[string]string{"FOO": "bar"})
	assert.Equal(t, "", env.GetEnv("MISSING"))
}

func TestMockEnvironment_GetEnv_NilMap(t *testing.T) {
	env := &MockEnvironment{}
	assert.Equal(t, "", env.GetEnv("ANY"))
}

// ---------------------------------------------------------------------------
// MockEnvironment - LookupEnv
// ---------------------------------------------------------------------------

func TestMockEnvironment_LookupEnv_Present(t *testing.T) {
	env := NewMockEnvironment(map[string]string{"KEY": "val"})
	val, ok := env.LookupEnv("KEY")
	assert.True(t, ok)
	assert.Equal(t, "val", val)
}

func TestMockEnvironment_LookupEnv_Missing(t *testing.T) {
	env := NewMockEnvironment(map[string]string{"KEY": "val"})
	_, ok := env.LookupEnv("NOPE")
	assert.False(t, ok)
}

func TestMockEnvironment_LookupEnv_NilMap(t *testing.T) {
	env := &MockEnvironment{}
	_, ok := env.LookupEnv("ANY")
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// MockEnvironment - HomeDir
// ---------------------------------------------------------------------------

func TestMockEnvironment_HomeDir(t *testing.T) {
	env := NewMockEnvironment(nil)
	home, err := env.HomeDir()
	require.NoError(t, err)
	assert.Equal(t, "/home/test", home)
}

func TestMockEnvironment_HomeDir_Error(t *testing.T) {
	env := &MockEnvironment{HomeError: errors.New("no home")}
	_, err := env.HomeDir()
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// MockEnvironment - ConfigDir, DataDir, CacheDir
// ---------------------------------------------------------------------------

func TestMockEnvironment_ConfigDir(t *testing.T) {
	env := NewMockEnvironment(nil)
	dir, err := env.ConfigDir()
	require.NoError(t, err)
	assert.Equal(t, "/home/test/.config", dir)
}

func TestMockEnvironment_DataDir(t *testing.T) {
	env := NewMockEnvironment(nil)
	dir, err := env.DataDir()
	require.NoError(t, err)
	assert.Equal(t, "/home/test/.local/share", dir)
}

func TestMockEnvironment_CacheDir(t *testing.T) {
	env := NewMockEnvironment(nil)
	dir, err := env.CacheDir()
	require.NoError(t, err)
	assert.Equal(t, "/home/test/.cache", dir)
}

// ---------------------------------------------------------------------------
// MockEnvironment - GOOS
// ---------------------------------------------------------------------------

func TestMockEnvironment_GOOS_Set(t *testing.T) {
	env := &MockEnvironment{OS: "darwin"}
	assert.Equal(t, "darwin", env.GOOS())
}

func TestMockEnvironment_GOOS_Default(t *testing.T) {
	env := &MockEnvironment{}
	assert.Equal(t, "linux", env.GOOS())
}

// ---------------------------------------------------------------------------
// MockEnvironment - LookPath
// ---------------------------------------------------------------------------

func TestMockEnvironment_LookPath_Found(t *testing.T) {
	env := &MockEnvironment{
		PathBinaries: map[string]string{"git": "/usr/bin/git"},
	}
	path, err := env.LookPath("git")
	require.NoError(t, err)
	assert.Equal(t, "/usr/bin/git", path)
}

func TestMockEnvironment_LookPath_NotFound(t *testing.T) {
	env := &MockEnvironment{
		PathBinaries: map[string]string{"git": "/usr/bin/git"},
	}
	_, err := env.LookPath("nonexistent")
	assert.ErrorIs(t, err, exec.ErrNotFound)
}

func TestMockEnvironment_LookPath_NilMap(t *testing.T) {
	env := &MockEnvironment{}
	_, err := env.LookPath("anything")
	assert.ErrorIs(t, err, exec.ErrNotFound)
}

// ---------------------------------------------------------------------------
// MockEnvironment - FileExists
// ---------------------------------------------------------------------------

func TestMockEnvironment_FileExists_InExistingFiles(t *testing.T) {
	env := &MockEnvironment{
		ExistingFiles: map[string]bool{"/tmp/file.txt": true},
	}
	assert.True(t, env.FileExists("/tmp/file.txt"))
}

func TestMockEnvironment_FileExists_InFiles(t *testing.T) {
	env := &MockEnvironment{
		Files: map[string][]byte{"/tmp/data.json": []byte("{}")},
	}
	assert.True(t, env.FileExists("/tmp/data.json"))
}

func TestMockEnvironment_FileExists_InExistingDirs(t *testing.T) {
	env := &MockEnvironment{
		ExistingDirs: map[string]bool{"/tmp/mydir": true},
	}
	assert.True(t, env.FileExists("/tmp/mydir"))
}

func TestMockEnvironment_FileExists_NotFound(t *testing.T) {
	env := &MockEnvironment{
		ExistingFiles: map[string]bool{"/tmp/other": true},
	}
	assert.False(t, env.FileExists("/tmp/missing"))
}

// ---------------------------------------------------------------------------
// MockEnvironment - IsDir
// ---------------------------------------------------------------------------

func TestMockEnvironment_IsDir_Found(t *testing.T) {
	env := &MockEnvironment{
		ExistingDirs: map[string]bool{"/tmp/dir": true},
	}
	assert.True(t, env.IsDir("/tmp/dir"))
}

func TestMockEnvironment_IsDir_NotFound(t *testing.T) {
	env := &MockEnvironment{
		ExistingDirs: map[string]bool{"/tmp/dir": true},
	}
	assert.False(t, env.IsDir("/tmp/other"))
}

func TestMockEnvironment_IsDir_NilMap(t *testing.T) {
	env := &MockEnvironment{}
	assert.False(t, env.IsDir("/any"))
}

// ---------------------------------------------------------------------------
// MockEnvironment - Exec
// ---------------------------------------------------------------------------

func TestMockEnvironment_Exec_HasOutput(t *testing.T) {
	env := &MockEnvironment{
		ExecOutputs: map[string][]byte{"git": []byte("v2.40.0\n")},
	}
	out, err := env.Exec(context.Background(), "git", "--version")
	require.NoError(t, err)
	assert.Equal(t, []byte("v2.40.0\n"), out)
}

func TestMockEnvironment_Exec_HasError(t *testing.T) {
	env := &MockEnvironment{
		ExecErrors: map[string]error{"bad": errors.New("command failed")},
	}
	_, err := env.Exec(context.Background(), "bad")
	assert.Error(t, err)
	assert.Equal(t, "command failed", err.Error())
}

func TestMockEnvironment_Exec_NotFound(t *testing.T) {
	env := &MockEnvironment{}
	_, err := env.Exec(context.Background(), "missing")
	assert.ErrorIs(t, err, exec.ErrNotFound)
}

func TestMockEnvironment_Exec_ErrorTakesPrecedence(t *testing.T) {
	// when both error and output are set for same command, error wins
	env := &MockEnvironment{
		ExecOutputs: map[string][]byte{"cmd": []byte("output")},
		ExecErrors:  map[string]error{"cmd": errors.New("fail")},
	}
	_, err := env.Exec(context.Background(), "cmd")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// MockEnvironment - ReadFile
// ---------------------------------------------------------------------------

func TestMockEnvironment_ReadFile_Exists(t *testing.T) {
	env := &MockEnvironment{
		Files: map[string][]byte{"/tmp/file.txt": []byte("hello")},
	}
	data, err := env.ReadFile("/tmp/file.txt")
	require.NoError(t, err)
	assert.Equal(t, []byte("hello"), data)
}

func TestMockEnvironment_ReadFile_NotFound(t *testing.T) {
	env := &MockEnvironment{
		Files: map[string][]byte{"/tmp/other": []byte("data")},
	}
	_, err := env.ReadFile("/tmp/missing")
	assert.Error(t, err)
}

func TestMockEnvironment_ReadFile_NilMap(t *testing.T) {
	env := &MockEnvironment{}
	_, err := env.ReadFile("/any")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// SystemEnvironment - minimal smoke tests
// ---------------------------------------------------------------------------

func TestSystemEnvironment_NonNil(t *testing.T) {
	env := NewSystemEnvironment()
	assert.NotNil(t, env)
}

func TestSystemEnvironment_GetEnv_Home(t *testing.T) {
	env := NewSystemEnvironment()
	// HOME (or USERPROFILE on Windows) should be set on any test runner
	home := env.GetEnv("HOME")
	if home == "" {
		home = env.GetEnv("USERPROFILE")
	}
	assert.NotEmpty(t, home, "expected HOME or USERPROFILE to be set")
}

func TestSystemEnvironment_GOOS_NonEmpty(t *testing.T) {
	env := NewSystemEnvironment()
	assert.NotEmpty(t, env.GOOS())
}

func TestSystemEnvironment_HomeDir(t *testing.T) {
	env := NewSystemEnvironment()
	home, err := env.HomeDir()
	require.NoError(t, err)
	assert.NotEmpty(t, home)
}

func TestSystemEnvironment_LookupEnv(t *testing.T) {
	env := NewSystemEnvironment()
	// PATH is always set
	val, ok := env.LookupEnv("PATH")
	assert.True(t, ok)
	assert.NotEmpty(t, val)
}

func TestSystemEnvironment_FileExists(t *testing.T) {
	env := NewSystemEnvironment()
	// current directory always exists
	assert.True(t, env.FileExists("."))
	assert.False(t, env.FileExists("/nonexistent_path_abc123xyz"))
}

func TestSystemEnvironment_IsDir(t *testing.T) {
	env := NewSystemEnvironment()
	assert.True(t, env.IsDir("."))
	assert.False(t, env.IsDir("/nonexistent_path_abc123xyz"))
}

// ---------------------------------------------------------------------------
// NewMockEnvironment defaults
// ---------------------------------------------------------------------------

func TestNewMockEnvironment_Defaults(t *testing.T) {
	env := NewMockEnvironment(nil)
	assert.Equal(t, "/home/test", env.Home)
	assert.Equal(t, "/home/test/.config", env.Config)
	assert.Equal(t, "/home/test/.local/share", env.Data)
	assert.Equal(t, "/home/test/.cache", env.Cache)
	assert.Equal(t, "linux", env.OS)
}
