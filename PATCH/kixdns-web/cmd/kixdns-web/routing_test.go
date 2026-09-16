package main

import "testing"

// TestServiceActionRouting pins the routing order for the service endpoints.
//
// This is a regression test for a real bug: "service/restart" also ends with
// "start", so a switch that tested HasSuffix(path, "start") first routed every
// restart to Start(), which then failed with "kixdns is already running" while
// kixdns was up. The restart path looked broken for as long as it existed.
func TestServiceActionRouting(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"/api/service/restart", "restart"},
		{"/api/service/start", "start"},
		{"/api/service/stop", "stop"},
		{"service/restart", "restart"},
		{"service/start", "start"},
		{"service/stop", "stop"},
	}
	for _, c := range cases {
		if got := serviceAction(c.path); got != c.want {
			t.Errorf("serviceAction(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

// TestServiceActionRejectsUnknown ensures an unrelated path yields no action
// rather than defaulting to something destructive.
func TestServiceActionRejectsUnknown(t *testing.T) {
	for _, p := range []string{"/api/service", "/api/service/", "/api/stats", "/api/service/restartx", ""} {
		if got := serviceAction(p); got != "" {
			t.Errorf("serviceAction(%q) = %q, want \"\"", p, got)
		}
	}
}
