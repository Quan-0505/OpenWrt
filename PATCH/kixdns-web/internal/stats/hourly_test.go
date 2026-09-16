package stats

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// realLines is a short slice of kixdns's own output: two requests, with all the
// extra events kixdns emits around them (cache lookup, forwarding, response).
const realLines = `2026-09-16T10:00:00.000000000Z DEBUG request started event="request_started" request_id=1 client=127.0.0.1:1 qname="a.example" qtype=A qclass=IN
2026-09-16T10:00:00.001000000Z DEBUG cache miss event="cache_miss" request_id=1
2026-09-16T10:00:00.002000000Z DEBUG cache lookup event="cache_lookup" request_id=1
2026-09-16T10:00:00.010000000Z  INFO forwarded event="dns_response" upstream=udp:127.0.0.1:1053 qname=a.example qtype=A cache=true
2026-09-16T10:00:00.011000000Z DEBUG request finished event="request_finished" request_id=1 status=Completed latency_us=1234
2026-09-16T10:00:05.000000000Z DEBUG request started event="request_started" request_id=2 client=127.0.0.1:2 qname="b.example" qtype=A qclass=IN
2026-09-16T10:00:05.001000000Z DEBUG cache hit event="cache_hit" request_id=2 kind=Fresh remaining_ttl_s=10 original_ttl_s=30
2026-09-16T10:00:05.002000000Z DEBUG request finished event="request_finished" request_id=2 status=Completed latency_us=200
`

// TestHourlyCountsQueriesNotLines pins the chart's meaning: the hourly trend is
// query volume, not log volume. kixdns emits several events per request, and
// counting every timestamped line made the chart's total several times the real
// query count.
func TestHourlyCountsQueriesNotLines(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "kixdns.log"), []byte(realLines), 0o644); err != nil {
		t.Fatal(err)
	}
	c := NewCollector(dir, filepath.Join(dir, "s.json"))
	// Pin the bucket zone. The log stamps are UTC ("…T10:00:00Z"), so without
	// this the expected hour would depend on the machine's own timezone — the
	// test would pass on a UTC box and fail on anything else.
	c.loc = time.UTC
	if err := c.Refresh(); err != nil {
		t.Fatal(err)
	}
	s := c.Snapshot()

	eventLines := strings.Count(realLines, "\n")
	started := strings.Count(realLines, `event="request_started"`)
	t.Logf("log has %d lines, %d request_started events", eventLines, started)

	if s.Queries != int64(started) {
		t.Errorf("queries=%d, want %d", s.Queries, started)
	}
	var hourly int64
	for _, v := range s.Hourly {
		hourly += v
	}
	if hourly != s.Queries {
		t.Errorf("hourly total=%d, want %d (= queries); the chart is counting lines, not queries", hourly, s.Queries)
	} else {
		t.Logf("hourly total=%d equals queries=%d", hourly, s.Queries)
	}
	if hourly == int64(eventLines) {
		t.Errorf("hourly total equals the raw line count (%d): the fix is not in effect", eventLines)
	}
	if v := s.Hourly[10]; v != 2 {
		t.Errorf("hourly[10]=%d, want 2 (both requests are at 10:00)", v)
	} else {
		t.Logf("hourly[10]=2 as expected")
	}
}
