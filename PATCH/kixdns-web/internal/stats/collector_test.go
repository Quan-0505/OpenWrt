package stats

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCollectorRefreshRealLog drives the exact code path the HTTP API uses
// (NewCollector -> Refresh -> Snapshot) over a real captured log, which the
// other tests bypass by calling consumeLine directly.
func TestCollectorRefreshRealLog(t *testing.T) {
	src := os.Getenv("REAL_LOG")
	if src == "" {
		t.Skip("set REAL_LOG=<path to a captured kixdns --debug log>")
	}
	raw, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	logPath := filepath.Join(dir, "kixdns.log")
	if err := os.WriteFile(logPath, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	statePath := filepath.Join(dir, "stats.json")

	c := NewCollector(dir, statePath)
	if err := c.Refresh(); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	snap := c.Snapshot()

	t.Logf("queries=%d cacheHit=%d cacheMiss=%d ratio=%.2f%%", snap.Queries, snap.CacheHit, snap.CacheMiss, snap.CacheRatio)
	t.Logf("avgUS=%d maxUS=%d avgUpUS=%d slow=%d", snap.AvgUS, snap.MaxUS, snap.AvgUpUS, snap.Slow)
	t.Logf("upOK=%d upFail=%d", snap.UpOK, snap.UpFail)
	t.Logf("domains=%v", snap.Domains)
	t.Logf("qtype=%v", snap.QType)

	if snap.AvgUS == 0 {
		t.Errorf("FAIL: AvgUS == 0 through the Collector path (consumeLine alone gave %d in the other test)", 436)
	}
	if snap.MaxUS == 0 {
		t.Errorf("FAIL: MaxUS == 0")
	}

	// Second refresh must be a no-op (offset already consumed).
	if err := c.Refresh(); err != nil {
		t.Fatal(err)
	}
	if s2 := c.Snapshot(); s2.Queries != snap.Queries {
		t.Errorf("FAIL: second Refresh changed queries %d -> %d (offset handling broken)", snap.Queries, s2.Queries)
	} else {
		t.Logf("second refresh: idempotent, queries still %d", s2.Queries)
	}

	// Append new lines and confirm only the delta is counted.
	extra := `2026-09-16T09:00:00.000000000Z DEBUG request started event="request_started" request_id=900 client=10.0.0.9:1111 qname="new.example" qtype=A qclass=IN background_refresh=false
2026-09-16T09:00:00.100000000Z DEBUG request finished event="request_finished" request_id=900 status=Completed latency_us=5000
`
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(extra)
	f.Close()

	if err := c.Refresh(); err != nil {
		t.Fatal(err)
	}
	s3 := c.Snapshot()
	if s3.Queries != snap.Queries+1 {
		t.Errorf("FAIL: incremental refresh queries %d -> %d, want +1", snap.Queries, s3.Queries)
	} else {
		t.Logf("incremental refresh: queries %d -> %d (delta applied)", snap.Queries, s3.Queries)
	}
	if s3.MaxUS != 5000 {
		t.Errorf("FAIL: appended latency not applied, MaxUS=%d want 5000", s3.MaxUS)
	} else {
		t.Logf("incremental refresh: MaxUS=%d (new line parsed)", s3.MaxUS)
	}
}
