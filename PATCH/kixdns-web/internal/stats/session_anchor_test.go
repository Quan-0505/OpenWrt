package stats

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestSessionAnchorsToProcessStart pins the requirement that the session figure
// agrees with the uptime badge.
//
// Both must derive from the process's own start time. Stamping the session with
// time.Now() when the collector happened to notice the process made it lag by
// however long detection took — measured 113 seconds on the device, so the row
// claimed a shorter run than the badge.
func TestSessionAnchorsToProcessStart(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "kixdns.log"), []byte(sessLine), 0o644); err != nil {
		t.Fatal(err)
	}
	procDir := filepath.Join(dir, "proc")
	if err := os.Mkdir(procDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Pretend the process started 600 seconds ago. procStart() reports the
	// /proc mtime, and procStartUnix reads that as epoch nanoseconds, so setting
	// the mtime to a real past instant is all that is needed.
	const agoSec = 600
	startWall := time.Now().Add(-agoSec * time.Second)
	if err := os.Chtimes(procDir, startWall, startWall); err != nil {
		t.Fatal(err)
	}

	c := NewCollector(dir, filepath.Join(dir, "s.json"))
	c.ProcDir = procDir
	if err := c.Refresh(); err != nil {
		t.Fatal(err)
	}

	got := c.Snapshot().SessionSeconds
	t.Logf("simulated process start %ds ago; session_s = %d", agoSec, got)
	if got < agoSec-5 || got > agoSec+5 {
		t.Errorf("session_s = %d, want ~%d: the session must count from the process start, not from when the collector looked", got, agoSec)
	}
}

// TestProcStartUnixZero guards the degenerate inputs.
func TestProcStartUnixZero(t *testing.T) {
	if got := procStartUnix(0); got != 0 {
		t.Errorf("procStartUnix(0) = %d, want 0", got)
	}
	if got := procStartUnix(-1); got != 0 {
		t.Errorf("procStartUnix(-1) = %d, want 0", got)
	}
}
