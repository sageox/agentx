//go:build windows

package agentx

// processAncestry is a stub on Windows. Walking the process tree there needs the
// Toolhelp32 snapshot API (golang.org/x/sys/windows), and agentx is intentionally
// dependency-free; ox's own internal/proc stubs Windows ancestry for the same
// reason. Buzz's buzz-acp harness is a macOS/Linux dev tool today, so on Windows
// BuzzAgent.Detect() cleanly falls back to its ORCHESTRATOR_ENV /
// BUZZ_ACP_AGENT_COMMAND env-var signals. Implement via Toolhelp32 if/when
// Buzz-on-Windows detection is needed.
func processAncestry() ([]string, error) {
	return nil, nil
}
