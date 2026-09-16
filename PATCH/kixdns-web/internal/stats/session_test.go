package stats

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const sessLine = `2026-09-16T09:00:00.000000000Z DEBUG request started event="request_started" request_id=1 client=10.0.0.1:1 qname="a.example" qtype=A qclass=IN` + "\n"

func fileSize(t *testing.T, path string) int64 {
	t.Helper()
	fi, err := os.Stat(path)
	if err != nil {
		return -1
	}
	return fi.Size()
}

// TestSessionBoundaryResetsOnRestart pins the per-session behaviour: when the
// monitored process is replaced, the counters must restart from zero rather than
// keep including lines that belong to the previous run.
func TestSessionBoundaryResetsOnRestart(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "kixdns.log")
	procDir := filepath.Join(dir, "proc")
	if err := os.Mkdir(procDir, 0o755); err != nil {
		t.Fatal(err)
	}
	base := time.Now().Truncate(time.Second)
	setProc := func(off int64) {
		ts := base.Add(time.Duration(off) * time.Second)
		if err := os.Chtimes(procDir, ts, ts); err != nil {
			t.Fatal(err)
		}
	}
	appendLines := func(n int) {
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < n; i++ {
			if _, err := f.WriteString(sessLine); err != nil {
				t.Fatal(err)
			}
		}
		f.Close()
	}

	// Session 1: three requests in the same file the collector is reading.
	setProc(1000)
	appendLines(3)

	c := NewCollector(dir, filepath.Join(dir, "s.json"))
	c.ProcDir = procDir
	if err := c.Refresh(); err != nil {
		t.Fatal(err)
	}
	t.Logf("DIAG after s1: procStart=%d state.ProcStart=%d offset=%d sessionStart=%d",
		c.procStart(), c.state.ProcStart, c.state.Offset, c.state.SessionStart)
	if got := c.Snapshot().Queries; got != 3 {
		t.Fatalf("session 1: queries=%d, want 3", got)
	}
	if c.state.SessionStartTS == 0 {
		t.Fatal("SessionStartTS was not recorded")
	}
	t.Logf("session 1: queries=3, session_s=%d", c.Snapshot().SessionSeconds)

	// Restart: new process token, new requests appended to the SAME file.
	setProc(2000)
	appendLines(2)
	t.Logf("DIAG before s2 read: procStart=%d (changed? %v) fileSize=%d offset=%d",
		c.procStart(), c.procStart() != c.state.ProcStart, fileSize(t, logPath), c.state.Offset)

	if err := c.Refresh(); err != nil {
		t.Fatal(err)
	}
	t.Logf("DIAG after s2: procStart=%d state.ProcStart=%d offset=%d sessionStart=%d queries=%d",
		c.procStart(), c.state.ProcStart, c.state.Offset, c.state.SessionStart, c.Snapshot().Queries)
	t.Logf("DIAG sizes: fileSize=%d (line len=%d)", fileSize(t, logPath), len(sessLine))
	s := c.Snapshot()
	if s.Queries != 2 {
		t.Errorf("after restart: queries=%d, want 2 (the first session's 3 must not carry over)", s.Queries)
	} else {
		t.Logf("after restart: queries=%d (reset, only the new session counted)", s.Queries)
	}
	if s.SessionSeconds < 0 {
		t.Errorf("SessionSeconds=%d, want >= 0", s.SessionSeconds)
	}

	// A restart with no traffic yet must read zero, not the previous total.
	setProc(3000)
	if err := c.Refresh(); err != nil {
		t.Fatal(err)
	}
	if got := c.Snapshot().Queries; got != 0 {
		t.Errorf("after restart with no traffic: queries=%d, want 0", got)
	} else {
		t.Logf("after restart with no traffic: queries=0")
	}

	// Traffic in the new session counts from zero again.
	appendLines(4)
	if err := c.Refresh(); err != nil {
		t.Fatal(err)
	}
	if got := c.Snapshot().Queries; got != 4 {
		t.Errorf("new session: queries=%d, want 4", got)
	} else {
		t.Logf("new session: queries=4")
	}
}

// TestSessionStartDefaultsToFileStart ensures a fresh state with no restart seen
// still counts the whole file, which is the pre-existing behaviour.
func TestSessionStartDefaultsToFileStart(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "kixdns.log")
	if err := os.WriteFile(logPath, []byte(sessLine+sessLine), 0o644); err != nil {
		t.Fatal(err)
	}
	c := NewCollector(dir, filepath.Join(dir, "s.json"))
	if err := c.Refresh(); err != nil {
		t.Fatal(err)
	}
	if got := c.Snapshot().Queries; got != 2 {
		t.Errorf("queries=%d, want 2", got)
	}
	if st := c.state.SessionStart; st != 0 {
		t.Errorf("SessionStart=%d, want 0 for a fresh state", st)
	}
	t.Logf("fresh state: queries=2, sessionStart=0, session_s=%d", c.Snapshot().SessionSeconds)
}
