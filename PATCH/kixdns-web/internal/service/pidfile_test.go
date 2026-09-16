package service

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// TestReadPIDFileRejectsForeignProcess pins the bug that made Restart fail with
// "kixdns is already running" while no kixdns was running: the pid file was
// trusted as long as any process held that pid.
func TestReadPIDFileRejectsForeignProcess(t *testing.T) {
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "kixdns.pid")

	// Point the manager at a binary name that cannot appear in our own cmdline.
	m := &Manager{Bin: "/usr/sbin/kixdns-imaginary-binary", PIDFile: pidFile}

	// Our own pid exists, but its cmdline is the test binary, not kixdns.
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := m.readPIDFile(); got != 0 {
		t.Errorf("readPIDFile() = %d, want 0: a live pid that is not kixdns must not be trusted", got)
	}
	t.Log("foreign live pid correctly rejected")

	// A pid that does not exist at all is likewise rejected.
	if err := os.WriteFile(pidFile, []byte("999999"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := m.readPIDFile(); got != 0 {
		t.Errorf("readPIDFile() = %d, want 0 for a dead pid", got)
	}
	t.Log("dead pid correctly rejected")

	// Garbage content is rejected rather than parsed loosely.
	if err := os.WriteFile(pidFile, []byte("not-a-pid\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := m.readPIDFile(); got != 0 {
		t.Errorf("readPIDFile() = %d, want 0 for non-numeric content", got)
	}
	t.Log("garbage content correctly rejected")
}

// TestMatchesKixdnsExcludesSelf pins the rule that keeps the watchdog script and
// kixdns-web from being mistaken for kixdns. A substring test is not enough:
// both of their command lines contain the string "kixdns", and treating either
// as kixdns would make Stop() kill the wrong process.
func TestMatchesKixdnsExcludesSelf(t *testing.T) {
	m := &Manager{Bin: "/data/ufi-tools/kixdns/kixdns"}
	cases := []struct {
		cmd  string
		want bool
	}{
		{"/data/ufi-tools/kixdns/kixdns run -c /data/ufi-tools/kixdns/pipeline.json --debug", true},
		{"kixdns run -c pipeline.json", false}, // different basename, must not match
		{"/data/ufi-tools/kixdns-web/kixdns-web -listen 0.0.0.0:8080 -bin /data/ufi-tools/kixdns/kixdns", false},
		{"/bin/sh /data/scripts/kixdns-guard.sh", false},
		{"/bin/sh /tmp/kixdns-guard.sh", false},
		{"/usr/sbin/dnsmasq -C /var/etc/dnsmasq.conf.lan_dns", false},
		{"", false},
	}
	for _, c := range cases {
		if got := m.matchesKixdns(c.cmd); got != c.want {
			t.Errorf("matchesKixdns(%.55q) = %v, want %v", c.cmd, got, c.want)
		}
	}
	t.Logf("checked %d command lines", len(cases))
}
