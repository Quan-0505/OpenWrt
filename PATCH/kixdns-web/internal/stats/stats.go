// Package stats aggregates kixdns observer events into dashboard statistics.
//
// Incremental strategy: only newly appended bytes of the active log file are
// parsed on each refresh, and aggregated counters are persisted as JSON so a
// restart does not lose history. This mirrors the OPNsense plugin's approach.
//
// Supported event lines (kixdns run --debug):
//
//	DEBUG request started event="request_started" request_id=1 client=127.0.0.1:51940 qname="www.qq.com" qtype=A qclass=IN background_refresh=false
//	DEBUG cache hit       event="cache_hit" request_id=2 kind=Fresh remaining_ttl_s=27 original_ttl_s=27
//	DEBUG cache miss      event="cache_miss" request_id=1
//	DEBUG request finished event="request_finished" request_id=1 status=Completed latency_us=36131
//	DEBUG upstream result  event="upstream_result" request_id=1 upstream="127.0.0.1:1053" outcome=Success latency_us=35261 rcode=No Error
package stats

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Counter is a name/count pair used by all top-N lists.
type Counter struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// State is the aggregated, persisted view of all parsed events.
type State struct {
	Offset int64  `json:"offset"`
	File   string `json:"file"`
	// ProcStart is a change-token (mtime of /proc/<pid>) for the monitored
	// process. When it changes, the log file belongs to a previous session
	// and Offset must be discarded.
	ProcStart int64 `json:"proc_start"`

	StartTS int64 `json:"start_ts"`
	LastTS  int64 `json:"last_ts"`

	Queries   int64 `json:"queries"`
	CacheHit  int64 `json:"cache_hit"`
	CacheMiss int64 `json:"cache_miss"`
	Blocked   int64 `json:"blocked"`
	UpOK      int64 `json:"up_ok"`
	UpFail    int64 `json:"up_fail"`
	UpSumUS   int64 `json:"up_sum_us"`
	SumUS     int64 `json:"sum_us"`
	MaxUS     int64 `json:"max_us"`
	Slow      int64 `json:"slow"`
	// CacheWritten counts responses kixdns stored in its cache (dns_response
	// with cache=true). Cumulative, not current occupancy.
	CacheWritten int64 `json:"cache_written"`

	QT map[string]int64 `json:"qtype"`
	RC map[string]int64 `json:"rcode"`
	IP map[string]int64 `json:"clients"`
	DM map[string]int64 `json:"domains"`
	UP map[string]int64 `json:"upstreams"`

	Hourly []int64 `json:"hourly"`
	Daily  []int64 `json:"daily"`

	dayKey string
}

const slowThresholdUS = 500000

func newState() *State {
	return &State{
		QT:     map[string]int64{},
		RC:     map[string]int64{},
		IP:     map[string]int64{},
		DM:     map[string]int64{},
		UP:     map[string]int64{},
		Hourly: make([]int64, 24),
		Daily:  make([]int64, 30),
	}
}

// ---------------------------------------------------------------------------
// line parsing
//
// Lines are tokenised into key=value pairs rather than matched field-by-field:
// kixdns values can contain spaces (rcode=No Error) and can be followed by
// further fields on the same line, which greedy per-field regexes get wrong.
// ---------------------------------------------------------------------------

// reTS extracts the RFC3339 timestamp prefix that every log line carries.
var reTS = regexp.MustCompile("^(\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2})")

func isKeyByte(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// splitKV parses "k=v k2=\"a b\" k3=x" into a map. Values keep inner spaces
// when quoted, and otherwise end at the last space before the next key token.
func splitKV(line string) map[string]string {
	out := make(map[string]string, 12)
	n := len(line)
	i := 0
	for i < n {
		// find start of a key: a run of key bytes followed by '='
		j := i
		for j < n && isKeyByte(line[j]) {
			j++
		}
		if j >= n || j == i || line[j] != '=' {
			i = j + 1
			continue
		}
		key := line[i:j]
		j++ // skip '='
		if j < n && line[j] == '"' {
			end := strings.IndexByte(line[j+1:], '"')
			if end < 0 {
				out[key] = line[j+1:]
				return out
			}
			out[key] = line[j+1 : j+1+end]
			i = j + 2 + end
			continue
		}
		start := j
		k := j
		lastWS := -1
		stop := false
		for k < n && !stop {
			switch line[k] {
			case ' ':
				lastWS = k
			case '"':
				k++
				for k < n && line[k] != '"' {
					k++
				}
			default:
				if lastWS >= 0 && isKeyByte(line[k]) {
					p := k
					for p < n && isKeyByte(line[p]) {
						p++
					}
					if p < n && line[p] == '=' {
						stop = true
						continue
					}
				}
			}
			k++
		}
		if lastWS >= 0 {
			out[key] = line[start:lastWS]
			i = lastWS + 1
		} else {
			out[key] = line[start:]
			i = n
		}
	}
	return out
}

// consumeLine folds one log line into the state.
func (s *State) consumeLine(line string, loc *time.Location) {
	if m := reTS.FindStringSubmatch(line); len(m) == 2 {
		if t, err := time.ParseInLocation("2006-01-02T15:04:05", m[1], loc); err == nil {
			s.Hourly[t.Hour()]++
			s.LastTS = t.Unix()
			if s.StartTS == 0 {
				s.StartTS = t.Unix()
			}
			dk := t.Format("2006-01-02")
			switch {
			case s.dayKey == "":
				s.dayKey = dk
			case s.dayKey != dk:
				copy(s.Daily, s.Daily[1:])
				s.Daily[len(s.Daily)-1] = 0
				s.dayKey = dk
			}
		}
	}

	kv := splitKV(line)
	ev := kv["event"]
	if ev == "" {
		return
	}

	// Field lookup trims surrounding whitespace. Refresh() reads lines with
	// bufio.Reader.ReadString('\n'), so the value of the last field on a line
	// would otherwise carry a trailing newline and fail numeric parsing.
	// This is why request_finished's trailing latency_us was silently dropped.
	get := func(k string) string { return strings.TrimSpace(kv[k]) }

	atoi := func(v string) (int64, bool) {
		if v == "" {
			return 0, false
		}
		n, err := strconv.ParseInt(v, 10, 64)
		return n, err == nil
	}

	switch ev {
	case "request_started":
		s.Queries++
		if d := get("qname"); d != "" {
			s.DM[d]++
		}
		if q := get("qtype"); q != "" {
			s.QT[q]++
		}
		if c := get("client"); c != "" {
			if i := strings.LastIndexByte(c, ':'); i > 0 {
				c = c[:i]
			}
			s.IP[c]++
		}
	case "cache_hit":
		s.CacheHit++
	case "cache_miss":
		s.CacheMiss++
	case "request_finished":
		if st := get("status"); st != "" && st != "Completed" {
			s.Blocked++
		}
		if us, ok := atoi(get("latency_us")); ok {
			s.SumUS += us
			if us > s.MaxUS {
				s.MaxUS = us
			}
			if us > slowThresholdUS {
				s.Slow++
			}
		}
	case "dns_response":
		// kixdns writes the response into its cache when it had to forward;
		// this event is emitted only on that path (cache hits do not emit it).
		if get("cache") == "true" {
			s.CacheWritten++
		}
	case "upstream_result":
		if u := get("upstream"); u != "" {
			s.UP[u]++
		}
		if rc := get("rcode"); rc != "" {
			s.RC[rc]++
		}
		switch get("outcome") {
		case "Success":
			s.UpOK++
		case "":
			// no outcome field on this line: not counted either way
		default:
			s.UpFail++
		}
		if us, ok := atoi(get("latency_us")); ok {
			s.UpSumUS += us
		}
	}
}

// ---------------------------------------------------------------------------
// collector
// ---------------------------------------------------------------------------

// Collector owns the on-disk aggregated state and serialises refreshes.
type Collector struct {
	LogDir    string // directory containing kixdns log files
	StatePath string // where aggregated state is persisted

	mu    sync.Mutex
	state *State
	loc   *time.Location

	// ProcDir, when set, is watched for restarts. Any change to its mtime
	// means the process was replaced, which invalidates the stored offset.
	ProcDir string
	lastErr string
}

// procStart returns a change-token for the monitored process, or 0 when it
// is not running.
func (c *Collector) procStart() int64 {
	if c.ProcDir == "" {
		return 0
	}
	fi, err := os.Stat(c.ProcDir)
	if err != nil {
		return 0
	}
	return fi.ModTime().UnixNano()
}

// SetProcPID points the collector at a live PID so restarts can be detected.
func (c *Collector) SetProcPID(pid int) {
	if pid > 0 {
		c.ProcDir = filepath.Join("/proc", strconv.Itoa(pid))
	} else {
		c.ProcDir = ""
	}
}

// NewCollector loads persisted state (if any) and returns a ready collector.
func NewCollector(logDir, statePath string) *Collector {
	c := &Collector{LogDir: logDir, StatePath: statePath, loc: time.Local, state: newState()}
	b, err := os.ReadFile(statePath)
	if err != nil {
		return c
	}
	var s State
	if json.Unmarshal(b, &s) != nil {
		return c
	}
	if s.QT == nil {
		s.QT = map[string]int64{}
	}
	if s.RC == nil {
		s.RC = map[string]int64{}
	}
	if s.IP == nil {
		s.IP = map[string]int64{}
	}
	if s.DM == nil {
		s.DM = map[string]int64{}
	}
	if s.UP == nil {
		s.UP = map[string]int64{}
	}
	if len(s.Hourly) != 24 {
		s.Hourly = make([]int64, 24)
	}
	if len(s.Daily) != 30 {
		s.Daily = make([]int64, 30)
	}
	c.state = &s
	return c
}

// ActiveLog returns the newest *.log file in LogDir.
func (c *Collector) ActiveLog() (string, os.FileInfo, error) {
	des, err := os.ReadDir(c.LogDir)
	if err != nil {
		return "", nil, err
	}
	var best string
	var bestInfo os.FileInfo
	for _, de := range des {
		if de.IsDir() || !strings.HasSuffix(de.Name(), ".log") {
			continue
		}
		info, err := de.Info()
		if err != nil {
			continue
		}
		if bestInfo == nil || info.ModTime().After(bestInfo.ModTime()) {
			best, bestInfo = filepath.Join(c.LogDir, de.Name()), info
		}
	}
	if bestInfo == nil {
		return "", nil, os.ErrNotExist
	}
	return best, bestInfo, nil
}

// Refresh parses newly appended bytes. Safe for concurrent use.
func (c *Collector) Refresh() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	file, info, err := c.ActiveLog()
	if err != nil {
		c.lastErr = "no log file in " + c.LogDir
		return nil // absence of logs is not an error for the caller
	}

	s := c.state
	// A restarted process keeps appending to the same file, so a stale offset
	// would silently skip everything written before it. Detect the restart
	// and re-read the current file from the start.
	if ps := c.procStart(); ps != 0 {
		if s.ProcStart != 0 && ps != s.ProcStart {
			s.Offset = 0
		}
		s.ProcStart = ps
	}

	// Rotation or truncation: restart from the beginning of the new file.
	if s.File != file || info.Size() < s.Offset {
		s.Offset = 0
		s.File = file
	}
	if info.Size() == s.Offset {
		return nil // nothing new
	}

	f, err := os.Open(file)
	if err != nil {
		c.lastErr = err.Error()
		return err
	}
	defer f.Close()

	if _, err := f.Seek(s.Offset, io.SeekStart); err != nil {
		c.lastErr = err.Error()
		return err
	}

	r := bufio.NewReaderSize(f, 256*1024) // kixdns lines can be long
	var consumed int64
	for {
		line, err := r.ReadString('\n')
		if len(line) > 0 {
			if line[len(line)-1] == '\n' {
				s.consumeLine(line, c.loc)
				consumed += int64(len(line))
			} else {
				break // partial trailing line: leave it for the next refresh
			}
		}
		if err != nil {
			break
		}
	}
	s.Offset += consumed
	c.lastErr = ""
	return c.save()
}

func (c *Collector) save() error {
	if c.StatePath == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(c.StatePath), 0o755); err != nil {
		return err
	}
	b, err := json.Marshal(c.state)
	if err != nil {
		return err
	}
	tmp := c.StatePath + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, c.StatePath)
}

// ---------------------------------------------------------------------------
// public snapshot
// ---------------------------------------------------------------------------

// Snapshot is the JSON payload consumed by the web UI.
type Snapshot struct {
	Generated int64 `json:"generated"`
	StartTS   int64 `json:"start_ts"`
	LastTS    int64 `json:"last_ts"`

	Queries    int64   `json:"queries"`
	CacheHit   int64   `json:"cache_hit"`
	CacheMiss  int64   `json:"cache_miss"`
	CacheRatio float64 `json:"cache_ratio"`
	Blocked    int64   `json:"blocked"`

	AvgUS   int64   `json:"avg_us"`
	MaxUS   int64   `json:"max_us"`
	AvgUpUS int64   `json:"avg_up_us"`
	UpOK    int64   `json:"up_ok"`
	UpFail  int64   `json:"up_fail"`
	Slow    int64   `json:"slow"`
	QPS     float64 `json:"qps"`
	// CacheWritten is a cumulative write count; kixdns exposes no live
	// cache-entry count, so the UI labels it accordingly.
	CacheWritten int64 `json:"cache_written"`

	Error string `json:"error,omitempty"`

	QType     []Counter `json:"qtype"`
	RCodes    []Counter `json:"rcodes"`
	Clients   []Counter `json:"clients"`
	Domains   []Counter `json:"domains"`
	Upstreams []Counter `json:"upstreams"`
	Hourly    []int64   `json:"hourly"`
	Daily     []int64   `json:"daily"`
}

func topN(m map[string]int64, n int) []Counter {
	out := make([]Counter, 0, len(m))
	for k, v := range m {
		out = append(out, Counter{Name: k, Count: v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Name < out[j].Name
	})
	if n > 0 && len(out) > n {
		out = out[:n]
	}
	return out
}

// Snapshot returns the current aggregated statistics.
func (c *Collector) Snapshot() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	s := c.state

	snap := Snapshot{
		Generated:    time.Now().Unix(),
		StartTS:      s.StartTS,
		LastTS:       s.LastTS,
		Queries:      s.Queries,
		CacheHit:     s.CacheHit,
		CacheMiss:    s.CacheMiss,
		Blocked:      s.Blocked,
		MaxUS:        s.MaxUS,
		UpOK:         s.UpOK,
		UpFail:       s.UpFail,
		Slow:         s.Slow,
		CacheWritten: s.CacheWritten,
		Error:        c.lastErr,
	}
	if s.Queries > 0 {
		snap.AvgUS = s.SumUS / s.Queries
	}
	if up := s.UpOK + s.UpFail; up > 0 {
		snap.AvgUpUS = s.UpSumUS / up
	}
	if tot := s.CacheHit + s.CacheMiss; tot > 0 {
		snap.CacheRatio = float64(s.CacheHit) / float64(tot) * 100
	}
	if span := s.LastTS - s.StartTS; span > 0 {
		snap.QPS = float64(s.Queries) / float64(span)
	}

	snap.QType = topN(s.QT, 0)
	snap.RCodes = topN(s.RC, 0)
	snap.Clients = topN(s.IP, 20)
	snap.Domains = topN(s.DM, 20)
	snap.Upstreams = topN(s.UP, 0)
	snap.Hourly = append([]int64(nil), s.Hourly...)
	snap.Daily = append([]int64(nil), s.Daily...)
	return snap
}

// Reset clears aggregated state (both in memory and on disk).
func (c *Collector) Reset() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state = newState()
	c.lastErr = ""
	return c.save()
}
