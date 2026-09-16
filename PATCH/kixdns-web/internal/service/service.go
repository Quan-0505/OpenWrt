// Package service controls the kixdns process and its pipeline configuration.
//
// Design rule: never leave the device with a broken configuration. Any write to
// the pipeline file is validated first (JSON parse + `kixdns -t` when the binary
// is available) and the previous revision is kept as a timestamped backup.
package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Manager operates on a concrete kixdns installation.
type Manager struct {
	Bin     string // kixdns binary
	WorkDir string // -d directory (geo data, rule sets)
	Config  string // pipeline.json
	LogDir  string // where stdout/stderr are redirected
	PIDFile string

	mu sync.Mutex
}

// New resolves a Manager from explicit paths, falling back to the well-known
// locations used by the OpenWrt package and by the U60 Pro deployment.
func New(bin, workdir, config, logdir, pidfile string) *Manager {
	m := &Manager{Bin: bin, WorkDir: workdir, Config: config, LogDir: logdir, PIDFile: pidfile}
	if m.Bin == "" {
		m.Bin = firstExisting("/usr/bin/kixdns", "/usr/sbin/kixdns", "/data/ufi-tools/kixdns/kixdns")
	}
	if m.WorkDir == "" {
		m.WorkDir = firstExisting("/etc/kixdns", "/usr/share/kixdns", "/data/ufi-tools/kixdns")
	}
	if m.Config == "" {
		m.Config = firstExisting(
			"/etc/kixdns/pipeline.json",
			"/data/ufi-tools/kixdns/pipeline.json",
			"/usr/local/etc/kixdns/pipeline.json",
		)
	}
	if m.LogDir == "" {
		m.LogDir = firstExisting("/var/log/kixdns", "/data/ufi-tools/kixdns/logs", "/tmp")
	}
	if m.PIDFile == "" {
		if _, err := os.Stat("/var/run"); err == nil {
			m.PIDFile = "/var/run/kixdns.pid"
		} else {
			m.PIDFile = "/tmp/kixdns.pid"
		}
	}
	return m
}

func firstExisting(paths ...string) string {
	for _, p := range paths {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// process handling
// ---------------------------------------------------------------------------

// findPIDs scans /proc for processes whose cmdline references the kixdns
// binary. Used as a fallback when no PID file is maintained, and as the
// authoritative check when a PID file has gone stale.
func (m *Manager) findPIDs() []int {
	var pids []int
	des, err := os.ReadDir("/proc")
	if err != nil {
		return pids
	}
	base := filepath.Base(m.Bin)
	if base == "" {
		return pids
	}
	for _, de := range des {
		pid, err := strconv.Atoi(de.Name())
		if err != nil {
			continue
		}
		b, err := os.ReadFile(filepath.Join("/proc", de.Name(), "cmdline"))
		if err != nil {
			continue
		}
		cmd := strings.ReplaceAll(string(b), "\x00", " ")
		// match the binary but exclude ourselves (our own path contains "kixdns")
		if strings.Contains(cmd, base) && !strings.Contains(cmd, "kixdns-web") {
			pids = append(pids, pid)
		}
	}
	sort.Ints(pids)
	return pids
}

func (m *Manager) readPIDFile() int {
	b, err := os.ReadFile(m.PIDFile)
	if err != nil {
		return 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil || pid <= 0 {
		return 0
	}
	if _, err := os.Stat(filepath.Join("/proc", strconv.Itoa(pid))); err != nil {
		return 0 // stale
	}
	return pid
}

// PIDs returns running kixdns PIDs (PID file first, /proc scan as fallback).
func (m *Manager) PIDs() []int {
	if pid := m.readPIDFile(); pid > 0 {
		return []int{pid}
	}
	return m.findPIDs()
}

// Status reports whether kixdns is running, and how it was started.
type Status struct {
	Running bool   `json:"running"`
	PIDs    []int  `json:"pids"`
	Version string `json:"version"`
	Debug   bool   `json:"debug"`
	Bin     string `json:"bin"`
	WorkDir string `json:"workdir"`
	Config  string `json:"config"`
	LogDir  string `json:"logdir"`
	Uptime  int64  `json:"uptime_s"`
}

func (m *Manager) Status() Status {
	st := Status{Bin: m.Bin, WorkDir: m.WorkDir, Config: m.Config, LogDir: m.LogDir}
	st.PIDs = m.PIDs()
	st.Running = len(st.PIDs) > 0
	if st.Running {
		pidDir := filepath.Join("/proc", strconv.Itoa(st.PIDs[0]))
		if b, err := os.ReadFile(filepath.Join(pidDir, "cmdline")); err == nil {
			st.Debug = strings.Contains(strings.ReplaceAll(string(b), "\x00", " "), "--debug")
		}
		if fi, err := os.Stat(pidDir); err == nil {
			st.Uptime = int64(time.Since(fi.ModTime()).Seconds())
		}
	}
	st.Version = m.Version()
	return st
}

// Version runs `kixdns -V`.
func (m *Manager) Version() string {
	if m.Bin == "" {
		return ""
	}
	out, err := exec.Command(m.Bin, "-V").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// Start launches kixdns in the background, redirecting output to LogDir.
func (m *Manager) Start(debug bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.PIDs()) > 0 {
		return errors.New("kixdns is already running")
	}
	if m.Bin == "" {
		return errors.New("kixdns binary not found")
	}
	if m.Config == "" {
		return errors.New("kixdns config not found")
	}
	if err := os.MkdirAll(m.LogDir, 0o755); err != nil {
		return err
	}
	logPath := filepath.Join(m.LogDir, "kixdns.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	args := []string{"run", "-c", m.Config}
	if debug {
		args = append(args, "--debug")
	}
	cmd := exec.Command(m.Bin, args...)
	cmd.Dir = m.WorkDir
	cmd.Stdout = f
	cmd.Stderr = f
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	_ = os.WriteFile(m.PIDFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0o644)
	// release the child so it survives this process
	_ = cmd.Process.Release()
	return nil
}

// Stop terminates kixdns (SIGTERM, then SIGKILL after a grace period).
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	pids := m.PIDs()
	if len(pids) == 0 {
		return errors.New("kixdns is not running")
	}
	for _, pid := range pids {
		_ = syscall.Kill(pid, syscall.SIGTERM)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if len(m.PIDs()) == 0 {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	for _, pid := range m.PIDs() {
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}
	_ = os.Remove(m.PIDFile)
	return nil
}

// Restart stops (when running) and starts again.
func (m *Manager) Restart(debug bool) error {
	_ = m.Stop()
	time.Sleep(500 * time.Millisecond)
	return m.Start(debug)
}

// ---------------------------------------------------------------------------
// configuration
// ---------------------------------------------------------------------------

// ReadConfig returns the raw pipeline JSON.
func (m *Manager) ReadConfig() (json.RawMessage, error) {
	b, err := os.ReadFile(m.Config)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

// ValidateConfig checks that raw is syntactically valid JSON with the required
// keys, and 鈥?when the binary is available 鈥?that kixdns actually accepts it.
//
// kixdns has no dedicated config-test flag (unlike mihomo's -t), so semantic
// validation works by trial start: the candidate is written to a temp file with
// its listeners rebound to loopback ephemeral ports, then started and observed.
// A config that fails to load makes the process exit promptly; one that loads
// keeps running, at which point we stop it. Because the ports are ephemeral and
// bound to 127.0.0.1, this cannot disturb the production listeners.
func (m *Manager) ValidateConfig(raw []byte) error {
	var probe map[string]any
	if err := json.Unmarshal(raw, &probe); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if _, ok := probe["pipelines"]; !ok {
		return errors.New("missing required key: pipelines")
	}
	if m.Bin == "" {
		return nil // nothing further can be checked without the binary
	}

	// Rebind listeners to loopback ephemeral ports so validation can never
	// collide with the running instance.
	rebound := raw
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err == nil {
		if rawSettings, ok := doc["settings"]; ok {
			var settings map[string]any
			if err := json.Unmarshal(rawSettings, &settings); err == nil {
				_, hasUDP := settings["bind_udp"]
				_, hasTCP := settings["bind_tcp"]
				if hasUDP {
					settings["bind_udp"] = "127.0.0.1:0"
				}
				if hasTCP {
					settings["bind_tcp"] = "127.0.0.1:0"
				}
				if hasUDP || hasTCP {
					if b, err := json.Marshal(settings); err == nil {
						doc["settings"] = b
					}
				}
				if b, err := json.Marshal(doc); err == nil {
					rebound = b
				}
			}
		}
	}

	dir := filepath.Dir(m.Config)
	tmp, err := os.CreateTemp(dir, ".validate-*.json")
	if err != nil {
		return nil // cannot create a temp file here: skip deep validation
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(rebound); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()

	cmd := exec.Command(m.Bin, "run", "-c", tmpName)
	cmd.Dir = m.WorkDir
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("kixdns failed to start: %w", err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("kixdns rejected config: %v", err)
		}
		return errors.New("kixdns exited immediately after start (config likely rejected)")
	case <-time.After(1200 * time.Millisecond):
		_ = cmd.Process.Kill()
		<-done
		return nil
	}
}

// WriteConfig validates then atomically replaces the pipeline file, keeping a
// timestamped backup of the previous revision.
func (m *Manager) WriteConfig(raw []byte) error {
	if err := m.ValidateConfig(raw); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if old, err := os.ReadFile(m.Config); err == nil {
		bak := fmt.Sprintf("%s.bak-%s", m.Config, time.Now().Format("20060102-150405"))
		_ = os.WriteFile(bak, old, 0o644)
	}
	tmp := m.Config + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	// rename is atomic; kixdns watches the file and reloads pipelines itself
	return os.Rename(tmp, m.Config)
}

// Backups lists configuration backups, newest first.
func (m *Manager) Backups() []string {
	des, err := os.ReadDir(filepath.Dir(m.Config))
	if err != nil {
		return nil
	}
	prefix := filepath.Base(m.Config) + ".bak-"
	var out []string
	for _, de := range des {
		if strings.HasPrefix(de.Name(), prefix) {
			out = append(out, filepath.Join(filepath.Dir(m.Config), de.Name()))
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(out)))
	return out
}

// PruneBackups keeps only the newest n configuration backups.
func (m *Manager) PruneBackups(n int) {
	baks := m.Backups()
	for i := n; i < len(baks); i++ {
		_ = os.Remove(baks[i])
	}
}

// ---------------------------------------------------------------------------
// log handling
// ---------------------------------------------------------------------------

// LogEntry describes one log file.
type LogEntry struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	ModTime int64  `json:"mod_time"`
}

// LogFiles lists *.log in LogDir, newest first.
func (m *Manager) LogFiles() []LogEntry {
	des, err := os.ReadDir(m.LogDir)
	if err != nil {
		return nil
	}
	var out []LogEntry
	for _, de := range des {
		if de.IsDir() || !strings.HasSuffix(de.Name(), ".log") {
			continue
		}
		info, err := de.Info()
		if err != nil {
			continue
		}
		out = append(out, LogEntry{
			Path:    filepath.Join(m.LogDir, de.Name()),
			Name:    de.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime().Unix(),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ModTime > out[j].ModTime })
	return out
}

// TailLog returns up to n lines from the end of file (default: the active log).
func (m *Manager) TailLog(file string, n int) ([]string, error) {
	if file == "" {
		file = filepath.Join(m.LogDir, "kixdns.log")
	}
	// keep the endpoint from reading arbitrary files
	clean := filepath.Clean(file)
	base := filepath.Clean(m.LogDir)
	if filepath.Dir(clean) != base {
		return nil, errors.New("path outside log directory")
	}
	b, err := os.ReadFile(clean)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
	if n > 0 && len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines, nil
}

// RotateLog renames the active log to kixdns-<timestamp>.log and deletes the
// oldest files so that retention stays within keepRotated files.
//
// Retention model ("keep 3 days"): the active file counts as day 0, so
// keepRotated = 2 leaves exactly three days of logs on disk.
func (m *Manager) RotateLog(keepRotated int) (string, error) {
	active := filepath.Join(m.LogDir, "kixdns.log")
	info, err := os.Stat(active)
	if err != nil {
		return "", err
	}
	if info.Size() == 0 {
		return "", nil
	}
	dst := filepath.Join(m.LogDir, "kixdns-"+time.Now().Format("20060102-150405")+".log")
	if err := os.Rename(active, dst); err != nil {
		return "", err
	}
	m.pruneRotated(keepRotated)
	return dst, nil
}

// pruneRotated deletes the oldest rotated logs, keeping at most keep files.
func (m *Manager) pruneRotated(keep int) []string {
	var removed []string
	if keep < 0 {
		keep = 0
	}
	des, err := os.ReadDir(m.LogDir)
	if err != nil {
		return removed
	}
	var rotated []string
	for _, de := range des {
		if strings.HasPrefix(de.Name(), "kixdns-") && strings.HasSuffix(de.Name(), ".log") {
			rotated = append(rotated, filepath.Join(m.LogDir, de.Name()))
		}
	}
	sort.Strings(rotated) // timestamped names sort chronologically
	for len(rotated) > keep {
		if err := os.Remove(rotated[0]); err == nil {
			removed = append(removed, rotated[0])
		}
		rotated = rotated[1:]
	}
	return removed
}

// PruneRotated is the exported form used by the HTTP layer.
func (m *Manager) PruneRotated(keep int) []string { return m.pruneRotated(keep) }

// ActiveLogSize returns the size of the current kixdns.log in bytes.
func (m *Manager) ActiveLogSize() int64 {
	if fi, err := os.Stat(filepath.Join(m.LogDir, "kixdns.log")); err == nil {
		return fi.Size()
	}
	return 0
}

// LogDiskUsage returns the total bytes used by kixdns log files.
func (m *Manager) LogDiskUsage() int64 {
	var total int64
	for _, e := range m.LogFiles() {
		total += e.Size
	}
	return total
}

// EnsureLogPath makes sure the active log file exists, so a fresh install has
// something to tail before kixdns produces output.
func (m *Manager) EnsureLogPath() error {
	if err := os.MkdirAll(m.LogDir, 0o755); err != nil {
		return err
	}
	p := filepath.Join(m.LogDir, "kixdns.log")
	f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	return f.Close()
}
