// Package dnsmasq reads dnsmasq's own DNS counters over ubus.
//
// Why this exists: on this device the lookup path is
//
//	client -> dnsmasq (cache) -> kixdns -> mihomo
//
// so kixdns only ever sees the queries dnsmasq had to forward. Everything
// dnsmasq answers from its own cache or from local/host entries never reaches
// kixdns, which makes the kixdns-side query count a *forwarded* count rather
// than a client-request count. dnsmasq exposes the missing half through ubus
// (`ubus call dnsmasq metrics`), and this package surfaces it.
//
// The values are read as a JSON object from `ubus call dnsmasq metrics`; the
// counters are cumulative since dnsmasq started, and dnsmasq offers no
// total-client-query counter, so the client total must be derived:
//
//	client queries ≈ dns_queries_forwarded + dns_local_answered
package dnsmasq

import (
	"encoding/json"
	"os"
	"os/exec"
	"sync"
	"time"
)

// Metrics mirrors the subset of dnsmasq's counters that matter here.
type Metrics struct {
	CacheInserted    int64 `json:"dns_cache_inserted"`
	QueriesForwarded int64 `json:"dns_queries_forwarded"`
	LocalAnswered    int64 `json:"dns_local_answered"`
	AuthAnswered     int64 `json:"dns_auth_answered"`
	StaleAnswered    int64 `json:"dns_stale_answered"`
	Unanswered       int64 `json:"dns_unanswered"`
	NoAnswer         int64 `json:"noanswer"`
}

// Reader polls dnsmasq metrics periodically and caches the last good sample.
type Reader struct {
	mu       sync.RWMutex
	last     *Metrics
	lastErr  string
	at       time.Time
	bin      string
	call     []string
	capacity int64
}

// New returns a Reader that shells out to ubus.
func New() *Reader {
	return &Reader{bin: "ubus", call: []string{"call", "dnsmasq", "metrics"}}
}

// Sample returns the most recent successful reading and when it was taken.
func (r *Reader) Sample() (*Metrics, time.Time, string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.last, r.at, r.lastErr
}

// Refresh runs one ubus call. Failures are recorded but never fatal: the
// dashboard must keep working on devices without ubus or without dnsmasq.
func (r *Reader) Refresh() {
	out, err := exec.Command(r.bin, r.call...).Output()
	if err != nil {
		r.setErr(err.Error())
		return
	}
	var m Metrics
	if err := json.Unmarshal(out, &m); err != nil {
		r.setErr("parse: " + err.Error())
		return
	}
	r.mu.Lock()
	r.last = &m
	r.at = time.Now()
	r.lastErr = ""
	r.mu.Unlock()
}

func (r *Reader) setErr(msg string) {
	r.mu.Lock()
	r.lastErr = msg
	r.mu.Unlock()
}

// Run polls until the channel is closed.
func (r *Reader) Run(every time.Duration, stop <-chan struct{}) {
	t := time.NewTicker(every)
	defer t.Stop()
	r.Refresh()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			r.Refresh()
		}
	}
}

// View is the JSON shape handed to the UI.
type View struct {
	Available bool   `json:"available"`
	Age       int64  `json:"age_s"`
	Error     string `json:"error,omitempty"`

	Forwarded  int64 `json:"forwarded"`
	Local      int64 `json:"local_answered"`
	CacheIns   int64 `json:"cache_inserted"`
	Auth       int64 `json:"auth_answered"`
	Stale      int64 `json:"stale_answered"`
	Unanswered int64 `json:"unanswered"`

	// Derived: dnsmasq gives no total-client-query counter.
	ClientTotal int64   `json:"client_total"`
	LocalRatio  float64 `json:"local_ratio"`

	// KixdnsCapacity is read from the pipeline config's cache_capacity. It is
	// reported here so the chain table can show kixdns cache writes against the
	// configured ceiling in one place.
	KixdnsCapacity int64 `json:"kixdns_capacity"`
}

// SetKixdnsCapacity records the cache capacity configured for kixdns.
func (r *Reader) SetKixdnsCapacity(n int64) {
	r.mu.Lock()
	r.capacity = n
	r.mu.Unlock()
}

// ReadCapacity extracts settings.cache_capacity from a kixdns pipeline config.
func ReadCapacity(configPath string) int64 {
	b, err := os.ReadFile(configPath)
	if err != nil {
		return 0
	}
	var doc struct {
		Settings struct {
			CacheCapacity int64 `json:"cache_capacity"`
		} `json:"settings"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return 0
	}
	return doc.Settings.CacheCapacity
}

// Snapshot converts the cached sample into the UI shape.
func (r *Reader) Snapshot() View {
	m, at, errMsg := r.Sample()
	r.mu.RLock()
	cap := r.capacity
	r.mu.RUnlock()

	if m == nil {
		return View{Available: false, Error: errMsg, KixdnsCapacity: cap}
	}
	v := View{
		Available:      true,
		Forwarded:      m.QueriesForwarded,
		Local:          m.LocalAnswered,
		CacheIns:       m.CacheInserted,
		Auth:           m.AuthAnswered,
		Stale:          m.StaleAnswered,
		Unanswered:     m.Unanswered + m.NoAnswer,
		KixdnsCapacity: cap,
	}
	if !at.IsZero() {
		v.Age = int64(time.Since(at).Seconds())
	}
	v.ClientTotal = v.Forwarded + v.Local
	if v.ClientTotal > 0 {
		v.LocalRatio = float64(v.Local) / float64(v.ClientTotal) * 100
	}
	return v
}
