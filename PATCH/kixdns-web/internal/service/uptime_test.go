package service

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

// TestProcUptimeSelf checks the /proc/<pid>/stat field arithmetic against the
// current process, whose start time is known to be a few seconds ago.
func TestProcUptimeSelf(t *testing.T) {
	up := procUptimeSeconds(os.Getpid())
	if up < 0 {
		t.Fatalf("negative uptime: %d", up)
	}
	t.Logf("self uptime = %ds", up)

	sys, err := systemUptimeSeconds()
	if err != nil {
		t.Fatalf("systemUptimeSeconds: %v", err)
	}
	t.Logf("system uptime = %ds", sys)
	if up > sys {
		t.Errorf("process uptime (%d) exceeds system uptime (%d)", up, sys)
	}

	// A freshly started test binary cannot be older than a few minutes.
	if up > 600 {
		t.Errorf("self uptime %ds is implausible for a just-started binary", up)
	}
}

// TestProcUptimeFieldIndex pins the field arithmetic: after the last ')',
// field 3 is state, so starttime (field 22) sits at index 19.
func TestProcUptimeFieldIndex(t *testing.T) {
	b, err := os.ReadFile("/proc/self/stat")
	if err != nil {
		t.Skip("no /proc/self/stat")
	}
	s := string(b)
	open := strings.IndexByte(s, '(')
	closeIdx := strings.LastIndexByte(s, ')')
	if open < 0 || closeIdx <= open {
		t.Fatalf("could not locate comm field in %q", s)
	}
	tail := s[closeIdx+1:]
	fields := strings.Fields(tail)
	t.Logf("comm=%q", s[open+1:closeIdx])
	t.Logf("field3(state)=%q  field%d=%q", fields[0], 3+19, fields[19])

	if fields[0] != "R" && fields[0] != "S" && fields[0] != "D" {
		t.Errorf("index 0 after ')' should be the state character, got %q", fields[0])
	}
	if _, err := strconv.ParseInt(fields[19], 10, 64); err != nil {
		t.Errorf("index 19 should be numeric starttime, got %q", fields[19])
	}
}

// TestProcUptimeBadPID must not panic on a missing or bogus pid.
func TestProcUptimeBadPID(t *testing.T) {
	for _, pid := range []int{0, -1, 999999} {
		if got := procUptimeSeconds(pid); got != 0 {
			t.Errorf("procUptimeSeconds(%d) = %d, want 0", pid, got)
		}
	}
}
