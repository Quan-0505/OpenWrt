// Command kixdns-web is a standalone AdGuardHome-style console for kixdns.
//
// It serves an embedded single-page dashboard plus a small JSON API, and works
// on both OpenWrt (uhttpd can reverse-proxy to it, or it can listen directly)
// and OPNsense/FreeBSD.
//
// Usage:
//
//	kixdns-web -listen 0.0.0.0:8080 \
//	           -bin /usr/bin/kixdns \
//	           -config /etc/kixdns/pipeline.json \
//	           -logdir /var/log/kixdns \
//	           -state /var/lib/kixdns/stats.json
package main

import (
	"crypto/subtle"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"kixdns-web/internal/dnsmasq"
	"kixdns-web/internal/service"
	"kixdns-web/internal/stats"
)

//go:embed all:web
var webFS embed.FS

type app struct {
	svc       *service.Manager
	stats     *stats.Collector
	dnsmasq   *dnsmasq.Reader
	statePath string
	logDir    string
	token     string // optional shared secret for the API
	readOnly  bool
	keepLogs  int
	logCap    int64
}

func main() {
	var (
		listen   = flag.String("listen", envOr("KIXDNS_WEB_LISTEN", "0.0.0.0:8080"), "HTTP listen address")
		bin      = flag.String("bin", os.Getenv("KIXDNS_BIN"), "path to the kixdns binary")
		workdir  = flag.String("workdir", os.Getenv("KIXDNS_WORKDIR"), "kixdns working directory (-d)")
		config   = flag.String("config", os.Getenv("KIXDNS_CONFIG"), "path to pipeline.json")
		logDir   = flag.String("logdir", os.Getenv("KIXDNS_LOGDIR"), "directory holding kixdns logs")
		state    = flag.String("state", os.Getenv("KIXDNS_WEB_STATE"), "aggregated stats state file")
		pidFile  = flag.String("pidfile", os.Getenv("KIXDNS_PIDFILE"), "kixdns PID file")
		token    = flag.String("token", os.Getenv("KIXDNS_WEB_TOKEN"), "optional shared secret (header X-Auth-Token or ?token=)")
		interval = flag.Duration("interval", 5*time.Second, "stats refresh interval")
		readOnly = flag.Bool("readonly", os.Getenv("KIXDNS_WEB_READONLY") == "1", "disable all mutating endpoints")
		keepLogs = flag.Int("keep-logs", 2, "rotated log files to keep (active + this = total days retained)")
		logCapMB = flag.Int64("log-cap-mb", 24, "rotate the active log when it exceeds this size (MB)")
		check    = flag.Bool("check", false, "validate configuration and exit")
	)
	flag.Parse()

	svcMgr := service.New(*bin, *workdir, *config, *logDir, *pidFile)
	if *state == "" {
		base := svcMgr.LogDir
		if base == "" {
			base = "/tmp"
		}
		*state = filepath.Join(base, "kixdns-web-stats.json")
	}

	if *check {
		fmt.Printf("kixdns   : %s (%s)\n", svcMgr.Bin, svcMgr.Version())
		fmt.Printf("workdir  : %s\n", svcMgr.WorkDir)
		fmt.Printf("config   : %s\n", svcMgr.Config)
		fmt.Printf("logdir   : %s\n", svcMgr.LogDir)
		fmt.Printf("state    : %s\n", *state)
		fmt.Printf("pidfile  : %s\n", svcMgr.PIDFile)
		if _, err := os.Stat(svcMgr.Config); err != nil {
			fmt.Printf("config check: MISSING (%v)\n", err)
			os.Exit(2)
		}
		if err := svcMgr.ValidateConfig(mustRead(svcMgr.Config)); err != nil {
			fmt.Printf("config check: FAILED (%v)\n", err)
			os.Exit(1)
		}
		fmt.Println("config check: OK")
		return
	}

	collector := stats.NewCollector(svcMgr.LogDir, *state)
	// Let the collector notice kixdns restarts: a restart keeps appending to
	// the same log file, and a stale byte offset would freeze the counters.
	if pids := svcMgr.PIDs(); len(pids) > 0 {
		collector.SetProcPID(pids[0])
	}

	a := &app{
		svc:       svcMgr,
		stats:     collector,
		dnsmasq:   dnsmasq.New(),
		statePath: *state,
		logDir:    svcMgr.LogDir,
		token:     *token,
		readOnly:  *readOnly,
		keepLogs:  *keepLogs,
		logCap:    *logCapMB * 1024 * 1024,
	}

	// dnsmasq's own counters: kixdns only sees what dnsmasq forwards, so the
	// client-side totals have to come from dnsmasq itself.
	stopDNS := make(chan struct{})
	defer close(stopDNS)
	go a.dnsmasq.Run(15*time.Second, stopDNS)

	// background aggregation + log rotation
	go func() {
		for {
			if pids := a.svc.PIDs(); len(pids) > 0 {
				a.stats.SetProcPID(pids[0])
			}
			if err := a.stats.Refresh(); err != nil {
				log.Printf("stats refresh: %v", err)
			}
			a.maybeRotate()
			time.Sleep(*interval)
		}
	}()

	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("embedded assets: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/api/", a.auth(http.HandlerFunc(a.api)))
	mux.Handle("/", a.auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// serve embedded assets; unknown paths fall back to the SPA entry
		p := strings.TrimPrefix(r.URL.Path, "/")
		if p == "" {
			p = "index.html"
		}
		if _, err := fs.Stat(sub, p); err != nil {
			p = "index.html"
		}
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFileFS(w, r, sub, p)
	})))

	log.Printf("kixdns-web listening on %s", *listen)
	log.Printf("  config=%s", svcMgr.Config)
	log.Printf("  logdir=%s state=%s", svcMgr.LogDir, *state)
	srv := &http.Server{
		Addr:              *listen,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func mustRead(p string) []byte {
	b, err := os.ReadFile(p)
	if err != nil {
		log.Fatalf("read %s: %v", p, err)
	}
	return b
}

// maybeRotate rotates the active log once it exceeds the configured cap.
func (a *app) maybeRotate() {
	if a.logCap <= 0 {
		return
	}
	active := filepath.Join(a.logDir, "kixdns.log")
	info, err := os.Stat(active)
	if err != nil || info.Size() < a.logCap {
		return
	}
	if dst, err := a.svc.RotateLog(a.keepLogs); err == nil && dst != "" {
		log.Printf("rotated log -> %s", dst)
		// kixdns holds the old fd; a restart is required for it to reopen the
		// active path. Do it only when the daemon was started with debug output.
		if st := a.svc.Status(); st.Running && st.Debug {
			_ = a.svc.Restart(true)
			log.Printf("kixdns restarted after rotation")
		}
	}
}

// auth enforces the optional shared secret.
func (a *app) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.token == "" {
			next.ServeHTTP(w, r)
			return
		}
		got := r.Header.Get("X-Auth-Token")
		if got == "" {
			got = r.URL.Query().Get("token")
		}
		if subtle.ConstantTimeCompare([]byte(got), []byte(a.token)) != 1 {
			w.Header().Set("WWW-Authenticate", `Bearer realm="kixdns-web"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ---------------------------------------------------------------------------
// API
// ---------------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func writeErr(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, map[string]string{"error": err.Error()})
}

func (a *app) api(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/")
	mutating := r.Method != http.MethodGet

	switch path {

	case "status":
		cfg, _ := a.svc.ReadConfig()
		writeJSON(w, 200, map[string]any{
			"service": a.svc.Status(),
			"paths": map[string]string{
				"bin":     a.svc.Bin,
				"workdir": a.svc.WorkDir,
				"config":  a.svc.Config,
				"logdir":  a.svc.LogDir,
				"state":   a.statePath,
				"pidfile": a.svc.PIDFile,
			},
			"readonly":  a.readOnly,
			"config_ok": cfg != nil,
			"log_files": a.svc.LogFiles(),
		})

	case "dnsmasq":
		writeJSON(w, 200, a.dnsmasq.Snapshot())

	case "stats":
		if err := a.stats.Refresh(); err != nil {
			writeErr(w, 500, err)
			return
		}
		writeJSON(w, 200, a.stats.Snapshot())

	case "stats/reset":
		if !a.allowMutate(w, mutating) {
			return
		}
		if err := a.stats.Reset(); err != nil {
			writeErr(w, 500, err)
			return
		}
		writeJSON(w, 200, map[string]string{"status": "reset"})

	case "config":
		if mutating {
			if !a.allowMutate(w, true) {
				return
			}
			body, err := readBody(r)
			if err != nil {
				writeErr(w, 400, err)
				return
			}
			if err := a.svc.WriteConfig(body); err != nil {
				writeErr(w, 400, err)
				return
			}
			writeJSON(w, 200, map[string]string{"status": "saved"})
			return
		}
		cfg, err := a.svc.ReadConfig()
		if err != nil {
			writeErr(w, 404, err)
			return
		}
		writeJSON(w, 200, map[string]any{"raw": string(cfg), "backups": a.svc.Backups()})

	case "config/validate":
		if !a.allowMutate(w, mutating) {
			return
		}
		body, err := readBody(r)
		if err != nil {
			writeErr(w, 400, err)
			return
		}
		if err := a.svc.ValidateConfig(body); err != nil {
			writeJSON(w, 200, map[string]any{"valid": false, "error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]any{"valid": true})

	case "service/start", "service/stop", "service/restart":
		if !a.allowMutate(w, mutating) {
			return
		}
		debug := r.URL.Query().Get("debug") == "1"
		var err error
		switch {
		case strings.HasSuffix(path, "start"):
			err = a.svc.Start(debug)
		case strings.HasSuffix(path, "stop"):
			err = a.svc.Stop()
		default:
			err = a.svc.Restart(debug)
		}
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		writeJSON(w, 200, map[string]any{"status": "ok", "service": a.svc.Status()})

	case "logs":
		n, _ := strconv.Atoi(r.URL.Query().Get("lines"))
		if n <= 0 || n > 20000 {
			n = 300
		}
		file := r.URL.Query().Get("file")
		lines, err := a.svc.TailLog(file, n)
		if err != nil {
			writeErr(w, 404, err)
			return
		}
		writeJSON(w, 200, map[string]any{
			"file":  file,
			"lines": lines,
			"files": a.svc.LogFiles(),
		})

	default:
		writeErr(w, 404, fmt.Errorf("unknown endpoint: %s", path))
	}
}

func (a *app) allowMutate(w http.ResponseWriter, mutating bool) bool {
	if a.readOnly {
		writeErr(w, http.StatusForbidden, fmt.Errorf("read-only mode"))
		return false
	}
	if !mutating {
		writeErr(w, http.StatusMethodNotAllowed, fmt.Errorf("POST required"))
		return false
	}
	return true
}

func readBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	buf := make([]byte, 0, 64*1024)
	tmp := make([]byte, 32*1024)
	for {
		n, err := r.Body.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
			if len(buf) > 4*1024*1024 {
				return nil, fmt.Errorf("body too large")
			}
		}
		if err != nil {
			break
		}
	}
	if len(buf) == 0 {
		return nil, fmt.Errorf("empty body")
	}
	return buf, nil
}
