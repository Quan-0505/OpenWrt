package stats

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"
	"time"
)

// TestConsumeRealLog feeds a captured kixdns --debug log through the parser and
// reports what was extracted. Run with REAL_LOG=<path>.
func TestConsumeRealLog(t *testing.T) {
	path := os.Getenv("REAL_LOG")
	if path == "" {
		t.Skip("set REAL_LOG=<path to a captured kixdns --debug log>")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	s := newState()
	loc := time.Local
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var lines, withLatency int
	for sc.Scan() {
		line := sc.Text()
		lines++
		if _, ok := splitKV(line)["latency_us"]; ok {
			withLatency++
		}
		s.consumeLine(line, loc)
	}

	ratio := 0.0
	if tot := s.CacheHit + s.CacheMiss; tot > 0 {
		ratio = float64(s.CacheHit) / float64(tot) * 100
	}
	avg := int64(0)
	if s.Queries > 0 {
		avg = s.SumUS / s.Queries
	}
	t.Logf("lines=%d  lines_with_latency_us=%d", lines, withLatency)
	t.Logf("queries=%d cacheHit=%d cacheMiss=%d ratio=%.2f%%", s.Queries, s.CacheHit, s.CacheMiss, ratio)
	t.Logf("sumUS=%d avgUS=%d maxUS=%d slow=%d", s.SumUS, avg, s.MaxUS, s.Slow)
	t.Logf("upOK=%d upFail=%d upSumUS=%d", s.UpOK, s.UpFail, s.UpSumUS)
	t.Logf("qtype=%v", s.QT)
	t.Logf("clients=%v", s.IP)
	t.Logf("upstreams=%v", s.UP)
	t.Logf("rcodes=%v", s.RC)
	t.Logf("domains=%v", s.DM)

	// ground truth expectations for the captured sample
	if s.SumUS == 0 {
		t.Errorf("FAIL: SumUS == 0 but %d lines carry latency_us", withLatency)
	}
	if s.Queries == 0 {
		t.Errorf("FAIL: no queries parsed")
	}
	if s.CacheHit+s.CacheMiss != s.Queries {
		t.Errorf("FAIL: cache_hit+cache_miss (%d) != queries (%d)", s.CacheHit+s.CacheMiss, s.Queries)
	}
}

// TestSplitKV pins the tokenizer behaviour on the real line shapes.
func TestSplitKV(t *testing.T) {
	cases := []struct {
		name string
		line string
		want map[string]string
	}{
		{
			name: "request_finished",
			line: `2026-09-16T08:42:09.286520597Z DEBUG request finished event="request_finished" request_id=1 status=Completed latency_us=1284`,
			want: map[string]string{"event": "request_finished", "request_id": "1", "status": "Completed", "latency_us": "1284"},
		},
		{
			name: "upstream_result_multiword_value",
			line: `2026-09-16T08:42:09.286389087Z DEBUG upstream result event="upstream_result" request_id=1 upstream="127.0.0.1:1053" transport=Udp outcome=Success via=Udp latency_us=901 rcode=No Error truncated=false`,
			want: map[string]string{"event": "upstream_result", "request_id": "1", "latency_us": "901", "rcode": "No Error", "truncated": "false", "transport": "Udp"},
		},
		{
			name: "cache_hit",
			line: `2026-09-16T08:22:34.333255503Z DEBUG cache hit event="cache_hit" request_id=2 kind=Fresh remaining_ttl_s=27 original_ttl_s=27`,
			want: map[string]string{"event": "cache_hit", "kind": "Fresh", "remaining_ttl_s": "27", "original_ttl_s": "27"},
		},
		{
			name: "request_started_with_client_port",
			line: `2026-09-16T08:22:34.293474513Z DEBUG request started event="request_started" request_id=1 listener="default" client=127.0.0.1:51940 qname="www.qq.com" qtype=A qclass=IN background_refresh=false`,
			want: map[string]string{"event": "request_started", "client": "127.0.0.1:51940", "qname": "www.qq.com", "qtype": "A", "qclass": "IN"},
		},
	}
	for _, c := range cases {
		got := splitKV(c.line)
		for k, want := range c.want {
			if got[k] != want {
				t.Errorf("%s: key %q = %q, want %q   (all=%v)", c.name, k, got[k], want, got)
			}
		}
	}
}

// TestSnapshotShape ensures the JSON payload has the keys the UI depends on.
func TestSnapshotShape(t *testing.T) {
	c := NewCollector(t.TempDir(), "")
	snap := c.Snapshot()
	b, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"queries", "cache_hit", "cache_miss", "cache_ratio", "avg_us", "max_us", "avg_up_us", "up_ok", "up_fail", "slow", "qps", "qtype", "rcodes", "clients", "domains", "upstreams", "hourly", "daily"} {
		if _, ok := m[k]; !ok {
			t.Errorf("snapshot missing key %q", k)
		}
	}
}
