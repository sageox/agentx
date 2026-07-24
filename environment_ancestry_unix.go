//go:build !windows

package agentx

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// processAncestry returns the executable base names of ancestor processes,
// starting from the current process's parent and walking toward PID 1.
//
// It takes a single `ps` snapshot (pid, ppid, comm) and walks the parent map in
// memory rather than exec-ing once per level, and it does not read /proc — so it
// works on macOS and the BSDs as well as Linux. This mirrors the approach in
// ox's internal/proc and is deliberately dependency-free (no golang.org/x/sys).
//
// Best-effort by contract: any failure returns (nil, err) and the caller falls
// back to env-var signals.
func processAncestry() ([]string, error) {
	out, err := exec.Command("ps", "-A", "-o", "pid=,ppid=,comm=").Output()
	if err != nil {
		return nil, err
	}
	return walkAncestry(string(out), os.Getppid()), nil
}

// walkAncestry parses `ps -A -o pid=,ppid=,comm=` output into a pid → parent map
// and returns the executable base names from startPPID up toward PID 1, capped at
// 20 levels. Extracted from processAncestry as a pure, deterministic function so
// the parsing and tree-walking logic is unit-testable without a real `ps`.
func walkAncestry(psOutput string, startPPID int) []string {
	type parent struct {
		ppid int
		name string
	}
	procs := make(map[int]parent)
	for _, line := range strings.Split(psOutput, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		ppid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		// comm can contain spaces (a full path on macOS); rejoin the remaining
		// fields, then reduce to the base name.
		name := strings.Join(fields[2:], " ")
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			name = name[idx+1:]
		}
		procs[pid] = parent{ppid: ppid, name: name}
	}

	var chain []string
	pid := startPPID
	for range 20 {
		if pid <= 1 {
			break
		}
		p, ok := procs[pid]
		if !ok {
			break
		}
		chain = append(chain, p.name)
		if p.ppid == pid { // defensive: never loop on a self-parent
			break
		}
		pid = p.ppid
	}
	return chain
}
