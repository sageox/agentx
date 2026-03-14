package agentx_test

import (
	"testing"

	"github.com/sageox/agentx"
)

func TestDetectOrigin(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		explicit agentx.SessionOrigin
		want     agentx.SessionOrigin
	}{
		{
			name: "default is human",
			want: agentx.OriginHuman,
		},
		{
			name:    "claude code task entry point means subagent",
			envVars: map[string]string{"CLAUDE_CODE_ENTRY_POINT": "task"},
			want:    agentx.OriginSubagent,
		},
		{
			name:     "subagent env takes precedence over explicit agent",
			envVars:  map[string]string{"CLAUDE_CODE_ENTRY_POINT": "task"},
			explicit: agentx.OriginAgent,
			want:     agentx.OriginSubagent,
		},
		{
			name:     "explicit agent origin",
			explicit: agentx.OriginAgent,
			want:     agentx.OriginAgent,
		},
		{
			name:     "explicit human origin",
			explicit: agentx.OriginHuman,
			want:     agentx.OriginHuman,
		},
		{
			name:    "CI=true means agent",
			envVars: map[string]string{"CI": "true"},
			want:    agentx.OriginAgent,
		},
		{
			name:    "GITHUB_ACTIONS means agent",
			envVars: map[string]string{"GITHUB_ACTIONS": "true"},
			want:    agentx.OriginAgent,
		},
		{
			name:    "GITLAB_CI means agent",
			envVars: map[string]string{"GITLAB_CI": "true"},
			want:    agentx.OriginAgent,
		},
		{
			name:    "JENKINS_URL means agent",
			envVars: map[string]string{"JENKINS_URL": "https://ci.example.com"},
			want:    agentx.OriginAgent,
		},
		{
			name:    "BUILDKITE means agent",
			envVars: map[string]string{"BUILDKITE": "true"},
			want:    agentx.OriginAgent,
		},
		{
			name:     "explicit overrides CI detection",
			envVars:  map[string]string{"CI": "true"},
			explicit: agentx.OriginHuman,
			want:     agentx.OriginHuman,
		},
		{
			name:    "non-matching entry point value is not subagent",
			envVars: map[string]string{"CLAUDE_CODE_ENTRY_POINT": "cli"},
			want:    agentx.OriginHuman,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := agentx.NewMockEnvironment(tt.envVars)
			got := agentx.DetectOrigin(env, tt.explicit)
			if got != tt.want {
				t.Errorf("DetectOrigin() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSessionOrigin_IsValid(t *testing.T) {
	tests := []struct {
		origin agentx.SessionOrigin
		want   bool
	}{
		{agentx.OriginHuman, true},
		{agentx.OriginSubagent, true},
		{agentx.OriginAgent, true},
		{agentx.SessionOrigin("invalid"), false},
		{agentx.SessionOrigin(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.origin), func(t *testing.T) {
			if got := tt.origin.IsValid(); got != tt.want {
				t.Errorf("SessionOrigin(%q).IsValid() = %v, want %v", tt.origin, got, tt.want)
			}
		})
	}
}

func TestValidOrigins(t *testing.T) {
	origins := agentx.ValidOrigins()
	if len(origins) != 3 {
		t.Fatalf("ValidOrigins() returned %d values, want 3", len(origins))
	}
	for _, o := range origins {
		if !o.IsValid() {
			t.Errorf("ValidOrigins() contains invalid origin %q", o)
		}
	}
}
