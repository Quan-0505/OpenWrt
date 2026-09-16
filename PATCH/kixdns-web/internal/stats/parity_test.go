package stats

import (
	"bufio"
	"os"
	"strings"
	"testing"
	"time"
)

// TestPathParity compares the two code paths line by line, to find where the
// Collector path diverges from the direct-parse path.
func TestPathParity(t *testing.T) {
	src := os.Getenv("REAL_LOG")
	if src == "" {
		t.Skip("set REAL_LOG=<path>")
	}
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}

	// path A: bufio.Scanner, Text() (no trailing \n) — what stats_test.go used
	sa := newState()
	var nA, latA int
	sc := bufio.NewScanner(strings.NewReader(string(b)))
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for sc.Scan() {
		line := sc.Text()
		nA++
		if _, ok := splitKV(line)["latency_us"]; ok {
			latA++
		}
		sa.consumeLine(line, time.Local)
	}

	// path B: bufio.Reader ReadString('\n') — what Refresh uses (keeps the \n)
	sb := newState()
	var nB, latB int
	r := bufio.NewReaderSize(strings.NewReader(string(b)), 256*1024)
	for {
		line, err := r.ReadString('\n')
		if len(line) > 0 {
			if line[len(line)-1] == '\n' {
				nB++
				if _, ok := splitKV(line)["latency_us"]; ok {
					latB++
				}
				sb.consumeLine(line, time.Local)
			} else {
				break
			}
		}
		if err != nil {
			break
		}
	}

	t.Logf("path A (Scanner):  lines=%d lat=%d queries=%d sumUS=%d maxUS=%d upSumUS=%d", nA, latA, sa.Queries, sa.SumUS, sa.MaxUS, sa.UpSumUS)
	t.Logf("path B (ReadString): lines=%d lat=%d queries=%d sumUS=%d maxUS=%d upSumUS=%d", nB, latB, sb.Queries, sb.SumUS, sb.MaxUS, sb.UpSumUS)

	if nA != nB {
		t.Errorf("line count differs: %d vs %d", nA, nB)
	}
	if sa.SumUS != sb.SumUS {
		t.Errorf("sumUS differs: %d vs %d", sa.SumUS, sb.SumUS)
	}
	if sb.SumUS == 0 {
		t.Errorf("FAIL: path B sumUS == 0 (the Collector path)")
		// dump the first request_finished line as seen by path B
		r2 := bufio.NewReaderSize(strings.NewReader(string(b)), 256*1024)
		for {
			line, err := r2.ReadString('\n')
			if strings.Contains(line, `event="request_finished"`) {
				kv := splitKV(line)
				t.Logf("first request_finished line (path B): %q", line)
				t.Logf("  tokens: %v", kv)
				t.Logf("  latency_us=%q", kv["latency_us"])
				break
			}
			if err != nil {
				break
			}
		}
	}
}
