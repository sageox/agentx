//go:build !windows

package agentx

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWalkAncestry(t *testing.T) {
	// A realistic `ps -A -o pid=,ppid=,comm=` snapshot: an ox process launched
	// under Claude, which was spawned by buzz-acp, under a login shell.
	snapshot := strings.Join([]string{
		"    1     0 launchd",
		"  200     1 zsh",
		"  300   200 buzz-acp",
		"  400   300 claude-agent-acp",
		"  500   400 /opt/homebrew/bin/ox", // full path -> base name
		"  600   500 ps",
	}, "\n")

	tests := []struct {
		name      string
		psOutput  string
		startPPID int
		want      []string
		why       string
	}{
		{
			name:      "full chain, PID 1 excluded, path base-named",
			psOutput:  snapshot,
			startPPID: 500,
			want:      []string{"ox", "claude-agent-acp", "buzz-acp", "zsh"},
			why:       "walk climbs toward init but stops before PID 1 (never an orchestrator); a full path is reduced to its base name",
		},
		{
			name:      "start mid-chain",
			psOutput:  snapshot,
			startPPID: 300,
			want:      []string{"buzz-acp", "zsh"},
			why:       "walk begins at startPPID, not the leaf, and still excludes PID 1",
		},
		{
			name:      "unknown start pid yields empty",
			psOutput:  snapshot,
			startPPID: 9999,
			want:      nil,
			why:       "a pid absent from the snapshot must not fabricate a chain",
		},
		{
			name:      "empty output",
			psOutput:  "",
			startPPID: 500,
			want:      nil,
			why:       "no ps data must yield an empty chain, not a panic",
		},
		{
			name: "malformed lines are skipped",
			psOutput: strings.Join([]string{
				"garbage header line",
				"  abc  200 notanumber-pid", // non-numeric pid -> skip
				"  300  xyz bad-ppid",       // non-numeric ppid -> skip
				"  200    1 zsh",
				"  300  200 buzz-acp",
			}, "\n"),
			startPPID: 300,
			want:      []string{"buzz-acp", "zsh"},
			why:       "unparseable rows must be ignored, valid rows still walked",
		},
		{
			name: "comm containing spaces is preserved then base-named",
			psOutput: strings.Join([]string{
				"  200    1 zsh",
				"  300  200 /Applications/My App.app/Contents/MacOS/My App",
			}, "\n"),
			startPPID: 300,
			want:      []string{"My App", "zsh"},
			why:       "comm may contain spaces; base name is everything after the last slash",
		},
		{
			name: "self-parent does not infinite-loop",
			psOutput: strings.Join([]string{
				"  300  300 stuck", // ppid == pid
			}, "\n"),
			startPPID: 300,
			want:      []string{"stuck"},
			why:       "a self-referential parent must terminate the walk after one hop",
		},
		{
			name:      "walk is capped at 20 levels",
			psOutput:  deepChain(30),
			startPPID: 30,
			want:      cappedNames(30, 20),
			why:       "an unexpectedly deep tree must not walk unbounded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := walkAncestry(tt.psOutput, tt.startPPID)
			assert.Equal(t, tt.want, got, tt.why)
		})
	}
}

// deepChain builds a linear ps snapshot of n processes: pid k has ppid k-1,
// name "p<k>", with p1's parent being PID 1's sentinel (0).
func deepChain(n int) string {
	var b strings.Builder
	for k := 1; k <= n; k++ {
		fmt.Fprintf(&b, "%d %d p%d\n", k, k-1, k)
	}
	return b.String()
}

// cappedNames returns the names the walk should collect starting at `start` and
// climbing (start, start-1, ...) for at most `cap` levels, stopping at pid<=1.
func cappedNames(start, cap int) []string {
	var names []string
	pid := start
	for range cap {
		if pid <= 1 {
			break
		}
		names = append(names, fmt.Sprintf("p%d", pid))
		pid--
	}
	return names
}
