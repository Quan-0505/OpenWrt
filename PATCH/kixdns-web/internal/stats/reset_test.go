package stats

import (
	"os"
	"path/filepath"
	"testing"
)

func appendTo(t *testing.T, path, s string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(s); err != nil {
		t.Fatal(err)
	}
	f.Close()
}

// TestResetDoesNotReReadHistory pins the fix for the reset button appearing
// broken: Reset() used to replace the whole state, which also cleared the read
// position, so the next refresh re-parsed the log from the beginning and the
// counters were refilled immediately.
func TestResetDoesNotReReadHistory(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "kixdns.log")
	if err := os.WriteFile(logPath, []byte(sessLine+sessLine+sessLine), 0o644); err != nil {
		t.Fatal(err)
	}
	c := NewCollector(dir, filepath.Join(dir, "s.json"))
	if err := c.Refresh(); err != nil {
		t.Fatal(err)
	}
	if got := c.Snapshot().Queries; got != 3 {
		t.Fatalf("setup: queries=%d, want 3", got)
	}
	before := c.state.ReadPos

	if err := c.Reset(); err != nil {
		t.Fatal(err)
	}
	if got := c.Snapshot().Queries; got != 0 {
		t.Fatalf("after reset: queries=%d, want 0", got)
	}
	if c.state.ReadPos != before {
		t.Errorf("reset moved ReadPos from %d to %d; that makes the next refresh re-read the log", before, c.state.ReadPos)
	}
	if c.state.SessionStart != 0 {
		t.Errorf("SessionStart=%d, want 0 (whole file is the current session)", c.state.SessionStart)
	}

	// A refresh with no new data must leave the counters at zero.
	if err := c.Refresh(); err != nil {
		t.Fatal(err)
	}
	if got := c.Snapshot().Queries; got != 0 {
		t.Errorf("after reset + idle refresh: queries=%d, want 0 (history was re-read)", got)
	} else {
		t.Log("reset + idle refresh: still 0")
	}

	// New traffic after the reset counts from zero.
	appendTo(t, logPath, sessLine)
	if err := c.Refresh(); err != nil {
		t.Fatal(err)
	}
	if got := c.Snapshot().Queries; got != 1 {
		t.Errorf("after reset + 1 new line: queries=%d, want 1", got)
	} else {
		t.Log("reset + 1 new line: queries=1")
	}
}

// TestResetAlsoClearsHourly mirrors the same guarantee for the trend chart.
func TestResetAlsoClearsHourly(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "kixdns.log")
	if err := os.WriteFile(logPath, []byte(sessLine+sessLine), 0o644); err != nil {
		t.Fatal(err)
	}
	c := NewCollector(dir, filepath.Join(dir, "s.json"))
	if err := c.Refresh(); err != nil {
		t.Fatal(err)
	}
	sum := func() int64 {
		var n int64
		for _, v := range c.Snapshot().Hourly {
			n += v
		}
		return n
	}
	if sum() != 2 {
		t.Fatalf("setup hourly sum=%d, want 2", sum())
	}
	if err := c.Reset(); err != nil {
		t.Fatal(err)
	}
	if sum() != 0 {
		t.Errorf("after reset hourly sum=%d, want 0", sum())
	}
	if err := c.Refresh(); err != nil {
		t.Fatal(err)
	}
	if sum() != 0 {
		t.Errorf("after reset + idle refresh hourly sum=%d, want 0", sum())
	}
	t.Log("hourly cleared and stayed clear")
}
